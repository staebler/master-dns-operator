// Copyright 2017 Istio Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package kube

import (
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"k8s.io/api/core/v1"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"

	"istio.io/istio/pilot/pkg/model"
	"istio.io/istio/pkg/log"
)

const (
	// NodeRegionLabel is the well-known label for kubernetes node region
	NodeRegionLabel = "failure-domain.beta.kubernetes.io/region"
	// NodeZoneLabel is the well-known label for kubernetes node zone
	NodeZoneLabel = "failure-domain.beta.kubernetes.io/zone"
	// IstioNamespace used by default for Istio cluster-wide installation
	IstioNamespace = "istio-system"
	// IstioConfigMap is used by default
	IstioConfigMap = "istio"
	// PrometheusScrape is the annotation used by prometheus to determine if service metrics should be scraped (collected)
	PrometheusScrape = "prometheus.io/scrape"
	// PrometheusPort is the annotation used to explicitly specify the port to use for scraping metrics
	PrometheusPort = "prometheus.io/port"
	// PrometheusPath is the annotation used to specify a path for scraping metrics. Default is "/metrics"
	PrometheusPath = "prometheus.io/path"
	// PrometheusPathDefault is the default value for the PrometheusPath annotation
	PrometheusPathDefault = "/metrics"
)

var (
	// experiment on getting some monitoring on config errors.
	k8sEvents = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "pilot_k8s_reg_events",
		Help: "Events from k8s registry.",
	}, []string{"type", "event"})
)

func init() {
	prometheus.MustRegister(k8sEvents)
}

// ControllerOptions stores the configurable attributes of a Controller.
type ControllerOptions struct {
	// Namespace the controller watches. If set to meta_v1.NamespaceAll (""), controller watches all namespaces
	WatchedNamespace string
	ResyncPeriod     time.Duration
	DomainSuffix     string

	// ClusterID identifies the remote cluster in a multicluster env.
	ClusterID string

	// XDSUpdater will push changes to the xDS server.
	XDSUpdater model.XDSUpdater

	// ConfigUpdater is used to request global config updates.
	ConfigUpdater model.ConfigUpdater

	stop chan struct{}
}

// Controller is a collection of synchronized resource watchers
// Caches are thread-safe
type Controller struct {
	domainSuffix string

	client    kubernetes.Interface
	queue     Queue
	services  cacheHandler
	endpoints cacheHandler
	nodes     cacheHandler

	pods *PodCache

	// Env is set by server to point to the environment, to allow the controller to
	// use env data and push status. It may be null in tests.
	Env *model.Environment

	// ClusterID identifies the remote cluster in a multicluster env.
	ClusterID string

	// XDSUpdater will push EDS changes to the ADS model.
	EDSUpdater model.XDSUpdater

	// ConfigUpdater is used to request global config updates.
	ConfigUpdater model.ConfigUpdater

	stop chan struct{}
}

type cacheHandler struct {
	informer cache.SharedIndexInformer
	handler  *ChainHandler
}

// NewController creates a new Kubernetes controller
// Created by bootstrap and multicluster (see secretcontroler).
func NewController(client kubernetes.Interface, options ControllerOptions) *Controller {
	log.Infof("Service controller watching namespace %q for service, endpoint, nodes and pods, refresh %d",
		options.WatchedNamespace, options.ResyncPeriod)

	// Queue requires a time duration for a retry delay after a handler error
	out := &Controller{
		domainSuffix:  options.DomainSuffix,
		client:        client,
		queue:         NewQueue(1 * time.Second),
		ClusterID:     options.ClusterID,
		EDSUpdater:    options.XDSUpdater,
		ConfigUpdater: options.ConfigUpdater,
	}

	out.services = out.createInformer(&v1.Service{}, "Service", options.ResyncPeriod,
		func(opts meta_v1.ListOptions) (runtime.Object, error) {
			return client.CoreV1().Services(options.WatchedNamespace).List(opts)
		},
		func(opts meta_v1.ListOptions) (watch.Interface, error) {
			return client.CoreV1().Services(options.WatchedNamespace).Watch(opts)
		})

	if out.EDSUpdater != nil {
		out.endpoints = out.createEDSInformer(&v1.Endpoints{}, "Endpoints", options.ResyncPeriod,
			func(opts meta_v1.ListOptions) (runtime.Object, error) {
				return client.CoreV1().Endpoints(options.WatchedNamespace).List(opts)
			},
			func(opts meta_v1.ListOptions) (watch.Interface, error) {
				return client.CoreV1().Endpoints(options.WatchedNamespace).Watch(opts)
			})
	} else {
		out.endpoints = out.createInformer(&v1.Endpoints{}, "Endpoints", options.ResyncPeriod,
			func(opts meta_v1.ListOptions) (runtime.Object, error) {
				return client.CoreV1().Endpoints(options.WatchedNamespace).List(opts)
			},
			func(opts meta_v1.ListOptions) (watch.Interface, error) {
				return client.CoreV1().Endpoints(options.WatchedNamespace).Watch(opts)
			})
	}

	out.nodes = out.createInformer(&v1.Node{}, "Node", options.ResyncPeriod,
		func(opts meta_v1.ListOptions) (runtime.Object, error) {
			return client.CoreV1().Nodes().List(opts)
		},
		func(opts meta_v1.ListOptions) (watch.Interface, error) {
			return client.CoreV1().Nodes().Watch(opts)
		})

	out.pods = newPodCache(out.createInformer(&v1.Pod{}, "Pod", options.ResyncPeriod,
		func(opts meta_v1.ListOptions) (runtime.Object, error) {
			return client.CoreV1().Pods(options.WatchedNamespace).List(opts)
		},
		func(opts meta_v1.ListOptions) (watch.Interface, error) {
			return client.CoreV1().Pods(options.WatchedNamespace).Watch(opts)
		}), out)

	return out
}

// notify is the first handler in the handler chain.
// Returning an error causes repeated execution of the entire chain.
func (c *Controller) notify(obj interface{}, event model.Event) error {
	if !c.HasSynced() {
		return errors.New("waiting till full synchronization")
	}
	return nil
}

// createInformer registers handlers for a specific event.
// Current implementation queues the events in queue.go, and the handler is run with
// some throttling.
// Used for Service, Endpoint, Node and Pod.
// See config/kube for CRD events.
// See config/ingress for Ingress objects
func (c *Controller) createInformer(
	o runtime.Object,
	otype string,
	resyncPeriod time.Duration,
	lf cache.ListFunc,
	wf cache.WatchFunc) cacheHandler {
	handler := &ChainHandler{funcs: []Handler{c.notify}}

	// TODO: finer-grained index (perf)
	informer := cache.NewSharedIndexInformer(
		&cache.ListWatch{ListFunc: lf, WatchFunc: wf}, o,
		resyncPeriod, cache.Indexers{})

	informer.AddEventHandler(
		cache.ResourceEventHandlerFuncs{
			// TODO: filtering functions to skip over un-referenced resources (perf)
			AddFunc: func(obj interface{}) {
				k8sEvents.With(prometheus.Labels{"type": otype, "event": "add"}).Add(1)
				c.queue.Push(Task{handler: handler.Apply, obj: obj, event: model.EventAdd})
			},
			UpdateFunc: func(old, cur interface{}) {
				if !reflect.DeepEqual(old, cur) {
					k8sEvents.With(prometheus.Labels{"type": otype, "event": "update"}).Add(1)
					c.queue.Push(Task{handler: handler.Apply, obj: cur, event: model.EventUpdate})
				} else {
					k8sEvents.With(prometheus.Labels{"type": otype, "event": "updateSame"}).Add(1)
				}
			},
			DeleteFunc: func(obj interface{}) {
				k8sEvents.With(prometheus.Labels{"type": otype, "event": "add"}).Add(1)
				c.queue.Push(Task{handler: handler.Apply, obj: obj, event: model.EventDelete})
			},
		})

	return cacheHandler{informer: informer, handler: handler}
}

// createEDSInformer is a special informer for Endpoints
func (c *Controller) createEDSInformer(
	o runtime.Object,
	otype string,
	resyncPeriod time.Duration,
	lf cache.ListFunc,
	wf cache.WatchFunc) cacheHandler {
	handler := &ChainHandler{funcs: []Handler{c.notify}}

	informer := cache.NewSharedIndexInformer(
		&cache.ListWatch{ListFunc: lf, WatchFunc: wf}, o,
		resyncPeriod, cache.Indexers{})

	informer.AddEventHandler(
		cache.ResourceEventHandlerFuncs{
			// TODO: filtering functions to skip over un-referenced resources (perf)
			AddFunc: func(obj interface{}) {
				k8sEvents.With(prometheus.Labels{"type": otype, "event": "add"}).Add(1)
				// TODO: we want this to propagate faster, without waiting in the queu. However
				// this requires pods to be updated before, and code to deal with eventual consistency,
				// otherwise enpdoint update is too fast.
				//c.updateEDS(obj.(*v1.Endpoints))
				c.queue.Push(Task{handler: handler.Apply, obj: obj, event: model.EventAdd})
			},
			UpdateFunc: func(old, cur interface{}) {
				// Avoid pushes if only resource version changed (kube-scheduller, cluster-autoscaller, etc)
				oldE := old.(*v1.Endpoints)
				curE := cur.(*v1.Endpoints)

				if !reflect.DeepEqual(oldE.Subsets, curE.Subsets) {
					k8sEvents.With(prometheus.Labels{"type": otype, "event": "update"}).Add(1)
					//c.updateEDS(cur.(*v1.Endpoints))
					c.queue.Push(Task{handler: handler.Apply, obj: cur, event: model.EventUpdate})
				} else {
					k8sEvents.With(prometheus.Labels{"type": otype, "event": "updateSame"}).Add(1)
				}
			},
			DeleteFunc: func(obj interface{}) {
				k8sEvents.With(prometheus.Labels{"type": otype, "event": "add"}).Add(1)
				// Deleting the endpoints results in an empty set from EDS perspective - only
				// deleting the service should delete the resources. The full sync replaces the
				// maps.
				//c.updateEDS(obj.(*v1.Endpoints))
				c.queue.Push(Task{handler: handler.Apply, obj: obj, event: model.EventDelete})
			},
		})

	return cacheHandler{informer: informer, handler: handler}
}

// HasSynced returns true after the initial state synchronization
func (c *Controller) HasSynced() bool {
	if !c.services.informer.HasSynced() ||
		!c.endpoints.informer.HasSynced() ||
		!c.pods.informer.HasSynced() ||
		!c.nodes.informer.HasSynced() {
		return false
	}
	return true
}

// Run all controllers until a signal is received
func (c *Controller) Run(stop <-chan struct{}) {
	go c.queue.Run(stop)

	go c.services.informer.Run(stop)
	go c.pods.informer.Run(stop)
	go c.nodes.informer.Run(stop)

	// To avoid endpoints without labels or ports, wait for sync.
	cache.WaitForCacheSync(stop, c.nodes.informer.HasSynced, c.pods.informer.HasSynced,
		c.services.informer.HasSynced)

	go c.endpoints.informer.Run(stop)

	<-stop
	log.Infof("Controller terminated")
}

// Stop the controller. Mostly for tests, to simplify the code (defer c.Stop())
func (c *Controller) Stop() {
	if c.stop != nil {
		c.stop <- struct{}{}
	}
}

// Services implements a service catalog operation
func (c *Controller) Services() ([]*model.Service, error) {
	list := c.services.informer.GetStore().List()
	out := make([]*model.Service, 0, len(list))

	for _, item := range list {
		if svc := convertService(*item.(*v1.Service), c.domainSuffix); svc != nil {
			out = append(out, svc)
		}
	}
	return out, nil
}

// GetService implements a service catalog operation
func (c *Controller) GetService(hostname model.Hostname) (*model.Service, error) {
	name, namespace, err := parseHostname(hostname)
	if err != nil {
		log.Infof("parseHostname(%s) => error %v", hostname, err)
		return nil, err
	}
	item, exists := c.serviceByKey(name, namespace)
	if !exists {
		return nil, nil
	}

	svc := convertService(*item, c.domainSuffix)
	return svc, nil
}

// serviceByKey retrieves a service by name and namespace
func (c *Controller) serviceByKey(name, namespace string) (*v1.Service, bool) {
	item, exists, err := c.services.informer.GetStore().GetByKey(KeyFunc(name, namespace))
	if err != nil {
		log.Infof("serviceByKey(%s, %s) => error %v", name, namespace, err)
		return nil, false
	}
	if !exists {
		return nil, false
	}
	return item.(*v1.Service), true
}

// GetPodAZ retrieves the AZ for a pod.
func (c *Controller) GetPodAZ(pod *v1.Pod) (string, bool) {
	// NodeName is set by the scheduler after the pod is created
	// https://github.com/kubernetes/community/blob/master/contributors/devel/api-conventions.md#late-initialization
	node, exists, err := c.nodes.informer.GetStore().GetByKey(pod.Spec.NodeName)
	if !exists || err != nil {
		log.Warnf("unable to get node %q for pod %q: %v", pod.Spec.NodeName, pod.Name, err)
		return "", false
	}
	region, exists := node.(*v1.Node).Labels[NodeRegionLabel]
	if !exists {
		return "", false
	}
	zone, exists := node.(*v1.Node).Labels[NodeZoneLabel]
	if !exists {
		return "", false
	}
	return fmt.Sprintf("%v/%v", region, zone), true
}

// ManagementPorts implements a service catalog operation
func (c *Controller) ManagementPorts(addr string) model.PortList {
	pod, exists := c.pods.getPodByIP(addr)
	if !exists {
		return nil
	}

	managementPorts, err := convertProbesToPorts(&pod.Spec)

	if err != nil {
		log.Infof("Error while parsing liveliness and readiness probe ports for %s => %v", addr, err)
	}

	// We continue despite the error because healthCheckPorts could return a partial
	// list of management ports
	return managementPorts
}

// WorkloadHealthCheckInfo implements a service catalog operation
func (c *Controller) WorkloadHealthCheckInfo(addr string) model.ProbeList {
	pod, exists := c.pods.getPodByIP(addr)
	if !exists {
		return nil
	}

	probes := make([]*model.Probe, 0)

	// Obtain probes from the readiness and liveness probes
	for _, container := range pod.Spec.Containers {
		if container.ReadinessProbe != nil && container.ReadinessProbe.Handler.HTTPGet != nil {
			p, err := convertProbePort(container, &container.ReadinessProbe.Handler)
			if err != nil {
				log.Infof("Error while parsing readiness probe port =%v", err)
			}
			probes = append(probes, &model.Probe{
				Port: p,
				Path: container.ReadinessProbe.Handler.HTTPGet.Path,
			})
		}
		if container.LivenessProbe != nil && container.LivenessProbe.Handler.HTTPGet != nil {
			p, err := convertProbePort(container, &container.LivenessProbe.Handler)
			if err != nil {
				log.Infof("Error while parsing liveness probe port =%v", err)
			}
			probes = append(probes, &model.Probe{
				Port: p,
				Path: container.LivenessProbe.Handler.HTTPGet.Path,
			})
		}
	}

	// Obtain probe from prometheus scrape
	if scrape := pod.Annotations[PrometheusScrape]; scrape == "true" {
		var port *model.Port
		path := PrometheusPathDefault
		if portstr := pod.Annotations[PrometheusPort]; portstr != "" {
			portnum, err := strconv.Atoi(portstr)
			if err != nil {
				log.Warna(err)
			} else {
				port = &model.Port{
					Port: portnum,
				}
			}
		}
		if pod.Annotations[PrometheusPath] != "" {
			path = pod.Annotations[PrometheusPath]
		}
		probes = append(probes, &model.Probe{
			Port: port,
			Path: path,
		})
	}

	return probes
}

// Instances returns the service instances associated with a service
// Called by GetIstioServiceAccounts to guess the list of authorized service accounts
// for a service.
//
func (c *Controller) Instances(hostname model.Hostname, ports []string,
	labelsList model.LabelsCollection) ([]*model.ServiceInstance, error) {
	// Get actual service by name
	name, namespace, err := parseHostname(hostname)
	if err != nil {
		log.Infof("parseHostname(%s) => error %v", hostname, err)
		return nil, err
	}

	item, exists := c.serviceByKey(name, namespace)
	if !exists {
		return nil, nil
	}

	// Locate all ports in the actual service
	svc := convertService(*item, c.domainSuffix)
	if svc == nil {
		return nil, nil
	}
	svcPorts := make(map[string]*model.Port)
	for _, port := range ports {
		if svcPort, exists := svc.Ports.Get(port); exists {
			svcPorts[port] = svcPort
		}
	}

	// TODO: single port service missing name
	for _, item := range c.endpoints.informer.GetStore().List() {
		ep := *item.(*v1.Endpoints)
		if ep.Name == name && ep.Namespace == namespace {
			var out []*model.ServiceInstance
			for _, ss := range ep.Subsets {
				for _, ea := range ss.Addresses {
					labels, _ := c.pods.labelsByIP(ea.IP)
					// check that one of the input labels is a subset of the labels
					if !labelsList.HasSubsetOf(labels) {
						continue
					}

					pod, exists := c.pods.getPodByIP(ea.IP)
					az, sa, uid := "", "", ""
					if exists {
						az, _ = c.GetPodAZ(pod)
						sa = kubeToIstioServiceAccount(pod.Spec.ServiceAccountName, pod.GetNamespace(), c.domainSuffix)
						uid = fmt.Sprintf("kubernetes://%s.%s", pod.Name, pod.Namespace)
					}

					// identify the port by name
					for _, port := range ss.Ports {
						if svcPort, exists := svcPorts[port.Name]; exists {
							out = append(out, &model.ServiceInstance{
								Endpoint: model.NetworkEndpoint{
									Address:     ea.IP,
									Port:        int(port.Port),
									ServicePort: svcPort,
									UID:         uid,
								},
								Service:          svc,
								Labels:           labels,
								AvailabilityZone: az,
								ServiceAccount:   sa,
							})
						}
					}
				}
			}
			return out, nil
		}
	}
	return nil, nil
}

// InstancesByPort implements a service catalog operation
func (c *Controller) InstancesByPort(hostname model.Hostname, reqSvcPort int,
	labelsList model.LabelsCollection) ([]*model.ServiceInstance, error) {
	// Get actual service by name
	name, namespace, err := parseHostname(hostname)
	if err != nil {
		log.Infof("parseHostname(%s) => error %v", hostname, err)
		return nil, err
	}

	item, exists := c.serviceByKey(name, namespace)
	if !exists {
		return nil, nil
	}

	// Locate all ports in the actual service

	svc := convertService(*item, c.domainSuffix)
	if svc == nil {
		return nil, nil
	}

	svcPortEntry, exists := svc.Ports.GetByPort(reqSvcPort)
	if !exists && reqSvcPort != 0 {
		return nil, nil
	}

	for _, item := range c.endpoints.informer.GetStore().List() {
		ep := *item.(*v1.Endpoints)
		if ep.Name == name && ep.Namespace == namespace {
			var out []*model.ServiceInstance
			for _, ss := range ep.Subsets {
				for _, ea := range ss.Addresses {
					labels, _ := c.pods.labelsByIP(ea.IP)
					// check that one of the input labels is a subset of the labels
					if !labelsList.HasSubsetOf(labels) {
						continue
					}

					pod, exists := c.pods.getPodByIP(ea.IP)
					az, sa, uid := "", "", ""
					if exists {
						az, _ = c.GetPodAZ(pod)
						sa = kubeToIstioServiceAccount(pod.Spec.ServiceAccountName, pod.GetNamespace(), c.domainSuffix)
						uid = fmt.Sprintf("kubernetes://%s.%s", pod.Name, pod.Namespace)
					}

					// identify the port by name. K8S EndpointPort uses the service port name
					for _, port := range ss.Ports {
						if port.Name == "" || // 'name optional if single port is defined'
							reqSvcPort == 0 || // return all ports (mostly used by tests/debug)
							svcPortEntry.Name == port.Name {
							out = append(out, &model.ServiceInstance{
								Endpoint: model.NetworkEndpoint{
									Address:     ea.IP,
									Port:        int(port.Port),
									ServicePort: svcPortEntry,
									UID:         uid,
								},
								Service:          svc,
								Labels:           labels,
								AvailabilityZone: az,
								ServiceAccount:   sa,
							})
						}
					}
				}
			}
			return out, nil
		}
	}
	return nil, nil
}

// GetProxyServiceInstances returns service instances co-located with a given proxy
func (c *Controller) GetProxyServiceInstances(proxy *model.Proxy) ([]*model.ServiceInstance, error) {
	proxyIP := proxy.IPAddress
	podNamespace := ""
	if pod, exists := c.pods.getPodByIP(proxyIP); exists {
		podNamespace = pod.Namespace
	}
	endpointsForPodInSameNS := make([]*model.ServiceInstance, 0)
	endpointsForPodInDifferentNS := make([]*model.ServiceInstance, 0)
	l := c.endpoints.informer.GetStore().List()
	for _, item := range l {
		ep := *item.(*v1.Endpoints)

		svcItem, exists := c.serviceByKey(ep.Name, ep.Namespace)
		if !exists {
			continue
		}
		svc := convertService(*svcItem, c.domainSuffix)
		if svc == nil {
			continue
		}

		endpoints := &endpointsForPodInSameNS
		if ep.Namespace != podNamespace {
			endpoints = &endpointsForPodInDifferentNS
		}
		for _, ss := range ep.Subsets {
			for _, port := range ss.Ports {
				svcPort, exists := svc.Ports.Get(port.Name)
				if !exists {
					continue
				}

				*endpoints = append(*endpoints, getEndpoints(ss.Addresses, proxyIP, c, port, svcPort, svc)...)
				nrEP := getEndpoints(ss.NotReadyAddresses, proxyIP, c, port, svcPort, svc)
				*endpoints = append(*endpoints, nrEP...)
				if len(nrEP) > 0 && c.Env != nil {
					c.Env.PushContext.Add(model.ProxyStatusEndpointNotReady, proxy.ID, proxy, "")
				}
			}
		}
	}

	// Put the endpointsForPodInSameNS in front of endpointsForPodInDifferentNS so that Pilot will
	// first use endpoints from endpointsForPodInSameNS. This makes sure if there are two endpoints
	// referring to the same IP/port, the one in endpointsForPodInSameNS will be used. (The other one
	// in endpointsForPodInDifferentNS will thus be rejected by Pilot).
	out := append(endpointsForPodInSameNS, endpointsForPodInDifferentNS...)
	if len(out) == 0 {
		if c.Env != nil {
			c.Env.PushContext.Add(model.ProxyStatusNoService, proxy.ID, proxy, "")
			status := c.Env.PushContext
			if status == nil {
				log.Infof("Empty list of services for pod %s %v", proxy.ID, c.Env)
			}
		} else {
			log.Infof("Missing env, empty list of services for pod %s", proxy.ID)
		}
	}
	return out, nil
}

func getEndpoints(addr []v1.EndpointAddress, proxyIP string, c *Controller,
	port v1.EndpointPort, svcPort *model.Port, svc *model.Service) []*model.ServiceInstance {

	var out []*model.ServiceInstance
	for _, ea := range addr {
		if proxyIP != ea.IP {
			continue
		}
		labels, _ := c.pods.labelsByIP(ea.IP)
		pod, exists := c.pods.getPodByIP(ea.IP)
		az, sa := "", ""
		if exists {
			az, _ = c.GetPodAZ(pod)
			sa = kubeToIstioServiceAccount(pod.Spec.ServiceAccountName, pod.GetNamespace(), c.domainSuffix)
		}
		out = append(out, &model.ServiceInstance{
			Endpoint: model.NetworkEndpoint{
				Address:     ea.IP,
				Port:        int(port.Port),
				ServicePort: svcPort,
			},
			Service:          svc,
			Labels:           labels,
			AvailabilityZone: az,
			ServiceAccount:   sa,
		})
	}
	return out
}

// GetIstioServiceAccounts returns the Istio service accounts running a serivce
// hostname. Each service account is encoded according to the SPIFFE VSID spec.
// For example, a service account named "bar" in namespace "foo" is encoded as
// "spiffe://cluster.local/ns/foo/sa/bar".
func (c *Controller) GetIstioServiceAccounts(hostname model.Hostname, ports []string) []string {
	saSet := make(map[string]bool)

	// Get the service accounts running the service, if it is deployed on VMs. This is retrieved
	// from the service annotation explicitly set by the operators.
	svc, err := c.GetService(hostname)
	if err != nil {
		// Do not log error here, as the service could exist in another registry
		return nil
	}
	if svc == nil {
		// Do not log error here as the service could exist in another registry
		return nil
	}

	// Get the service accounts running service within Kubernetes. This is reflected by the pods that
	// the service is deployed on, and the service accounts of the pods.
	instances, err := c.Instances(hostname, ports, model.LabelsCollection{})
	if err != nil {
		log.Warnf("Instances(%s) error: %v", hostname, err)
		return nil
	}
	for _, si := range instances {
		if si.ServiceAccount != "" {
			saSet[si.ServiceAccount] = true
		}
	}

	for _, serviceAccount := range svc.ServiceAccounts {
		sa := serviceAccount
		saSet[sa] = true
	}

	saArray := make([]string, 0, len(saSet))
	for sa := range saSet {
		saArray = append(saArray, sa)
	}

	return saArray
}

// AppendServiceHandler implements a service catalog operation
func (c *Controller) AppendServiceHandler(f func(*model.Service, model.Event)) error {
	c.services.handler.Append(func(obj interface{}, event model.Event) error {
		svc := *obj.(*v1.Service)

		// Do not handle "kube-system" services
		if svc.Namespace == meta_v1.NamespaceSystem {
			return nil
		}

		log.Infof("Handle service %s in namespace %s", svc.Name, svc.Namespace)

		if c.EDSUpdater != nil {
			hostname := svc.Name + "." + svc.Namespace
			ports := map[string]uint32{}
			portsByNum := map[uint32]string{}

			for _, port := range svc.Spec.Ports {
				ports[port.Name] = uint32(port.Port)
				portsByNum[uint32(port.Port)] = port.Name
			}
			// EDS needs the port mapping.
			c.EDSUpdater.SvcUpdate(c.ClusterID, hostname, ports, portsByNum)
			// Bypass convertService and the cache invalidation.

			// Shortcut the conversion
			c.ConfigUpdater.ConfigUpdate(true)
			return nil
		}

		if svcConv := convertService(svc, c.domainSuffix); svcConv != nil {
			f(svcConv, event)
		}
		return nil
	})
	return nil
}

// AppendInstanceHandler implements a service catalog operation
func (c *Controller) AppendInstanceHandler(f func(*model.ServiceInstance, model.Event)) error {
	if c.endpoints.handler == nil {
		return nil
	}
	c.endpoints.handler.Append(func(obj interface{}, event model.Event) error {
		ep := obj.(*v1.Endpoints)

		// Do not handle "kube-system" endpoints
		if ep.Namespace == meta_v1.NamespaceSystem {
			return nil
		}

		if c.EDSUpdater != nil {
			c.updateEDS(ep)
		} else {
			log.Infof("Handle endpoint %s in namespace %s -> %v", ep.Name, ep.Namespace, ep.Subsets)
			if item, exists := c.serviceByKey(ep.Name, ep.Namespace); exists {
				if svc := convertService(*item, c.domainSuffix); svc != nil {
					// TODO: we're passing an incomplete instance to the
					// handler since endpoints is an aggregate structure
					f(&model.ServiceInstance{Service: svc}, event)
				}
			}
		}
		return nil
	})
	return nil
}

func (c *Controller) updateEDS(ep *v1.Endpoints) {
	hostname := serviceHostname(ep.Name, ep.Namespace, c.domainSuffix)

	endpoints := []*model.IstioEndpoint{}
	for _, ss := range ep.Subsets {
		for _, ea := range ss.Addresses {
			//podName := ea.TargetRef.Name

			pod, exists := c.pods.getPodByIP(ea.IP)
			if !exists {
				log.Warnf("Endpoint without pod %s %v", ea.IP, ep)
				if c.Env != nil {
					c.Env.PushContext.Add(model.EndpointNoPod, string(hostname), nil, ea.IP)
				}
				// Request a global config update - after debounce we may find an endpoint.
				if c.ConfigUpdater != nil {
					c.ConfigUpdater.ConfigUpdate(true)
				}
				// TODO: keep them in a list, and check when pod events happen !
				continue
			}

			labels := map[string]string(convertLabels(pod.ObjectMeta))

			uid := fmt.Sprintf("kubernetes://%s.%s", pod.Name, pod.Namespace)

			// EDS and ServiceEntry use name for service port - ADS will need to
			// map to numbers.
			for _, port := range ss.Ports {
				endpoints = append(endpoints, &model.IstioEndpoint{
					Address:         ea.IP,
					EndpointPort:    uint32(port.Port),
					ServicePortName: port.Name,
					Labels:          labels,
					UID:             uid,
					ServiceAccount:  kubeToIstioServiceAccount(pod.Spec.ServiceAccountName, pod.GetNamespace(), c.domainSuffix),
				})
			}
		}
	}

	// TODO: Endpoints include the service labels, maybe we can use them ?
	// nodeName is also included, not needed

	log.Infof("Handle EDS endpoint %s in namespace %s -> %v %v", ep.Name, ep.Namespace, ep.Subsets, endpoints)

	err := c.EDSUpdater.EDSUpdate(c.ClusterID, string(hostname), endpoints)
	if err != nil {
		// Request a global push if we failed to do EDS only
		c.ConfigUpdater.ConfigUpdate(true)
	} else {
		c.ConfigUpdater.ConfigUpdate(false)
	}
}

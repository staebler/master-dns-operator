package rds

//Licensed under the Apache License, Version 2.0 (the "License");
//you may not use this file except in compliance with the License.
//You may obtain a copy of the License at
//
//http://www.apache.org/licenses/LICENSE-2.0
//
//Unless required by applicable law or agreed to in writing, software
//distributed under the License is distributed on an "AS IS" BASIS,
//WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//See the License for the specific language governing permissions and
//limitations under the License.
//
// Code generated by Alibaba Cloud SDK Code Generator.
// Changes may cause incorrect behavior and will be lost if the code is regenerated.

import (
	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/requests"
	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/responses"
)

// DescribeLogicDBInstanceTopology invokes the rds.DescribeLogicDBInstanceTopology API synchronously
// api document: https://help.aliyun.com/api/rds/describelogicdbinstancetopology.html
func (client *Client) DescribeLogicDBInstanceTopology(request *DescribeLogicDBInstanceTopologyRequest) (response *DescribeLogicDBInstanceTopologyResponse, err error) {
	response = CreateDescribeLogicDBInstanceTopologyResponse()
	err = client.DoAction(request, response)
	return
}

// DescribeLogicDBInstanceTopologyWithChan invokes the rds.DescribeLogicDBInstanceTopology API asynchronously
// api document: https://help.aliyun.com/api/rds/describelogicdbinstancetopology.html
// asynchronous document: https://help.aliyun.com/document_detail/66220.html
func (client *Client) DescribeLogicDBInstanceTopologyWithChan(request *DescribeLogicDBInstanceTopologyRequest) (<-chan *DescribeLogicDBInstanceTopologyResponse, <-chan error) {
	responseChan := make(chan *DescribeLogicDBInstanceTopologyResponse, 1)
	errChan := make(chan error, 1)
	err := client.AddAsyncTask(func() {
		defer close(responseChan)
		defer close(errChan)
		response, err := client.DescribeLogicDBInstanceTopology(request)
		if err != nil {
			errChan <- err
		} else {
			responseChan <- response
		}
	})
	if err != nil {
		errChan <- err
		close(responseChan)
		close(errChan)
	}
	return responseChan, errChan
}

// DescribeLogicDBInstanceTopologyWithCallback invokes the rds.DescribeLogicDBInstanceTopology API asynchronously
// api document: https://help.aliyun.com/api/rds/describelogicdbinstancetopology.html
// asynchronous document: https://help.aliyun.com/document_detail/66220.html
func (client *Client) DescribeLogicDBInstanceTopologyWithCallback(request *DescribeLogicDBInstanceTopologyRequest, callback func(response *DescribeLogicDBInstanceTopologyResponse, err error)) <-chan int {
	result := make(chan int, 1)
	err := client.AddAsyncTask(func() {
		var response *DescribeLogicDBInstanceTopologyResponse
		var err error
		defer close(result)
		response, err = client.DescribeLogicDBInstanceTopology(request)
		callback(response, err)
		result <- 1
	})
	if err != nil {
		defer close(result)
		callback(nil, err)
		result <- 0
	}
	return result
}

// DescribeLogicDBInstanceTopologyRequest is the request struct for api DescribeLogicDBInstanceTopology
type DescribeLogicDBInstanceTopologyRequest struct {
	*requests.RpcRequest
	OwnerId              requests.Integer `position:"Query" name:"OwnerId"`
	ResourceOwnerAccount string           `position:"Query" name:"ResourceOwnerAccount"`
	ResourceOwnerId      requests.Integer `position:"Query" name:"ResourceOwnerId"`
}

// DescribeLogicDBInstanceTopologyResponse is the response struct for api DescribeLogicDBInstanceTopology
type DescribeLogicDBInstanceTopologyResponse struct {
	*responses.BaseResponse
	DBInstanceId          int                                    `json:"DBInstanceId" xml:"DBInstanceId"`
	DBInstanceName        string                                 `json:"DBInstanceName" xml:"DBInstanceName"`
	DBInstanceStatus      int                                    `json:"DBInstanceStatus" xml:"DBInstanceStatus"`
	DBInstanceStatusDesc  string                                 `json:"DBInstanceStatusDesc" xml:"DBInstanceStatusDesc"`
	DBInstanceConnType    string                                 `json:"DBInstanceConnType" xml:"DBInstanceConnType"`
	DBInstanceDescription string                                 `json:"DBInstanceDescription" xml:"DBInstanceDescription"`
	Engine                string                                 `json:"Engine" xml:"Engine"`
	EngineVersion         string                                 `json:"EngineVersion" xml:"EngineVersion"`
	Items                 ItemsInDescribeLogicDBInstanceTopology `json:"Items" xml:"Items"`
}

// CreateDescribeLogicDBInstanceTopologyRequest creates a request to invoke DescribeLogicDBInstanceTopology API
func CreateDescribeLogicDBInstanceTopologyRequest() (request *DescribeLogicDBInstanceTopologyRequest) {
	request = &DescribeLogicDBInstanceTopologyRequest{
		RpcRequest: &requests.RpcRequest{},
	}
	request.InitWithApiInfo("Rds", "2014-08-15", "DescribeLogicDBInstanceTopology", "rds", "openAPI")
	return
}

// CreateDescribeLogicDBInstanceTopologyResponse creates a response to parse from DescribeLogicDBInstanceTopology response
func CreateDescribeLogicDBInstanceTopologyResponse() (response *DescribeLogicDBInstanceTopologyResponse) {
	response = &DescribeLogicDBInstanceTopologyResponse{
		BaseResponse: &responses.BaseResponse{},
	}
	return
}

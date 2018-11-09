package live

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

// DescribeLiveRecordNotifyConfig invokes the live.DescribeLiveRecordNotifyConfig API synchronously
// api document: https://help.aliyun.com/api/live/describeliverecordnotifyconfig.html
func (client *Client) DescribeLiveRecordNotifyConfig(request *DescribeLiveRecordNotifyConfigRequest) (response *DescribeLiveRecordNotifyConfigResponse, err error) {
	response = CreateDescribeLiveRecordNotifyConfigResponse()
	err = client.DoAction(request, response)
	return
}

// DescribeLiveRecordNotifyConfigWithChan invokes the live.DescribeLiveRecordNotifyConfig API asynchronously
// api document: https://help.aliyun.com/api/live/describeliverecordnotifyconfig.html
// asynchronous document: https://help.aliyun.com/document_detail/66220.html
func (client *Client) DescribeLiveRecordNotifyConfigWithChan(request *DescribeLiveRecordNotifyConfigRequest) (<-chan *DescribeLiveRecordNotifyConfigResponse, <-chan error) {
	responseChan := make(chan *DescribeLiveRecordNotifyConfigResponse, 1)
	errChan := make(chan error, 1)
	err := client.AddAsyncTask(func() {
		defer close(responseChan)
		defer close(errChan)
		response, err := client.DescribeLiveRecordNotifyConfig(request)
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

// DescribeLiveRecordNotifyConfigWithCallback invokes the live.DescribeLiveRecordNotifyConfig API asynchronously
// api document: https://help.aliyun.com/api/live/describeliverecordnotifyconfig.html
// asynchronous document: https://help.aliyun.com/document_detail/66220.html
func (client *Client) DescribeLiveRecordNotifyConfigWithCallback(request *DescribeLiveRecordNotifyConfigRequest, callback func(response *DescribeLiveRecordNotifyConfigResponse, err error)) <-chan int {
	result := make(chan int, 1)
	err := client.AddAsyncTask(func() {
		var response *DescribeLiveRecordNotifyConfigResponse
		var err error
		defer close(result)
		response, err = client.DescribeLiveRecordNotifyConfig(request)
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

// DescribeLiveRecordNotifyConfigRequest is the request struct for api DescribeLiveRecordNotifyConfig
type DescribeLiveRecordNotifyConfigRequest struct {
	*requests.RpcRequest
	SecurityToken string           `position:"Query" name:"SecurityToken"`
	DomainName    string           `position:"Query" name:"DomainName"`
	OwnerId       requests.Integer `position:"Query" name:"OwnerId"`
}

// DescribeLiveRecordNotifyConfigResponse is the response struct for api DescribeLiveRecordNotifyConfig
type DescribeLiveRecordNotifyConfigResponse struct {
	*responses.BaseResponse
	RequestId              string                 `json:"RequestId" xml:"RequestId"`
	LiveRecordNotifyConfig LiveRecordNotifyConfig `json:"LiveRecordNotifyConfig" xml:"LiveRecordNotifyConfig"`
}

// CreateDescribeLiveRecordNotifyConfigRequest creates a request to invoke DescribeLiveRecordNotifyConfig API
func CreateDescribeLiveRecordNotifyConfigRequest() (request *DescribeLiveRecordNotifyConfigRequest) {
	request = &DescribeLiveRecordNotifyConfigRequest{
		RpcRequest: &requests.RpcRequest{},
	}
	request.InitWithApiInfo("live", "2016-11-01", "DescribeLiveRecordNotifyConfig", "live", "openAPI")
	return
}

// CreateDescribeLiveRecordNotifyConfigResponse creates a response to parse from DescribeLiveRecordNotifyConfig response
func CreateDescribeLiveRecordNotifyConfigResponse() (response *DescribeLiveRecordNotifyConfigResponse) {
	response = &DescribeLiveRecordNotifyConfigResponse{
		BaseResponse: &responses.BaseResponse{},
	}
	return
}

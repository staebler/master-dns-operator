package r_kvstore

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

// ModifyCertification invokes the r_kvstore.ModifyCertification API synchronously
// api document: https://help.aliyun.com/api/r-kvstore/modifycertification.html
func (client *Client) ModifyCertification(request *ModifyCertificationRequest) (response *ModifyCertificationResponse, err error) {
	response = CreateModifyCertificationResponse()
	err = client.DoAction(request, response)
	return
}

// ModifyCertificationWithChan invokes the r_kvstore.ModifyCertification API asynchronously
// api document: https://help.aliyun.com/api/r-kvstore/modifycertification.html
// asynchronous document: https://help.aliyun.com/document_detail/66220.html
func (client *Client) ModifyCertificationWithChan(request *ModifyCertificationRequest) (<-chan *ModifyCertificationResponse, <-chan error) {
	responseChan := make(chan *ModifyCertificationResponse, 1)
	errChan := make(chan error, 1)
	err := client.AddAsyncTask(func() {
		defer close(responseChan)
		defer close(errChan)
		response, err := client.ModifyCertification(request)
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

// ModifyCertificationWithCallback invokes the r_kvstore.ModifyCertification API asynchronously
// api document: https://help.aliyun.com/api/r-kvstore/modifycertification.html
// asynchronous document: https://help.aliyun.com/document_detail/66220.html
func (client *Client) ModifyCertificationWithCallback(request *ModifyCertificationRequest, callback func(response *ModifyCertificationResponse, err error)) <-chan int {
	result := make(chan int, 1)
	err := client.AddAsyncTask(func() {
		var response *ModifyCertificationResponse
		var err error
		defer close(result)
		response, err = client.ModifyCertification(request)
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

// ModifyCertificationRequest is the request struct for api ModifyCertification
type ModifyCertificationRequest struct {
	*requests.RpcRequest
	ResourceOwnerId      requests.Integer `position:"Query" name:"ResourceOwnerId"`
	InstanceId           string           `position:"Query" name:"InstanceId"`
	SecurityToken        string           `position:"Query" name:"SecurityToken"`
	ResourceOwnerAccount string           `position:"Query" name:"ResourceOwnerAccount"`
	OwnerAccount         string           `position:"Query" name:"OwnerAccount"`
	OwnerId              requests.Integer `position:"Query" name:"OwnerId"`
	NoCertification      requests.Boolean `position:"Query" name:"NoCertification"`
}

// ModifyCertificationResponse is the response struct for api ModifyCertification
type ModifyCertificationResponse struct {
	*responses.BaseResponse
	RequestId string `json:"RequestId" xml:"RequestId"`
}

// CreateModifyCertificationRequest creates a request to invoke ModifyCertification API
func CreateModifyCertificationRequest() (request *ModifyCertificationRequest) {
	request = &ModifyCertificationRequest{
		RpcRequest: &requests.RpcRequest{},
	}
	request.InitWithApiInfo("R-kvstore", "2015-01-01", "ModifyCertification", "redisa", "openAPI")
	return
}

// CreateModifyCertificationResponse creates a response to parse from ModifyCertification response
func CreateModifyCertificationResponse() (response *ModifyCertificationResponse) {
	response = &ModifyCertificationResponse{
		BaseResponse: &responses.BaseResponse{},
	}
	return
}

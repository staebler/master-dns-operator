package dm

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

// QueryDomainByParam invokes the dm.QueryDomainByParam API synchronously
// api document: https://help.aliyun.com/api/dm/querydomainbyparam.html
func (client *Client) QueryDomainByParam(request *QueryDomainByParamRequest) (response *QueryDomainByParamResponse, err error) {
	response = CreateQueryDomainByParamResponse()
	err = client.DoAction(request, response)
	return
}

// QueryDomainByParamWithChan invokes the dm.QueryDomainByParam API asynchronously
// api document: https://help.aliyun.com/api/dm/querydomainbyparam.html
// asynchronous document: https://help.aliyun.com/document_detail/66220.html
func (client *Client) QueryDomainByParamWithChan(request *QueryDomainByParamRequest) (<-chan *QueryDomainByParamResponse, <-chan error) {
	responseChan := make(chan *QueryDomainByParamResponse, 1)
	errChan := make(chan error, 1)
	err := client.AddAsyncTask(func() {
		defer close(responseChan)
		defer close(errChan)
		response, err := client.QueryDomainByParam(request)
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

// QueryDomainByParamWithCallback invokes the dm.QueryDomainByParam API asynchronously
// api document: https://help.aliyun.com/api/dm/querydomainbyparam.html
// asynchronous document: https://help.aliyun.com/document_detail/66220.html
func (client *Client) QueryDomainByParamWithCallback(request *QueryDomainByParamRequest, callback func(response *QueryDomainByParamResponse, err error)) <-chan int {
	result := make(chan int, 1)
	err := client.AddAsyncTask(func() {
		var response *QueryDomainByParamResponse
		var err error
		defer close(result)
		response, err = client.QueryDomainByParam(request)
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

// QueryDomainByParamRequest is the request struct for api QueryDomainByParam
type QueryDomainByParamRequest struct {
	*requests.RpcRequest
	OwnerId              requests.Integer `position:"Query" name:"OwnerId"`
	ResourceOwnerAccount string           `position:"Query" name:"ResourceOwnerAccount"`
	ResourceOwnerId      requests.Integer `position:"Query" name:"ResourceOwnerId"`
	PageNo               requests.Integer `position:"Query" name:"PageNo"`
	PageSize             requests.Integer `position:"Query" name:"PageSize"`
	KeyWord              string           `position:"Query" name:"KeyWord"`
	Status               requests.Integer `position:"Query" name:"Status"`
}

// QueryDomainByParamResponse is the response struct for api QueryDomainByParam
type QueryDomainByParamResponse struct {
	*responses.BaseResponse
	RequestId  string                   `json:"RequestId" xml:"RequestId"`
	TotalCount int                      `json:"TotalCount" xml:"TotalCount"`
	PageNumber int                      `json:"PageNumber" xml:"PageNumber"`
	PageSize   int                      `json:"PageSize" xml:"PageSize"`
	Data       DataInQueryDomainByParam `json:"data" xml:"data"`
}

// CreateQueryDomainByParamRequest creates a request to invoke QueryDomainByParam API
func CreateQueryDomainByParamRequest() (request *QueryDomainByParamRequest) {
	request = &QueryDomainByParamRequest{
		RpcRequest: &requests.RpcRequest{},
	}
	request.InitWithApiInfo("Dm", "2015-11-23", "QueryDomainByParam", "", "")
	return
}

// CreateQueryDomainByParamResponse creates a response to parse from QueryDomainByParam response
func CreateQueryDomainByParamResponse() (response *QueryDomainByParamResponse) {
	response = &QueryDomainByParamResponse{
		BaseResponse: &responses.BaseResponse{},
	}
	return
}

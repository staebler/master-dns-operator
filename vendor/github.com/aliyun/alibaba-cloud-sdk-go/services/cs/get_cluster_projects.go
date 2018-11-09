package cs

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

// GetClusterProjects invokes the cs.GetClusterProjects API synchronously
// api document: https://help.aliyun.com/api/cs/getclusterprojects.html
func (client *Client) GetClusterProjects(request *GetClusterProjectsRequest) (response *GetClusterProjectsResponse, err error) {
	response = CreateGetClusterProjectsResponse()
	err = client.DoAction(request, response)
	return
}

// GetClusterProjectsWithChan invokes the cs.GetClusterProjects API asynchronously
// api document: https://help.aliyun.com/api/cs/getclusterprojects.html
// asynchronous document: https://help.aliyun.com/document_detail/66220.html
func (client *Client) GetClusterProjectsWithChan(request *GetClusterProjectsRequest) (<-chan *GetClusterProjectsResponse, <-chan error) {
	responseChan := make(chan *GetClusterProjectsResponse, 1)
	errChan := make(chan error, 1)
	err := client.AddAsyncTask(func() {
		defer close(responseChan)
		defer close(errChan)
		response, err := client.GetClusterProjects(request)
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

// GetClusterProjectsWithCallback invokes the cs.GetClusterProjects API asynchronously
// api document: https://help.aliyun.com/api/cs/getclusterprojects.html
// asynchronous document: https://help.aliyun.com/document_detail/66220.html
func (client *Client) GetClusterProjectsWithCallback(request *GetClusterProjectsRequest, callback func(response *GetClusterProjectsResponse, err error)) <-chan int {
	result := make(chan int, 1)
	err := client.AddAsyncTask(func() {
		var response *GetClusterProjectsResponse
		var err error
		defer close(result)
		response, err = client.GetClusterProjects(request)
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

// GetClusterProjectsRequest is the request struct for api GetClusterProjects
type GetClusterProjectsRequest struct {
	*requests.RoaRequest
	ClusterId string `position:"Path" name:"ClusterId"`
}

// GetClusterProjectsResponse is the response struct for api GetClusterProjects
type GetClusterProjectsResponse struct {
	*responses.BaseResponse
}

// CreateGetClusterProjectsRequest creates a request to invoke GetClusterProjects API
func CreateGetClusterProjectsRequest() (request *GetClusterProjectsRequest) {
	request = &GetClusterProjectsRequest{
		RoaRequest: &requests.RoaRequest{},
	}
	request.InitWithApiInfo("CS", "2015-12-15", "GetClusterProjects", "/clusters/[ClusterId]/projects", "", "")
	request.Method = requests.GET
	return
}

// CreateGetClusterProjectsResponse creates a response to parse from GetClusterProjects response
func CreateGetClusterProjectsResponse() (response *GetClusterProjectsResponse) {
	response = &GetClusterProjectsResponse{
		BaseResponse: &responses.BaseResponse{},
	}
	return
}

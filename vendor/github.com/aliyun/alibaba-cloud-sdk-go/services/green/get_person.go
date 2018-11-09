package green

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

// GetPerson invokes the green.GetPerson API synchronously
// api document: https://help.aliyun.com/api/green/getperson.html
func (client *Client) GetPerson(request *GetPersonRequest) (response *GetPersonResponse, err error) {
	response = CreateGetPersonResponse()
	err = client.DoAction(request, response)
	return
}

// GetPersonWithChan invokes the green.GetPerson API asynchronously
// api document: https://help.aliyun.com/api/green/getperson.html
// asynchronous document: https://help.aliyun.com/document_detail/66220.html
func (client *Client) GetPersonWithChan(request *GetPersonRequest) (<-chan *GetPersonResponse, <-chan error) {
	responseChan := make(chan *GetPersonResponse, 1)
	errChan := make(chan error, 1)
	err := client.AddAsyncTask(func() {
		defer close(responseChan)
		defer close(errChan)
		response, err := client.GetPerson(request)
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

// GetPersonWithCallback invokes the green.GetPerson API asynchronously
// api document: https://help.aliyun.com/api/green/getperson.html
// asynchronous document: https://help.aliyun.com/document_detail/66220.html
func (client *Client) GetPersonWithCallback(request *GetPersonRequest, callback func(response *GetPersonResponse, err error)) <-chan int {
	result := make(chan int, 1)
	err := client.AddAsyncTask(func() {
		var response *GetPersonResponse
		var err error
		defer close(result)
		response, err = client.GetPerson(request)
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

// GetPersonRequest is the request struct for api GetPerson
type GetPersonRequest struct {
	*requests.RoaRequest
	ClientInfo string `position:"Query" name:"ClientInfo"`
}

// GetPersonResponse is the response struct for api GetPerson
type GetPersonResponse struct {
	*responses.BaseResponse
}

// CreateGetPersonRequest creates a request to invoke GetPerson API
func CreateGetPersonRequest() (request *GetPersonRequest) {
	request = &GetPersonRequest{
		RoaRequest: &requests.RoaRequest{},
	}
	request.InitWithApiInfo("Green", "2018-05-09", "GetPerson", "/green/sface/person", "green", "openAPI")
	request.Method = requests.POST
	return
}

// CreateGetPersonResponse creates a response to parse from GetPerson response
func CreateGetPersonResponse() (response *GetPersonResponse) {
	response = &GetPersonResponse{
		BaseResponse: &responses.BaseResponse{},
	}
	return
}

package core

import (
	"fmt"
	"net/http"
	"net/url"

	"github.com/blendlabs/go-exception"
	"github.com/blendlabs/go-request"
)

type mockedResponse struct {
	ResponseBody []byte
	StatusCode   int
	Error        error
}

var isMocked bool
var mocks map[string]mockedResponse

// MockError simulates a mocked error
func MockError(verb string, url string) {
	Mock(
		verb,
		url,
		mockedResponse{
			StatusCode: http.StatusInternalServerError,
			Error:      exception.New("Error! This is from service_request#MockError. If you don't want an error don't mock it."),
		},
	)
}

// Mock mocks a response for a given verb to a given url.
func Mock(verb string, url string, res mockedResponse) {
	isMocked = true
	if mocks == nil {
		mocks = map[string]mockedResponse{}
	}
	storedURL := fmt.Sprintf("%s_%s", verb, url)
	mocks[storedURL] = res
}

// MockResponseFromBinary mocks a request from a byte array response.
func MockResponseFromBinary(verb string, url string, statusCode int, response []byte) {
	Mock(
		verb,
		url,
		mockedResponse{
			ResponseBody: response,
			StatusCode:   statusCode,
		},
	)
}

// MockResponseFromString mocks a request from a string response.
func MockResponseFromString(verb string, url string, statusCode int, response string) {
	MockResponseFromBinary(verb, url, statusCode, []byte(response))
}

// ClearMockedResponses clears and disables response mocking
func ClearMockedResponses() {
	isMocked = false
	mocks = map[string]mockedResponse{}
}

// NewExternalRequest Creates a new external request
func NewExternalRequest() *request.HttpRequest {
	req := request.NewRequest().WithMockedResponse(func(verb string, workingURL *url.URL) (bool, *request.HttpResponseMeta, []byte, error) {
		if isMocked {
			storedURL := fmt.Sprintf("%s_%s", verb, workingURL.String())
			if mockResponse, ok := mocks[storedURL]; ok {
				meta := &request.HttpResponseMeta{}
				meta.StatusCode = mockResponse.StatusCode
				meta.ContentLength = int64(len(mockResponse.ResponseBody))
				return true, meta, mockResponse.ResponseBody, mockResponse.Error
			}
			panic(fmt.Sprintf("attempted to make service request w/o mocking endpoint: %s", workingURL.String()))
		} else {
			return false, nil, nil, nil
		}
	})
	return req
}

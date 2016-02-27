package util

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"testing"

	"github.com/blendlabs/go-assert"
)

func TestGetIP(t *testing.T) {
	assert := assert.New(t)

	hdr := http.Header{}
	hdr.Set("X-Forwarded-For", "1")
	r := http.Request{
		Header: hdr,
	}
	assert.Equal("1", GetIP(&r))

	hdr = http.Header{}
	hdr.Set("X-FORWARDED-FOR", "1")
	r = http.Request{
		Header: hdr,
	}
	assert.Equal("1", GetIP(&r))

	hdr = http.Header{}
	hdr.Set("X-FORWARDED-FOR", "1,2,3")
	r = http.Request{
		Header: hdr,
	}
	assert.Equal("1", GetIP(&r))

	hdr = http.Header{}
	hdr.Set("X-Real-Ip", "1")
	r = http.Request{
		Header: hdr,
	}
	assert.Equal("1", GetIP(&r))

	hdr = http.Header{}
	hdr.Set("X-REAL-IP", "1")
	r = http.Request{
		Header: hdr,
	}
	assert.Equal("1", GetIP(&r))

	hdr = http.Header{}
	hdr.Set("X-REAL-IP", "1,2,3")
	r = http.Request{
		Header: hdr,
	}
	assert.Equal("1", GetIP(&r))

	r = http.Request{
		RemoteAddr: "1:1",
	}
	assert.Equal("1", GetIP(&r))
}

func createTestHttpRequest(paramLocation, paramKey, paramValue string) *http.Request {
	urlStr := "http://localhost/unit/test"
	if paramLocation == "query" {
		urlStr = fmt.Sprintf("http://localhost/unit/test?%s=%s", paramKey, paramValue)
	}

	req, _ := http.NewRequest("GET", urlStr, strings.NewReader(""))

	if paramLocation == "header" {
		req.Header.Add(paramKey, paramValue)
	}

	if paramLocation == "cookie" {
		formatted := fmt.Sprintf("%s=%s", paramKey, paramValue)
		req.Header.Add("Cookie", formatted)
	}

	if paramLocation == "formvalue" {
		req.Form = url.Values{}
		req.Form.Add(paramKey, paramValue)
	}

	return req
}

func TestGetParamByName(t *testing.T) {
	assert := assert.New(t)
	req := createTestHttpRequest("query", "test", "test")
	value := GetParamByName(req, "test")
	assert.NotEmpty(value, "query string")

	req = createTestHttpRequest("header", "test", "test")
	value = GetParamByName(req, "test")
	assert.NotEmpty(value, "header")

	req = createTestHttpRequest("cookie", "test", "test")
	value = GetParamByName(req, "test")
	assert.NotEmpty(value, "cookie")

	req = createTestHttpRequest("formvalue", "test", "test")
	value = GetParamByName(req, "test")
	assert.NotEmpty(value, "formvalue")
}

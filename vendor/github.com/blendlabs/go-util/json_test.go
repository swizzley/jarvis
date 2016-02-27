package util

import (
	"bytes"
	"io/ioutil"
	"strings"
	"testing"

	"github.com/blendlabs/go-assert"
)

type MockObject struct {
	Id    string `json:"id"`
	Email string `json:"email"`
}

func TestJson(t *testing.T) {
	assert := assert.New(t)

	obj_str := "{ \"id\" : \"test\", \"email\" : \"foo@bar.com\" }"
	obj := MockObject{}
	DeserializeJson(&obj, obj_str)
	assert.Equal("test", obj.Id)
	assert.Equal("foo@bar.com", obj.Email)

	new_obj := MockObject{}
	DeserializeJsonFromReader(&new_obj, bytes.NewBufferString(obj_str))
	assert.Equal("test", new_obj.Id)
	assert.Equal("foo@bar.com", new_obj.Email)

	serialized := SerializeJson(new_obj)
	assert.True(strings.Contains(serialized, "foo@bar.com"))

	serialized_reader := SerializeJsonAsReader(new_obj)
	serialized_reader_contents, reader_err := ioutil.ReadAll(serialized_reader)
	assert.Nil(reader_err)
	serialized_reader_str := string(serialized_reader_contents)
	assert.True(strings.Contains(serialized_reader_str, "foo@bar.com"))
}

func TestDeserializePostBody(t *testing.T) {
	assert := assert.New(t)

	obj_str := "{ \"id\" : \"test\", \"email\" : \"foo@bar.com\" }"
	obj_bytes := []byte(obj_str)
	obj_reader := bytes.NewReader(obj_bytes)
	obj_reader_closer := ioutil.NopCloser(obj_reader)

	mo := MockObject{}
	deserializeErr := DeserializeJsonFromReadCloser(&mo, obj_reader_closer)
	assert.Nil(deserializeErr)

	assert.Equal("test", mo.Id)
}

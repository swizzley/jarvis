package util

import (
	"bytes"
	"io/ioutil"
	"strings"
	"testing"

	"github.com/blendlabs/go-assert"
)

type mockObject struct {
	ID    string `json:"id"`
	Email string `json:"email"`
}

func TestJSON(t *testing.T) {
	assert := assert.New(t)

	objStr := "{ \"id\" : \"test\", \"email\" : \"foo@bar.com\" }"
	obj := mockObject{}
	DeserializeJSON(&obj, objStr)
	assert.Equal("test", obj.ID)
	assert.Equal("foo@bar.com", obj.Email)

	newObject := mockObject{}
	DeserializeJSONFromReader(&newObject, bytes.NewBufferString(objStr))
	assert.Equal("test", newObject.ID)
	assert.Equal("foo@bar.com", newObject.Email)

	serialized := SerializeJSON(obj)
	assert.True(strings.Contains(serialized, "foo@bar.com"))

	serializedReader := SerializeJSONAsReader(obj)
	serializedContents, err := ioutil.ReadAll(serializedReader)
	assert.Nil(err)
	serializedStr := string(serializedContents)
	assert.True(strings.Contains(serializedStr, "foo@bar.com"))
}

func TestDeserializePostBody(t *testing.T) {
	assert := assert.New(t)

	objStr := "{ \"id\" : \"test\", \"email\" : \"foo@bar.com\" }"
	objBytes := []byte(objStr)
	objReader := bytes.NewReader(objBytes)
	objReaderCloser := ioutil.NopCloser(objReader)

	mo := mockObject{}
	deserializeErr := DeserializeJSONFromReadCloser(&mo, objReaderCloser)
	assert.Nil(deserializeErr)

	assert.Equal("test", mo.ID)
}

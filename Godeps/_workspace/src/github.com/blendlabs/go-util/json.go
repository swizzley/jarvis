package util

import (
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"

	"github.com/wcharczuk/jarvis-cli/Godeps/_workspace/src/github.com/blendlabs/go-exception"
)

func ReadJsonFile(path string) string {
	bytes, _ := ioutil.ReadFile(path)
	return string(bytes)
}

func DeserializeJson(object interface{}, body string) error {
	decoder := json.NewDecoder(bytes.NewBufferString(body))
	return exception.Wrap(decoder.Decode(object))
}

func DeserializeJsonFromReader(object interface{}, body io.Reader) error {
	decoder := json.NewDecoder(body)
	return exception.Wrap(decoder.Decode(object))
}

func DeserializeJsonFromReadCloser(object interface{}, body io.ReadCloser) error {
	defer body.Close()
	bodyBytes, err := ioutil.ReadAll(body)
	if err != nil {
		return exception.Wrap(err)
	}

	decoder := json.NewDecoder(bytes.NewBuffer(bodyBytes))
	return exception.Wrap(decoder.Decode(object))
}

func SerializeJson(object interface{}) string {
	b, _ := json.Marshal(object)
	return string(b)
}

func SerializeJsonPretty(object interface{}, prefix, indent string) string {
	b, _ := json.MarshalIndent(object, prefix, indent)
	return string(b)
}

func SerializeJsonAsReader(object interface{}) io.Reader {
	b, _ := json.Marshal(object)
	return bytes.NewBufferString(string(b))
}

package util

import (
	"bytes"
	"encoding/xml"
	"io"
)

type charsetReader func(charset string, input io.Reader) (io.Reader, error)

func DeserializeXml(object interface{}, body string) error {
	return DeserializeXmlFromReader(object, bytes.NewBufferString(body))
}

func DeserializeXmlFromReader(object interface{}, reader io.Reader) error {
	decoder := xml.NewDecoder(reader)
	return decoder.Decode(object)
}

func DeserializeXmlFromReaderWithCharsetReader(object interface{}, body io.Reader, charsetReader charsetReader) error {
	decoder := xml.NewDecoder(body)
	decoder.CharsetReader = charsetReader
	return decoder.Decode(object)
}

func SerializeXml(object interface{}) string {
	b, _ := xml.Marshal(object)
	return string(b)
}

func SerializeXmlToReader(object interface{}) io.Reader {
	b, _ := xml.Marshal(object)
	return bytes.NewBufferString(string(b))
}

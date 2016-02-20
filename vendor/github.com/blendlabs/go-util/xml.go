package util

import (
	"bytes"
	"encoding/xml"
	"io"
	"regexp"
)

var cdataPrefix []byte = []byte("<![CDATA[")
var cdataSuffix []byte = []byte("]]>")

func EncodeCdata(data []byte) []byte {
	return bytes.Join([][]byte{cdataPrefix, data, cdataSuffix}, []byte{})
}

var cdataRe *regexp.Regexp = regexp.MustCompile("<!\\[CDATA\\[(.*?)\\]\\]>")

func DecodeCdata(cdata []byte) []byte {
	matches := cdataRe.FindAllSubmatch(cdata, 1)
	if len(matches) == 0 {
		return cdata
	}

	return matches[0][1]
}

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

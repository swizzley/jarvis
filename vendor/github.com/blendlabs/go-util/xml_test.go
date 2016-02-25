package util

import (
	"encoding/xml"
	"io/ioutil"
	"testing"

	"github.com/blendlabs/go-assert"
)

type MockXmlObject struct {
	XMLName       xml.Name `xml:"Person"`
	Id            string   `xml:"id,attr"`
	Email         string   `xml:"Email"`
	StreetAddress string   `xml:"Address>Street"`
}

func TestXml(t *testing.T) {
	assert := assert.New(t)

	obj_str := `<Person id="test"><Email>foo@bar.com</Email><Address><Street>123 Road St</Street></Address></Person>`
	obj := MockXmlObject{}
	DeserializeXml(&obj, obj_str)
	assert.Equal("test", obj.Id)
	assert.Equal("foo@bar.com", obj.Email)
	assert.Equal("123 Road St", obj.StreetAddress)

	serialized := SerializeXml(obj)
	assert.Equal(obj_str, serialized)

	serialized_reader := SerializeXmlToReader(obj)
	serialized_reader_contents, reader_err := ioutil.ReadAll(serialized_reader)
	assert.Nil(reader_err)
	serialized_reader_str := string(serialized_reader_contents)
	assert.Equal(obj_str, serialized_reader_str)
}

func TestCdata(t *testing.T) {
	assert := assert.New(t)

	assert.Equal("<![CDATA[test]]>", string(EncodeCdata([]byte("test"))))

	assert.Equal("<![CDATA[test", string(DecodeCdata([]byte("<![CDATA[test"))))
	assert.Len(DecodeCdata([]byte("<![CDATA[]]>")), 0)
	assert.Equal("test", string(DecodeCdata([]byte("<![CDATA[test]]>"))))
	assert.Equal("test", string(DecodeCdata([]byte(" <![CDATA[test]]>"))))
	assert.Equal("<![CDATA[test", string(DecodeCdata([]byte("<![CDATA[<![CDATA[test]]>]]>"))))
	assert.Equal("one", string(DecodeCdata([]byte("<![CDATA[one]]><![CDATA[two]]>"))))
}

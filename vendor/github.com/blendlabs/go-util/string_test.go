package util

import (
	"testing"

	"github.com/blendlabs/go-assert"
)

func TestRandomString(t *testing.T) {
	assert := assert.New(t)
	str := RandomString(10)
	assert.Equal(len(str), 10)
}

func TestCaseInsensitiveEquals(t *testing.T) {
	assert := assert.New(t)
	assert.True(CaseInsensitiveEquals("foo", "FOO"))
	assert.True(CaseInsensitiveEquals("foo123", "FOO123"))
	assert.True(CaseInsensitiveEquals("!foo123", "!foo123"))
	assert.False(CaseInsensitiveEquals("foo", "bar"))
}

func TestRegexMatch(t *testing.T) {
	assert := assert.New(t)

	result := RegexMatch("a", "b")
	assert.Equal("", result)
}

func TestParse(t *testing.T) {
	assert := assert.New(t)

	good := ParseFloat64("3.14")
	bad := ParseFloat64("I Am Dog")
	assert.Equal(3.14, good)
	assert.Equal(0.0, bad)

	good32 := ParseFloat32("3.14")
	bad32 := ParseFloat32("I Am Dog")
	assert.Equal(3.14, good32)
	assert.Equal(0.0, bad32)

	goodInt := ParseInt("3")
	badInt := ParseInt("I Am Dog")
	assert.Equal(3, goodInt)
	assert.Equal(0.0, badInt)

	strGood := Float64ToString(3.14)
	assert.Equal("3.14", strGood)

	strGood = Float32ToString(3.14)
	assert.Equal("3.14", strGood)

	strGood = IntToString(3)
	assert.Equal("3", strGood)
}

func TestTrimWhitespace(t *testing.T) {
	assert := assert.New(t)

	tests := []KeyValuePair{
		KeyValuePair{"test", "test"},
		KeyValuePair{" test", "test"},
		KeyValuePair{"test ", "test"},
		KeyValuePair{" test ", "test"},
		KeyValuePair{"\ttest", "test"},
		KeyValuePair{"test\t", "test"},
		KeyValuePair{"\ttest\t", "test"},
		KeyValuePair{" \ttest\t ", "test"},
		KeyValuePair{" \ttest\n\t ", "test\n"},
	}

	for _, test := range tests {
		result := TrimWhitespace(test.Key)
		assert.Equal(test.Value, result)
	}
}

func TestIsCamelCase(t *testing.T) {
	assert := assert.New(t)

	assert.True(IsCamelCase("McDonald"))
	assert.True(IsCamelCase("mcDonald"))
	assert.False(IsCamelCase("mcdonald"))
	assert.False(IsCamelCase("MCDONALD"))
}

func TestCombinePathComponents(t *testing.T) {
	assert := assert.New(t)

	value := CombinePathComponents("foo")
	assert.Equal("foo", value)

	value = CombinePathComponents("/foo")
	assert.Equal("foo", value)

	value = CombinePathComponents("foo/")
	assert.Equal("foo", value)

	value = CombinePathComponents("/foo/")
	assert.Equal("foo", value)

	value = CombinePathComponents("foo", "bar")
	assert.Equal("foo/bar", value)

	value = CombinePathComponents("foo/", "bar")
	assert.Equal("foo/bar", value)

	value = CombinePathComponents("foo/", "/bar")
	assert.Equal("foo/bar", value)

	value = CombinePathComponents("/foo/", "/bar")
	assert.Equal("foo/bar", value)

	value = CombinePathComponents("/foo/", "/bar/")
	assert.Equal("foo/bar", value)

	value = CombinePathComponents("foo", "bar", "baz")
	assert.Equal("foo/bar/baz", value)

	value = CombinePathComponents("foo/", "bar/", "baz")
	assert.Equal("foo/bar/baz", value)

	value = CombinePathComponents("foo/", "bar/", "baz/")
	assert.Equal("foo/bar/baz", value)

	value = CombinePathComponents("foo/", "/bar/", "/baz")
	assert.Equal("foo/bar/baz", value)

	value = CombinePathComponents("/foo/", "/bar/", "/baz")
	assert.Equal("foo/bar/baz", value)

	value = CombinePathComponents("/foo/", "/bar/", "/baz/")
	assert.Equal("foo/bar/baz", value)
}

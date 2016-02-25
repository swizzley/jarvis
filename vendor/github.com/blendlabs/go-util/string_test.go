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

	str_good := Float64ToString(3.14)
	assert.Equal("3.14", str_good)

	str_good = Float32ToString(3.14)
	assert.Equal("3.14", str_good)

	str_good = IntToString(3)
	assert.Equal("3", str_good)
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

	brian_is_a_pedant := CombinePathComponents("foo")
	assert.Equal("foo", brian_is_a_pedant)

	brian_is_a_pedant2 := CombinePathComponents("/foo")
	assert.Equal("foo", brian_is_a_pedant2)

	brian_is_a_pedant3 := CombinePathComponents("foo/")
	assert.Equal("foo", brian_is_a_pedant3)

	brian_is_a_pedant4 := CombinePathComponents("/foo/")
	assert.Equal("foo", brian_is_a_pedant4)

	dual_test1 := CombinePathComponents("foo", "bar")
	assert.Equal("foo/bar", dual_test1)

	dual_test2 := CombinePathComponents("foo/", "bar")
	assert.Equal("foo/bar", dual_test2)

	dual_test3 := CombinePathComponents("foo/", "/bar")
	assert.Equal("foo/bar", dual_test3)

	dual_test4 := CombinePathComponents("/foo/", "/bar")
	assert.Equal("foo/bar", dual_test4)

	dual_test5 := CombinePathComponents("/foo/", "/bar/")
	assert.Equal("foo/bar", dual_test5)

	test1 := CombinePathComponents("foo", "bar", "baz")
	assert.Equal("foo/bar/baz", test1)

	test2 := CombinePathComponents("foo/", "bar/", "baz")
	assert.Equal("foo/bar/baz", test2)

	test3 := CombinePathComponents("foo/", "bar/", "baz/")
	assert.Equal("foo/bar/baz", test3)

	test4 := CombinePathComponents("foo/", "/bar/", "/baz")
	assert.Equal("foo/bar/baz", test4)

	test5 := CombinePathComponents("/foo/", "/bar/", "/baz")
	assert.Equal("foo/bar/baz", test5)

	test6 := CombinePathComponents("/foo/", "/bar/", "/baz/")
	assert.Equal("foo/bar/baz", test6)
}

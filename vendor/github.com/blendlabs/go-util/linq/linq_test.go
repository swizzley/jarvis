package linq

import (
	"strings"
	"testing"

	"github.com/blendlabs/go-assert"
	"github.com/blendlabs/go-util"
)

func TestAny(t *testing.T) {
	assert := assert.New(t)

	objs := []util.KeyValuePair{util.KeyValuePair{Key: "Foo", Value: "Bar"}, util.KeyValuePair{Key: "Foo2", Value: "Baz"}, util.KeyValuePair{Key: "Foo3", Value: 3}}

	hasBar := Any(objs, func(value interface{}) bool {
		if kvpValue, isKvp := value.(util.KeyValuePair); isKvp {
			if stringValue, isString := kvpValue.Value.(string); isString {
				return stringValue == "Bar"
			} else {
				return false
			}
		} else {
			return false
		}
	})
	assert.True(hasBar)
	hasFizz := Any(objs, func(value interface{}) bool {
		if kvpValue, isKvp := value.(util.KeyValuePair); isKvp {
			if stringValue, isString := kvpValue.Value.(string); isString {
				return stringValue == "Fizz"
			} else {
				return false
			}
		} else {
			return false
		}
	})
	assert.False(hasFizz)
	assert.True(Any(objs, DeepEqual(util.KeyValuePair{Key: "Foo3", Value: 3})))
}

func TestAnyOfStringIgnoreCase(t *testing.T) {
	assert := assert.New(t)
	items := []string{"foo", "FOO", "bar", "Bar", "BAR", "baZ"}
	assert.True(AnyOfString(items, EqualsCaseInsenitive("baz")))
	assert.False(AnyOfString(items, EqualsCaseInsenitive("will")))
}

func TestAll(t *testing.T) {
	assert := assert.New(t)

	objs := []util.KeyValuePair{util.KeyValuePair{Key: "Foo", Value: "Bar"}, util.KeyValuePair{Key: "Foo2", Value: "Baz"}, util.KeyValuePair{Key: "Foo3", Value: 3}}

	hasBar := All(objs, func(value interface{}) bool {
		if kvpValue, isKvp := value.(util.KeyValuePair); isKvp {
			return strings.HasPrefix(kvpValue.Key, "Foo")
		} else {
			return false
		}
	})
	assert.True(hasBar)
	hasFizz := All(objs, func(value interface{}) bool {
		if kvpValue, isKvp := value.(util.KeyValuePair); isKvp {
			return strings.HasPrefix(kvpValue.Key, "Fizz")
		} else {
			return false
		}
	})
	assert.False(hasFizz)
}

func TestFirst(t *testing.T) {
	assert := assert.New(t)

	items := []float64{6.1, 3.2, 4.0, 12.4, 912.4, 912.3, 3.14}
	bigNumber := FirstOfFloat(items, func(v float64) bool {
		return v > 900
	})

	assert.NotNil(bigNumber)
	assert.Equal(912.4, *bigNumber)
}

func TestLast(t *testing.T) {
	assert := assert.New(t)

	items := []float64{6.1, 3.2, 4.0, 12.4, 912.4, 912.3, 3.14}
	bigNumber := LastOfFloat(items, func(v float64) bool {
		return v > 900
	})

	assert.NotNil(bigNumber)
	assert.Equal(912.3, *bigNumber)
}

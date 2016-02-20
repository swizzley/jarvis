package util

import (
	"testing"

	"github.com/blendlabs/go-assert"
)

type SubType struct {
	Id   int    `json:"id"`
	Name string `json:"name"`
}

type TestType struct {
	Id        int    `json:"id"`
	Name      string `json:"name"`
	NotTagged string
	Tagged    string    `json:"is_tagged"`
	SubTypes  []SubType `json:"children"`
}

func TestDecomposeToPostData(t *testing.T) {
	assert := assert.New(t)

	my_obj := TestType{}
	my_obj.Id = 123
	my_obj.Name = "Test Object"
	my_obj.NotTagged = "Not Tagged"
	my_obj.Tagged = "Is Tagged"
	my_obj.SubTypes = append([]SubType{}, SubType{1, "One"})
	my_obj.SubTypes = append(my_obj.SubTypes, SubType{2, "Two"})
	my_obj.SubTypes = append(my_obj.SubTypes, SubType{3, "Three"})
	my_obj.SubTypes = append(my_obj.SubTypes, SubType{4, "Four"})

	post_datums := DecomposeToPostData(my_obj)

	assert.NotEmpty(post_datums)

	assert.Equal("id", post_datums[0].Key)
	assert.Equal("123", post_datums[0].Value)

	assert.Equal("name", post_datums[1].Key)
	assert.Equal("Test Object", post_datums[1].Value)

	assert.Equal("NotTagged", post_datums[2].Key)
	assert.Equal("Not Tagged", post_datums[2].Value)

	assert.Equal("is_tagged", post_datums[3].Key)
	assert.Equal("Is Tagged", post_datums[3].Value)

	assert.Equal("children[0].id", post_datums[4].Key)
	assert.Equal("1", post_datums[4].Value)

	assert.Equal("children[0].name", post_datums[5].Key)
	assert.Equal("One", post_datums[5].Value)

	assert.Equal("children[1].id", post_datums[6].Key)
	assert.Equal("2", post_datums[6].Value)
}

func TestDecomposeToPostDataAsJson(t *testing.T) {
	assert := assert.New(t)

	my_obj := TestType{}
	my_obj.Id = 123
	my_obj.Name = "Test Object"
	my_obj.NotTagged = "Not Tagged"
	my_obj.Tagged = "Is Tagged"
	my_obj.SubTypes = append([]SubType{}, SubType{1, "One"})
	my_obj.SubTypes = append(my_obj.SubTypes, SubType{2, "Two"})
	my_obj.SubTypes = append(my_obj.SubTypes, SubType{3, "Three"})
	my_obj.SubTypes = append(my_obj.SubTypes, SubType{4, "Four"})

	post_datums := DecomposeToPostDataAsJson(my_obj)

	assert.NotEmpty(post_datums)
	assert.Equal("id", post_datums[0].Key)
	assert.Equal("123", post_datums[0].Value)

	assert.Equal("name", post_datums[1].Key)
	assert.Equal("Test Object", post_datums[1].Value)

	assert.Equal("NotTagged", post_datums[2].Key)
	assert.Equal("Not Tagged", post_datums[2].Value)

	assert.Equal("is_tagged", post_datums[3].Key)
	assert.Equal("Is Tagged", post_datums[3].Value)

	assert.Equal("children", post_datums[4].Key)
	assert.NotEmpty(post_datums[4].Value)

	verify := []SubType{}
	verifyErr := DeserializeJson(&verify, post_datums[4].Value)
	assert.Nil(verifyErr)
	assert.Equal(1, verify[0].Id)
}

type TestType2 struct {
	SomeVal    string `coalesce:"SomeVal2"`
	SomeVal2   string
	OtherVal   string `coalesce:"OtherVal2,OtherVal3"`
	OtherVal2  string
	OtherVal3  string
	StructVal  SubType `coalesce:"StructVal2"`
	StructVal2 SubType
}

func TestCoalesceFieldsNoChange(t *testing.T) {
	assert := assert.New(t)

	testVal := TestType2{
		SomeVal:    "Foo1",
		SomeVal2:   "Foo2",
		OtherVal:   "Foo3",
		OtherVal2:  "Foo4",
		OtherVal3:  "Foo5",
		StructVal:  SubType{1, "Name"},
		StructVal2: SubType{2, "Name2"},
	}

	CoalesceFields(&testVal)
	assert.Equal("Foo1", testVal.SomeVal)
	assert.Equal("Foo3", testVal.OtherVal)
	assert.Equal(SubType{1, "Name"}, testVal.StructVal)
	// omit values to check that values coalesce
}

func TestCoalesceFieldsSimple(t *testing.T) {
	assert := assert.New(t)
	testVal2 := TestType2{
		SomeVal2:   "Foo2",
		OtherVal2:  "Foo4",
		OtherVal3:  "Foo5",
		StructVal2: SubType{2, "Name2"},
	}
	CoalesceFields(&testVal2)
	assert.Equal("Foo2", testVal2.SomeVal)
	assert.Equal("Foo4", testVal2.OtherVal)
	assert.Equal(SubType{2, "Name2"}, testVal2.StructVal)
}

func TestCoalesceFieldsStructAndMultiples(t *testing.T) {
	assert := assert.New(t)
	// omit values to check that values coalesce
	testVal3 := TestType2{
		SomeVal:   "Foo1",
		OtherVal3: "Foo5",
		StructVal: SubType{1, "Name"},
	}
	CoalesceFields(&testVal3)
	assert.Equal("Foo5", testVal3.OtherVal)
}

type TestType3 struct {
	Sub  SubType2 `coalesce:"Sub2"`
	Sub2 SubType2
}

type SubType2 struct {
	Val1 string `coalesce:"Val2"`
	Val2 string
}

func TestCoalesceFieldsNested(t *testing.T) {
	assert := assert.New(t)
	t1 := TestType3{
		Sub: SubType2{"", "foo"},
	}

	CoalesceFields(&t1)
	assert.Equal("foo", t1.Sub.Val1)

	t2 := TestType3{
		Sub2: SubType2{"", "foo2"},
	}

	CoalesceFields(&t2)
	assert.Equal("foo2", t2.Sub.Val1)
}

type TestType4 struct {
	Subs []SubType2
}

func TestCoalesceFieldsArray(t *testing.T) {
	assert := assert.New(t)
	t1 := TestType4{[]SubType2{{"", "foo"}, SubType2{"foo2", ""}}}
	CoalesceFields(&t1)
	assert.Equal("foo", t1.Subs[0].Val1)
	assert.Equal("foo2", t1.Subs[1].Val1)
}

func TestPatchObject(t *testing.T) {
	assert := assert.New(t)

	my_obj := TestType{}
	my_obj.Id = 123
	my_obj.Name = "Test Object"
	my_obj.NotTagged = "Not Tagged"
	my_obj.Tagged = "Is Tagged"
	my_obj.SubTypes = append([]SubType{}, SubType{1, "One"})
	my_obj.SubTypes = append(my_obj.SubTypes, SubType{2, "Two"})
	my_obj.SubTypes = append(my_obj.SubTypes, SubType{3, "Three"})
	my_obj.SubTypes = append(my_obj.SubTypes, SubType{4, "Four"})

	patch_data := make(map[string]interface{})
	patch_data["is_tagged"] = "Is Not Tagged"

	patch_err := PatchObject(&my_obj, patch_data)
	assert.Nil(patch_err)
	assert.Equal("Is Not Tagged", my_obj.Tagged)
}

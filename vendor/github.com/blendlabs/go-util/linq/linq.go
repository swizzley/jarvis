package linq

import (
	"fmt"
	"reflect"
	"strconv"
	"time"

	"github.com/blendlabs/go-util"
)

type Predicate func(item interface{}) bool
type PredicateOfInt func(item int) bool
type PredicateOfFloat func(item float64) bool
type PredicateOfString func(item string) bool
type PredicateOfTime func(item time.Time) bool

func ReturnsTrue() Predicate {
	return func(_ interface{}) bool {
		return true
	}
}

func ReturnsTrueOfString() PredicateOfString {
	return func(_ string) bool {
		return true
	}
}

func ReturnsTrueOfInt() PredicateOfInt {
	return func(_ int) bool {
		return true
	}
}

func ReturnsTrueOfFloat() PredicateOfFloat {
	return func(_ float64) bool {
		return true
	}
}

func ReturnsTrueOfTime() PredicateOfTime {
	return func(_ time.Time) bool {
		return true
	}
}

func ReturnsFalse() Predicate {
	return func(_ interface{}) bool {
		return false
	}
}

func ReturnsFalseOfString() PredicateOfString {
	return func(_ string) bool {
		return false
	}
}

func ReturnsFalseOfInt() PredicateOfInt {
	return func(_ int) bool {
		return false
	}
}

func ReturnsFalseOfFloat() PredicateOfFloat {
	return func(_ float64) bool {
		return false
	}
}

func ReturnsFalseOfTime() PredicateOfTime {
	return func(_ time.Time) bool {
		return false
	}
}

func DeepEqual(shouldBe interface{}) Predicate {
	return func(value interface{}) bool {
		return reflect.DeepEqual(shouldBe, value)
	}
}

func EqualsOfInt(shouldBe int) PredicateOfInt {
	return func(value int) bool {
		return shouldBe == value
	}
}

func EqualsOfFloat(shouldBe float64) PredicateOfFloat {
	return func(value float64) bool {
		return shouldBe == value
	}
}

func EqualsOfString(shouldBe string) PredicateOfString {
	return func(value string) bool {
		return shouldBe == value
	}
}

func EqualsCaseInsenitive(shouldBe string) PredicateOfString {
	return func(value string) bool {
		return util.CaseInsensitiveEquals(shouldBe, value)
	}
}

type MapAction func(item interface{}) interface{}

func ToInt(item interface{}) interface{} {
	if itemAsString, isString := item.(string); isString {
		if intValue, intValueErr := strconv.Atoi(itemAsString); intValueErr == nil {
			return intValue
		}
	}

	return nil
}

func ToFloat(item interface{}) interface{} {
	if itemAsString, isString := item.(string); isString {
		if floatValue, floatValueErr := strconv.ParseFloat(itemAsString, 64); floatValueErr == nil {
			return floatValue
		}
	}

	return nil
}

func ToString(item interface{}) interface{} {
	return fmt.Sprintf("%v", item)
}

func Any(target interface{}, predicate Predicate) bool {
	t := reflect.TypeOf(target)
	for t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	v := reflect.ValueOf(target)
	for v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	if t.Kind() != reflect.Slice {
		return false
	}

	for x := 0; x < v.Len(); x++ {
		obj := v.Index(x).Interface()
		if predicate == nil || predicate(obj) {
			return true
		}
	}
	return false
}

func AnyOfInt(target []int, predicate PredicateOfInt) bool {
	v := reflect.ValueOf(target)

	for x := 0; x < v.Len(); x++ {
		obj := v.Index(x).Interface().(int)
		if predicate == nil || predicate(obj) {
			return true
		}
	}
	return false
}

func AnyOfFloat(target []float64, predicate PredicateOfFloat) bool {
	v := reflect.ValueOf(target)

	for x := 0; x < v.Len(); x++ {
		obj := v.Index(x).Interface().(float64)
		if predicate == nil || predicate(obj) {
			return true
		}
	}
	return false
}

func AnyOfString(target []string, predicate PredicateOfString) bool {
	v := reflect.ValueOf(target)

	for x := 0; x < v.Len(); x++ {
		obj := v.Index(x).Interface().(string)
		if predicate == nil || predicate(obj) {
			return true
		}
	}
	return false
}

func AnyOfTime(target []time.Time, predicate PredicateOfTime) bool {
	v := reflect.ValueOf(target)

	for x := 0; x < v.Len(); x++ {
		obj := v.Index(x).Interface().(time.Time)
		if predicate == nil || predicate(obj) {
			return true
		}
	}
	return false
}

func All(target interface{}, predicate Predicate) bool {
	t := reflect.TypeOf(target)
	for t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	v := reflect.ValueOf(target)
	for v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	if t.Kind() != reflect.Slice {
		return false
	}

	for x := 0; x < v.Len(); x++ {
		obj := v.Index(x).Interface()
		if !predicate(obj) {
			return false
		}
	}
	return true
}

func AllOfInt(target []int, predicate PredicateOfInt) bool {
	v := reflect.ValueOf(target)

	for x := 0; x < v.Len(); x++ {
		obj := v.Index(x).Interface().(int)
		if !predicate(obj) {
			return false
		}
	}
	return true
}

func AllOfFloat(target []float64, predicate PredicateOfFloat) bool {
	v := reflect.ValueOf(target)

	for x := 0; x < v.Len(); x++ {
		obj := v.Index(x).Interface().(float64)
		if !predicate(obj) {
			return false
		}
	}
	return true
}

func AllOfString(target []string, predicate PredicateOfString) bool {
	v := reflect.ValueOf(target)

	for x := 0; x < v.Len(); x++ {
		obj := v.Index(x).Interface().(string)
		if !predicate(obj) {
			return false
		}
	}
	return true
}

func AllOfTime(target []time.Time, predicate PredicateOfTime) bool {
	v := reflect.ValueOf(target)

	for x := 0; x < v.Len(); x++ {
		obj := v.Index(x).Interface().(time.Time)
		if !predicate(obj) {
			return false
		}
	}
	return true
}

func First(target interface{}, predicate Predicate) interface{} {
	t := reflect.TypeOf(target)
	for t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	v := reflect.ValueOf(target)
	for v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	if t.Kind() != reflect.Slice {
		return false
	}

	for x := 0; x < v.Len(); x++ {
		obj := v.Index(x).Interface()
		if predicate == nil || predicate(obj) {
			return obj
		}
	}
	return nil
}

func FirstOfInt(target []int, predicate PredicateOfInt) *int {
	v := reflect.ValueOf(target)

	for x := 0; x < v.Len(); x++ {
		obj := v.Index(x).Interface().(int)
		if predicate == nil || predicate(obj) {
			return &obj
		}
	}
	return nil
}

func FirstOfFloat(target []float64, predicate PredicateOfFloat) *float64 {
	v := reflect.ValueOf(target)

	for x := 0; x < v.Len(); x++ {
		obj := v.Index(x).Interface().(float64)
		if predicate == nil || predicate(obj) {
			return &obj
		}
	}
	return nil
}

func FirstOfString(target []string, predicate PredicateOfString) *string {
	v := reflect.ValueOf(target)

	for x := 0; x < v.Len(); x++ {
		obj := v.Index(x).Interface().(string)
		if predicate == nil || predicate(obj) {
			return &obj
		}
	}
	return nil
}

func FirstOfTime(target []time.Time, predicate PredicateOfTime) *time.Time {
	v := reflect.ValueOf(target)

	for x := 0; x < v.Len(); x++ {
		obj := v.Index(x).Interface().(time.Time)
		if predicate == nil || predicate(obj) {
			return &obj
		}
	}
	return nil
}

func Last(target interface{}, predicate Predicate) interface{} {
	t := reflect.TypeOf(target)
	for t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	v := reflect.ValueOf(target)
	for v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	if t.Kind() != reflect.Slice {
		return false
	}

	for x := v.Len() - 1; x > 0; x-- {
		obj := v.Index(x).Interface()
		if predicate == nil || predicate(obj) {
			return obj
		}
	}

	return nil
}

func LastOfInt(target []int, predicate PredicateOfInt) *int {
	v := reflect.ValueOf(target)

	for x := v.Len() - 1; x > 0; x-- {
		obj := v.Index(x).Interface().(int)
		if predicate == nil || predicate(obj) {
			return &obj
		}
	}
	return nil
}

func LastOfFloat(target []float64, predicate PredicateOfFloat) *float64 {
	v := reflect.ValueOf(target)

	for x := v.Len() - 1; x > 0; x-- {
		obj := v.Index(x).Interface().(float64)
		if predicate == nil || predicate(obj) {
			return &obj
		}
	}
	return nil
}

func LastOfString(target []string, predicate PredicateOfString) *string {
	v := reflect.ValueOf(target)

	for x := v.Len() - 1; x > 0; x-- {
		obj := v.Index(x).Interface().(string)
		if predicate == nil || predicate(obj) {
			return &obj
		}
	}
	return nil
}

func LastOfTime(target []time.Time, predicate PredicateOfTime) *time.Time {
	v := reflect.ValueOf(target)

	for x := v.Len() - 1; x > 0; x-- {
		obj := v.Index(x).Interface().(time.Time)
		if predicate == nil || predicate(obj) {
			return &obj
		}
	}
	return nil
}

func Select(target interface{}, mapFn MapAction) []interface{} {
	t := reflect.TypeOf(target)
	for t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	v := reflect.ValueOf(target)
	for v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	if t.Kind() != reflect.Slice {
		panic("cannot map non-slice.")
	}

	values := []interface{}{}
	for x := 0; x < v.Len(); x++ {
		obj := v.Index(x).Interface()
		values = append(values, mapFn(obj))
	}
	return values
}

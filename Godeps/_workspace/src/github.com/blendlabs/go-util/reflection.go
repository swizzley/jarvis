package util

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/wcharczuk/jarvis-cli/Godeps/_workspace/src/github.com/blendlabs/go-exception"
)

func FollowValuePointer(v reflect.Value) interface{} {
	if v.Kind() == reflect.Ptr && v.IsNil() {
		return nil
	}

	val := v
	for val.Kind() == reflect.Ptr {
		val = val.Elem()
	}
	return val.Interface()
}

func FollowValue(v reflect.Value) reflect.Value {
	for v.Kind() == reflect.Ptr || v.Kind() == reflect.Interface {
		v = v.Elem()
	}
	return v
}

func ReflectValue(obj interface{}) reflect.Value {
	v := reflect.ValueOf(obj)
	for v.Kind() == reflect.Ptr || v.Kind() == reflect.Interface {
		v = v.Elem()
	}
	return v
}

func ReflectType(obj interface{}) reflect.Type {
	t := reflect.TypeOf(obj)
	for t.Kind() == reflect.Ptr || t.Kind() == reflect.Interface {
		t = t.Elem()
	}

	return t
}

func MakeNew(t reflect.Type) interface{} {
	return reflect.New(t).Interface()
}

func MakeSliceOfType(t reflect.Type) interface{} {
	return reflect.New(reflect.SliceOf(t)).Interface()
}

func TypeName(obj interface{}) string {
	t := ReflectType(obj)
	return t.Name()
}

func GetValueByName(target interface{}, field_name string) interface{} {
	target_value := ReflectValue(target)
	field := target_value.FieldByName(field_name)
	return field.Interface()
}

func GetFieldByNameOrJsonTag(target_value reflect.Type, field_name string) *reflect.StructField {
	for index := 0; index < target_value.NumField(); index++ {
		field := target_value.Field(index)

		if field.Name == field_name {
			return &field
		} else {
			tag := field.Tag
			json_tag := tag.Get("json")
			if strings.Contains(json_tag, field_name) {
				return &field
			}
		}
	}

	return nil
}

func SetValueByName(target interface{}, field_name string, field_value interface{}) error {
	target_value := ReflectValue(target)
	target_type := ReflectType(target)
	relevant_field := GetFieldByNameOrJsonTag(target_type, field_name)

	if relevant_field == nil {
		return exception.New(fmt.Sprintf("Invalid field for %s : `%s`", target_type.Name(), field_name))
	}

	field := target_value.FieldByName(relevant_field.Name)
	field_type := field.Type()
	if field.CanSet() {
		value_reflected := ReflectValue(field_value)
		if value_reflected.IsValid() {
			if value_reflected.Type().AssignableTo(field_type) {
				if field.Kind() == reflect.Ptr && value_reflected.CanAddr() {
					field.Set(value_reflected.Addr())
				} else {
					field.Set(value_reflected)
				}
			} else {
				if field.Kind() == reflect.Ptr {
					if value_reflected.CanAddr() {
						converted_value := value_reflected.Convert(field_type.Elem())
						if converted_value.CanAddr() {
							field.Set(converted_value.Addr())
						}
					}
				} else {
					converted_value := value_reflected.Convert(field_type)
					field.Set(converted_value)
				}
			}
		} else {
			return exception.New(fmt.Sprintf("Invalid field for %s : `%s`", target_type.Name(), field_name))
		}
	} else {
		return exception.New(fmt.Sprintf("Cannot set field for %s : `%s`", target_type.Name(), field_name))
	}
	return nil
}

func PatchObject(obj interface{}, patch_values map[string]interface{}) error {
	for key, value := range patch_values {
		set_err := SetValueByName(obj, key, value)
		if set_err != nil {
			return set_err
		}
	}
	return nil
}

func DecomposeToPostData(object interface{}) []KeyValuePairOfString {
	kvps := []KeyValuePairOfString{}

	obj_type := ReflectType(object)
	obj_value := ReflectValue(object)

	number_of_fields := obj_type.NumField()
	for index := 0; index < number_of_fields; index++ {
		field := obj_type.Field(index)
		value_field := obj_value.Field(index)

		kvp := KeyValuePairOfString{}

		if !field.Anonymous {
			tag := field.Tag.Get("json")
			if len(tag) != 0 {
				if strings.Contains(tag, ",") {
					parts := strings.Split(tag, ",")
					kvp.Key = parts[0]
				} else {
					kvp.Key = tag
				}
			} else {
				kvp.Key = field.Name
			}

			if field.Type.Kind() == reflect.Slice {
				//do something special
				for sub_index := 0; sub_index < value_field.Len(); sub_index++ {
					item_at_index := value_field.Index(sub_index).Interface()
					for _, prop := range DecomposeToPostData(item_at_index) {
						if len(prop.Value) != 0 { //this is a gutcheck, it shouldn't be needed
							ikvp := KeyValuePairOfString{}
							ikvp.Key = fmt.Sprintf("%s[%d].%s", kvp.Key, sub_index, prop.Key)
							ikvp.Value = prop.Value
							kvps = append(kvps, ikvp)
						}
					}
				}
			} else {
				value := FollowValuePointer(value_field)
				if value != nil {
					kvp.Value = fmt.Sprintf("%v", value)
					if len(kvp.Value) != 0 {
						kvps = append(kvps, kvp)
					}
				}
			}
		}
	}

	return kvps
}

func DecomposeToPostDataAsJson(object interface{}) []KeyValuePairOfString {
	kvps := []KeyValuePairOfString{}

	obj_type := ReflectType(object)
	obj_value := ReflectValue(object)

	number_of_fields := obj_type.NumField()
	for index := 0; index < number_of_fields; index++ {
		field := obj_type.Field(index)
		value_field := obj_value.Field(index)

		kvp := KeyValuePairOfString{}

		if !field.Anonymous {
			tag := field.Tag.Get("json")
			if len(tag) != 0 {
				if strings.Contains(tag, ",") {
					parts := strings.Split(tag, ",")
					kvp.Key = parts[0]
				} else {
					kvp.Key = tag
				}
			} else {
				kvp.Key = field.Name
			}

			value_dereferenced := FollowValue(value_field)
			value := FollowValuePointer(value_field)
			if value != nil {
				if value_dereferenced.Kind() == reflect.Slice || value_dereferenced.Kind() == reflect.Map {
					kvp.Value = SerializeJson(value)
				} else {
					kvp.Value = fmt.Sprintf("%v", value)
				}
			}

			if len(kvp.Value) != 0 {
				kvps = append(kvps, kvp)
			}
		}
	}

	return kvps
}

// checks if a value is a zero value or its types default value
func isZero(v reflect.Value) bool {
	if !v.IsValid() {
		return true
	}
	switch v.Kind() {
	case reflect.Func, reflect.Map, reflect.Slice:
		return v.IsNil()
	case reflect.Array:
		z := true
		for i := 0; i < v.Len(); i++ {
			z = z && isZero(v.Index(i))
		}
		return z
	case reflect.Struct:
		z := true
		for i := 0; i < v.NumField(); i++ {
			z = z && isZero(v.Field(i))
		}
		return z
	}
	// Compare other types directly:
	z := reflect.Zero(v.Type())
	return v.Interface() == z.Interface()
}

// Given a the name of a type variable, determines if the variable is exported
// by checking if first variable is capitalilzed
func isExported(fieldName string) bool {
	return fieldName != "" && strings.ToUpper(fieldName)[0] == fieldName[0]
}

func CoalesceFields(object interface{}) {
	object_value := ReflectValue(object)
	object_type := ReflectType(object)
	if object_type.Kind() == reflect.Struct {
		number_of_fields := object_value.NumField()
		for index := 0; index < number_of_fields; index++ {
			field := object_type.Field(index)
			field_value := object_value.Field(index)
			// only alter the field if it is exported (uppercase variable name) and is not already a non-zero value
			if isExported(field.Name) && isZero(field_value) {
				alternate_field_names := strings.Split(field.Tag.Get("coalesce"), ",")

				// find the first non-zero value in the list of backup values
				for j := 0; j < len(alternate_field_names); j++ {
					alternate_field_name := alternate_field_names[j]
					alternate_value := object_value.FieldByName(alternate_field_name)
					// will panic if trying to set a non-exported value or a zero value, so ignore those
					if isExported(alternate_field_name) && !isZero(alternate_value) {
						field_value.Set(alternate_value)
						break
					}
				}
			}
			// recurse, in case nested values of this field need to be set as well
			if isExported(field.Name) && !isZero(field_value) {
				CoalesceFields(field_value.Addr().Interface())
			}
		}
	} else if object_type.Kind() == reflect.Array || object_type.Kind() == reflect.Slice {
		arr_len := object_value.Len()
		for i := 0; i < arr_len; i++ {
			CoalesceFields(object_value.Index(i).Addr().Interface())
		}
	}
}

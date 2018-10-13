package resource

import (
	"reflect"

	"github.com/pkg/errors"
)

// DeletableResources converts given raw resources for a given resource type
// into a format that can be deleted by the Terraform API.
func (aws AWS) DeletableResources(resType TerraformResourceType, resources interface{}) (DeletableResources, error) {
	deletableResources := DeletableResources{}
	reflectResources := reflect.ValueOf(resources)

	for i := 0; i < reflectResources.Len(); i++ {
		deleteID, err := getDeleteID(resType)
		if err != nil {
			return nil, err
		}

		field, err := findField(deleteID, reflect.Indirect(reflectResources.Index(i)))
		if err != nil {
			return nil, err
		}

		deletableResources = append(deletableResources, &DeletableResource{
			Type: resType,
			ID:   field.Elem().String(),
			Tags: findTags(reflectResources.Index(i)),
		})
	}

	return deletableResources, nil
}

func findField(name string, v reflect.Value) (reflect.Value, error) {
	field := v.FieldByName(name)

	if !field.IsValid() {
		return reflect.Value{}, errors.Errorf("Field %s does not exist", name)
	}
	return field, nil
}

//
//func findSlice(name string, v reflect.Value) (reflect.Value, error) {
//	if v.Type().Kind() != reflect.Struct {
//		return reflect.Value{}, errors.Errorf("Input not a struct: %s", v)
//	}
//	slice := v.FieldByName(name)
//
//	if !slice.IsValid() {
//		return reflect.Value{}, errors.Errorf("Slice %s does not exist", name)
//	}
//	return slice, nil
//}

// findTags finds findTags via reflection in the describe output.
func findTags(res reflect.Value) map[string]string {
	tags := map[string]string{}

	ts := reflect.Indirect(res).FieldByName("Tags")
	if !ts.IsValid() {
		ts = reflect.Indirect(res).FieldByName("TagSet")
	}

	if ts.IsValid() {
		for i := 0; i < ts.Len(); i++ {
			key := reflect.Indirect(ts.Index(i)).FieldByName("Key").Elem()
			value := reflect.Indirect(ts.Index(i)).FieldByName("Value").Elem()
			tags[key.String()] = value.String()
		}
	}
	return tags
}

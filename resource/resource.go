package resource

import (
	"reflect"

	"log"

	"github.com/pkg/errors"
)

func List(a ApiDesc) (Resources, interface{}) {
	descOut := invoke(a.DescribeFn, a.DescribeFnInput)

	foundRes, err := findSlice(a.DescribeOutputName[0], descOut.Elem())
	if err != nil {
		log.Fatal(err)
	}

	if len(a.DescribeOutputName) == 2 {
		// find resources in the case the output structs are nested
		// (e.g. "Reservations" -> "Instances")
		foundRes, err = findSlice(a.DescribeOutputName[1], foundRes.Index(0).Elem())
		if err != nil {
			log.Fatal(err)
		}
	}

	res := Resources{}

	for i := 0; i < foundRes.Len(); i++ {
		field, err := findField(a, reflect.Indirect(foundRes.Index(i)))
		if err != nil {
			continue
		}

		res = append(res, &Resource{
			Type: a.TerraformType,
			Id:   field.Elem().String(),
			Tags: Tags(foundRes.Index(i)),
		})

	}

	return res, descOut.Interface()
}

// invoke is used to generically call any describe function fn
// of the aws-go-sdk API, where arg is a describe input
// (e.g. DescribeAutoScalingGroupsInput). Invoke returns
// a generic describe output and awserr.Error.
func invoke(fn interface{}, arg interface{}) reflect.Value {
	inputs := []reflect.Value{
		reflect.ValueOf(arg),
	}

	outputs := reflect.ValueOf(fn).Call(inputs)
	return outputs[0]
}

func findField(a ApiDesc, v reflect.Value) (reflect.Value, error) {
	field := v.FieldByName(a.DeleteId)

	if !field.IsValid() {
		return reflect.Value{}, errors.Errorf("Field %s does not exist", a.DeleteId)
	}
	return field, nil
}

func findSlice(name string, v reflect.Value) (reflect.Value, error) {
	if v.Type().Kind() != reflect.Struct {
		return reflect.Value{}, errors.Errorf("Input not a struct: %s", v)
	}
	slice := v.FieldByName(name)

	if !slice.IsValid() {
		return reflect.Value{}, errors.Errorf("Slice %s does not exist", name)
	}
	return slice, nil
}

func Tags(res reflect.Value) map[string]string {
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
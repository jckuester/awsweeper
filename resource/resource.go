package resource

import (
	"reflect"

	"log"

	"github.com/pkg/errors"
)

// List lists all AWS resources based on a given
// API description
func List(a APIDesc) (Resources, interface{}) {
	descOut := invoke(a.Describe, a.DescribeInput)
	descOutRes, err := findSlice(a.DescribeOutputName[0], descOut.Elem())
	if err != nil {
		log.Fatal(err)
	}

	res := Resources{}

	if len(a.DescribeOutputName) == 2 {
		// find resources in the case the output is a nested struct
		// (e.g. "Reservations" -> "Instances")
		for i := 0; i < descOutRes.Len(); i++ {
			nestedDescOut, err := findSlice(a.DescribeOutputName[1], descOutRes.Index(i).Elem())
			if err != nil {
				log.Fatal(err)
			}

			for i := 0; i < nestedDescOut.Len(); i++ {
				field, err := findField(a, reflect.Indirect(nestedDescOut.Index(i)))
				if err != nil {
					log.Fatal(err)
				}

				res = append(res, &Resource{
					Type: a.TerraformType,
					ID:   field.Elem().String(),
					Tags: tags(nestedDescOut.Index(i)),
				})
			}
		}
		return res, descOut.Interface()
	}

	for i := 0; i < descOutRes.Len(); i++ {
		field, err := findField(a, reflect.Indirect(descOutRes.Index(i)))
		if err != nil {
			log.Fatal(err)
		}

		res = append(res, &Resource{
			Type: a.TerraformType,
			ID:   field.Elem().String(),
			Tags: tags(descOutRes.Index(i)),
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

func findField(a APIDesc, v reflect.Value) (reflect.Value, error) {
	field := v.FieldByName(a.DeleteID)

	if !field.IsValid() {
		return reflect.Value{}, errors.Errorf("Field %s does not exist", a.DeleteID)
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

func tags(res reflect.Value) map[string]string {
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

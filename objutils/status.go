package objutils

import (
	"fmt"
	"reflect"

	"sigs.k8s.io/controller-runtime/pkg/client"
)

func GetStatusField(obj client.Object) (interface{}, error) {
	val := reflect.ValueOf(obj)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}
	statusField := val.FieldByName("Status")
	if !statusField.IsValid() {
		return nil, fmt.Errorf("no Status field found")
	}
	return statusField.Interface(), nil
}

func SetStatusField(obj client.Object, status interface{}) error {
	val := reflect.ValueOf(obj)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}
	statusField := val.FieldByName("Status")
	if !statusField.IsValid() || !statusField.CanSet() {
		return fmt.Errorf("cannot set Status field")
	}
	statusField.Set(reflect.ValueOf(status))
	return nil
}

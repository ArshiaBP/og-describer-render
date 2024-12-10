package describer

import (
	"encoding/json"
	"reflect"
)

type JSONAllFieldsMarshaller struct {
	Value interface{}
}

func (x JSONAllFieldsMarshaller) MarshalJSON() (res []byte, err error) {
	var val = x.Value
	v := reflect.ValueOf(x.Value)
	if !v.IsValid() {
		return json.Marshal(val)
	}
	return json.Marshal(val)
}

func (x *JSONAllFieldsMarshaller) UnmarshalJSON(data []byte) (err error) {
	v := reflect.ValueOf(x.Value)
	if !v.IsValid() {
		return nil
	}
	val := reflect.New(v.Type())
	err = json.Unmarshal(data, val.Interface())
	if err != nil {
		return err
	}
	newVal := reflect.New(v.Type())
	if !val.Elem().Type().AssignableTo(newVal.Elem().Type()) {
		return nil
	}
	newVal.Elem().Set(val.Elem())
	x.Value = newVal.Elem().Interface()
	return nil
}

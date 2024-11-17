package applescript

import (
	"bytes"
	"fmt"
	"reflect"
)

type marshaler struct {
	buf bytes.Buffer
}

func Marshal(v any) ([]byte, error) {
	val := reflect.ValueOf(v)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	if val.Kind() != reflect.Struct {
		return nil, fmt.Errorf("unsupported type: %T", v)
	}

	m := &marshaler{}
	if err := m.marshalStruct(val); err != nil {
		return nil, err
	}

	return m.buf.Bytes(), nil
}

func (m *marshaler) marshalStruct(val reflect.Value) error {
	typ := val.Type()
	m.buf.WriteString("{")

	first := true
	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)
		fieldVal := val.Field(i)

		tag := field.Tag.Get("applescript")
		if tag == "-" {
			continue
		}
		if tag == "" {
			tag = field.Name
		}

		if isEmpty(fieldVal) {
			continue
		}

		if !first {
			m.buf.WriteString(", ")
		}
		first = false

		if err := m.marshalField(tag, fieldVal); err != nil {
			return err
		}
	}

	m.buf.WriteString("}")
	return nil
}

func (m *marshaler) marshalSlice(val reflect.Value) error {
	m.buf.WriteString("{")

	first := true
	for i := 0; i < val.Len(); i++ {
		if !first {
			m.buf.WriteString(", ")
		}
		first = false

		if err := m.marshalField("", val.Index(i)); err != nil {
			return err
		}
	}

	m.buf.WriteString("}")
	return nil
}

func (m *marshaler) marshalField(fieldName string, fieldVal reflect.Value) error {
	if fieldName != "" {
		m.buf.WriteString(fieldName)
		m.buf.WriteString(": ")
	}

	switch fieldVal.Type() {
	case reflect.TypeOf(""):
		m.buf.WriteString(fmt.Sprintf("%q", fieldVal.String()))
	case reflect.TypeOf(int(0)), reflect.TypeOf(int8(0)), reflect.TypeOf(int16(0)), reflect.TypeOf(int32(0)), reflect.TypeOf(int64(0)):
		m.buf.WriteString(fmt.Sprintf("%d", fieldVal.Int()))
	case reflect.TypeOf(float32(0)), reflect.TypeOf(float64(0)):
		m.buf.WriteString(fmt.Sprintf("%f", fieldVal.Float()))
	case reflect.TypeOf(bool(false)):
		m.buf.WriteString(fmt.Sprintf("%t", fieldVal.Bool()))
	default:
		switch fieldVal.Type().Kind() {
		case reflect.String:
			m.buf.WriteString(fieldVal.String())
		case reflect.Struct:
			return m.marshalStruct(fieldVal)
		case reflect.Slice, reflect.Array:
			return m.marshalSlice(fieldVal)
		default:
			return fmt.Errorf("unsupported type: %T", fieldVal.Interface())
		}
	}

	return nil
}

package main

import (
	"fmt"
	"reflect"
)

func i2s(data interface{}, out interface{}) error {

	err := iter(data, out)

	if err != nil {
		return err
	}
	return nil
}

func iter(data, out interface{}) error {
	outVal := reflect.ValueOf(out)

	if outVal.Kind() != reflect.Ptr {
		fmt.Println(outVal.Kind())
		return fmt.Errorf("Expected pointer")
	}

	inType := reflect.TypeOf(data)
	inVal := reflect.ValueOf(data)

	o := outVal.Elem()

	fmt.Println(o.Kind())
	switch o.Kind() {
	case reflect.Struct:
		if inType.Kind() != reflect.Map {
			fmt.Println(inType.Kind())
			return fmt.Errorf("Expected map")
		}

		for _, k := range inVal.MapKeys() {
			fmt.Println(k.String())
			field := o.FieldByName(k.String())
			fmt.Println(field.Kind())
			switch field.Kind() {
			case reflect.Int:
				if reflect.TypeOf(inVal.MapIndex(k).Interface()).Kind() != reflect.Float64 {
					return fmt.Errorf("Expected float")
				}
				val := reflect.ValueOf(inVal.MapIndex(k).Interface()).Float()
				field.SetInt(int64(val))
			case reflect.String:
				if reflect.TypeOf(inVal.MapIndex(k).Interface()).Kind() != reflect.String {
					return fmt.Errorf("Expected string")
				}

				val := reflect.ValueOf(inVal.MapIndex(k).Interface()).String()
				field.SetString(val)
			case reflect.Bool:
				if reflect.TypeOf(inVal.MapIndex(k).Interface()).Kind() != reflect.Bool {
					return fmt.Errorf("Expected bool")
				}
				val := reflect.ValueOf(inVal.MapIndex(k).Interface()).Bool()
				field.SetBool(val)
			default:
				fmt.Println(field.CanAddr())
				err := iter(inVal.MapIndex(k).Interface(), field.Addr().Interface())
				if err != nil {
					return err
				}
				fmt.Println("AAAA")
			}
		}
	case reflect.Slice:
		if inType.Kind() != reflect.Slice {
			fmt.Println(inType.Kind())
			return fmt.Errorf("Expected slice")
		}
		for i := 0; i < inVal.Len(); i++ {
			fmt.Println(o.Kind())
			fmt.Println(o.Type().Elem())
			elType := o.Type().Elem()
			fmt.Println(reflect.TypeOf(elType))
			newEl := reflect.New(elType)

			fmt.Println(inVal.Index(i).CanAddr())

			err := iter(inVal.Index(i).Interface(), newEl.Interface())

			if err != nil {
				return err
			}
			fmt.Println(reflect.ValueOf(55))
			fmt.Println(newEl)
			fmt.Println(reflect.ValueOf(newEl.Elem()))
			o.Set(reflect.Append(o, newEl.Elem()))
			fmt.Println(o.Len())
		}
	}

	return nil
}



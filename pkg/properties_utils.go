package pkg

import (
	"errors"
	"fmt"
	"reflect"
)

// CopyProperties copy properties to dest
func CopyProperties(dst, src interface{}) (err error) {
	// Prevention of accidents panic
	defer func() {
		if e := recover(); e != nil {
			err = fmt.Errorf("%v", e)
		}
	}()

	dstType, dstValue := reflect.TypeOf(dst), reflect.ValueOf(dst)
	srcType, srcValue := reflect.TypeOf(src), reflect.ValueOf(src)

	// dst must be struct pointer
	if dstType.Kind() != reflect.Ptr || dstType.Elem().Kind() != reflect.Struct {
		return errors.New("dst type should be a struct pointer")
	}

	// src must be struct or struct pointer
	if srcType.Kind() == reflect.Ptr {
		srcType, srcValue = srcType.Elem(), srcValue.Elem()
	}
	if srcType.Kind() != reflect.Struct {
		return errors.New("src type should be a struct or a struct pointer")
	}

	dstType, dstValue = dstType.Elem(), dstValue.Elem()

	propertyNums := dstType.NumField()

	for i := 0; i < propertyNums; i++ {
		property := dstType.Field(i)
		propertyValue := srcValue.FieldByName(property.Name)

		if !propertyValue.IsValid() || property.Type != propertyValue.Type() {
			continue
		}

		if dstValue.Field(i).CanSet() {
			dstValue.Field(i).Set(propertyValue)
		}
	}

	return nil
}

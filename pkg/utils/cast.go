package utils

import (
	"fmt"
	"reflect"
)

var ErrNilParam = fmt.Errorf("cast error: got nil param")

func SafeCast[T any](param any) (T, error) {
	var getT T

	if param == nil {
		return getT, ErrNilParam
	}

	v, ok := param.(T)
	if !ok {
		return v, fmt.Errorf("cast error: got type: %s, want type: %s", reflect.TypeOf(param).String(), reflect.TypeOf(getT).String())
	}

	return v, nil
}

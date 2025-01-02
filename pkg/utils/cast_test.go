package utils

import (
	"reflect"
	"testing"
)

func TestInvoicesSafeCast(t *testing.T) {
	cast, err := SafeCast[int](12334)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(cast, reflect.TypeOf(cast).String())

	_, err = SafeCast[string](nil)
	t.Log(err)
}

package logger

import (
	"testing"
)

func TestAnyToStr(t *testing.T) {

	tests := []struct {
		T    any
		TStr string
	}{
		{10, "10"},
		{-10, "-10"},
		{true, "true"},
		{false, "false"},
		{"test", "test"},
		{"", ""},
		{nil, "<nil>"},
		{struct{}{}, "{}"},

		{struct {
			Z string
			F int
		}{"test", 10}, "{test 10}"},

		// {make(chan int), "<nil>"},
		{[]int{1, 2, 3}, "[1 2 3]"},
	}

	for _, x := range tests {
		res := AnyToStr(x.T)
		if x.TStr != res {
			t.Log(x.T)
			t.Fatalf("failed: %s != %s", x.TStr, res)
		}

	}

}

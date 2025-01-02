package nats

import (
	"testing"

	"github.com/brianvoe/gofakeit/v7"
)

func TestHelpersIsError(t *testing.T) {
	tests := []struct {
		data    []byte
		isValid bool
	}{
		{
			data:    []byte(""), // null string
			isValid: false,
		},
		{
			data:    []byte("error"), // != 'error:'
			isValid: false,
		},
		{
			data:    []byte("error:\t\t"),
			isValid: true,
		}, {
			data:    []byte("error:"),
			isValid: true,
		}, {
			data:    []byte("error: "),
			isValid: true,
		},
		{
			data:    []byte("error: " + gofakeit.LetterN(100)),
			isValid: true,
		},
		{
			data:    []byte("error: " + gofakeit.LetterN(100<<1)),
			isValid: true,
		},
		{
			data:    []byte("error: " + gofakeit.LetterN(1000<<1)),
			isValid: true,
		},
	}

	for _, i := range tests {
		is, errmsg := HelpersIsError(i.data)
		if i.isValid != is {
			t.Fatalf("i.isValid != is: %s", string(i.data))
		}

		t.Log("ERROR_MSG:", errmsg)
	}

	for range 10000 {
		is, _ := HelpersIsError([]byte("error: " + gofakeit.LetterN(1000<<1)))
		if !is {
			t.Fatal("!is")
		}
	}
}

func TestHelpersInvoiceGetTxHash(t *testing.T) {
	tests := []struct {
		data    []byte
		isValid bool
		txHash  string
	}{
		{
			[]byte("ok: 0x1234567890123456789012345678901234567890123456789012345678901234"),
			true,
			"0x1234567890123456789012345678901234567890123456789012345678901234",
		},
		{
			[]byte(""),
			false,
			"",
		}, {
			[]byte("ok: "),
			false,
			"",
		}, {
			[]byte("ok:   "),
			false,
			"",
		}, {
			[]byte(gofakeit.GlobalFaker.BuzzWord()),
			false,
			"",
		}, {
			[]byte("error: hello 8r892ru892ru2r82389ru289ru2"),
			false,
			"",
		},
	}

	for _, i := range tests {
		txHash, err := HelpersInvoiceGetTxHash(i.data)
		if i.isValid && err != nil {
			t.Fatalf("i.isValid && err != nil")
		}
		if err != nil && !i.isValid {
			continue
		}

		if i.txHash != txHash {
			t.Fatalf("i.txHash != txHash: %s", string(txHash))
		}
	}

}

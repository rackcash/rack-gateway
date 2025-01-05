package service

import (
	"fmt"
	"net/http"
	"testing"
)

func TestFindOrNew(t *testing.T) {

	s := NewQrCodesService()

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		qr, err := s.FindOrNew("test")
		if err != nil {
			t.Error(err)
		}

		w.Header().Add("Content-Type", "image/png")
		w.Write([]byte(qr))
	})

	if err := http.ListenAndServe(":9999", nil); err != nil {
		fmt.Println("Error starting server:", err)
	}

	// fmt.Println()

}

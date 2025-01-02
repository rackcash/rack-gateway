package failover

import (
	"fmt"
	"net/http"
	"net/url"
	"testing"
	"time"
)

func TestFailover(t *testing.T) {
	f := New(CheckConnection,
		OptCheckUrlBeforeAdding(true),
		OptCheckUrlDelay(10*time.Second),
		OptMaxAttempts(10),
	)

	tests := []*url.URL{
		// {Host: "google.com", Scheme: "https"},
		{Host: "0.0.0.0:8080", Scheme: "http"},
		{Host: "0.0.0.0:8081", Scheme: "http"},
		{Host: "0.0.0.0:8081", Scheme: "http"},
		{Host: "0.0.0.0:8081", Scheme: "http"},
		{Host: "0.0.0.0:8082", Scheme: "http"},
		{Host: "0.0.0.0:8082", Scheme: "http"},
		{Host: "0.0.0.0:8083", Scheme: "http"},
		{Host: "0.0.0.0:8084", Scheme: "http"},
		{Host: "0.0.0.0:8085", Scheme: "http"},
		{Host: "0.0.0.0:8086", Scheme: "http"},
		{Host: "0.0.0.0:8087", Scheme: "http"},
	}

	err := f.AddUrl(tests[0])
	if err != nil {
		t.Fatal(err)
	}

	err = f.AddUrl(tests[0]) // duplicate
	if err != nil {
		t.Fatal(err)
	}

	err = f.AddUrl(tests[1])
	if err != nil {
		t.Fatal(err)
	}

	err = f.AddUrls(tests...)
	if err != nil {
		t.Fatal(err)
	}

	time.Sleep(5 * time.Second)
	// fmt.Println("STATUS", f.Request(Request, OptReqOnErr(ReqOnErrReconnectNext), OptMaxAttempts(4)))
	fmt.Println("STATUS", f.Request(Request, OptReqOnErr(ReqOnErrReconnectNext), OptMaxAttempts(2)))
	// f.Request(Request)
	// f.Request(Request)
	// f.Request(Request)

}

func CheckConnection(url *url.URL) error {
	_, err := http.Get(url.String())
	if err != nil {
		fmt.Println(err)
		return err
	}
	return nil
}

func Request(url *url.URL) error {
	_, err := http.Get("http://" + url.Host)
	if err != nil {
		return err
	}

	return nil
}

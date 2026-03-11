package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestSimple(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		qs := r.URL.Query()
		fmt.Fprintf(w, "hello, %s", qs.Get("name"))
	}))
	defer ts.Close()

	r, err := getReflectedParams(ts.URL + "?name=Mr%20Naughty")
	t.Logf("params reflected: %#v", r)

	if err != nil {
		t.Fatalf("expected nil error from getReflectedParams(), have %s", err)
	}

	if len(r) != 1 {
		t.Errorf("wanted length 1 for returned keys, have %d", len(r))
	}
}

func TestAppend(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		qs := r.URL.Query()
		// Reflect the value of the parameter so checkAppend can see it
		fmt.Fprintf(w, "hello, %s", qs.Get("name"))
	}))
	defer ts.Close()

	// Testing if "somerandomvalue" is reflected when appended to the 'name' param
	r, err := checkAppend(ts.URL+"?name=Mr%20Naughty", "name", "somerandomvalue")

	if err != nil {
		t.Fatalf("expected nil error from checkAppend(), have %s", err)
	}

	if !r {
		t.Errorf("wanted checkAppend() to return true (reflection detected), but it didn't")
	}
}

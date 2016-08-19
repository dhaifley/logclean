package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

func TestELKClientGetIndexes(t *testing.T) {
	fs := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.Header().Set("Content-Type", "text/plain")
		fmt.Fprintln(w, "yellow testindex-1983.02.02    some stuff\n")
	}))
	defer fs.Close()

	tr := &http.Transport{
		Proxy: func(req *http.Request) (*url.URL, error) {
			return url.Parse(fs.URL)
		},
	}

	hc := &http.Client{Transport: tr}
	cli := &ELKClient{"testindex", 20, hc}

	result := ""
	ci := cli.GetIndexes()
	for ri := range ci {
		if ri.Err != nil {
			t.Error(ri.Err)
		}

		result = ri.Msg
	}

	expected := "testindex-1983.02.02"
	if string(result) != expected {
		t.Errorf("Index expected: %q, got: %q", expected, result)
	}
}

func TestELKClientDeleteIndex(t *testing.T) {
	fs := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.Header().Set("Content-Type", "text/plain")
		fmt.Fprintln(w, "test\n")
	}))
	defer fs.Close()

	tr := &http.Transport{
		Proxy: func(req *http.Request) (*url.URL, error) {
			return url.Parse(fs.URL)
		},
	}

	hc := &http.Client{Transport: tr}
	cli := &ELKClient{"testindex", 20, hc}

	result := ""
	ri := cli.DeleteIndex("testindex-1983.02.02")
	if ri.Err != nil {
		t.Error(ri.Err)
	}

	result = ri.Msg
	expected := "testindex-1983.02.02"
	if string(result) != expected {
		t.Errorf("Index expected: %q, got: %q", expected, result)
	}
}

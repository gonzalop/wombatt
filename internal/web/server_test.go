package web

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"
)

type Example struct {
	Hola string `name:"Hola"`
}

func TestServer(t *testing.T) {
	srv := NewServer(":5000", "/test")
	defer func() {
		_ = srv.Shutdown(context.Background())
	}()
	// Create a channel to signal server start

	// Publish before Start is a no-op
	srv.Publish("hola", &Example{Hola: "this is a test"})
	if err := srv.Start(); err != nil {
		t.Fatalf("%v", err)
	}
	client := &http.Client{}
	f := func(url string, status int) string {
		resp, err := client.Get(url)
		if err != nil {
			t.Errorf("response error: %v", err)
			return ""
		}
		defer resp.Body.Close()
		data, err := io.ReadAll(resp.Body)
		if err != nil {
			t.Errorf("error in HTTP client: %v", err)
		} else if resp.StatusCode != status {
			t.Errorf("%v %d", string(data), len(data))
		}
		return string(data)
	}

	_ = f("http://127.0.0.1:5000/test/hola", http.StatusNotFound)
	srv.Publish("hola", &Example{Hola: "this is a test"})
	s := f("http://127.0.0.1:5000/test/hola", http.StatusOK)
	if !strings.HasPrefix(s, "Hola: this is a test\n") {
		t.Errorf("got '%v'; want '%v'", s, "Hola: this is a test")
	}
	if !strings.Contains(s, "last_updated: ") {
		t.Errorf("last_updated not found. got: '%v'", s)
	}
	s = f("http://127.0.0.1:5000/test/hola?format=json&fields=Hola", http.StatusOK)
	obj := &Example{}
	err := json.Unmarshal([]byte(s), &obj)
	if err != nil {
		t.Errorf("JSON error: %v", err)
	}
	if obj.Hola != "this is a test" {
		t.Errorf("got: '%v'", obj.Hola)
	}
	srv.Publish("hola", nil)
}

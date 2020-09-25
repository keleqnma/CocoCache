package cococache

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"testing"
	"time"
)

var testDB = map[string]string{
	"Tom":  "630",
	"Jack": "589",
	"Sam":  "567",
}

func serveHTTP() {
	NewCache(2<<10, GetterFunc(
		func(key string) ([]byte, error) {
			log.Println("[SlowDB] search key", key)
			if v, ok := testDB[key]; ok {
				return []byte(v), nil
			}
			return nil, fmt.Errorf("%s not exist", key)
		}))

	addr := "localhost:9999"
	peers := NewHTTPPool(addr)
	log.Println("cococache is running at", addr)

	if err := http.ListenAndServe(addr, peers); err != nil {
		log.Fatal(err)
	}
}

func TestHTTPGet(t *testing.T) {
	go serveHTTP()
	time.Sleep(time.Second)

	resp, err := http.Get("http://localhost:9999/_cococache/Tom")
	if err != nil {
		t.Fatal(err)
	}
	if resp != nil {
		buf, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			t.Fatal(err)
		}
		if string(buf) != "630" {
			err = fmt.Errorf("expected:%v, received:%v", "630", string(buf))
			t.Fatal(err)
		}
		resp.Body.Close()
	}

	searchKey := "wejbdnwqdx"
	resp, err = http.Get("http://localhost:9999/_cococache/" + searchKey)
	if err != nil {
		t.Fatal(err)
	}
	if resp != nil {
		buf, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(string(buf), "not exist") {
			err = fmt.Errorf("expected:%v, received:%v", fmt.Errorf("%s not exist", searchKey), string(buf))
			t.Fatal(err)
		}
		resp.Body.Close()
	}
}

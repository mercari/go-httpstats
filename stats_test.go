package httpstats

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestSimple(t *testing.T) {
	mw, err := New()
	if err != nil {
		t.Fatal(err)
	}
	testsrv := httptest.NewServer(
		mw.WrapHandleFunc(
			http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Write([]byte("hello"))
			})))
	resp, err := http.Get(testsrv.URL)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	d, err := mw.Data()
	if err != nil {
		t.Fatal(err)
	}
	if d.Request.Count != 1 {
		t.Errorf("Count: expected 1 but actual %d", d.Request.Count)
	}
	if d.Request.StatusCount[200] != 1 {
		t.Errorf("StatusCount[200]: expected 1 but actual %d", d.Request.StatusCount[200])
	}

}

func Test500(t *testing.T) {
	mw, err := New()
	if err != nil {
		t.Fatal(err)
	}
	testsrv500 := httptest.NewServer(
		mw.WrapHandleFunc(
			http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				http.Error(w, "", http.StatusInternalServerError)
			})))
	resp, err := http.Get(testsrv500.URL)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	d, err := mw.Data()
	if err != nil {
		t.Fatal(err)
	}
	if d.Request.Count != 1 {
		t.Errorf("Count: expected 1 but actual %d", d.Request.Count)
	}
	if d.Request.StatusCount[500] != 1 {
		t.Errorf("StatusCount[500]: expected 1 but actual %d", d.Request.StatusCount[500])
	}
}

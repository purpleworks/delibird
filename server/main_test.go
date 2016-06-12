package main

import (
	"net/http"
	"net/http/httptest"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestMain(t *testing.T) {
	ts := httptest.NewServer(App())
	defer ts.Close()

	Convey("HTTP Request E2E test", t, func() {
		Convey("#tracking", func() {
			Convey("Response has 404(not found) status code when invalid path", func() {
				res, err := http.Get(ts.URL + "/track")
				if err != nil {
					t.Fatal(err)
				}
				So(res.StatusCode, ShouldEqual, http.StatusNotFound)
			})
		})
	})
}

package couriers

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/purpleworks/delibird"
	. "github.com/smartystreets/goconvey/convey"
)

type SampleCourier struct{}

func (t SampleCourier) Code() string {
	return "SampleCourier"
}
func (t SampleCourier) Name() string {
	return "SampleCourier"
}
func (t SampleCourier) Parse(invoice string) (delibird.Track, *delibird.ApiError) {
	return delibird.Track{}, nil
}

// test html mock file
func readTestResponseFile(filename string) string {
	prefix := "."

	if _, currentPath, _, ok := runtime.Caller(0); ok {
		prefix = filepath.Dir(currentPath)
	}

	b, _ := ioutil.ReadFile(fmt.Sprintf("%s/testhtml/%s", prefix, filename))

	return string(b)
}

func TestCourier(t *testing.T) {
	Convey("Courier create test", t, func() {
		Convey("Return error when invalid code", func() {
			_, err := New("TEST")
			So(err, ShouldNotBeNil)
		})

		Convey("Return Courier object when valid code", func() {
			RegisterCourier("SampleCourier", &SampleCourier{})

			courier, err := New("SampleCourier")
			So(courier.Code(), ShouldEqual, "SampleCourier")
			So(courier.Name(), ShouldEqual, "SampleCourier")
			So(err, ShouldBeNil)
		})
	})
}

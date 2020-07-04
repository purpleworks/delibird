package couriers

import (
	"fmt"
	"strings"
	"testing"

	"github.com/jarcoal/httpmock"
	. "github.com/smartystreets/goconvey/convey"
)

// func TestFunc(t *testing.T) {
// 	convey.Convey("starting value", t, func() {
// 		x := 0
// 		convey.Convey("incremented", func() {
// 			x++

// 			convey.Convey("shouldEqual", func() {
// 				convey.So(x, convey.ShouldEqual, 1)
// 			})
// 		})
// 	})
// }

func TestParse(t *testing.T) {
	slxstr := "slx"
	slxstr = strings.ToUpper(slxstr)
	courier, _ := New(slxstr)

	// create http mock
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	invalidTrackingNumber := "slx_invalid"
	completeTrackingNumber := "slx_complete"

	httpmock.RegisterResponder("GET", fmt.Sprintf(Slx{}.TrackingUrl(), invalidTrackingNumber),
		httpmock.NewStringResponder(200, readTestResponseFile(invalidTrackingNumber+".html")))
	httpmock.RegisterResponder("GET", fmt.Sprintf(Slx{}.TrackingUrl(), completeTrackingNumber),
		httpmock.NewStringResponder(200, readTestResponseFile(completeTrackingNumber+".html")))

	// data, err := courier.Parse(completeTrackingNumber)

	// if err != nil {
	// 	fmt.Println(err)
	// }
	// fmt.Println(data.Sender)

	Convey("starting value", t, func() {
		Convey("invalid", func() {
			_, err := courier.Parse(invalidTrackingNumber)

			fmt.Println(err)
		})
		Convey("Complete courier test", func() {
			data, _ := courier.Parse(completeTrackingNumber)

			So(data.Sender, ShouldEqual, "알○딘")

		})
	})
}

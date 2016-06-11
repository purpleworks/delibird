package delibird

import (
	"fmt"
	"github.com/jarcoal/httpmock"
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

func TestCJCourier(t *testing.T) {
	// create cj courier
	courier, _ := NewCourier("CJ")

	// create http mock
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	invalidTrackingNumber := "cj_invalid"
	startTrackingNumber := "cj_start"
	completeTrackingNumber := "cj_complete"

	httpmock.RegisterResponder("GET", fmt.Sprintf(Cj{}.TrackingUrl(), invalidTrackingNumber),
		httpmock.NewStringResponder(200, readTestResponseFile(invalidTrackingNumber+".html")))
	httpmock.RegisterResponder("GET", fmt.Sprintf(Cj{}.TrackingUrl(), startTrackingNumber),
		httpmock.NewStringResponder(200, readTestResponseFile(startTrackingNumber+".html")))
	httpmock.RegisterResponder("GET", fmt.Sprintf(Cj{}.TrackingUrl(), completeTrackingNumber),
		httpmock.NewStringResponder(200, readTestResponseFile(completeTrackingNumber+".html")))

	Convey("CJ test", t, func() {
		Convey("Invalid tracking number test", func() {
			_, err := courier.Parse(invalidTrackingNumber)

			So(err, ShouldNotBeNil)
		})

		Convey("Start courier test", func() {
			data, _ := courier.Parse(startTrackingNumber)

			So(data.StatusCode, ShouldEqual, DeleveryStart)
			So(data.Sender, ShouldEqual, "홈*럼")
			So(data.Receiver, ShouldEqual, "테*트")
			So(data.Signer, ShouldEqual, "")
			So(data.CompanyCode, ShouldEqual, "CJ")
		})

		Convey("Complete courier test", func() {
			data, _ := courier.Parse(completeTrackingNumber)

			So(data.StatusCode, ShouldEqual, DeleveryComplete)
			So(data.Sender, ShouldEqual, "홈*럼")
			So(data.Receiver, ShouldEqual, "테*트")
			So(data.Signer, ShouldEqual, "본인")
			So(data.CompanyCode, ShouldEqual, "CJ")
		})
	})
}

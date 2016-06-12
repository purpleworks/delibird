package delibird

import (
	"fmt"
	"testing"

	"github.com/jarcoal/httpmock"
	. "github.com/smartystreets/goconvey/convey"
)

func TestKpostCourier(t *testing.T) {
	// create epost courier
	courier, _ := NewCourier("KPOST")

	// create http mock
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	invalidTrackingNumber := "kpost_invalid"
	invalidTrackingNumber2 := "kpost_invalid_2"
	startTrackingNumber := "kpost_start"
	completeTrackingNumber := "kpost_complete"

	httpmock.RegisterResponder("GET", fmt.Sprintf(Kpost{}.TrackingUrl(), invalidTrackingNumber),
		httpmock.NewStringResponder(200, readTestResponseFile(invalidTrackingNumber+".html")))
	httpmock.RegisterResponder("GET", fmt.Sprintf(Kpost{}.TrackingUrl(), invalidTrackingNumber2),
		httpmock.NewStringResponder(200, readTestResponseFile(invalidTrackingNumber2+".html")))
	httpmock.RegisterResponder("GET", fmt.Sprintf(Kpost{}.TrackingUrl(), startTrackingNumber),
		httpmock.NewStringResponder(200, readTestResponseFile(startTrackingNumber+".html")))
	httpmock.RegisterResponder("GET", fmt.Sprintf(Kpost{}.TrackingUrl(), completeTrackingNumber),
		httpmock.NewStringResponder(200, readTestResponseFile(completeTrackingNumber+".html")))

	Convey("KPOST test", t, func() {
		Convey("Invalid tracking number test", func() {
			_, err := courier.Parse(invalidTrackingNumber)

			So(err, ShouldNotBeNil)
		})

		Convey("Invalid tracking number test2", func() {
			_, err := courier.Parse(invalidTrackingNumber2)

			So(err, ShouldNotBeNil)
		})

		Convey("Start courier test", func() {
			data, _ := courier.Parse(startTrackingNumber)

			So(data.StatusCode, ShouldEqual, Unloading)
			So(data.Sender, ShouldEqual, "홈*럼")
			So(data.Receiver, ShouldEqual, "테*트")
			So(data.CompanyCode, ShouldEqual, "KPOST")
		})

		Convey("Complete courier test", func() {
			data, _ := courier.Parse(completeTrackingNumber)

			So(data.StatusCode, ShouldEqual, DeleveryComplete)
			So(data.Sender, ShouldEqual, "홈*럼")
			So(data.Receiver, ShouldEqual, "테*트")
			So(data.Signer, ShouldEqual, "테*트님 - 본인")
			So(data.CompanyCode, ShouldEqual, "KPOST")
		})
	})
}

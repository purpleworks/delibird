package delibird

import (
	"fmt"
	"testing"

	"github.com/jarcoal/httpmock"
	. "github.com/smartystreets/goconvey/convey"
)

func TestKGCourier(t *testing.T) {
	// create kg courier
	courier, _ := NewCourier("KG")

	// create http mock
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	invalidTrackingNumber := "kg_invalid"
	invalidTrackingNumber2 := "kg_invalid2"
	completeTrackingNumber := "kg_complete"

	httpmock.RegisterResponder("GET", fmt.Sprintf(Kg{}.TrackingUrl(), invalidTrackingNumber),
		httpmock.NewStringResponder(200, readTestResponseFile(invalidTrackingNumber+".html")))
	httpmock.RegisterResponder("GET", fmt.Sprintf(Kg{}.TrackingUrl(), invalidTrackingNumber2),
		httpmock.NewStringResponder(200, readTestResponseFile(invalidTrackingNumber2+".html")))
	httpmock.RegisterResponder("GET", fmt.Sprintf(Kg{}.TrackingUrl(), completeTrackingNumber),
		httpmock.NewStringResponder(200, readTestResponseFile(completeTrackingNumber+".html")))

	Convey("KG test", t, func() {
		Convey("Invalid tracking number test", func() {
			_, err := courier.Parse(invalidTrackingNumber)

			So(err, ShouldNotBeNil)
		})
		Convey("Invalid tracking number test2", func() {
			_, err := courier.Parse(invalidTrackingNumber2)

			So(err, ShouldNotBeNil)
		})

		Convey("Complete courier test", func() {
			data, _ := courier.Parse(completeTrackingNumber)

			So(data.StatusCode, ShouldEqual, DeleveryComplete)
			So(data.Sender, ShouldEqual, "웨일런샵 님")
			So(data.Receiver, ShouldEqual, "김예준 님")
			So(data.CompanyCode, ShouldEqual, "KG")
		})
	})
}

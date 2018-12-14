package couriers

import (
	"fmt"
	"testing"

	"github.com/jarcoal/httpmock"
	"github.com/purpleworks/delibird"
	. "github.com/smartystreets/goconvey/convey"
)

func TestHanjinCourier(t *testing.T) {
	// create hanjin courier
	courier, _ := New("HANJIN")

	// create http mock
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	invalidTrackingNumber := "hanjin_invalid"
	completeTrackingNumber := "hanjin_complete"

	httpmock.RegisterResponder("GET", fmt.Sprintf(Hanjin{}.TrackingUrl(), invalidTrackingNumber),
		httpmock.NewStringResponder(200, readTestResponseFile(invalidTrackingNumber+".html")))
	httpmock.RegisterResponder("GET", fmt.Sprintf(Hanjin{}.TrackingUrl(), completeTrackingNumber),
		httpmock.NewStringResponder(200, readTestResponseFile(completeTrackingNumber+".html")))

	Convey("HANJIN test", t, func() {
		Convey("Invalid tracking number test", func() {
			_, err := courier.Parse(invalidTrackingNumber)

			So(err, ShouldNotBeNil)
		})

		Convey("Complete courier test", func() {
			data, _ := courier.Parse(completeTrackingNumber)

			So(data.StatusCode, ShouldEqual, delibird.DeleveryComplete)
			So(data.Sender, ShouldEqual, "위****** 님")
			So(data.Receiver, ShouldEqual, "김* 님")
			So(data.Signer, ShouldEqual, "김*(본인)")
			So(data.CompanyCode, ShouldEqual, "HANJIN")
		})
	})
}

package couriers

import (
	"fmt"
	"testing"

	"github.com/jarcoal/httpmock"
	"github.com/purpleworks/delibird"
	. "github.com/smartystreets/goconvey/convey"
)

func TestSlxCourier(t *testing.T) {
	// create slx courier
	courier, _ := New("SLX")

	// create http mock
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	invalidTrackingNumber := "slx_invalid"
	//startTrackingNumber := "slx_start"
	completeTrackingNumber := "slx_complete"

	httpmock.RegisterResponder("GET", fmt.Sprintf(Slx{}.TrackingUrl(), invalidTrackingNumber),
		httpmock.NewStringResponder(200, readTestResponseFile(invalidTrackingNumber+".html")))
	//httpmock.RegisterResponder("GET", fmt.Sprintf(Slx{}.TrackingUrl(), startTrackingNumber),
	//	httpmock.NewStringResponder(200, readTestResponseFile(startTrackingNumber+".html")))
	httpmock.RegisterResponder("GET", fmt.Sprintf(Slx{}.TrackingUrl(), completeTrackingNumber),
		httpmock.NewStringResponder(200, readTestResponseFile(completeTrackingNumber+".html")))

	Convey("SLX test", t, func() {
		Convey("Invalid tracking number test", func() {
			_, err := courier.Parse(invalidTrackingNumber)
			So(err, ShouldNotBeNil)
		})

		//Convey("Start courier test", func() {
		//	data, _ := courier.Parse(startTrackingNumber)
		//
		//	So(data.StatusCode, ShouldEqual, delibird.Ready)
		//	So(data.Sender, ShouldEqual, "홈*럼")
		//	So(data.Receiver, ShouldEqual, "테*트")
		//	So(data.Signer, ShouldEqual, "")
		//	So(data.CompanyCode, ShouldEqual, "CJ")
		//})

		Convey("Complete courier test", func() {
			data, _ := courier.Parse(completeTrackingNumber)

			So(data.StatusCode, ShouldEqual, delibird.DeleveryComplete)
			So(data.Sender, ShouldEqual, "홈*럼")
			So(data.Receiver, ShouldEqual, "테*트")
			So(data.Signer, ShouldEqual, "본인")
			So(data.CompanyCode, ShouldEqual, "SLX")
		})
	})
}

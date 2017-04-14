package couriers

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"strings"

	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/purpleworks/delibird"
	"golang.org/x/text/encoding/korean"
	"golang.org/x/text/transform"
)

type Logen struct{}

func init() {
	RegisterCourier("LOGEN", &Logen{})
}

func (t Logen) Code() string {
	return "LOGEN"
}

func (t Logen) Name() string {
	return "로젠택배"
}

func (t Logen) TrackingUrl() string {
	return "https://www.ilogen.com/iLOGEN.Web.New/TRACE/TraceDetail.aspx?slipno=%s&gubun=fromview"
}

func (t Logen) Parse(trackingNumber string) (delibird.Track, *delibird.ApiError) {
	track := delibird.Track{}

	body, err := t.getHtml(trackingNumber)

	//buf := new(bytes.Buffer)
	//buf.ReadFrom(body)
	//log.Print(buf.String())

	if err != nil {
		return track, delibird.NewApiError(delibird.RequestPageError, err.Error())
	}

	doc, err := goquery.NewDocumentFromReader(body)
	if err != nil {
		return track, delibird.NewApiError(delibird.ParseError, err.Error())
	}

	noTrackingStatusLength := doc.Find(".table_01").Length()
	if noTrackingStatusLength > 0 {
		return track, delibird.NewApiError(delibird.NoTrackingInfo, "등록되지 않은 운송장이거나 배송준비중입니다.")
	}

	track = delibird.Track{
		TrackingNumber: trackingNumber,
		CompanyCode:    t.Code(),
		CompanyName:    t.Name(),
		Sender:         strings.TrimSpace(doc.Find("#tbSndCustNm").Eq(0).AttrOr("value", "-")),
		Receiver:       strings.TrimSpace(doc.Find("#tbRcvCustNm").Eq(0).AttrOr("value", "-")),
		Signer:         strings.TrimSpace(doc.Find("#tbSignGubun").Eq(0).AttrOr("value", "-")),
	}

	history := []delibird.History{}

	//배송정보
	doc.Find("table table").Eq(1).Find("td table table tr").Each(func(i int, s *goquery.Selection) {
		if i > 0 {
			dateText := strings.TrimSpace(s.Find("td").Eq(0).Text())
			if dateText == "" {
				return
			}

			date, err := time.Parse("2006-01-02 15:04", dateText)
			if err != nil {
				log.Printf("%s - %s", dateText, err)
			} else {
				statusText := strings.TrimSpace(s.Find("td").Eq(2).Text())
				if i == 1 {
					track.StatusCode = t.getStatus(statusText)
					track.StatusText = statusText
				}
				history = append(history,
					delibird.History{
						Date:       date.Add(-time.Hour * 9).Unix(),
						DateText:   date.Format("2006-01-02 15:04"),
						Area:       strings.TrimSpace(s.Find("td").Eq(1).Text()),
						StatusCode: t.getStatus(statusText),
						StatusText: statusText,
					})

				track.StatusCode = t.getStatus(statusText)
				track.StatusText = statusText
			}
		}
	})
	track.History = history

	return track, nil
}

func (t Logen) getHtml(trackingNumber string) (io.Reader, error) {
	url := fmt.Sprintf(t.TrackingUrl(), trackingNumber)
	res, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	convertedBody := transform.NewReader(bytes.NewReader(body), korean.EUCKR.NewDecoder())

	return convertedBody, nil
}

func (t Logen) getStatus(status_text string) delibird.TrackingStatus {
	switch status_text {
	case "터미널입고":
		return delibird.Loading
	case "터미널출고":
		return delibird.Unloading
	case "배송출고":
		return delibird.DeleveryStart
	case "배송완료":
		return delibird.DeleveryComplete
	case "미배달":
		return delibird.DoNotDelevery
	}

	return delibird.UnknownStatus
}

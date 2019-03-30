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

type Slx struct{}

func init() {
	RegisterCourier("SLX", &Slx{})
}

func (t Slx) Code() string {
	return "SLX"
}

func (t Slx) Name() string {
	return "SLX"
}

func (t Slx) TrackingUrl() string {
	return "http://aladin.slx.co.kr/Tracking.slx?param1=%s"
}

func (t Slx) Parse(trackingNumber string) (delibird.Track, *delibird.ApiError) {
	track := delibird.Track{}

	body, err := t.getHtml(trackingNumber)

	if err != nil {
		return track, delibird.NewApiError(delibird.RequestPageError, err.Error())
	}

	doc, err := goquery.NewDocumentFromReader(body)
	if err != nil {
		return track, delibird.NewApiError(delibird.ParseError, err.Error())
	}

	hasTrainNo := strings.TrimSpace(doc.Find(".train_no span").First().Text())
	if len(hasTrainNo) == 0 {
		return track, delibird.NewApiError(delibird.NoTrackingInfo, "등록되지 않은 운송장이거나 배송준비중입니다.")
	}

	track = delibird.Track{
		TrackingNumber: trackingNumber,
		CompanyCode:    t.Code(),
		CompanyName:    t.Name(),
		//document.querySelectorAll("table")[1].querySelectorAll("td")[1];
		Sender:   strings.TrimSpace(doc.Find("table").Eq(1).Find("td").Eq(1).Text()),
		Receiver: strings.TrimSpace(doc.Find("table").Eq(1).Find("td").Eq(5).Text()),
		Signer:   strings.TrimSpace(doc.Find("table").Eq(1).Find("td").Eq(7).Text()),
	}

	history := []delibird.History{}

	//배송정보
	//document.querySelectorAll("table")[2].querySelectorAll("tr")[1];
	doc.Find("table").Eq(2).Find("tr").Each(func(i int, s *goquery.Selection) {
		if i > 0 {
			dateText := strings.TrimSpace(s.Find("td").Eq(0).Text())
			if dateText == "" {
				return
			}

			timeText := strings.TrimSpace(s.Find("td").Eq(1).Text())
			if timeText == "" {
				return
			}

			date, err := time.Parse("2006/01/02 15:04", dateText+" "+timeText)
			if err != nil {
				log.Printf("%s - %s", dateText, err)
			} else {
				statusText := strings.TrimSpace(s.Find("td").Eq(4).Text())
				history = append(history,
					delibird.History{
						Date:       date.Add(-time.Hour * 9).Unix(),
						DateText:   date.Format("2006-01-02 15:04"),
						Area:       strings.TrimSpace(s.Find("td").Eq(2).Text()),
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

func (t Slx) getHtml(trackingNumber string) (io.Reader, error) {
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

func (t Slx) getStatus(status_text string) delibird.TrackingStatus {
	switch status_text {
	case "집하":
		return delibird.Ready
	case "입고":
		return delibird.Loading
	case "배송 출발":
		return delibird.DeleveryStart
	case "배송 완료":
		return delibird.DeleveryComplete
	}

	return delibird.UnknownStatus
}

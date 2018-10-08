package couriers

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/purpleworks/delibird"
	"golang.org/x/text/encoding/korean"
	"golang.org/x/text/transform"
)

type Cj struct{}

func init() {
	RegisterCourier("CJ", &Cj{})
}

func (t Cj) Code() string {
	return "CJ"
}

func (t Cj) Name() string {
	return "CJ대한통운"
}

func (t Cj) TrackingUrl() string {
	return "http://nexs.cjgls.com/web/info.jsp?slipno=%s"
}

func (t Cj) Parse(trackingNumber string) (delibird.Track, *delibird.ApiError) {
	track := delibird.Track{}

	body, err := t.getHtml(trackingNumber)

	if err != nil {
		return track, delibird.NewApiError(delibird.RequestPageError, err.Error())
	}

	doc, err := goquery.NewDocumentFromReader(body)
	if err != nil {
		return track, delibird.NewApiError(delibird.ParseError, err.Error())
	}

	trackingStatus := doc.Find("table").Eq(0).Find("tbody tr td").Text()
	if strings.Index(trackingStatus, "미등록운송장") > -1 {
		return track, delibird.NewApiError(delibird.NoTrackingInfo, "등록되지 않은 운송장이거나 배송준비중입니다.")
	}

	track = delibird.Track{
		TrackingNumber: trackingNumber,
		CompanyCode:    t.Code(),
		CompanyName:    t.Name(),
		Sender:         strings.TrimSpace(doc.Find("table").Eq(2).Find("tbody tr").Eq(1).Find("td").Eq(0).Text()),
		Receiver:       strings.TrimSpace(doc.Find("table").Eq(2).Find("tbody tr").Eq(1).Find("td").Eq(1).Text()),
		Signer:         strings.TrimSpace(doc.Find("table").Eq(2).Find("tbody tr").Eq(1).Find("td").Eq(3).Text()),
	}

	history := []delibird.History{}

	numberReg, _ := regexp.Compile(`\(([0-9-]+)\)`)

	//배송정보
	doc.Find("table").Eq(4).Find("tbody tr").Each(func(i int, s *goquery.Selection) {
		dateText := strings.TrimSpace(s.Find("td").Eq(0).Text()) + " " + strings.TrimSpace(s.Find("td").Eq(1).Text())
		if i > 0 && strings.Index(dateText, "Tel :") <= 0 {
			date, err := time.Parse("2006-01-02 15:04:05", dateText)
			if err != nil {
				log.Print(err)
			} else {
				statusText := strings.TrimSpace(s.Find("td").Eq(5).Text())
				if i == 1 {
					track.StatusCode = t.getStatus(statusText)
					track.StatusText = statusText
				}
				tel := ""
				tels := numberReg.FindStringSubmatch(s.Find("td table tr td").Eq(1).Text())
				if len(tels) > 1 {
					tel = tels[1]
				}
				if tel == "--" {
					tel = ""
				}
				history = append([]delibird.History{
					delibird.History{
						Date:       date.Add(-time.Hour * 9).Unix(),
						DateText:   date.Format("2006-01-02 15:04"),
						Area:       strings.TrimSpace(s.Find("td table tr td").Eq(0).Text()),
						Tel:        tel,
						StatusCode: t.getStatus(statusText),
						StatusText: statusText,
					},
				}, history...)

			}
		}
	})
	track.History = history

	return track, nil
}

func (t Cj) getHtml(trackingNumber string) (io.Reader, error) {
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

func (t Cj) getStatus(status_text string) delibird.TrackingStatus {
	switch status_text {
	case "SM입고":
		return delibird.Ready
	case "집화처리":
		return delibird.PickupComplete
	case "간선상차":
		return delibird.Loading
	case "간선하차":
		return delibird.Unloading
	case "배달출발":
		return delibird.DeleveryStart
	case "배달완료":
		return delibird.DeleveryComplete
	case "미배달":
		return delibird.DoNotDelevery
	}

	return delibird.UnknownStatus
}

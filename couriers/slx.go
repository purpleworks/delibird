package couriers

import (
	"bytes"
	"delibird"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	//"github.com/purpleworks/delibird"
)

// SLX : slx 당일 택배
type Slx struct{}

func init() {
	RegisterCourier("SLX", &Slx{})
}

// Code : SLX
func (t Slx) Code() string {
	return "SLX"
}

// Name : SLX 당일 택배
func (t Slx) Name() string {
	return "SLX 당일택배"
}

// TrackingUrl : slx url
func (t Slx) TrackingUrl() string {
	return "http://www.slx.co.kr/delivery/delivery_number.php?param1=%s"
}

func (t Slx) getHtml(trackingNumber string) (io.Reader, error) {
	url := fmt.Sprintf(t.TrackingUrl(), trackingNumber)
	res, err := http.Get(url)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	return bytes.NewBuffer(body), nil
}

// Parse : parsing html
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

	trackingStatus := doc.Find("div.cont").Find("div h4").Text()
	if strings.Index(trackingStatus, "검색된 결과가 없습니다. ") > -1 { // strings.index  문자열이 없으면 return -1
		return track, delibird.NewApiError(delibird.NoTrackingInfo, "등록되지 않은 운송장이거나 배송준비중입니다.")
	}

	track = delibird.Track{
		TrackingNumber: trackingNumber,
		CompanyCode:    t.Code(),
		CompanyName:    t.Name(),
		Sender:         strings.TrimSpace(doc.Find("div.tbl_type02").Eq(0).Find("table tbody tr").Eq(1).Find("td").Eq(0).Text()),
		Receiver:       strings.TrimSpace(doc.Find("div.tbl_type02").Eq(0).Find("table tbody tr").Eq(2).Find("td").Eq(0).Text()),
		//Signer:         strings.TrimSpace(doc.Find("div.tbl_type02").Eq(0).Find("table tbody tr").Eq(1).Find("td").Eq(0).Text()),
	}

	history := []delibird.History{}

	//배송정보
	size := doc.Find("div.tbl_type02").Eq(1).Find("table tbody tr").Length()
	doc.Find("div.tbl_type02").Eq(1).Find("table tbody tr").Each(func(i int, s *goquery.Selection) {
		dateText := strings.TrimSpace(s.Find("td").Eq(0).Text())
		if i > 0 {
			date, err := time.Parse("2006.01.02 15:04", dateText)
			if err != nil {
				log.Print(err)
			} else {
				statusText := strings.TrimSpace(s.Find("td").Eq(4).Text())
				if i == (size - 1) {
					track.StatusCode = t.getStatus(statusText)
					track.StatusText = statusText
				}
				tel := strings.TrimSpace(s.Find("td").Eq(3).Text())

				history = append(history,
					[]delibird.History{
						delibird.History{
							Date:       date.Add(-time.Hour * 9).Unix(),
							DateText:   date.Format("2006-01-02 15:04"),
							Area:       strings.TrimSpace(s.Find("td").Eq(1).Text()),
							Tel:        tel,
							StatusCode: t.getStatus(statusText),
							StatusText: statusText,
						},
					}...,
				)

			}
		}
	})
	track.History = history

	return track, nil
}

func (t Slx) getStatus(status_text string) delibird.TrackingStatus {
	switch status_text {
	case "집하":
		return delibird.PickupComplete
	case "도착":
	case "배달전":
		return delibird.DeleveryStart
	case "배달완료":
		return delibird.DeleveryComplete
	case "미배달":
		return delibird.DoNotDelevery
	}

	return delibird.UnknownStatus
}

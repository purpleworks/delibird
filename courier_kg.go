package delibird

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
)

type Kg struct{}

func init() {
	RegisterCourier("KG", &Kg{})
}

func (t Kg) Code() string {
	return "KG"
}

func (t Kg) Name() string {
	return "KG로지스"
}

func (t Kg) TrackingUrl() string {
	return "http://www.kglogis.co.kr/contents/waybill.jsp?item_no=%s"
}

func (t Kg) Parse(trackingNumber string) (Track, *ApiError) {
	track := Track{}

	body, err := t.getHtml(trackingNumber)

	if err != nil {
		return track, NewApiError(RequestPageError, err.Error())
	}

	doc, err := goquery.NewDocumentFromReader(body)
	if err != nil {
		return track, NewApiError(ParseError, err.Error())
	}

	if strings.HasPrefix(strings.TrimSpace(doc.Text()), "alert") {
		return track, NewApiError(NoTrackingInfo, "등록되지 않은 운송장이거나 배송준비중입니다.")
	}

	if strings.TrimSpace(doc.Find("table").Eq(1).Find("tbody tr td").Eq(0).Text()) == "물품 이동경로정보가 없습니다." {
		return track, NewApiError(NoTrackingInfo, "등록되지 않은 운송장이거나 배송준비중입니다.")
	}

	track = Track{
		TrackingNumber: trackingNumber,
		CompanyCode:    t.Code(),
		CompanyName:    t.Name(),
		Sender:         strings.TrimSpace(doc.Find("table").Eq(0).Find("tbody tr").Eq(1).Find("td span").Eq(0).Text()),
		Receiver:       strings.TrimSpace(doc.Find("table").Eq(0).Find("tbody tr").Eq(2).Find("td span").Eq(0).Text()),
		Signer:         "",
	}

	history := []History{}

	numberReg, _ := regexp.Compile("[^0-9-]")

	//배송정보
	doc.Find("table").Eq(1).Find("tbody tr").Each(func(i int, s *goquery.Selection) {
		dateText := strings.TrimSpace(s.Find("td").Eq(0).Text()) + " " + strings.TrimSpace(s.Find("td").Eq(1).Text())
		date, err := time.Parse("2006.01.02 15:04", dateText)
		if err != nil {
			log.Fatal(err)
		} else {
			statusText := strings.TrimSpace(s.Find("td span").Eq(3).Text())

			track.StatusCode = t.getStatus(statusText)
			track.StatusText = statusText

			area_tel := strings.Split(strings.TrimSpace(s.Find("td span").Eq(2).Text()), "/")
			tel := numberReg.ReplaceAllString(area_tel[1], "")
			if tel == "--" {
				tel = ""
			}
			history = append(history,
				History{
					Date:       date.Add(-time.Hour * 9).Unix(),
					DateText:   date.Format("2006-01-02 15:04"),
					Area:       strings.TrimSpace(area_tel[0]),
					Tel:        tel,
					StatusCode: t.getStatus(statusText),
					StatusText: statusText,
				})

		}
	})
	track.History = history

	return track, nil
}

func (t Kg) getHtml(trackingNumber string) (io.Reader, error) {
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

	return bytes.NewBuffer(body), nil
}

func (t Kg) getStatus(status_text string) TrackingStatus {
	switch status_text {
	case "집하":
		return PickupComplete
	case "간선입고":
		return Loading
	case "간선출고":
		return Unloading
	case "배송출발":
		return DeleveryStart
	case "배송완료":
		return DeleveryComplete
	case "미배달":
		return DoNotDelevery
	}

	return UnknownStatus
}

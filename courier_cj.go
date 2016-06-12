package delibird

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/djimenez/iconv-go"
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

func (t Cj) Parse(trackingNumber string) (Track, *ApiError) {
	track := Track{}

	body, err := t.getHtml(trackingNumber)

	if err != nil {
		return track, NewApiError(RequestPageError, err.Error())
	}

	doc, err := goquery.NewDocumentFromReader(body)
	if err != nil {
		return track, NewApiError(ParseError, err.Error())
	}

	trackingStatus := doc.Find("table").Eq(0).Find("tbody tr td").Text()
	if strings.Index(trackingStatus, "미등록운송장") > -1 {
		return track, NewApiError(NoTrackingInfo, "등록되지 않은 운송장이거나 배송준비중입니다.")
	}

	track = Track{
		TrackingNumber: trackingNumber,
		CompanyCode:    t.Code(),
		CompanyName:    t.Name(),
		Sender:         strings.TrimSpace(doc.Find("table").Eq(2).Find("tbody tr").Eq(1).Find("td").Eq(0).Text()),
		Receiver:       strings.TrimSpace(doc.Find("table").Eq(2).Find("tbody tr").Eq(1).Find("td").Eq(1).Text()),
		Signer:         strings.TrimSpace(doc.Find("table").Eq(2).Find("tbody tr").Eq(1).Find("td").Eq(3).Text()),
	}

	history := []History{}

	numberReg, _ := regexp.Compile("[^0-9-]")

	//배송정보
	doc.Find("table").Eq(4).Find("tbody tr").Each(func(i int, s *goquery.Selection) {
		dateText := strings.TrimSpace(s.Find("td").Eq(0).Text()) + " " + strings.TrimSpace(s.Find("td").Eq(1).Text())
		if i > 0 && strings.Index(dateText, "Tel :") <= 0 {
			date, err := time.Parse("2006-01-02 15:04:05", dateText)
			if err != nil {
				log.Fatal(err)
			} else {
				statusText := strings.TrimSpace(s.Find("td").Eq(5).Text())
				if i == 1 {
					track.StatusCode = t.getStatus(statusText)
					track.StatusText = statusText
				}
				tel := numberReg.ReplaceAllString(s.Find("td table tr td").Eq(1).Text(), "")
				if tel == "--" {
					tel = ""
				}
				history = append([]History{
					History{
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

func (t Cj) getHtml(trackingNumber string) (*iconv.Reader, error) {
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

	convertedBody, err := iconv.NewReader(bytes.NewReader(body), "euc-kr", "utf-8")
	if err != nil {
		return nil, err
	}

	return convertedBody, nil
}

func (t Cj) getStatus(status_text string) TrackingStatus {
	switch status_text {
	case "SM입고":
		return Ready
	case "집화처리":
		return PickupComplete
	case "간선상차":
		return Loading
	case "간선하차":
		return Unloading
	case "배달출발":
		return DeleveryStart
	case "배달완료":
		return DeleveryComplete
	case "미배달":
		return DoNotDelevery
	}

	return UnknownStatus
}

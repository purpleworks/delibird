package delibird

import (
	"bytes"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"regexp"
	"strings"
	"time"
)

type Kpost struct{}

func init() {
	RegisterCourier("KPOST", &Kpost{})
}

func (t Kpost) Code() string {
	return "KPOST"
}

func (t Kpost) Name() string {
	return "우체국택배"
}

func (t Kpost) TrackingUrl() string {
	return "https://service.epost.go.kr/trace.RetrieveDomRigiTraceList.comm?sid1=%s&displayHeader=N"
}

func (t Kpost) Parse(trackingNumber string) (Track, *ApiError) {
	track := Track{}

	body, err := t.getHtml(trackingNumber)

	if err != nil {
		return track, NewApiError(RequestPageError, err.Error())
	}

	doc, err := goquery.NewDocumentFromReader(body)
	if err != nil {
		return track, NewApiError(ParseError, err.Error())
	}

	senderInfo, _ := doc.Find(".shipping_area table").Eq(0).Find("tbody tr").Eq(1).Find("td").Eq(1).Html()
	sender := ""
	if tmp := strings.Split(senderInfo, "<br/>"); len(tmp) == 2 {
		sender = tmp[0]
	}

	if sender == "" {
		return track, NewApiError(NoTrackingInfo, "등록되지 않은 운송장이거나 배송준비중입니다.")
	}

	receiverInfo, _ := doc.Find(".shipping_area table").Eq(0).Find("tbody tr").Eq(1).Find("td").Eq(2).Html()
	receiver := ""
	if tmp := strings.Split(receiverInfo, "<br/>"); len(tmp) == 2 {
		receiver = tmp[0]
	}

	signerInfo, _ := doc.Find(".shipping_area table").Eq(2).Find("tbody tr").Last().Find("td").Eq(3).Html()
	signer := ""
	if match := regexp.MustCompile("\\(수령인:(.*)\\)").FindAllStringSubmatch(signerInfo, -1); len(match) > 0 {
		signer = match[0][1]
	}
	track = Track{
		TrackingNumber: trackingNumber,
		CompanyCode:    t.Code(),
		CompanyName:    t.Name(),
		Sender:         sender,
		Receiver:       receiver,
		Signer:         signer,
	}

	history := []History{}

	//배송정보
	doc.Find(".shipping_area table").Eq(2).Find("tbody tr").Each(func(i int, s *goquery.Selection) {
		if i > 0 {
			dateText := strings.TrimSpace(s.Find("td").Eq(0).Text()) + " " + strings.TrimSpace(s.Find("td").Eq(1).Text())
			date, err := time.Parse("2006.01.02 15:04", dateText)
			if err != nil {
				log.Fatal(err)
			} else {
				statusText := strings.TrimSpace(s.Find("td").Eq(3).Text())
				if strings.Contains(statusText, "\n") {
					statusText = statusText[0:strings.Index(statusText, "\n")]
				}

				if i == doc.Find(".shipping_area table").Eq(2).Find("tbody tr").Size()-1 {
					track.StatusCode = t.getStatus(statusText)
					track.StatusText = statusText
				}

				// TODO: popup?
				tel := ""
				history = append(history,
					History{
						Date:       date.Add(-time.Hour * 9).Unix(),
						DateText:   date.Format("2006-01-02 15:04"),
						Area:       strings.TrimSpace(s.Find("td table tr td").Eq(0).Text()),
						Tel:        tel,
						StatusCode: t.getStatus(statusText),
						StatusText: statusText,
					})
			}

		}
	})
	track.History = history

	return track, nil
}

func (t Kpost) getHtml(trackingNumber string) (io.Reader, error) {
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

func (t Kpost) getStatus(status_text string) TrackingStatus {
	switch status_text {
	case "접수":
		return Ready
	case "발송":
		return Loading
	case "도착":
		return Unloading
	case "배달준비":
		return DeleveryStart
	case "배달완료":
		return DeleveryComplete
	case "미배달":
		return DoNotDelevery
	}

	return UnknownStatus
}

/**
 * add hanjin
 *
 * @author TOTALSOFT (admin@totalsoft.co.kr)
 */
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

type Hanjin struct{}

func init() {
	RegisterCourier("HANJIN", &Hanjin{})
}

func (t Hanjin) Code() string {
	return "HANJIN"
}

func (t Hanjin) Name() string {
	return "한진택배"
}

func (t Hanjin) TrackingUrl() string {
	return "https://www.hanjin.co.kr/delivery_html/inquiry/result_waybill.jsp?wbl_num=%s"
}

func (t Hanjin) Re(sentence string) string {
	re_leadclose_whtsp := regexp.MustCompile(`^[\s\p{Zs}]+|[\s\p{Zs}]+$`)
	re_inside_whtsp := regexp.MustCompile(`[\s\p{Zs}]{2,}`)
	final := re_leadclose_whtsp.ReplaceAllString(sentence, "")
	final = re_inside_whtsp.ReplaceAllString(final, " ")

	return final
}

func (t Hanjin) Parse(trackingNumber string) (delibird.Track, *delibird.ApiError) {
	track := delibird.Track{}

	body, err := t.getHtml(trackingNumber)

	if err != nil {
		return track, delibird.NewApiError(delibird.RequestPageError, err.Error())
	}

	doc, err := goquery.NewDocumentFromReader(body)
	if err != nil {
		return track, delibird.NewApiError(delibird.ParseError, err.Error())
	}

	trackingStatus := doc.Text()
	if strings.Index(trackingStatus, "result_error.jsp") > -1 {
		return track, delibird.NewApiError(delibird.NoTrackingInfo, "등록되지 않은 운송장이거나 배송준비중입니다.")
	}

	signerInfo, _ := doc.Find("table").Eq(1).Find("tbody tr td").Eq(-1).Html()
	signerInfo = t.Re(signerInfo)
	signer := ""
	if match := regexp.MustCompile("\u003cb\u003e수령인 : (.*) \u003c/b\u003e").FindAllStringSubmatch(signerInfo, -1); len(match) > 0 {
		signer = match[0][1]
	}

	track = delibird.Track{
		TrackingNumber: trackingNumber,
		CompanyCode:    t.Code(),
		CompanyName:    t.Name(),
		Sender:         t.Re(doc.Find("table").Eq(0).Find("tbody tr td").Eq(3).Text()),
		Receiver:       t.Re(doc.Find("table").Eq(0).Find("tbody tr td").Eq(4).Text()),
		Signer:         signer,
	}

	history := []delibird.History{}

	numberReg, _ := regexp.Compile("[^0-9-]")

	//배송정보
	doc.Find("table").Eq(1).Find("tbody tr").Each(func(i int, s *goquery.Selection) {
		dateText := strings.TrimSpace(s.Find("td").Eq(0).Text()) + " " + strings.TrimSpace(s.Find("td").Eq(1).Text())

		if strings.Index(dateText, "수령인") > -1 {
			return
		}

		date, err := time.Parse("2006-01-02 15:04", dateText)

		if err != nil {
			log.Print(err)
		} else {
			statusText := t.Re(s.Find("td").Eq(3).Text())
			if len(statusText) > 0 {
				track.StatusCode = t.getStatus(statusText)
				track.StatusText = statusText
			}
			tel := numberReg.ReplaceAllString(s.Find("td").Eq(4).Text(), "")
			if tel == "-" {
				tel = ""
			}
			history = append([]delibird.History{
				delibird.History{
					Date:       date.Add(-time.Hour * 9).Unix(),
					DateText:   date.Format("2006-01-02 15:04"),
					Area:       strings.TrimSpace(s.Find("td").Eq(2).Text()),
					Tel:        tel,
					StatusCode: t.getStatus(statusText),
					StatusText: statusText,
				},
			}, history...)

		}
	})
	track.History = history

	return track, nil
}

func (t Hanjin) getHtml(trackingNumber string) (io.Reader, error) {
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

func (t Hanjin) strpos(haystack string, needle string) bool {
	if strings.Index(haystack, needle) > -1 {
		return true
	}

	return false
}

func (t Hanjin) getStatus(status_text string) delibird.TrackingStatus {
	switch {
	case t.strpos(status_text, "터미널에 입고"):
		return delibird.Ready
	case t.strpos(status_text, "터미널에 도착"), t.strpos(status_text, "이동중"):
		return delibird.Loading
	case t.strpos(status_text, "배송준비중"):
		return delibird.Unloading
	case t.strpos(status_text, "배송출발"):
		return delibird.DeleveryStart
	case t.strpos(status_text, "배송완료"):
		return delibird.DeleveryComplete
	case t.strpos(status_text, "미배달"):
		return delibird.DoNotDelevery
	}

	return delibird.UnknownStatus
}

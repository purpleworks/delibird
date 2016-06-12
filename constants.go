package delibird

type TrackingStatus int

const (
	// 알수 없음
	UnknownStatus TrackingStatus = -1
	// 접수 대기
	Pending TrackingStatus = 1
	// 영업점 접수 (SM입고)
	Ready TrackingStatus = 2
	// 집화처리
	PickupComplete TrackingStatus = 3
	// 간선상차 / 물건 실음 / 중간 집화지 출발
	Loading TrackingStatus = 4
	// 간선하차 / 분류 / 중간 집화지 도착
	Unloading TrackingStatus = 5
	// 배송출발
	DeleveryStart TrackingStatus = 51
	// 배송완료
	DeleveryComplete TrackingStatus = 91
	// 미배달
	DoNotDelevery TrackingStatus = 99
)

const (
	NoCode           string = "NO_CODE_AVAILABLE"
	NoTrackingInfo   string = "NO_TRACKING_INFO"
	ParseError       string = "PARSE_ERROR"
	RequestPageError string = "REQUEST_PAGE_ERROR"
)

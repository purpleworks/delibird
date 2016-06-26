package couriers

import (
	"fmt"
	"github.com/purpleworks/delibird"
	"reflect"
)

// Courier is the interface representing the standardized methods to
// parse shipment tracking html
type Courier interface {
	// Parse html to tracking object
	Parse(invoice string) (delibird.Track, *delibird.ApiError)
	// Courier code
	Code() string
	// Courier name
	Name() string
}

var courierMap = map[string]Courier{}

// NewCourier creates courier object by courier company code
func New(name string) (Courier, *delibird.ApiError) {
	if value, ok := courierMap[name]; ok {
		courier := reflect.New(reflect.TypeOf(value).Elem()).Interface().(Courier)
		return courier, nil
	}

	return nil, delibird.NewApiError(delibird.NoCode, fmt.Sprintf("%v is not supported.", name))
}

// RegisterCourier register new courier
func RegisterCourier(name string, courier Courier) {
	courierMap[name] = courier
}

package env

import (
	"fmt"
	"os"
	"reflect"
	_ "regexp"
	"strconv"
	"strings"
)

// Var struct
type Var struct {
	Name     string
	Key      string
	Type     reflect.Type
	Value    reflect.Value
	Required bool
	Default  reflect.Value
	Options  []reflect.Value
}

// NewVar returns a new Var
func NewVar(field reflect.StructField) (*Var, error) {
	// spew.Dump(new(Var).Default == reflect.ValueOf(nil))
	newVar := &Var{} //Default: reflect.ValueOf(nil)}
	newVar.Parse(field)

	value, err := convert(newVar.Type, os.Getenv(newVar.Key))
	if err != nil {
		return newVar, err
	}
	newVar.SetValue(value)

	if value == reflect.ValueOf(nil) {
		if newVar.Required {
			return newVar, fmt.Errorf("%s required", newVar.Key)
		}

		// Check if we have a default value to set, otherwise set the type's zero value
		if newVar.Default != reflect.ValueOf(nil) {
			// fmt.Println("setting default:", newVar.Default.String())
			newVar.SetValue(newVar.Default)
		} else {
			// fmt.Println("No default; setting zero value")
			newVar.SetValue(reflect.Zero(newVar.Type))
		}
	}

	if len(newVar.Options) > 0 {
		if !newVar.optionsContains(newVar.Value) {
			return newVar, fmt.Errorf(`%v="%v" not in allowed options: %v`, newVar.Key, newVar.Value, newVar.Options)
		}
	}

	return newVar, nil
}

func (v *Var) optionsContains(s reflect.Value) bool {
	for _, v := range v.Options {
		if s.Interface() == v.Interface() {
			return true
		}
	}
	return false
}

// SetValue sets Var.Value
func (v *Var) SetValue(value reflect.Value) {
	v.Value = value
}

// SetName sets Var.Name
func (v *Var) SetName(value string) {
	v.Name = value
}

// SetType sets Var.Type
func (v *Var) SetType(value reflect.Type) {
	v.Type = value
}

// SetRequired sets Var.Required
func (v *Var) SetRequired(value bool) {
	v.Required = value
}

// SetDefault sets Var.Default
func (v *Var) SetDefault(value reflect.Value) {
	v.Default = value
}

// SetOptions sets Var.Options
func (v *Var) SetOptions(values []reflect.Value) {
	v.Options = values
}

// SetKey sets Var.Key
func (v *Var) SetKey(value string) {
	// src := []byte(value)
	// regex := regexp.MustCompile("[0-9A-Za-z]+")
	// chunks := regex.FindAll(src, -1)
	// for i, val := range chunks {
	//
	// }
	v.Key = strings.ToUpper(value)
}

// Parse parses the struct tags of each field
func (v *Var) Parse(field reflect.StructField) error {
	v.SetName(field.Name)
	v.SetType(field.Type)
	v.SetKey(v.Name)

	tag := field.Tag.Get("env")

	if tag == "" {
		return nil
	}

	tagParams := strings.Split(tag, " ")
	for _, tagParam := range tagParams {
		var key, value string

		option := strings.Split(tagParam, "=")
		key = option[0]
		if len(option) > 1 {
			value = option[1]
		}

		switch key {
		case "key":
			// override the default key if one is specified
			v.SetKey(value)
		case "required":
			// val, _ := strconv.ParseBool(value)
			// if val != false {
			v.SetRequired(true)
			// }
		case "default":
			d, err := convert(v.Type, value)
			if err != nil {
				return err
			}
			v.SetDefault(d)
		case "options":
			in := strings.Split(value, ",")
			// var values []reflect.Value
			values := make([]reflect.Value, len(in))
			for k, val := range in {
				v1, err := convert(v.Type, val)
				if err != nil {
					return err
				}
				values[k] = v1
			}
			v.SetOptions(values)
		}
	}

	return nil
}

// Convert a string into the specified type. Return the type's zero value
// if we receive an empty string
func convert(t reflect.Type, value string) (reflect.Value, error) {
	if value == "" {
		return reflect.ValueOf(nil), nil
	}

	switch t.Kind() {
	case reflect.String:
		return reflect.ValueOf(value), nil
		// ptr.Elem()
		// ptr = reflect.ValueOf(value).Elem().Convert(reflect.String)
		// return reflect.ValueOf(value), nil
	case reflect.Int:
		return parseInt(value)
	case reflect.Bool:

		return parseBool(value)
	}

	return reflect.ValueOf(nil), conversionError(value, `unsupported `+t.Kind().String())
}

type errConversion struct {
	Value string
	Type  string
}

func (e *errConversion) Error() string {
	return fmt.Sprintf(`could not convert value "%s" into %s type`, e.Value, e.Type)
}

func conversionError(v, t string) *errConversion {
	return &errConversion{Value: v, Type: t}
}

func parseInt(value string) (reflect.Value, error) {
	if value == "" {
		return reflect.Zero(reflect.TypeOf(reflect.Int)), nil
	}
	i, err := strconv.Atoi(value)
	if err != nil {
		return reflect.ValueOf(nil), conversionError(value, "int")
	}
	return reflect.ValueOf(i), nil
}

func parseBool(value string) (reflect.Value, error) {
	if value == "" {
		return reflect.Zero(reflect.TypeOf(reflect.Int)), nil
	}
	b, err := strconv.ParseBool(value)
	if err != nil {
		return reflect.ValueOf(nil), conversionError(value, "bool")
	}
	return reflect.ValueOf(b), nil
}

func parseFloat(value string) (reflect.Value, error) {
	b, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return reflect.ValueOf(nil), conversionError(value, "float64")
	}
	return reflect.ValueOf(b), nil
}

package env

// package main

import (
	"errors"
	"reflect"
)

// Env struct
type Env struct {
	Value  reflect.Value // Value is the value of an interface or pointer
	Prefix string
	Vars   []*Var
}

// Process takes a struct, and maps environment variables to its fields.
// Errors returned from underlying functions will bubble up to the surface.
func Process(v interface{}) error {
	_, err := NewEnv(v)
	if err != nil {
		return err
	}

	return nil
}

// MustProcess maps environment variables to the fields of struct v.
// If any errors are returned, this function will panic.
func MustProcess(v interface{}) {
	_, err := NewEnv(v)
	if err != nil {
		panic(err)
	}
}

// FieldNames returns the name of all struct fields as aa slice of strings
func (e *Env) FieldNames() []string {
	fieldType := e.Type()

	var fieldNames []string
	for i := 0; i < fieldType.NumField(); i++ {
		field := fieldType.Field(i)
		fieldNames = append(fieldNames, field.Name)
	}
	return fieldNames
}

// Type returns the type of e.Value
func (e *Env) Type() reflect.Type {
	return e.Value.Type()
}

var errInvalidValue = errors.New("expected value must be a pointer to a struct")

// New is a shortcut wrapper around NewEnv
func New(v interface{}) (*Env, error) {
	return NewEnv(v)
}

// NewEnv returns a new Env
func NewEnv(v interface{}) (*Env, error) {
	e := &Env{}

	if reflect.TypeOf(v).Kind() != reflect.Ptr {
		return nil, errInvalidValue
	}
	if reflect.ValueOf(v).Elem().Kind() != reflect.Struct {
		return nil, errInvalidValue
	}

	e.SetValue(v)

	if err := e.Parse(); err != nil {
		return e, err
	}

	return e, nil
}

// SetValue sets Value of Env e
func (e *Env) SetValue(v interface{}) {
	if reflect.TypeOf(v).Kind() == reflect.Ptr {
		e.Value = reflect.ValueOf(v).Elem()
	} else {
		e.Value = reflect.ValueOf(v)
	}
}

// SetPrefix sets prefix of Env e
func (e *Env) SetPrefix(prefix string) {
	e.Prefix = prefix
}

// Parse parses the config struct into valid Vars
func (e *Env) Parse() error {
	for _, name := range e.FieldNames() {
		field, _ := e.Value.Type().FieldByName(name)
		v, err := NewVar(field)

		if err != nil {
			return err
		}
		e.Value.FieldByName(name).Set(v.Value)
		e.Vars = append(e.Vars, v)
	}

	return nil
}

// Copyright 2021 Flamego. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

// Package inject provides utilities for mapping and injecting dependencies in
// various ways.
package inject

import (
	"fmt"
	"reflect"
)

// Injector represents an interface for mapping and injecting dependencies into
// structs and function arguments.
type Injector interface {
	Applicator
	Invoker
	TypeMapper
	// SetParent sets the parent of the injector. If the injector cannot find a
	// dependency in its Type map it will check its parent before returning an
	// error.
	SetParent(Injector)
}

// Applicator represents an interface for mapping dependencies to a struct.
type Applicator interface {
	// Apply maps dependencies in the Type map to each field in the struct that is
	// tagged with "inject". Returns an error if the injection fails.
	Apply(interface{}) error
}

// Invoker represents an interface for calling functions via reflection.
type Invoker interface {
	// Invoke attempts to call the `interface{}` provided as a function, providing
	// dependencies for function arguments based on Type. Returns a slice of
	// reflect.Value representing the returned values of the function. Returns an
	// error if the injection fails.
	Invoke(interface{}) ([]reflect.Value, error)
}

// FastInvoker represents an interface in order to avoid the calling function
// via reflection.
type FastInvoker interface {
	// Invoke attempts to call the `interface{}` provided as interface method,
	// providing dependencies for function arguments based on Type. Returns a slice
	// of reflect.Value representing the returned values of the function. Returns an
	// error if the injection fails.
	Invoke([]interface{}) ([]reflect.Value, error)
}

// IsFastInvoker returns true if the `handler` implements FastInvoker.
func IsFastInvoker(handler interface{}) bool {
	_, ok := handler.(FastInvoker)
	return ok
}

// TypeMapper represents an interface for mapping `interface{}` values based on
// type.
type TypeMapper interface {
	// Map maps the `interface{}` values based on their immediate type from
	// reflect.TypeOf.
	Map(...interface{}) TypeMapper
	// MapTo maps the `interface{}` value based on the pointer of an Interface
	// provided. This is really only useful for mapping a value as an interface, as
	// interfaces cannot at this time be referenced directly without a pointer.
	MapTo(interface{}, interface{}) TypeMapper
	// Set provides a possibility to directly insert a mapping based on type and
	// value. This makes it possible to directly map type arguments not possible to
	// instantiate with reflect like unidirectional channels.
	Set(reflect.Type, reflect.Value) TypeMapper
	// Value returns the reflect.Value that is mapped to the reflect.Type. It
	// returns a zeroed reflect.Value if the Type has not been mapped.
	Value(reflect.Type) reflect.Value
}

type injector struct {
	values map[reflect.Type]reflect.Value
	parent Injector
}

// InterfaceOf dereferences a pointer to an Interface type. It panics if value
// is not an pointer to an interface.
func InterfaceOf(value interface{}) reflect.Type {
	t := reflect.TypeOf(value)
	for t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	if t.Kind() != reflect.Interface {
		panic("called inject.InterfaceOf with a value that is not a pointer to an interface. (*MyInterface)(nil)")
	}
	return t
}

// New returns a new Injector.
func New() Injector {
	return &injector{
		values: make(map[reflect.Type]reflect.Value),
	}
}

// Invoke attempts to call the interface{} provided as a function,
// providing dependencies for function arguments based on Type.
// Returns a slice of reflect.Value representing the returned values of the function.
// Returns an error if the injection fails.
// It panics if f is not a function
func (inj *injector) Invoke(f interface{}) ([]reflect.Value, error) {
	t := reflect.TypeOf(f)
	switch v := f.(type) {
	case FastInvoker:
		return inj.fastInvoke(v, t, t.NumIn())
	default:
		return inj.callInvoke(f, t, t.NumIn())
	}
}

func (inj *injector) fastInvoke(f FastInvoker, t reflect.Type, numIn int) ([]reflect.Value, error) {
	var in []interface{}
	if numIn > 0 {
		in = make([]interface{}, numIn) // Panic if t is not kind of Func
		var argType reflect.Type
		var val reflect.Value
		for i := 0; i < numIn; i++ {
			argType = t.In(i)
			val = inj.Value(argType)
			if !val.IsValid() {
				return nil, fmt.Errorf("value not found for type %v", argType)
			}

			in[i] = val.Interface()
		}
	}
	return f.Invoke(in)
}

func (inj *injector) callInvoke(f interface{}, t reflect.Type, numIn int) ([]reflect.Value, error) {
	var in []reflect.Value
	if numIn > 0 {
		in = make([]reflect.Value, numIn)
		var argType reflect.Type
		var val reflect.Value
		for i := 0; i < numIn; i++ {
			argType = t.In(i)
			val = inj.Value(argType)
			if !val.IsValid() {
				return nil, fmt.Errorf("value not found for type %v", argType)
			}

			in[i] = val
		}
	}
	return reflect.ValueOf(f).Call(in), nil
}

func (inj *injector) Apply(val interface{}) error {
	v := reflect.ValueOf(val)

	for v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	if v.Kind() != reflect.Struct {
		return nil // Should not panic here ?
	}

	t := v.Type()

	for i := 0; i < v.NumField(); i++ {
		f := v.Field(i)
		structField := t.Field(i)
		_, ok := structField.Tag.Lookup("inject")
		if f.CanSet() && ok {
			ft := f.Type()
			v := inj.Value(ft)
			if !v.IsValid() {
				return fmt.Errorf("value not found for type %v", ft)
			}

			f.Set(v)
		}

	}
	return nil
}

func (inj *injector) Map(values ...interface{}) TypeMapper {
	for _, val := range values {
		inj.values[reflect.TypeOf(val)] = reflect.ValueOf(val)
	}
	return inj
}

func (inj *injector) MapTo(val, ifacePtr interface{}) TypeMapper {
	inj.values[InterfaceOf(ifacePtr)] = reflect.ValueOf(val)
	return inj
}

func (inj *injector) Set(typ reflect.Type, val reflect.Value) TypeMapper {
	inj.values[typ] = val
	return inj
}

func (inj *injector) Value(t reflect.Type) reflect.Value {
	val := inj.values[t]

	if val.IsValid() {
		return val
	}

	// No concrete types found, try to find implementors if t is an interface.
	if t.Kind() == reflect.Interface {
		for k, v := range inj.values {
			if k.Implements(t) {
				val = v
				break
			}
		}
	}

	// Still no type found, try to look it up on the parent
	if !val.IsValid() && inj.parent != nil {
		val = inj.parent.Value(t)
	}

	return val
}

func (inj *injector) SetParent(parent Injector) {
	inj.parent = parent
}

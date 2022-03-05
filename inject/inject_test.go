// Copyright 2021 Flamego. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package inject

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

type specialString interface{}

type testStruct struct {
	Dep1 string        `inject:"" json:"-"`
	Dep2 specialString `inject:""`
	Dep3 string
}

type greeter struct {
	Name string
}

func (g *greeter) String() string {
	return "Hello, My name is" + g.Name
}

type myFastInvoker func(string)

func (myFastInvoker) Invoke([]interface{}) ([]reflect.Value, error) {
	return nil, nil
}

func TestInjector_Invoke(t *testing.T) {
	t.Run("invoke functions", func(t *testing.T) {
		inj := New()

		dep := "some dependency"
		inj.Map(dep)
		dep2 := "another dep"
		inj.MapTo(dep2, (*specialString)(nil))
		dep3 := make(chan *specialString)
		dep4 := make(chan *specialString)
		typRecv := reflect.ChanOf(reflect.RecvDir, reflect.TypeOf(dep3).Elem())
		typSend := reflect.ChanOf(reflect.SendDir, reflect.TypeOf(dep4).Elem())
		inj.Set(typRecv, reflect.ValueOf(dep3))
		inj.Set(typSend, reflect.ValueOf(dep4))

		_, err := inj.Invoke(func(d1 string, d2 specialString, d3 <-chan *specialString, d4 chan<- *specialString) {
			assert.Equal(t, dep, d1)
			assert.Equal(t, dep2, d2)
			assert.Equal(t, reflect.TypeOf(dep3).Elem(), reflect.TypeOf(d3).Elem())
			assert.Equal(t, reflect.TypeOf(dep4).Elem(), reflect.TypeOf(d4).Elem())
			assert.Equal(t, reflect.RecvDir, reflect.TypeOf(d3).ChanDir())
			assert.Equal(t, reflect.SendDir, reflect.TypeOf(d4).ChanDir())
		})
		assert.Nil(t, err)

		_, err = inj.Invoke(myFastInvoker(func(string) {}))
		assert.Nil(t, err)
	})

	t.Run("invoke functions with return values", func(t *testing.T) {
		inj := New()

		dep := "some dependency"
		inj.Map(dep)
		dep2 := "another dep"
		inj.MapTo(dep2, (*specialString)(nil))

		result, err := inj.Invoke(func(d1 string, d2 specialString) string {
			assert.Equal(t, dep, d1)
			assert.Equal(t, dep2, d2)
			return "Hello world"
		})
		assert.Nil(t, err)

		assert.Equal(t, "Hello world", result[0].String())
	})
}

func TestInjector_Apply(t *testing.T) {
	inj := New()
	inj.Map("a dep").MapTo("another dep", (*specialString)(nil))

	s := testStruct{}
	assert.Nil(t, inj.Apply(&s))

	assert.Equal(t, "a dep", s.Dep1)
	assert.Equal(t, "another dep", s.Dep2)
}

func TestInjector_InterfaceOf(t *testing.T) {
	iType := InterfaceOf((*specialString)(nil))
	assert.Equal(t, reflect.Interface, iType.Kind())

	iType = InterfaceOf((**specialString)(nil))
	assert.Equal(t, reflect.Interface, iType.Kind())

	defer func() {
		assert.NotNil(t, recover())
	}()
	InterfaceOf((*testing.T)(nil))
}

func TestInjector_Set(t *testing.T) {
	inj := New()

	typ := reflect.TypeOf("string")
	typSend := reflect.ChanOf(reflect.SendDir, typ)
	typRecv := reflect.ChanOf(reflect.RecvDir, typ)

	// Instantiating unidirectional channels is not possible using reflect, see body
	// of reflect.MakeChan for detail.
	chanRecv := reflect.MakeChan(reflect.ChanOf(reflect.BothDir, typ), 0)
	chanSend := reflect.MakeChan(reflect.ChanOf(reflect.BothDir, typ), 0)

	inj.Set(typSend, chanSend)
	inj.Set(typRecv, chanRecv)

	assert.True(t, inj.Value(typSend).IsValid())
	assert.True(t, inj.Value(typRecv).IsValid())
	assert.False(t, inj.Value(chanSend.Type()).IsValid())
}

func TestInjector_GetVal(t *testing.T) {
	inj := New()
	inj.Map("some dependency")

	assert.True(t, inj.Value(reflect.TypeOf("string")).IsValid())
	assert.False(t, inj.Value(reflect.TypeOf(11)).IsValid())
}

func TestInjector_SetParent(t *testing.T) {
	inj := New()
	inj.MapTo("another dep", (*specialString)(nil))

	inj2 := New()
	inj2.SetParent(inj)

	assert.True(t, inj2.Value(InterfaceOf((*specialString)(nil))).IsValid())
}

func TestInjector_Implementors(t *testing.T) {
	inj := New()

	g := &greeter{"Jeremy"}
	inj.Map(g)

	assert.True(t, inj.Value(InterfaceOf((*fmt.Stringer)(nil))).IsValid())
}

func TestIsFastInvoker(t *testing.T) {
	assert.True(t, IsFastInvoker(myFastInvoker(nil)))
}

func BenchmarkInjector_Invoke(b *testing.B) {
	inj := New()
	inj.Map("some dependency").MapTo("another dep", (*specialString)(nil))

	fn := func(d1 string, d2 specialString) string { return "something" }
	for i := 0; i < b.N; i++ {
		_, _ = inj.Invoke(fn)
	}
}

type testFastInvoker func(d1 string, d2 specialString) string

func (f testFastInvoker) Invoke(args []interface{}) ([]reflect.Value, error) {
	f(args[0].(string), args[1].(specialString))
	return nil, nil
}

func BenchmarkInjector_FastInvoke(b *testing.B) {
	inj := New()
	inj.Map("some dependency").MapTo("another dep", (*specialString)(nil))

	fn := testFastInvoker(func(d1 string, d2 specialString) string { return "something" })
	for i := 0; i < b.N; i++ {
		_, _ = inj.Invoke(fn)
	}
}

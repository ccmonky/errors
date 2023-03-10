package errors

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
	"sync"

	"github.com/ccmonky/log"
)

var (
	// ErrorAttr used to attach another error on input error, usually used to attach meta error
	ErrorAttr = NewAttr[error]("error", WithAttrDescription("error as an attr"))

	// CtxAttr used to attach context.Context on error
	CtxAttr = NewAttr[context.Context]("ctx", WithAttrDescription("context.Context as an attr"))

	// Meta used to attach meta to Error
	MetaAttr = NewAttr[*Meta]("meta", WithAttrDescription("meta as an attr"))

	// Message used to attach message to Error
	MessageAttr = NewAttr[string]("msg", WithAttrDescription("message as an attr"))

	// Status used as meta value stands for http status,
	// Status.Get returns http status if err is MetaError with status attached,
	// otherwise return 500 if err != nil else return 200
	StatusAttr = NewAttr[int]("status",
		WithAttrDefault(func(err error) any {
			if err != nil {
				return http.StatusInternalServerError
			}
			return http.StatusOK
		}),
		WithAttrDescription("http status as an attr"))

	// Caller used as meta value stands for runtime.Caller info
	CallerAttr = NewAttr[string]("caller", WithAttrDescription("caller as an attr"))

	// Stack attach `github.com/pkg/errors.stack` on error
	StackAttr = NewAttr[*stack]("stack", WithAttrDescription("github.com/pkg/errors.stack as an attr"))
)

/*
Attr defines an value extension for `Meta`, it can create an MetaOption by `With` method and get the value by `Get` method

Usage:

	// define a new Attr
	var Status = NewAttr[int](WithAttrDefault(func(err error){
		if err != nil {
			return http.StatusInternalServerError
		}
		return http.StatusOK
	}))

	// use `Attr.With`
	var NotFound = NewMetaError("not_found(5)", "not found", source, Status.With(http.StatusNotFound))

	// use `Attr.Get`
	log.Println(Status.Get(err))

*/
type Attr[T any] struct {
	key          *string
	defaultValue func(error) any
	description  string
}

func (a *Attr[T]) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]any{
		"type":                   reflect.TypeOf(new(T)).Elem().String(),
		"name":                   *a.key,
		"description":            a.description,
		"has_default_value_func": a.defaultValue != nil,
	})
}

// NewAttr creates a new `Attr`
func NewAttr[T any](name string, opts ...AttrOption) *Attr[T] {
	options := AttrOptions{}
	for _, opt := range opts {
		opt(&options)
	}
	attr := Attr[T]{
		key:          NewAttrKey(name).(*string),
		defaultValue: options.DefaultValueFunc,
		description:  options.Description,
	}
	if options.DoNotRegister {
		return &attr
	}
	_, ok := nameAttrs.LoadOrStore(*attr.key, &attr)
	if ok {
		if options.PanicOnDuplicateNames {
			log.Panicf("Attr(%s:%s) with same name already exists", name, options.Description)
			return nil
		} else {
			nameAttrs.Store(*attr.key, &attr) // NOTE: override by default
		}
	}
	_, ok = attrs.LoadOrStore(attr.key, &attr)
	if ok {
		// NOTE: Attr's key generated by NewAttrKey internally, if duplicate, then NewAttrKey is problematic!
		log.Fatalf("Attr(%s:%s) with same key already exists", name, options.Description)
	}
	return &attr
}

// AttrOptions defines `Attr` constructor options
type AttrOptions struct {
	// DefaultValueFunc Attr's default value function
	DefaultValueFunc func(error) any

	// Description Attr's description
	Description string

	// DoNotRegister do not register the new created attr into registry, default to false
	DoNotRegister bool

	// PanicOnDuplicateNames duplicate names are allowed by default
	PanicOnDuplicateNames bool
}

// AttrOption defines `Attr` constructor option
type AttrOption func(*AttrOptions)

// WithAttrDefault specify `Attr` default value function
func WithAttrDefault(fn func(error) any) AttrOption {
	return func(options *AttrOptions) {
		options.DefaultValueFunc = fn
	}
}

// WithAttrDefault specify `Attr` description
func WithAttrDescription(desc string) AttrOption {
	return func(options *AttrOptions) {
		options.Description = desc
	}
}

// WithAttrDoNotRegister specify `Attr` DoNotRegister
func WithAttrDoNotRegister(noReg bool) AttrOption {
	return func(options *AttrOptions) {
		options.DoNotRegister = noReg
	}
}

// WithAttrPanicOnDuplicateNames specify `Attr` PanicOnDuplicateNames
func WithAttrPanicOnDuplicateNames(noPanic bool) AttrOption {
	return func(options *AttrOptions) {
		options.PanicOnDuplicateNames = noPanic
	}
}

// Key returns the internal key of Attr
func (a *Attr[T]) Name() string {
	return *a.key
}

// Description returns the description of Attr
func (a *Attr[T]) Description() string {
	return a.description
}

// Key returns the internal key of Attr
func (a *Attr[T]) Key() any {
	return a.key
}

func (a *Attr[T]) With(err error, value T) error {
	if err == nil {
		return nil
	}
	return &valueError{err, a.key, value}
}

func (a *Attr[T]) Option(value T) Option {
	return func(err error) error {
		return a.With(err, value)
	}
}

// Get get value specified by Attr's internal key from error if implement MetaError, otherwise return default value
func (a *Attr[T]) Get(err error) T {
	if ce, ok := err.(Error); ok {
		value := ce.Value(a.key)
		if value != nil {
			if tv, ok := value.(T); ok {
				return tv
			} else {
				log.Panicf("attr %v got invalid type value, expect %T, got %T", *a.key, *new(T), value)
			}
		}
	}
	if a.defaultValue != nil {
		return a.defaultValue(err).(T)
	}
	return *new(T)
}

// GetAny return attr value of err as any
func (a *Attr[T]) GetAny(err error) any {
	return a.Get(err)
}

// GetAll get all values specified by Attr's internal key from error
func (a *Attr[T]) GetAll(err error) []T {
	var all []T
	values := GetAll(err, a.key)
	for _, value := range values {
		if tv, ok := value.(T); ok {
			all = append(all, tv)
		} else {
			log.Panicf("attr %v got invalid type value, expect %T, got %T", *a.key, *new(T), value)
		}
	}
	return all
}

// NewAttrKey used to creates a new Attr's internal key
// https://github.com/golang/go/issues/33742
func NewAttrKey(name string) any {
	return &name
}

// MustGetAttrByKey get attr by key, panic if not found
func MustGetAttrByKey[T any](key any) *Attr[T] {
	a, err := GetAttrByKey[T](key)
	if err != nil {
		log.Panic(err)
	}
	return a
}

// GetAttrByKey get attr by type and key from attrs registry
func GetAttrByKey[T any](key any) (*Attr[T], error) {
	v, ok := attrs.Load(key)
	if !ok {
		return nil, WithError(fmt.Errorf("attr(name=%s) with key(%v) not found", *key.(*string), key), NotFound)
	}
	a, ok := v.(*Attr[T])
	if !ok {
		return nil, WithError(fmt.Errorf("attr(name=%s) with type %T not found", *key.(*string), *new(T)), NotFound)
	}
	return a, nil
}

// MustGetAttrByKey get attr by name, panic if not found
func MustGetAttrByName[T any](name string) *Attr[T] {
	a, err := GetAttrByName[T](name)
	if err != nil {
		log.Panic(err)
	}
	return a
}

// GetAttrByName get attr by type and name from attrs registry
// NOTE: the result may not be what you want, since the same name is allowed to be overwritten by default.
func GetAttrByName[T any](name string) (*Attr[T], error) {
	v, ok := nameAttrs.Load(name)
	if !ok {
		return nil, WithError(fmt.Errorf("attr with name %s not found", name), NotFound)
	}
	a, ok := v.(*Attr[T])
	if !ok {
		return nil, WithError(fmt.Errorf("attr(name=%s) with type %T not found", name, *new(T)), NotFound)
	}
	return a, nil
}

// AllAttrs return all registered Attrs
func AllAttrs() map[string]any {
	m := make(map[string]any)
	fn := func(key, value any) bool {
		m[fmt.Sprintf("%s:%v", *key.(*string), key)] = value
		return true
	}
	attrs.Range(fn)
	return m
}

// AttrInterface abstract Attr's minimal interface used for `Attrs`
type AttrInterface interface {
	Key() any
	Name() string
	Description() string
	GetAny(error) any
}

// Attrs is a group of AttrInterface, which usually used to extractor attrs's values
type Attrs []AttrInterface

// NewAttrs create new Attrs
func NewAttrs(attrs ...AttrInterface) Attrs {
	return Attrs(attrs)
}

// Map returns the name:value map of Attrs according to Attrs's order
func (as Attrs) Map(err error) map[string]any {
	var m = make(map[string]any, len(as)+5) // NOTE: 5 means flatten meta(4)+status(1) in most common scenarios
	for _, a := range as {
		v := Get(err, a.Key())
		m[*a.Key().(*string)] = v
		switch me := v.(type) {
		case *Meta:
			m[MetaAttrAppFieldName] = me.app()
			m[MetaAttrSourceFieldName] = me.source
			m[MetaAttrCodeFieldName] = me.code
			m[MetaAttrMessageFieldName] = me.msg
		case error:
			mm := Map(me)
			for kk, vv := range mm {
				m[kk] = vv
			}
		}
	}
	return m
}

var (
	attrs     sync.Map // map[*string]*Attr
	nameAttrs sync.Map // map[string]*Attr
)

var (
	_ AttrInterface = (*Attr[error])(nil)
)

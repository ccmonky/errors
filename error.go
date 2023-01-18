package errors

import (
	"fmt"
	"io"
	"reflect"
	"sync"

	"github.com/ccmonky/log"
)

// helper functions for attrs
var (
	WithError   = ErrorAttr.With
	ErrorOption = ErrorAttr.Option

	WithMeta   = MetaAttr.With
	MetaOption = MetaAttr.Option

	WithCtx   = CtxAttr.With
	CtxOption = CtxAttr.Option

	WithStatus   = StatusAttr.With
	StatusOption = StatusAttr.Option

	WithCaller   = CallerAttr.With
	CallerOption = CallerAttr.Option

	StackOption   = StackAttr.Option
	MessageOption = MessageAttr.Option
)

// NewMetaError define a new error with meta attached
func NewMetaError(source, code, msg string, opts ...Option) error {
	ce := MetaAttr.With(empty, NewMeta(source, code, msg))
	for _, opt := range opts {
		ce = opt(ce)
	}
	return ce
}

// With used to attach multiple values on error with options
func With(err error, opts ...Option) error {
	if err == nil {
		return nil
	}
	for _, opt := range opts {
		err = opt(err)
	}
	return err
}

func WithMetaNx(err, metaErr error, opts ...Option) error {
	return nil
}

// WithErrorOptions
func WithErrorOptions(err, optsErr error) error {
	if err == nil {
		return nil
	}
	for _, opt := range Options(optsErr) {
		err = opt(err)
	}
	return err
}

// Option used to attach value on error
type Option func(error) error

// ContextError is a error with context values
type ContextError interface {
	error

	// Value get value specified by key in the Meta
	Value(key any) any
}

type emptyContextError struct{}

func (e *emptyContextError) String() string {
	return ""
}

func (e *emptyContextError) Error() string {
	return ""
}

func (*emptyContextError) Value(key any) any {
	return nil
}

// Empty returns a non-nil, empty ContextError. It has no values.
// It is typically used by the main function, initialization, and tests,
// and as the top-level ContextError when define a new ContextError.
func Empty() ContextError {
	return empty
}

// WithValue returns a copy of parent in which the value associated with key is
// val.
//
// The provided key must be comparable and should not be of type
// string or any other built-in type to avoid collisions between
// packages using meta. Users of WithValue should define their own
// types for keys. To avoid allocating when assigning to an
// interface{}, context keys often have concrete type
// struct{}. Alternatively, exported context key variables' static
// type should be a pointer or interface.
func WithValue(err error, key, val any) error {
	if err == nil {
		log.Panicln("cannot create Error from nil parent")
	}
	if key == nil {
		log.Panicln("nil key")
	}
	if !reflect.TypeOf(key).Comparable() {
		log.Panicln("key is not comparable")
	}
	return &valueError{err, key, val}
}

// A valueError carries a key-value pair. It implements Value for that key and
// delegates all other calls to the embedded Error.
type valueError struct {
	error
	key, val any
}

func (e *valueError) Value(key any) any {
	if e.key == key {
		return e.val
	}
	if vg, ok := e.val.(valueGetter); ok {
		value := vg.Value(key)
		if value != nil {
			return value
		}
	}
	return Get(e.error, key)
}

func Get(m error, key any) any {
	for {
		switch tm := m.(type) {
		case *valueError:
			if key == tm.key {
				return tm.val
			}
			if vg, ok := tm.val.(valueGetter); ok {
				value := vg.Value(key)
				if value != nil {
					return value
				}
			}
			m = tm.error
		case *emptyContextError:
			return nil
		default:
			return nil
		}
	}
}

type valueGetter interface {
	Value(any) any
}

// String returns string representation for valueMeta
// NOTE: if key type is *string, will use *key when print
func (e *valueError) Error() string {
	return formatByMode(e)
}

func (e *valueError) String() string {
	return formatByMode(e)
}

func formatByMode(e *valueError) string {
	formatModeLock.RLock()
	mode := formatMode
	formatModeLock.RUnlock()
	errStr := e.error.Error() + ":"
	if e.error == empty {
		errStr = ""
	}
	switch mode {
	case Simplified:
		return errStr + "*={*}"
	case NoValue:
		if sp, ok := e.key.(*string); ok {
			return errStr + *sp + "={*}"
		}
		return errStr + fmt.Sprintf("%v={*}", e.key)
	default:
		if sp, ok := e.key.(*string); ok {
			return errStr + fmt.Sprintf("%s={%v}", *sp, e.val)
		}
		return errStr + fmt.Sprintf("%v={%v}", e.key, e.val)
	}
}

// Format implement fmt.Formatter
func (e *valueError) Format(s fmt.State, verb rune) {
	switch verb {
	case 'v':
		if s.Flag('+') {
			fmt.Fprintf(s, "%+v\n", e.error)
			var key = e.key
			if sp, ok := e.key.(*string); ok {
				key = *sp
			}
			io.WriteString(s, fmt.Sprintf("%+v={%+v}", key, e.val))
			return
		}
		fallthrough
	case 's', 'q':
		io.WriteString(s, e.Error())
	}
}

// Unwrap used to response to `errors.Unwrap`
func (e *valueError) Unwrap() error {
	return e.error
}

// Is implement errors.Is for valueError, to test if a `valueError` wrap from target, used for error assertion
//
// Usage:
//
//     err := errors.With(errors.Errorf("xxx"), errors.ErrorOption(errors.NotFound), errors.MessageOption("..."))
//     errors.Is(err, errros.NotFound) // true
//
// NOTE:
// 1. if err chain first wrapped with a error err1, then wrapped with a second err2, errors.Is(err, err1) is also true!
func (e *valueError) Is(target error) bool {
	return ErrorAttr.Get(e) == target
}

// IsError used to test if err is a error, return true only if target == Cause(err) || target == ErrorAttr.Get(err)
func IsError(err, target error) bool {
	if Cause(err) == target || ErrorAttr.Get(err) == target {
		return true
	}
	return false
}

// UnwrapAll unwrap to get error, key and value
func (e *valueError) UnwrapAll() (any, any, error) {
	return e.key, e.val, e.error
}

func Options(err error) []Option {
	var unwrapOpts []Option
	for err != nil {
		uv, ok := err.(interface {
			UnwrapAll() (any, any, error)
		})
		if !ok {
			break
		}
		var k, v any
		k, v, err = uv.UnwrapAll()
		unwrapOpts = append(unwrapOpts, func(e error) error {
			return &valueError{e, k, v}
		})
	}
	var opts = make([]Option, 0, len(unwrapOpts))
	for i := len(unwrapOpts) - 1; i >= 0; i-- {
		opts = append(opts, unwrapOpts[i])
	}
	return opts
}

// Cause returns the underlying cause of the error, if possible.
// An error value has a cause if it implements the following
// interface:
//
//     type unwrap interface {
//            Unwrap() error
//     }
//
// If the error does not implement unwrap, the original error will
// be returned. If the error is nil, nil will be returned without further
// investigation.
func Cause(err error) error {
	for err != nil {
		u, ok := err.(interface {
			Unwrap() error
		})
		if !ok {
			break
		}
		err = u.Unwrap()
	}
	return err
}

// FormatMode used to format Meta value wrapper
type FormatMode int

const (
	// Default format value meta as `meta.String():key=value`
	Default FormatMode = iota

	// Simplified format value meta as `meta.String():*=*`
	Simplified

	// Default format value meta as `meta.String():key=*`
	NoValue
)

func SetFormatMode(mode FormatMode) {
	formatModeLock.Lock()
	defer formatModeLock.Unlock()
	if mode != Default && mode != Simplified && mode != NoValue {
		mode = Default
	}
	formatMode = mode
}

var (
	empty = new(emptyContextError)
)

var (
	formatMode     = Default
	formatModeLock sync.RWMutex
)

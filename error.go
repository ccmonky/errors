package errors

import (
	"fmt"
	"io"
	"reflect"
	"sync"

	"github.com/ccmonky/log"
)

var (
	WithMeta = MetaAttr.With
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

// Option used to attach value on error
type Option func(error) error

// WithError wrap error's options if it implement `UnwrapAll()(k,v any, error)`, otherwise nothing changed
func WithError(err, unwrapAllErr error) error {
	for _, opt := range Options(unwrapAllErr) {
		err = opt(err)
	}
	return err
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
	return value(e.error, key)
}

func value(m error, key any) any {
	for {
		switch tm := m.(type) {
		case *valueError:
			if key == tm.key {
				return tm.val
			}
			m = tm.error
		case *emptyContextError:
			return nil
		case ContextError:
			return tm.Value(key)
		default:
			return nil
		}
	}
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
	switch mode {
	case Simplified:
		if e.error == empty {
			return "*={*}"
		}
		return fmt.Sprintf("%v:*={*}", e.error)
	case NoValue:
		if e.error == empty {
			if sp, ok := e.key.(*string); ok {
				return fmt.Sprintf("%s={*}", *sp)
			}
			return fmt.Sprintf("%v={*}", e.key)
		}
		if sp, ok := e.key.(*string); ok {
			return fmt.Sprintf("%v:%s={*}", e.error, *sp)
		}
		return fmt.Sprintf("%v:%v={*}", e.error, e.key)
	default:
		if e.error == empty {
			if sp, ok := e.key.(*string); ok {
				return fmt.Sprintf("%s={%v}", *sp, e.val)
			}
			return fmt.Sprintf("%v={%v}", e.key, e.val)
		}
		if sp, ok := e.key.(*string); ok {
			return fmt.Sprintf("%v:%s={%v}", e.error, *sp, e.val)
		}
		return fmt.Sprintf("%s:%v={%v}", e.error, e.key, e.val)
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

// UnwrapAll unwrap to get error, key and value
func (e *valueError) UnwrapAll() (any, any, error) {
	return e.key, e.val, e.error
}

// Is implement errors.Is for metaError, to test if a `MetaError` wrap from target, used for error assertion
// Usage:
//
// err := metaerrors.Adapt(errors.New("origin error"), metaerrors.NotFound) // error suggestion
// err = errors.WithMessage(err, "wrapper")                                 // error wrapper
// err = metaerrors.Adapt(err, metaerrors.Unknown)                          // error fallback
func (e *valueError) Is(target error) bool {
	return MetaAttr.Get(e) == MetaAttr.Get(target)
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

var (
	empty = new(emptyContextError)
)

var (
	formatMode     = Default
	formatModeLock sync.RWMutex
)

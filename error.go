package errors

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"reflect"
	"sync"

	"github.com/ccmonky/inithook"
	"github.com/ccmonky/log"
)

func init() {
	inithook.RegisterAttrSetter(inithook.AppName, "errors", func(ctx context.Context, value string) error {
		return SetAppName(value)
	})
}

// Option used to attach value on error
type Option func(error) error

// Error is a error with context values
type Error interface {
	error

	// Value get value specified by key in the Meta
	Value(key any) any
}

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

// With used to attach multiple values on error with options
func With(err error, opts ...Option) error {
	if err == nil {
		return nil
	}
	for _, opt := range opts {
		if opt == nil {
			continue
		}
		err = opt(err)
	}
	return err
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

type emptyError struct{}

func (e *emptyError) String() string {
	return ""
}

func (e *emptyError) Error() string {
	return ""
}

func (*emptyError) Value(key any) any {
	return nil
}

// Empty returns a non-nil, empty Error. It has no values.
// It is typically used by the main function, initialization, and tests,
// and as the top-level Error when define a new Error.
func Empty() Error {
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

func (e *valueError) MarshalJSON() ([]byte, error) {
	key := fmt.Sprintf("%v", e.key)
	if ks, ok := e.key.(*string); ok {
		key = *ks
	}
	data, err := json.Marshal(e.error)
	if err != nil {
		return nil, WithMessagef(err, "marshal error %v failed", e.error)
	}
	var rawError json.RawMessage = data
	data, err = json.Marshal(e.val)
	if err != nil {
		return nil, WithMessagef(err, "marshal val %v failed", e.val)
	}
	var rawValue json.RawMessage = data
	if e.error == empty {
		return json.Marshal(map[string]any{
			"key":   key,
			"value": rawValue,
		})
	}
	return json.Marshal(map[string]any{
		"error": rawError,
		"key":   key,
		"value": rawValue,
	})
}

func (e *valueError) Value(key any) any {
	return Get(e, key)
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
		case *emptyError:
			return nil
		default:
			return nil
		}
	}
}

// Values get all values of key recursively
func (e *valueError) Values(key any) []any {
	return GetAll(e, key)
}

// GetAll get all values of key recursively
func GetAll(m error, key any) []any {
	var all []any
	for {
		switch tm := m.(type) {
		case *valueError:
			if key == tm.key {
				all = append(all, tm.val)
			}
			if vg, ok := tm.val.(valuesGetter); ok {
				values := vg.Values(key)
				if len(values) > 0 {
					all = append(all, values...)
				}
			} else {
				if vg, ok := tm.val.(valueGetter); ok {
					value := vg.Value(key)
					if value != nil {
						all = append(all, value)
					}
				}
			}
			m = tm.error
		case *emptyError:
			return reverse(all)
		default:
			return reverse(all)
		}
	}
}

type valueGetter interface {
	Value(any) any
}

type valuesGetter interface {
	Values(any) []any
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

func (e *valueError) App() string {
	value := e.Value(MetaAttr.Key())
	if m, ok := value.(*Meta); ok {
		return m.app()
	}
	return ""
}

func (e *valueError) Source() string {
	value := e.Value(MetaAttr.Key())
	if m, ok := value.(*Meta); ok {
		return m.source
	}
	return ""
}

func (e *valueError) Code() string {
	value := e.Value(MetaAttr.Key())
	if m, ok := value.(*Meta); ok {
		return m.code
	}
	return ""
}

func (e *valueError) Message() string {
	value := e.Value(MetaAttr.Key())
	if m, ok := value.(*Meta); ok {
		return m.msg
	}
	return ""
}

// UnwrapAll unwrap to get error, key and value
func (e *valueError) UnwrapAll() (any, any, error) {
	return e.key, e.val, e.error
}

// IsCauseOrLatest used to test if err is a error, return true only if target == Cause(err) || target == ErrorAttr.Get(err)
func IsCauseOrLatest(err, target error) bool {
	if Cause(err) == target || ErrorAttr.Get(err) == target {
		return true
	}
	return false
}

// Options used to get `[]Option` attached on err
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
	return reverse(unwrapOpts)
}

// Map unwrap all to get `name:value` map of all attrs
//
// NOTE:
// 1. unlike Get or `GetAll`, Map do not traversal the value of `valueError`
// 2. result will not contain the value if key type is not *string
// 3. if meta exists, then app, source and message fields will be added into result
// 4. if key's name duplicates, the result will only contains the latest value
func Map(err error) map[string]any {
	var kvs []kv
	for err != nil {
		uv, ok := err.(interface {
			UnwrapAll() (any, any, error)
		})
		if !ok {
			break
		}
		var k, v any
		k, v, err = uv.UnwrapAll()
		if ksPtr, ok := k.(*string); ok {
			kvs = append(kvs, kv{*ksPtr, v})
		}
	}
	var m = make(map[string]any, len(kvs)+5) // NOTE: 5 means flatten meta(4)+status(1) in most common scenarios
	for _, kv := range reverse(kvs) {
		m[kv.k] = kv.v
		switch me := kv.v.(type) {
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

type kv struct {
	k string
	v any
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

// GetAllErrors get all errors contained in err, all errors attached by `WithError` + Cause
func GetAllErrors(err error) []error {
	var errs = []error{
		Cause(err),
	}
	return append(errs, ErrorAttr.GetAll(err)...)
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

// SetFormatMode set mode for formatting the `valueError`
func SetFormatMode(mode FormatMode) {
	formatModeLock.Lock()
	defer formatModeLock.Unlock()
	if mode != Default && mode != Simplified && mode != NoValue {
		mode = Default
	}
	formatMode = mode
}

var (
	empty = new(emptyError)

	app     string
	appLock sync.RWMutex

	formatMode     = Default
	formatModeLock sync.RWMutex
)

func reverse[T any](in []T) []T {
	var out = make([]T, 0, len(in))
	for i := len(in) - 1; i >= 0; i-- {
		out = append(out, in[i])
	}
	return out
}

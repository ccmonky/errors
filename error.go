package errors

import (
	"context"
	"fmt"
	"io"
	"reflect"
	"runtime"
	"strings"
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

// WithMetaNx wrap err with a metaErr(if not contains meta, metaErr will be `Unknown`) and extra options
func WithMetaNx(err error, guard MetaError) error {
	if err == nil {
		return nil
	}
	if guard == nil {
		guard = Unknown
	}
	if guard.App() != AppName() {
		log.Panicf("gurad meta error's app(%s) != current app name(%s)\n", guard.App(), AppName())
		return err
	}
	dyn := MetaAttr.Get(err)
	if dyn == nil {
		return With(WithError(err, guard), addCaller())
	}
	if dyn.App() != AppName() { // NOTE: maybe upstream case || bad dynamic case
		e := MataMapping()(dyn) // NOTE: try to map to current app's meta error by source & code
		if e != nil {
			return With(WithError(err, guard), addCaller())
		}
		log.Panicf("can not found current app's meta error for %s:%s\n", err, dyn.Source(), dyn.Code())
		return err
	}
	return With(err, addCaller())
}

func addCaller() Option {
	noCallerForNxLock.RLock()
	add := !noCallerForNx
	noCallerForNxLock.RUnlock()
	if add {
		return CallerAttr.Option(caller(3))
	}
	return nil
}

func caller(skip int) string {
	pc, _, _, _ := runtime.Caller(skip)
	path := runtime.FuncForPC(pc).Name()
	parts := strings.Split(path, "/")
	return parts[len(parts)-1]
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
		case *emptyError:
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
	empty = new(emptyError)

	app     string
	appLock sync.RWMutex

	formatMode     = Default
	formatModeLock sync.RWMutex

	noCallerForNx     bool
	noCallerForNxLock sync.RWMutex

	metaMapping     = mappingBySourceCode
	metaMappingLock sync.RWMutex
)

func MataMapping() func(*Meta) MetaError {
	metaMappingLock.RLock()
	defer metaMappingLock.RUnlock()
	return metaMapping
}

func SetMataMapping(fn func(*Meta) MetaError) {
	metaMappingLock.Lock()
	defer metaMappingLock.Unlock()
	metaMapping = fn
}

// mappingBySourceCode map meta to another by code and source in current app codes
func mappingBySourceCode(upstream *Meta) MetaError {
	metaErrorsLock.RLock()
	defer metaErrorsLock.RUnlock()
	for id, me := range metaErrors {
		if MetaID(AppName(), upstream.Source(), upstream.Code()) == id {
			return me
		}
	}
	return nil
}

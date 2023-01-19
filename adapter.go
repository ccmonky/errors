package errors

import (
	"runtime"
	"strconv"
	"strings"

	"github.com/ccmonky/log"
)

// Adapt defaultAdapter's Adapt
func Adapt(err error, guard MetaError) error {
	return defaultAdapter.Adapt(err, guard)
}

// Adapter provides `Adapt` mainly to support suggestive error, error delivery path, error overrides ...
type Adapter interface {
	Adapt(err error, guard MetaError) error
}

// NewAdapter creates a new Adapter, usually no need to create a new one, just use the default `Adapt` function is enough
func NewAdapter(opts ...AdapterOption) Adapter {
	a := adapter{}
	for _, opt := range opts {
		opt(&a)
	}
	return &a
}

// AdapterOption default adapter implementation control option
type AdapterOption func(*adapter)

// WithAddCaller add caller for default adapter implementation
func WithAddCaller() AdapterOption {
	return func(a *adapter) {
		a.AddCaller = true
	}
}

// WithCallerSkip specify caller skip depth for default adapter implementation
func WithCallerSkip(skip int) AdapterOption {
	return func(a *adapter) {
		a.CallerSkip = skip
	}
}

// WithCallerFunc specify the caller func which returns the caller info for default adapter implementation
func WithCallerFunc(fn func(int) string) AdapterOption {
	return func(a *adapter) {
		a.CallerFunc = fn
	}
}

// WithDefaultOptions specify the default meta options to append when `Adapt` for default adapter implementation
func WithDefaultOptions(opts ...Option) AdapterOption {
	return func(a *adapter) {
		a.DefaultOptions = opts
	}
}

// WithMetaMappingFunc specify the default meta mappping function
func WithMetaMappingFunc(fn func(*Meta) MetaError) AdapterOption {
	return func(a *adapter) {
		a.MetaMappingFunc = fn
	}
}

type adapter struct {
	AddCaller       bool
	CallerSkip      int
	CallerFunc      func(skip int) string
	DefaultOptions  []Option
	MetaMappingFunc func(*Meta) MetaError
}

// Adapt append guard into err if err is not MetaError, otherwise only apply adapter's caller & default options,
// Adapt will first find the latest `Meta` dyn in error, if exists, only apply extra options, but these is a special case,
// that is, if dyn's app != current app name, then it will be considered as an Meta casted from upstream, so it can not be
// used directly, and should be mapping to current app's meta, the default mapping is by source+code, unless you specify
// a mapping funciton
func (a *adapter) Adapt(err error, guard MetaError) error {
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
	var opts []Option
	if a.AddCaller {
		callerName := a.CallerFunc(a.CallerSkip)
		opts = append(opts, CallerAttr.Option(callerName))
	}
	opts = append(opts, a.DefaultOptions...)
	dyn := MetaAttr.Get(err)
	if dyn == nil {
		return With(WithError(err, guard), opts...)
	}
	if dyn.App() != AppName() { // NOTE: maybe upstream case || bad dynamic case
		e := a.MetaMappingFunc(dyn) // NOTE: try to map to current app's meta error by source & code
		if e != nil {
			return With(WithError(err, guard), opts...)
		}
		log.Panicf("can not found current app's meta error for %s:%s\n", err, dyn.Source(), dyn.Code())
		return err
	}
	return With(err, opts...)
}

func caller(skip int) string {
	pc, _, line, _ := runtime.Caller(skip)
	path := runtime.FuncForPC(pc).Name()
	parts := strings.Split(path, "/")
	return parts[len(parts)-1] + ":" + strconv.Itoa(line)
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

var defaultAdapter = NewAdapter(
	WithAddCaller(),
	WithCallerSkip(3),
	WithCallerFunc(caller),
	WithMetaMappingFunc(mappingBySourceCode))

var (
	_ Adapter = (*adapter)(nil)
)

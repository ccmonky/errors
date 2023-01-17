package errors

import (
	"fmt"
	"runtime"
	"strings"
)

// Adapt defaultAdapter's Adapt
func Adapt(err error, defaultCtxErr ContextError, opts ...Option) error {
	return defaultAdapter.Adapt(err, defaultCtxErr, opts...)
}

// Adapter provides `Adapt` mainly to support suggestive error, error delivery path, error overrides ...
type Adapter interface {
	Adapt(err error, defaultCtxErr ContextError, opts ...Option) error
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

type adapter struct {
	AddCaller      bool
	CallerSkip     int
	CallerFunc     func(skip int) string
	DefaultOptions []Option
}

// Adapt append defaultMeta into err if err is not ContextError, otherwise only apply opts into err's exist Meta
// note, Adapt will first find the latest Meta(override scenario) specified by meta.App() and meta.Code()
// there are also some connor scenarios rarely used, Adapt will try best to action as expected.
func (a *adapter) Adapt(err error, defaultCtxErr ContextError, opts ...Option) error {
	if err == nil {
		return nil
	}
	if a.AddCaller {
		callerName := a.CallerFunc(a.CallerSkip)
		opts = append(opts, CallerAttr.Option(callerName))
	}
	opts = append(opts, a.DefaultOptions...)
	if dyn := MetaAttr.Get(err); dyn != nil {
		static := GetMeta(dyn.ID())
		if static != nil {
			// NOTE: why wrap again?
			// 1. if len(opts)>0, e.g., add caller info, should wrap again to add dynamic options.
			// 2. if app want to override the previous meta, e.g., attach more values, should wrap again to use the new meta.
			// so, when dynamic opts exists or source changed, should wrap again!
			if dyn == static {
				// NOTE:
				// - if dyn wrap from static, then dyn is likely with more values than meta.
				// - if no dynamic options(opts), then use dyn directly.
				return With(err, opts...)
			} else {
				return With(WithMeta(err, static), opts...)
			}
		} else {
			// cases:
			// 1. not registered meta found, maybe a dynamic meta, rare case
			// 2. upstream meta(e.g. from headers), support meta mapping?
			if dyn.app() == AppName() { // dynamic case
				return With(err, opts...)
			} else { // upstream case
				mm := MappingByCodeSource(dyn)
				if mm == nil {
					mm = refresh(MetaAttr.Get(defaultCtxErr)) // NOTE: mapping meta not found, and should not use upstream meta!
				}
				return With(WithMeta(err, mm), opts...)
			}
		}
	}
	return With(WithMeta(err, refresh(MetaAttr.Get(defaultCtxErr))), opts...)
}

// refresh used for find the latest overrided  meta
func refresh(dyn *Meta) *Meta {
	if dyn == nil {
		return MetaAttr.Get(Unknown)
	}
	static := GetMeta(dyn.ID())
	// cases:
	// 1. if meta == nil, only dyn can be used.
	// 2. if dyn wrap from static, then dyn is likely with more values than me.Meta.
	// 3. if dyn does not wrap from static, then use static is reasonable, but dyn's dyanmic values will be missing!
	if static == nil || dyn != static {
		return dyn
	}
	return static
}

// MetaMapping function used to map a meta to another
type MetaMapping func(*Meta) *Meta

// MappingByCodeSource map meta to another by code and source in current app codes
func MappingByCodeSource(upstream *Meta) *Meta {
	metasLock.RLock()
	defer metasLock.RUnlock()
	for id, meta := range metas {
		if fmt.Sprintf("%s:%s:%s", meta.app(), upstream.Source(), upstream.Code()) == id {
			return meta
		}
	}
	return nil
}

func caller(skip int) string {
	pc, _, _, _ := runtime.Caller(skip)
	path := runtime.FuncForPC(pc).Name()
	parts := strings.Split(path, "/")
	return parts[len(parts)-1]
}

var defaultAdapter = NewAdapter(WithAddCaller(), WithCallerSkip(3), WithCallerFunc(caller))

var (
	_ Adapter = (*adapter)(nil)
)

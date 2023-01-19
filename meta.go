package errors

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"sync"

	"github.com/ccmonky/log"
)

type MetaError interface {
	error
	App() string
	Source() string
	Code() string
	Message() string
}

// NewMetaError define a new error with meta attached
func NewMetaError(source, code, msg string, opts ...Option) MetaError {
	e := MetaAttr.With(empty, newMeta(source, code, msg))
	for _, opt := range opts {
		e = opt(e)
	}
	me := e.(MetaError)
	err := RegisterMetaError(me)
	if err != nil {
		log.Panicln(err, "register meta error failed")
		return nil
	}
	return me
}

type Meta struct {
	app    func() string
	source string
	code   string
	msg    string
}

// newMeta creates a new Meta
func newMeta(source, code, msg string) *Meta {
	m := Meta{
		app:    AppName,
		source: source,
		code:   code,
		msg:    msg,
	}
	return &m
}

func (e *Meta) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]any{
		"meta.app":     e.app(),
		"meta.source":  e.source,
		"meta.code":    e.code,
		"meta.message": e.msg,
	})
}

func (e *Meta) App() string {
	return e.app()
}

func (e *Meta) Source() string {
	return e.source
}

func (e *Meta) Code() string {
	return e.code
}

func (e *Meta) Message() string {
	return e.msg
}

func (e *Meta) ID() string {
	return fmt.Sprintf("%s:%s:%s", e.app(), e.source, e.code)
}

func (e *Meta) String() string {
	return fmt.Sprintf("source=%s;code=%s", pkgName(e.source), e.code)
}

// Format implement fmt.Formatter
func (e *Meta) Format(s fmt.State, verb rune) {
	switch verb {
	case 'v':
		if s.Flag('+') {
			io.WriteString(s, fmt.Sprintf("%s:%s:%s:%s", e.App(), e.Source(), e.Code(), e.Message()))
			return
		}
		fallthrough
	case 's', 'q':
		io.WriteString(s, e.String())
	}
}

func pkgName(source string) string {
	parts := strings.Split(source, "/")
	return parts[len(parts)-1]
}

// RegisterMetaError register MetaError into metaErrors registry, return error is exists
func RegisterMetaError(me MetaError) error {
	if me == nil {
		return New("register nil meta error, just ignore")
	}
	if me.App() != AppName() {
		return Errorf("meta error app(%s) != current app name %s", me.App(), AppName())
	}
	id := MetaID(me.App(), me.Source(), me.Code())
	if me.Source() == "" || me.Code() == "" || me.Message() == "" {
		return Errorf("meta error(%s): source, code and msg can not be empty", id)
	}
	metaErrorsLock.Lock()
	defer metaErrorsLock.Unlock()
	if _, ok := metaErrors[id]; ok {
		return Errorf("meta error %s already exists; use with_xxx to rebind", id)
	}
	metaErrors[id] = me
	return nil
}

// GetMetaError get Meta according to app, source and code
func GetMetaError(id string) MetaError {
	metaErrorsLock.RLock()
	defer metaErrorsLock.RUnlock()
	return metaErrors[id]
}

// AllMetaErrors return all registered MetaErrors as a map with key is `app:source:code`
func AllMetaErrors() map[metaID]MetaError {
	var result = map[metaID]MetaError{}
	metaErrorsLock.RLock()
	defer metaErrorsLock.RUnlock()
	for id, me := range metaErrors {
		result[id] = me
	}
	return result
}

// AppName return current app name, use `SetAppName` or `inithook.AppName` to set app name
func AppName() string {
	appLock.RLock()
	defer appLock.RUnlock()
	return app
}

// SetAppName set app name and register Meta to new app namespace
func SetAppName(appNew string) error {
	appLock.Lock()
	appOld := app
	app = appNew
	appLock.Unlock()
	metaErrorsLock.Lock()
	if appOld != appNew {
		newMetaErrors := make(map[metaID]MetaError, len(metaErrors))
		for _, me := range metaErrors {
			idNew := MetaID(me.App(), me.Source(), me.Code())
			if _, ok := metaErrors[idNew]; ok {
				return Errorf("meta error(%s) already exists", idNew)
			}
			newMetaErrors[idNew] = me
		}
		for id, me := range newMetaErrors {
			metaErrors[id] = me
		}
	}
	metaErrorsLock.Unlock()
	return nil
}

// MetaID returns a unique id of `Meta`
func MetaID(app, source, code string) string {
	return fmt.Sprintf("%s:%s:%s", app, source, code)
}

var (
	metaErrors     = make(map[metaID]MetaError)
	metaErrorsLock sync.RWMutex
)

type metaID = string

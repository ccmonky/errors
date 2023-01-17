package errors

import (
	"context"
	"fmt"
	"io"
	"log"
	"strings"
	"sync"

	"github.com/ccmonky/inithook"
)

// NewMeta creates a new Meta
func NewMeta(source, code, msg string) *Meta {
	m := Meta{
		app:    AppName,
		source: source,
		code:   code,
		msg:    msg,
	}
	return &m
}

type Meta struct {
	app    func() string
	source string
	code   string
	msg    string
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

// RegisterMeta register Meta into metas registry, news will be override olds,
// it's best for app register Meta before any `Adapt` calls, usually in init phase.
func RegisterMeta(meta *Meta) {
	if meta == nil {
		log.Println("register nil Meta")
		return
	}
	metasLock.Lock()
	defer metasLock.Unlock()
	id := meta.ID()
	if _, ok := metas[id]; ok {
		if !OverridableAttr.GetAny(meta) { // FIXME:
			log.Panicf("Meta %s already exists and no overridable attr found", id)
		}
	}
	metas[id] = meta
}

// GetMeta get Meta according to app and code
func GetMeta(id string) *Meta {
	metasLock.RLock()
	defer metasLock.RUnlock()
	return metas[id]
}

// metas return all registered metas as a map with key is `app:code`
func Metas() map[metaID]*Meta {
	var result = map[metaID]*Meta{}
	metasLock.RLock()
	defer metasLock.RUnlock()
	for id, Meta := range metas {
		result[id] = Meta
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
func SetAppName(appNew string) {
	appLock.Lock()
	appOld := app
	app = appNew
	appLock.Unlock()
	metasLock.Lock()
	if appOld != appNew {
		newmetas := make(map[metaID]*Meta, len(metas))
		for _, meta := range metas {
			idNew := meta.ID()
			if _, ok := metas[idNew]; ok {
				if !OverridableAttr.GetAny(meta) {
					log.Panicf("Meta %s already exists and no overridable attr found", idNew)
				}
			}
			newmetas[idNew] = meta
		}
		for id, Meta := range newmetas {
			metas[id] = Meta
		}
	}
	metasLock.Unlock()
}

func SetFormatMode(mode FormatMode) {
	formatModeLock.Lock()
	defer formatModeLock.Unlock()
	if mode != Default && mode != Simplified && mode != NoValue {
		formatMode = Default
	}
	formatMode = mode
}

func init() {
	inithook.RegisterAttrSetter(inithook.AppName, "errors", func(ctx context.Context, value string) error {
		SetAppName(value)
		return nil
	})
}

var (
	app     string
	appLock sync.RWMutex

	metas     = make(map[metaID]*Meta)
	metasLock sync.RWMutex
)

type metaID = string

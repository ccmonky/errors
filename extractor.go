package errors

var (
	MetaAttrAppFieldName     string = "meta.app"
	MetaAttrSourceFieldName  string = "meta.source"
	MetaAttrCodeFieldName    string = "meta.code"
	MetaAttrMessageFieldName string = "meta.message"
)

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

// GetApp get app name from error if err is ContextError, otherwise return empty string
func GetApp(err error) string {
	m := MetaAttr.Get(err)
	if m != nil {
		return m.app()
	}
	return ""
}

// GetSource returns error source if err is ContextError, otherwise return empty string
func GetSource(err error) string {
	m := MetaAttr.Get(err)
	if m != nil {
		return m.source
	}
	return ""
}

// GetCode returns error code if err is ContextError, otherwise return Unknown.Code() if err != nil else return Ok.Code()
func GetCode(err error) string {
	m := MetaAttr.Get(err)
	if m != nil {
		return m.code
	}
	return ""
}

// GetMessage returns error message if err is ContextError, otherwise return Unknown.Message() if err != nil else return Ok.Message()
func GetMessage(err error) string {
	m := MetaAttr.Get(err)
	if m != nil {
		return m.msg
	}
	return ""
}

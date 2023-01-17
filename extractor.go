package errors

var (
	MetaAppAttrName     string = "app_name"
	MetaSourceAttrName  string = "source"
	MetaCodeAttrName    string = "code"
	MetaMessageAttrName string = "message"
)

// Attrs returns meta attrs specified by attrFuncs and buitins(app, code, message, status)
func Attrs(err error, attrFuncs ...AttrFunc) map[string]any {
	m := make(map[string]any, 4+len(attrFuncs))
	m[MetaAppAttrName] = GetApp(err)
	m[MetaSourceAttrName] = GetSource(err)
	m[MetaCodeAttrName] = GetCode(err)
	m[MetaMessageAttrName] = GetMessage(err)
	var fns = make([]AttrFunc, 0, 1+len(attrFuncs))
	fns = append(fns, func(err error) (string, any) { return StatusAttr.Name(), StatusAttr.Get(err) })
	fns = append(fns, attrFuncs...)
	for _, fn := range fns {
		k, v := fn(err)
		m[k] = v
	}
	return m
}

// AttrFunc defines attr getter function from error
type AttrFunc func(error) (attr string, value any)

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
	if err != nil {
		return "unknown(2)" //Unknown.Code()
	}
	return "success(0)" //OK.Code()
}

// GetMessage returns error message if err is ContextError, otherwise return Unknown.Message() if err != nil else return Ok.Message()
func GetMessage(err error) string {
	m := MetaAttr.Get(err)
	if m != nil {
		return m.msg
	}
	if err != nil {
		return "server throws an exception" //Unknown.Message()
	}
	return "success" // OK.Message()
}

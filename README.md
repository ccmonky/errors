# errors

like pkg/errors, but can wrap and retrieve anything

## Requirements

- [Go 1.18 or newer](https://golang.org/dl/)

## Usage

- error definition

```go
var source = "your_package_path"
var NotFound = NewMetaError(source, "not_found(5)", "not found", status(http.StatusNotFound))
```

- error override

```go
import "github.com/ccmonky/errors"

errors.NotFound = errors.With(errors.NotFound, errors.StatusOption(400), ...)
```

- error(meta) suggestion

```go
err := errors.New("xxx")
err = errors.WithError(err, errors.NotFound)
```

- error(meta) fallback

```go
// NOTE: do notknow whether error has carried meta
err = errors.Adapt(err, errors.Unknown)
```

- error enforce

```go
// NOTE: attach errors.Unavailable on err whether err has carried meta or not
err = errors.WithError(err, errors.Unavailable)
```

- error attr

```go
// define attr
CountAttr := errors.NewAttr[int]("count", errors.WithAttrDescription("xxx count as an attr"))

// use attr
err = CountAttr.With(errors.New("xxx"), 100)
// or
err = errors.With(errors.New("xxx"), CountAttr.Option(100))
// get count
count := CountAttr.Get(err)
count == 100 // true
```

- error wrap

```go
err := errors.New("xxx")
err = errors.WithError(err, errors.NotFound)
err = errors.WithMessage(err, "wrapper")
err = errors.WithCaller(err, "caller...")
err = errors.WithSatus(err, 206)
//...
```

- error unwrap

```go
// unwrap error
originErr := errors.New("xxx")
err := Cause(errors.With(err, errors.ErrorOption(errors.NotFound), errors.StatusOption(409), errors.CallerOption("caller...")))
log.Print(originErr == err) // true

// unwrap attr
errors.StatusAttr.Get(err) == 409
errors.ErrorAttr.Get(err) == errors.NotFound
// NOTE: you can also get attr in attr's value(as long as it implement `interface{ Value(any)any}`, like this
errors.MetaAttr.Get(err).Code() == "not_found(5)" 
```

- error assert

```go
originErr := errors.New("xxx")
err := errors.WithError(originErr, errors.NotFound)
err = errors.WithMessage(err, "xxx")

errors.Is(err, originErr)                 // true
errors.Is(err, errors.NotFound)           // true
errors.Is(err, errors.AlreadyExists)      // false

errors.IsCauseOrLatest(err, originErr)            // true
errors.IsCauseOrLatest(err, errors.NotFound)      // true
errors.IsCauseOrLatest(err, errors.AlreadyExists) // false

err = errors.WithError(err, errors.AlreadyExists)
errors.Is(err, originErr)                 // true
errors.Is(err, errors.NotFound)           // true
errors.Is(err, errors.AlreadyExists)      // true

errors.IsCauseOrLatest(err, originErr)            // true
errors.IsCauseOrLatest(err, errors.NotFound)      // false
errors.IsCauseOrLatest(err, errors.AlreadyExists) // true

err = errors.Adapt(err, errors.Unknown)
errors.Is(err, originErr)                 // true
errors.Is(err, errors.NotFound)           // true
errors.Is(err, errors.AlreadyExists)      // true
errors.Is(err, errors.Unknown)            // false

errors.IsCauseOrLatest(err, originErr)            // true
errors.IsCauseOrLatest(err, errors.NotFound)      // false
errors.IsCauseOrLatest(err, errors.AlreadyExists) // true
errors.IsCauseOrLatest(err, errors.Unknown)       // false
```

- error collection

```go
err := errors.WithError(errors.New("e1"), errors.New("e2"))
err = errors.WithError(err, errors.New("e3"))
log.Println(errors.GetAllErrors(err))
// Output: [e1 e2 e3]
```

- error attr extraction

```go
err := errors.WithError(errors.New("xxx"), errors.NotFound)
err = errors.WithMessage(err, "wrapper1")
err = errors.WithCaller(err, "caller1")
err = errors.WithError(err, errors.AlreadyExists)
err = errors.WithMessage(err, "wrapper2")
err = errors.WithCaller(err, "caller2")
var ki int
err = errors.WithValue(err, &ki, "will not in map")
var ks = "ks"
err = errors.WithValue(err, &ks, "ks")

// errors.Map(err): get all values
// NOTE: error and *Meta will be flatten
m := errors.Map(err)
assert.Equalf(t, 10, len(m), "m length")
assert.Equalf(t, "myapp", m["meta.app"], "meta.app")
assert.Equalf(t, "github.com/ccmonky/errors", m["meta.source"], "meta.source")
assert.Equalf(t, "already_exists(6)", m["meta.code"], "meta.code")
assert.Equalf(t, "already exists", m["meta.message"], "meta.message")
assert.Equalf(t, "ks", m["ks"], "ks")
assert.Equalf(t, "wrapper2", m["msg"], "msg")
assert.Equalf(t, "caller2", m["caller"], "caller")
assert.Equalf(t, 409, m["status"], "status")
assert.Equalf(t, "meta={source=errors;code=already_exists(6)}:status={409}", fmt.Sprint(m["error"]), "error")
assert.Equalf(t, "source=errors;code=already_exists(6)", fmt.Sprint(m["meta"]), "meta")

// errors.Attrs: get values of specified attrs
// NOTE: error and *Meta will be flatten
m := errors.NewAttrs(errors.ErrorAttr, errors.MessageAttr).Map(err)
assert.Equalf(t, 8, len(m), "m length")
assert.Equalf(t, "myapp", m["meta.app"], "meta.app")
assert.Equalf(t, "github.com/ccmonky/errors", m["meta.source"], "meta.source")
assert.Equalf(t, "already_exists(6)", m["meta.code"], "meta.code")
assert.Equalf(t, "already exists", m["meta.message"], "meta.message")
assert.Equalf(t, "wrapper2", m["msg"], "msg")
assert.Equalf(t, 409, m["status"], "status")
assert.Equalf(t, "meta={source=errors;code=already_exists(6)}:status={409}", fmt.Sprint(m["error"]), "error")
assert.Equalf(t, "source=errors;code=already_exists(6)", fmt.Sprint(m["meta"]), "meta")
```

- error admin

```go
// list all registered meta errors(generated by `NewMetaError`)
mes, _ := errors.AllMetaErrors()
log.Println(string(data))

// list all registered error attrs(generated by `NewAttr`)
attrs, _ := errors.AllAttrs()
log.Println(string(data))
```

meta errors marshal result like this:

```json
{
    ":github.com/ccmonky/errors:aborted(10)": {
        "error": {
            "key": "meta",
            "value": {
                "meta.app": "myapp",
                "meta.code": "aborted(10)",
                "meta.message": "operation was aborted",
                "meta.source": "github.com/ccmonky/errors"
            }
        },
        "key": "status",
        "value": 409
    },
    ":github.com/ccmonky/errors:already_exists(6)": {
        "error": {
            "key": "meta",
            "value": {
                "meta.app": "myapp",
                "meta.code": "already_exists(6)",
                "meta.message": "already exists",
                "meta.source": "github.com/ccmonky/errors"
            }
        },
        "key": "status",
        "value": 409
    }
    // ...
}
```

attrs json marshal result like this:

```json
{
    "caller:0xc000110ed0": {
        "description": "caller as an attr",
        "has_default_value_func": false,
        "name": "caller",
        "type": "string"
    },
    "ctx:0xc000110dd0": {
        "description": "context.Context as an attr",
        "has_default_value_func": false,
        "name": "ctx",
        "type": "context.Context"
    },
    "error:0xc000110d70": {
        "description": "error as an attr",
        "has_default_value_func": false,
        "name": "error",
        "type": "error"
    },
    "meta:0xc000110e10": {
        "description": "meta as an attr",
        "has_default_value_func": false,
        "name": "meta",
        "type": "*errors.Meta"
    },
    "msg:0xc000110e50": {
        "description": "message as an attr",
        "has_default_value_func": false,
        "name": "msg",
        "type": "string"
    },
    "stack:0xc000110f10": {
        "description": "github.com/pkg/errors.stack as an attr",
        "has_default_value_func": false,
        "name": "stack",
        "type": "*errors.stack"
    },
    "status:0xc000110e90": {
        "description": "http status as an attr",
        "has_default_value_func": true,
        "name": "status",
        "type": "int"
    }
}
```

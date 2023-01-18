# errors

like pkg/errors, but can wrap any and retrieve

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

- error(meta) guard

```go
// NOTE: do notknow whether error has carried meta
err = errors.WithMetaNx(err, errors.Unknown)
```

- error force

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

errors.IsError(err, originErr)            // true
errors.IsError(err, errors.NotFound)      // true
errors.IsError(err, errors.AlreadyExists) // false

err = errors.WithError(err, errors.AlreadyExists)
errors.Is(err, originErr)                 // true
errors.Is(err, errors.NotFound)           // true
errors.Is(err, errors.AlreadyExists)      // true

errors.IsError(err, originErr)            // true
errors.IsError(err, errors.NotFound)      // false
errors.IsError(err, errors.AlreadyExists) // true

err = errors.WithMetaNx(err, errors.Unknown)
errors.Is(err, originErr)                 // true
errors.Is(err, errors.NotFound)           // true
errors.Is(err, errors.AlreadyExists)      // true
errors.Is(err, errors.Unknown)            // false

errors.IsError(err, originErr)            // true
errors.IsError(err, errors.NotFound)      // false
errors.IsError(err, errors.AlreadyExists) // true
errors.IsError(err, errors.Unknown)       // false
```

- error extraction

```go
...
```

- error list

```go
errors.MetaErrors()
```

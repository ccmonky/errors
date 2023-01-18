package errors_test

import (
	"context"
	"testing"

	"github.com/ccmonky/errors"
	"github.com/ccmonky/inithook"
	"github.com/stretchr/testify/assert"
)

func init() {
	err := inithook.ExecuteAttrSetters(context.Background(), inithook.AppName, "myapp")
	if err != nil {
		panic(err)
	}
}

func TestError(t *testing.T) {
	assert.Equalf(t, "meta={source=errors;code=not_found(5)}:status={404}", errors.NotFound.Error(), "notFound error")
	originErr := errors.Errorf("xxx")
	err := errors.WithError(originErr, errors.NotFound)
	err = errors.MessageAttr.With(err, "wrapper")
	err = errors.CallerAttr.With(err, "TestError")
	var kkk string
	ctx := context.WithValue(context.TODO(), &kkk, "vvv")
	err = errors.CtxAttr.With(err, ctx)
	assert.Equalf(t, "xxx:error={meta={source=errors;code=not_found(5)}:status={404}}:msg={wrapper}:caller={TestError}:ctx={context.TODO.WithValue(type *string, val vvv)}", err.Error(), "err chain")
	assert.Truef(t, errors.Is(err, originErr), "err is originErr")
	assert.Truef(t, errors.Is(err, errors.NotFound), "err is notfound")
	assert.Truef(t, !errors.Is(err, errors.AlreadyExists), "err is not alreadyexists")
	assert.Equalf(t, "wrapper", errors.Get(err, errors.MessageAttr.Key()), "err message")
	assert.Equalf(t, "TestError", errors.Get(err, errors.CallerAttr.Key()), "err caller")
	assert.Equalf(t, ctx, errors.Get(err, errors.CtxAttr.Key()), "ctx")
	assert.Equalf(t, "vvv", errors.Get(err, &kkk), "kkk value")
	meta := errors.MetaAttr.Get(err)
	assert.NotNilf(t, meta, "err meta")
	assert.Equalf(t, "github.com/ccmonky/errors", meta.Source(), "err meta source")
	assert.Equalf(t, "not_found(5)", meta.Code(), "err meta code")
	assert.Equalf(t, "not found", meta.Message(), "err meta message")
	assert.Equalf(t, 404, errors.StatusAttr.Get(err), "err status")
	assert.Equalf(t, "xxx", errors.Cause(err).Error(), "err cause")
	err = errors.WithError(err, errors.AlreadyExists)
	assert.Equalf(t, "xxx:error={meta={source=errors;code=not_found(5)}:status={404}}:msg={wrapper}:caller={TestError}:ctx={context.TODO.WithValue(type *string, val vvv)}:error={meta={source=errors;code=already_exists(6)}:status={409}}", err.Error(), "err with alreadyexists")
	assert.Truef(t, errors.Is(err, originErr), "err is originErr")
	assert.Truef(t, errors.Is(err, errors.NotFound), "err is notfound") // NOTE: also true
	assert.Truef(t, errors.Is(err, errors.AlreadyExists), "err is not alreadyexists")
	meta = errors.MetaAttr.Get(err)
	assert.NotNilf(t, meta, "err meta")
	assert.Equalf(t, "github.com/ccmonky/errors", meta.Source(), "err meta source 2")
	assert.Equalf(t, "already_exists(6)", meta.Code(), "err meta code 2")
	assert.Equalf(t, "already exists", meta.Message(), "err meta message 2")
	assert.Truef(t, errors.IsError(err, originErr), "err is originErr")
	assert.Truef(t, !errors.IsError(err, errors.NotFound), "err is not notfound")
	assert.Truef(t, errors.IsError(err, errors.AlreadyExists), "err is alreadyexists")
	assert.Truef(t, errors.Get(err, errors.ErrorAttr.Key()) == errors.AlreadyExists, "get err is alreadyexists")
	assert.Truef(t, errors.Get(err, errors.ErrorAttr.Key()) != errors.NotFound, "get err is not notfound")
	err = errors.WithMetaNx(err, errors.FailedPrecondition)
	assert.Equalf(t, "xxx:error={meta={source=errors;code=not_found(5)}:status={404}}:msg={wrapper}:caller={TestError}:ctx={context.TODO.WithValue(type *string, val vvv)}:error={meta={source=errors;code=already_exists(6)}:status={409}}:caller={errors_test.TestError}", err.Error(), "err with alreadyexists")
	assert.Truef(t, errors.Is(err, originErr), "err is originErr after with FailedPrecondition")
	assert.Truef(t, errors.Is(err, errors.NotFound), "err is notfound after with FailedPrecondition") // NOTE: also true
	assert.Truef(t, errors.Is(err, errors.AlreadyExists), "err is not alreadyexists after with FailedPrecondition")
	assert.Truef(t, !errors.Is(err, errors.FailedPrecondition), "err is not alreadyexists after with FailedPrecondition")
	assert.Truef(t, errors.IsError(err, originErr), "err is originErr after with FailedPrecondition")
	assert.Truef(t, !errors.IsError(err, errors.NotFound), "err is not notfound  after with FailedPrecondition")
	assert.Truef(t, errors.IsError(err, errors.AlreadyExists), "err is alreadyexists after with FailedPrecondition")
	assert.Truef(t, !errors.IsError(err, errors.FailedPrecondition), "err is not alreadyexists after with FailedPrecondition")
	assert.Truef(t, errors.Get(err, errors.ErrorAttr.Key()) == errors.AlreadyExists, "get err is alreadyexists after with FailedPrecondition")
	assert.Truef(t, errors.Get(err, errors.ErrorAttr.Key()) != errors.NotFound, "get err is not notfound after with FailedPrecondition")
	assert.Truef(t, errors.Get(err, errors.ErrorAttr.Key()) != errors.FailedPrecondition, "get err is not notfound after with FailedPrecondition")
}

func BenchmarkError(b *testing.B) {
	err := errors.WithError(errors.New("xxx"), errors.NotFound)
	err = errors.WithMessage(err, "wrapper1")
	err = errors.WithCaller(err, "caller1")
	err = errors.WithMessage(err, "wrapper2")
	var k string
	err = errors.WithCtx(err, context.WithValue(context.TODO(), &k, "v"))
	err = errors.WithMessage(err, "wrapper3")
	err = errors.WithStatus(err, 409)
	err = errors.WithMessage(err, "wrapper4")
	err = errors.WithValue(err, "k4", "v4")
	err = errors.WithMessage(err, "wrapper5")
	err = errors.WithValue(err, "k5", "v5")
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		// goos: darwin
		// goarch: amd64
		// pkg: github.com/ccmonky/errors
		// cpu: Intel(R) Core(TM) i7-9750H CPU @ 2.60GHz
		// BenchmarkError-12    	 5816956	       204.1 ns/op	       0 B/op	       0 allocs/op
		errors.MetaAttr.Get(err)

		// goos: darwin
		// goarch: amd64
		// pkg: github.com/ccmonky/errors
		// cpu: Intel(R) Core(TM) i7-9750H CPU @ 2.60GHz
		// BenchmarkError-12    	45875535	        29.66 ns/op	       0 B/op	       0 allocs/op
		//errors.MessageAttr.Get(err)

		// goos: darwin
		// goarch: amd64
		// pkg: github.com/ccmonky/errors
		// cpu: Intel(R) Core(TM) i7-9750H CPU @ 2.60GHz
		// BenchmarkError-12    	 9519049	       123.2 ns/op	       0 B/op	       0 allocs/op
		//errors.Cause(err)
	}
}

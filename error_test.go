package errors_test

import (
	"context"
	"fmt"
	"log"
	"strings"
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
	err = errors.Adapt(err, errors.FailedPrecondition)
	assert.Equalf(t, "xxx:error={meta={source=errors;code=not_found(5)}:status={404}}:msg={wrapper}:caller={TestError}:ctx={context.TODO.WithValue(type *string, val vvv)}:error={meta={source=errors;code=already_exists(6)}:status={409}}:caller={errors_test.TestError:60}", err.Error(), "err with alreadyexists")
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

func TestValues(t *testing.T) {
	err := errors.WithError(errors.New("xxx"), errors.NotFound)
	err = errors.WithMessage(err, "wrapper1")
	err = errors.WithCaller(err, "caller1")
	err = errors.WithError(err, errors.AlreadyExists)
	err = errors.WithMessage(err, "wrapper2")
	err = errors.WithCaller(err, "caller2")
	var k string
	err = errors.WithCtx(err, context.WithValue(context.TODO(), &k, "v1"))
	err = errors.WithMessage(err, "wrapper3")
	err = errors.WithCtx(err, context.WithValue(context.TODO(), &k, "v2"))
	assert.Equalf(t, "meta={source=errors;code=not_found(5)}:status={404}|meta={source=errors;code=already_exists(6)}:status={409}", join(errors.ErrorAttr.GetAll(err)...), "all errors")
	assert.Equalf(t, "source=errors;code=not_found(5)|source=errors;code=already_exists(6)", join(errors.MetaAttr.GetAll(err)...), "all metas")
	assert.Equalf(t, "wrapper1|wrapper2|wrapper3", join(errors.MessageAttr.GetAll(err)...), "all messages")
	assert.Equalf(t, "caller1|caller2", join(errors.CallerAttr.GetAll(err)...), "all callers")
	assert.Equalf(t, "v1|v2", join(errors.GetAll(err, &k)...), "all kvs")
}

func join[T any](values ...T) string {
	var ss []string
	for _, v := range values {
		ss = append(ss, fmt.Sprintf("%v", v))
	}
	return strings.Join(ss, "|")
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
		// BenchmarkError-12    	45875535	        29.66 ns/op	       0 B/op	       0 allocs/op
		//errors.MessageAttr.Get(err)

		// goos: darwin
		// goarch: amd64
		// pkg: github.com/ccmonky/errors
		// cpu: Intel(R) Core(TM) i7-9750H CPU @ 2.60GHz
		// BenchmarkError-12    	 9519049	       123.2 ns/op	       0 B/op	       0 allocs/op
		//errors.Cause(err)

		// goos: darwin
		// goarch: amd64
		// pkg: github.com/ccmonky/errors
		// cpu: Intel(R) Core(TM) i7-9750H CPU @ 2.60GHz
		// BenchmarkError-12    	 5816956	       204.1 ns/op	       0 B/op	       0 allocs/op
		//errors.MetaAttr.Get(err)

		// goos: darwin
		// goarch: amd64
		// pkg: github.com/ccmonky/errors
		// cpu: Intel(R) Core(TM) i7-9750H CPU @ 2.60GHz
		// BenchmarkError-12    	 1795410	       633.0 ns/op	      48 B/op	       3 allocs/op
		errors.ErrorAttr.GetAll(err)

		// goos: darwin
		// goarch: amd64
		// pkg: github.com/ccmonky/errors
		// cpu: Intel(R) Core(TM) i7-9750H CPU @ 2.60GHz
		// BenchmarkError-12    	  965583	      1228 ns/op	     560 B/op	       9 allocs/op
		//errors.MessageAttr.GetAll(err)
	}
}

func BenchmarkErrorAggregation(b *testing.B) {
	err := errors.WithError(errors.New("e1"), errors.New("e2"))
	err = errors.WithError(err, errors.New("e3"))
	err = errors.WithError(err, errors.New("e4"))
	err = errors.WithError(err, errors.New("e5"))
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		// goos: darwin
		// goarch: amd64
		// pkg: github.com/ccmonky/errors
		// cpu: Intel(R) Core(TM) i7-9750H CPU @ 2.60GHz
		// BenchmarkErrorAggregation-12    	 1597552	       726.4 ns/op	     288 B/op	       7 allocs/op
		errors.ErrorAttr.GetAll(err)
	}
}

func BenchmarkContext(b *testing.B) {
	ctx := context.WithValue(context.Background(), "k1", "v1")
	ctx = context.WithValue(ctx, "k2", "v2")
	ctx = context.WithValue(ctx, "k3", "v3")
	ctx = context.WithValue(ctx, "k4", "v4")
	ctx = context.WithValue(ctx, "k5", "v5")
	ctx = context.WithValue(ctx, "k6", "v6")
	ctx = context.WithValue(ctx, "k7", "v7")
	ctx = context.WithValue(ctx, "k8", "v8")
	ctx = context.WithValue(ctx, "k9", "v9")
	ctx = context.WithValue(ctx, "k10", "v10")
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		// goos: darwin
		// goarch: amd64
		// pkg: github.com/ccmonky/errors
		// cpu: Intel(R) Core(TM) i7-9750H CPU @ 2.60GHz
		// BenchmarkContext-12    	75000534	        17.96 ns/op	       0 B/op	       0 allocs/op
		ctx.Value("k9")

		// goos: darwin
		// goarch: amd64
		// pkg: github.com/ccmonky/errors
		// cpu: Intel(R) Core(TM) i7-9750H CPU @ 2.60GHz
		// BenchmarkContext-12    	10423622	       118.1 ns/op	       0 B/op	       0 allocs/op
		//ctx.Value("k1")
	}
}

func TestX(t *testing.T) {
	err := errors.WithError(errors.New("e1"), errors.New("e2"))
	err = errors.WithError(err, errors.New("e3"))
	log.Println(errors.GetAllErrors(err))
	t.Fatal(1)
}

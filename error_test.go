package errors_test

import (
	"context"
	"testing"

	"github.com/ccmonky/errors"
	"github.com/stretchr/testify/assert"
)

func TestContextError(t *testing.T) {
	assert.Equalf(t, "meta={source=errors;code=not_found(5)}:status={404}", errors.NotFound.Error(), "notFound error")
	originErr := errors.Errorf("xxx")
	err := errors.WithError(originErr, errors.NotFound)
	err = errors.MessageAttr.With(err, "wrapper")
	err = errors.CallerAttr.With(err, "TestContextError")
	var kkk string
	err = errors.CtxAttr.With(err, context.WithValue(context.TODO(), &kkk, "vvv"))
	assert.Equalf(t, "xxx:error={meta={source=errors;code=not_found(5)}:status={404}}:msg={wrapper}:caller={TestContextError}:ctx={context.TODO.WithValue(type *string, val vvv)}", err.Error(), "err chain")
	assert.Truef(t, errors.Is(err, originErr), "err is originErr")
	assert.Truef(t, errors.Is(err, errors.NotFound), "err is notfound")
	assert.Truef(t, !errors.Is(err, errors.AlreadyExists), "err is not alreadyexists")
	assert.Equalf(t, "wrapper", errors.Get(err, errors.MessageAttr.Key()), "err message")
	assert.Equalf(t, "TestContextError", errors.Get(err, errors.CallerAttr.Key()), "err caller")
	assert.Equalf(t, "vvv", errors.Get(err, &kkk), "kkk value")
	meta := errors.MetaAttr.Get(err)
	assert.NotNilf(t, meta, "err meta")
	assert.Equalf(t, "github.com/ccmonky/errors", meta.Source(), "err meta source")
	assert.Equalf(t, "not_found(5)", meta.Code(), "err meta code")
	assert.Equalf(t, "not found", meta.Message(), "err meta message")
	assert.Equalf(t, 404, errors.StatusAttr.Get(err), "err status")
	assert.Equalf(t, "xxx", errors.Cause(err).Error(), "err cause")
	err = errors.WithError(err, errors.AlreadyExists)
	assert.Equalf(t, "xxx:error={meta={source=errors;code=not_found(5)}:status={404}}:msg={wrapper}:caller={TestContextError}:ctx={context.TODO.WithValue(type *string, val vvv)}:error={meta={source=errors;code=already_exists(6)}:status={409}}", err.Error(), "err with alreadyexists")
	assert.Truef(t, errors.Is(err, originErr), "err is originErr")
	assert.Truef(t, errors.Is(err, errors.NotFound), "err is notfound") // NOTE: also true
	assert.Truef(t, errors.Is(err, errors.AlreadyExists), "err is not alreadyexists")
	meta = errors.MetaAttr.Get(err)
	assert.NotNilf(t, meta, "err meta")
	assert.Equalf(t, "github.com/ccmonky/errors", meta.Source(), "err meta source 2")
	assert.Equalf(t, "already_exists(6)", meta.Code(), "err meta code 2")
	assert.Equalf(t, "already exists", meta.Message(), "err meta message 2")
	assert.Truef(t, errors.IsMetaError(err, originErr), "err is originErr")
	assert.Truef(t, !errors.IsMetaError(err, errors.NotFound), "err is notfound") // NOTE: also true
	assert.Truef(t, errors.IsMetaError(err, errors.AlreadyExists), "err is not alreadyexists")
}

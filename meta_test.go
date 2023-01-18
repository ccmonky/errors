package errors_test

import (
	"testing"

	"github.com/ccmonky/errors"
	"github.com/stretchr/testify/assert"
)

func TestNewMetaError(t *testing.T) {
	assert.Panicsf(t, func() { errors.NewMetaError("", "", "") }, "empty code")
	assert.Panicsf(t, func() { errors.NewMetaError("1", "", "") }, "empty source")
	assert.Panicsf(t, func() { errors.NewMetaError("1", "1", "") }, "empty message")

	me := errors.NewMetaError("1", "1", "1")
	assert.Equalf(t, "myapp", me.App(), "app 1 new")
	assert.Equalf(t, "1", me.Source(), "source 1 new")
	assert.Equalf(t, "1", me.Code(), "code 1 new")
	assert.Equalf(t, "1", me.Message(), "msg 1 new")
	for k, me := range errors.MetaErrors() {
		t.Log(k, me)
	}

	me = errors.GetMetaError("myapp:1:1")
	assert.Equalf(t, "myapp", me.App(), "app 1 get")
	assert.Equalf(t, "1", me.Source(), "source 1 get")
	assert.Equalf(t, "1", me.Code(), "code 1 get")
	assert.Equalf(t, "1", me.Message(), "msg 1 get")

	me = errors.GetMetaError("myapp:github.com/ccmonky/errors:not_found(5)")
	assert.Equalf(t, "myapp", me.App(), "notfound app")
	assert.Equalf(t, "github.com/ccmonky/errors", me.Source(), "notfound source")
	assert.Equalf(t, "not_found(5)", me.Code(), "notfound code")
	assert.Equalf(t, "not found", me.Message(), "not found msg")
}

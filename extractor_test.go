package errors_test

import (
	"fmt"
	"testing"

	"github.com/ccmonky/errors"
	"github.com/stretchr/testify/assert"
)

func TestAttrs(t *testing.T) {
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
}

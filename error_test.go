package errors_test

import (
	"fmt"
	"testing"

	"github.com/ccmonky/errors"
)

func TestContextError(t *testing.T) {
	ce := errors.NewMetaError("errors_test", "1", "unknown", errors.StatusAttr.Option(500))
	ce = errors.MessageAttr.With(ce, "wrapper")
	t.Log(ce)
	nf := errors.NewMetaError("errors_test", "2", "not found", errors.StatusAttr.Option(404))
	t.Log(nf)
	ce = errors.WithError(ce, nf)
	t.Log(ce)
	err := errors.With(fmt.Errorf("xxx"), errors.Options(nf)...)
	err = errors.MessageAttr.With(err, "wrapper")
	err = errors.CallerAttr.With(err, "TestContextError")
	t.Log(err)
	t.Log(errors.MetaAttr.Get(err), errors.Is(err, nf))
	t.Fatal(1)
}

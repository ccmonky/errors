package errors_test

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/ccmonky/errors"
	"github.com/stretchr/testify/assert"
)

func TestAttr(t *testing.T) {
	data, err := json.Marshal(errors.StatusAttr)
	assert.Nilf(t, err, "marshal SattusAttr")
	assert.JSONEq(t, `{"description":"http status as an attr","has_default_value_func":true,"name":"status","type":"int"}`, string(data), "status json")
	data, err = json.Marshal(errors.MessageAttr)
	assert.Nilf(t, err, "marshal MessageAttr")
	assert.JSONEq(t, `{"description":"message as an attr","has_default_value_func":false,"name":"msg","type":"string"}`, string(data), "message json")
	data, err = json.Marshal(errors.MetaAttr)
	assert.Nilf(t, err, "marshal MetaAttr")
	assert.JSONEq(t, `{"description":"meta as an attr","has_default_value_func":false,"name":"meta","type":"*errors.Meta"}`, string(data), "meta json")
	data, err = json.Marshal(errors.ErrorAttr)
	assert.Nilf(t, err, "marshal ErrorAttr")
	assert.JSONEq(t, `{"description":"error as an attr","has_default_value_func":false,"name":"error","type":"error"}`, string(data), "error json")
	_, err = json.Marshal(errors.AllAttrs())
	assert.Nilf(t, err, "marshal all attrs")
	//t.Log(string(data))
}

func TestNewAttrKey(t *testing.T) {
	var err = errors.New("xxx")
	callerKey1 := errors.NewAttrKey("caller")
	err = errors.WithValue(err, callerKey1, "caller1")
	callerKey2 := errors.NewAttrKey("caller")
	err = errors.WithValue(err, callerKey2, "caller2")
	err = errors.WithValue(err, callerKey1, "caller3")
	assert.Equalf(t, "caller3", errors.Get(err, callerKey1), "caller key 1")
	assert.Equalf(t, "caller2", errors.Get(err, callerKey2), "caller key 2")
}

func TestGetAttr(t *testing.T) {
	assert.Equalf(t, "caller", errors.MustGetAttrByName[string]("caller").Name(), "caller name")
	assert.Equalf(t, "status", errors.MustGetAttrByName[int]("status").Name(), "status name")
	_, err := errors.GetAttrByName[*int]("status")
	assert.Truef(t, nil != err, "status with bad type: %v", err)
	_, err = errors.GetAttrByName[int]("_status")
	assert.Truef(t, nil != err, "status with bad name: %v", err)
}

func TestStatus(t *testing.T) {
	assert.Equalf(t, 200, errors.StatusAttr.Get(nil), "err nil")
	err := errors.WithMessage(errors.New("xxx"), "wrapper")
	assert.Equalf(t, 500, errors.StatusAttr.Get(err), "err no status")
	err = errors.WithError(err, errors.NotFound)
	assert.Equalf(t, 404, errors.StatusAttr.Get(err), "err with status")
}

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

func BenchmarkNewAttrStrKey(b *testing.B) {
	b.ReportAllocs()
	meta := errors.NotFound
	ptrKey := errors.NewAttrKey("bench")
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			meta.(errors.Error).Value(ptrKey)
		}
	})
}

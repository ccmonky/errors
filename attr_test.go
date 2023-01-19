package errors_test

import (
	"encoding/json"
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

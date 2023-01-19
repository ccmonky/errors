package errors_test

import (
	"testing"

	"github.com/ccmonky/errors"
	"github.com/stretchr/testify/assert"
)

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

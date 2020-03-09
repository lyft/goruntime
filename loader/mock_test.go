package loader

import (
	"testing"
)

func intValue(t *testing.T, x interface{}) uint64 {
	switch v := x.(type) {
	case int:
		return uint64(v)
	case int8:
		return uint64(v)
	case int16:
		return uint64(v)
	case int32:
		return uint64(v)
	case int64:
		return uint64(v)
	case uint:
		return uint64(v)
	case uint8:
		return uint64(v)
	case uint16:
		return uint64(v)
	case uint32:
		return uint64(v)
	case uint64:
		return uint64(v)
	default:
		t.Fatalf("invalid int type: %[1]T - %#[1]v", x)
	}
	panic("unreachable")
}

func TestMockWithValues(t *testing.T) {
	kvs := map[string]interface{}{
		"int":    int(1),
		"int8":   int8(1),
		"int16":  int16(1),
		"int32":  int32(1),
		"int64":  int64(1),
		"uint":   uint(1),
		"uint8":  uint8(1),
		"uint16": uint16(1),
		"uint32": uint32(1),
		"uint64": uint64(1),
		"string": "string",
	}
	var a []interface{}
	for k, v := range kvs {
		a = append(a, k, v)
	}
	opt := MockWithValues(a...)

	ents := NewMock(opt).Snapshot().Entries()
	for key, e := range kvs {
		if key == "string" {
			val := ents[key].StringValue
			exp := e.(string)
			if val != exp {
				t.Errorf("%s: got: %#v want: %#v", key, val, exp)
			}
			continue
		}
		exp := intValue(t, e)
		val := ents[key].Uint64Value
		if val != exp {
			t.Errorf("%s: got: %#v want: %#v", key, val, exp)
		}
	}
}

func TestMockWithValues_Invalid(t *testing.T) {
	testPanic := func(t *testing.T, kvs ...interface{}) {
		t.Helper()
		defer func() {
			if recover() == nil {
				t.Errorf("panic: got: %t want: %t kvs: %v", false, true, kvs)
			}
		}()
		MockWithValues(kvs)
	}

	testPanic(t, "key")
	testPanic(t, 1, 1)
	testPanic(t, "k", 1, "a")
	testPanic(t, "k", complex64(1))
	testPanic(t, "k", "v")
}

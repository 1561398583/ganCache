package cache3

import (
	"github.com/pkg/errors"
	"testing"
)

func TestGetterFunc_Get(t *testing.T) {
	m := make(map[string]string)
	m["key1"] = "value1"
	m["key2"] = "value2"
	finder := func(key string) ([]byte, error){
		v, ok := m[key]
		if !ok {
			return nil, errors.New("not found")
		}
		return []byte(v), nil
	}
	group := NewCacheWithGetter(2<<10, FindFunc(finder))
	bd1, err := group.Get("key1")
	if err != nil || string(bd1) != "value1" {
		t.Errorf("getter failed")
	}
	//“key1”第一次已经存入cache，第二次就应该从cache中取，不该回调getter
	m["key1"] = "no"
	bd2, err := group.Get("key1")
	if err != nil || string(bd2) != "value1" {
		t.Errorf("expect bd.b : value1 , but %s", string(bd2))
	}
	_, err = group.Get("key3")
	if err == nil {
		t.Errorf("expect err != nil but not")
	}
}


package mcpack

import (
	"testing"

	json "github.com/json-iterator/go"
)

func TestDecode(t *testing.T) {
	equalFun := func(in, out interface{}) (equal bool) {
		resp, err := json.Marshal(in)
		if err != nil {
			return false
		}
		resp2, err := json.Marshal(out)
		if err != nil {
			return false
		}
		if string(resp) != string(resp2) {
			return false
		}
		return true
	}

	tests := []struct {
		name string
		in   interface{}
	}{
		{
			name: "string",
			in: map[string]interface{}{
				"username": "js",
				"age":      "10",
			},
		},
		{
			name: "int",
			in: map[string]interface{}{
				"username": 1,
				"age":      2,
			},
		},
		{
			name: "int-multi",
			in: map[string]interface{}{
				"username": 1,
				"age":      2,
				"testjson": map[string]interface {
				}{
					"a": 1,
					"b": 2,
				},
			},
		},
		{
			name: "multi",
			in: map[string]interface{}{
				"username": "js",
				"age":      10,
				"testjson": map[string]interface {
				}{
					"a": "1",
					"b": 2,
				},
			},
		},
	}
	for _, test := range tests {
		resp, err := Marshal(test.in)
		if err != nil {
			t.Error("expect nil , got error: ", err.Error())
			continue
		}
		out, err := Decode(resp)
		if err != nil {
			t.Error("expect nil , got error: ", err.Error())
			continue
		}
		if false == equalFun(test.in, out) {
			t.Errorf("not equal, got: %+v want: %+v", out, test.in)
		}
	}
}

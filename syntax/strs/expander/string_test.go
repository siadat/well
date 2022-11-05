package expander_test

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/siadat/well/syntax/strs/expander"
)

func TestEncodeToString(tt *testing.T) {
	var testCases = []struct {
		src    string
		want   string
		err    string
		values map[string]interface{}
	}{
		{
			src:  `ls  -lash --directory -C ./something`,
			want: `ls  -lash --directory -C ./something`,
		},
		{
			src:  `actual \«double\» and \‹single\› guillemets and backslashes \ \\ \`,
			want: `actual «double» and ‹single› guillemets and backslashes \ \\ \`,
		},
		{
			src: `unclosed open «guillemet`,
			err: `unclosed container`,
		},
		{
			src: `double guillemet ‹closed››`,
			err: `unexpected token RSINGLE_GUILLEMET("›")`,
		},
		{
			src: `double quote "closed""`,
			err: `unclosed container`,
		},
		{
			src:    `echo "Hello ${name}!"`,
			want:   `echo "Hello sina!"`,
			values: map[string]interface{}{"name": "sina"},
		},
		{
			src:    `jq «.${key:%q} | .»`,
			want:   `jq ".\"a long key\" | ."`,
			values: map[string]interface{}{"key": "a long key"},
		},
		{
			src:    `hello ${key}`,
			err:    `variable key is <nil>`,
			values: map[string]interface{}{},
		},
		{
			src:  `hello {key}`,
			want: `hello {key}`, // allow raw { and }
		},
		{
			src:    `abc ${key:%Q}`,
			want:   `abc $'a long key'`,
			values: map[string]interface{}{"key": "a long key"},
		},
		{
			src:    `jq «.«${key}» | .»`,
			want:   `jq ".\"a long key\" | ."`,
			values: map[string]interface{}{"key": "a long key"},
		},
		{
			src:  `abc « «1» ««2»» »`,
			want: `abc " \"1\" \"\\\"2\\\"\" "`,
		},
		{
			src:  `abc « «1» ««2»» »`,
			want: `abc " \"1\" \"\\\"2\\\"\" "`,
		},
		{
			src:    `echo ‹echo ‹echo ‹${name}›››`,
			want:   `echo $'echo $\'echo $\\\'O\\\\\\\'Reilly\\\'\''`,
			values: map[string]interface{}{"name": "O'Reilly"},
		},
	}

	for _, tc := range testCases {
		var src = tc.src
		var got, err = expander.ParseAndEncodeToString(src, expander.MappingFuncFromMap(tc.values), true)
		if tc.err == "" {
			if err != nil {
				tt.Fatalf("expected no error, got: %v", err)
			}
		} else {
			if err == nil || err.Error() != tc.err {
				tt.Fatalf("expected error %q, got: %v", tc.err, err)
			}
		}
		if diff := cmp.Diff(tc.want, got); diff != "" {
			tt.Fatalf("case failed src=%q (-want +got):\n%s", src, diff)
		}
	}
}

package exec_test

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/siadat/well/exec"
)

func TestEncodeToString(tt *testing.T) {
	var testCases = []struct {
		src    string
		want   string
		values map[string]interface{}
	}{
		{
			src:  `ls  -lash --directory -C ./something`,
			want: `ls  -lash --directory -C ./something`,
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
		var got = exec.EncodeToString(src, exec.MappingFuncFromMap(tc.values))
		if diff := cmp.Diff(tc.want, got); diff != "" {
			tt.Fatalf("case failed src=%q (-want +got):\n%s", src, diff)
		}
	}
}

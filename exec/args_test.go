package exec_test

import (
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/siadat/well/exec"
	"github.com/siadat/well/parser"
)

func TestEncodeToCmdArgs(tt *testing.T) {
	var testCases = []struct {
		src    string
		want   []string
		values map[string]interface{}
	}{
		{
			src:  `ls  -lash --directory -C ./something`,
			want: []string{"ls", "-lash", "--directory", "-C", "./something"},
		},
		{
			src:    `echo "Hello ${name}!"`,
			want:   []string{"echo", "Hello sina!"},
			values: map[string]interface{}{"name": "sina"},
		},
		{
			src:    `jq «.${key:%q} | .»`,
			want:   []string{"jq", `."a long key" | .`},
			values: map[string]interface{}{"key": "a long key"},
		},
		{
			src:    `jq «.«${key}» | .»`,
			want:   []string{"jq", `."a long key" | .`},
			values: map[string]interface{}{"key": "a long key"},
		},
		{
			src:    `jq .« ${key} »`, // The '.' is outside «»
			want:   []string{"jq", `. a long key `},
			values: map[string]interface{}{"key": "a long key"},
		},
		{
			src:  `abc « «1» ««2»» »`,
			want: []string{"abc", ` "1" "\"2\"" `},
		},
		{
			src:  `a «"b"»`,
			want: []string{"a", `"b"`},
		},
		{
			src:  `a '"b"'`,
			want: []string{"a", `"b"`},
		},
		// {
		// 	src:    `${fileA} '"${fileB}"'`, // and without <">
		// 	want:   []string{"file A", "file B"},
		// 	values: map[string]interface{}{"fileA": `file A`, "fileB": `file B`},
		// },
		// {
		// 	src:    `${fileA} ${fileB}`,
		// 	want:   []string{"file", "A", "file", "B"},
		// 	values: map[string]interface{}{"fileA": `file A`, "fileB": `file B`},
		// },
		// {
		// 	src:    `${file1} ${file2}`,
		// 	want:   []string{"file", "1", "file", "2"},
		// 	values: map[string]interface{}{"file1": `file 1`, "file2": `file 2`},
		// },
	}

	for _, tc := range testCases {
		var p = parser.NewParser()
		var src = tc.src
		var node, err = p.Parse(strings.NewReader(src))
		if err != nil {
			tt.Fatalf("test case failed src=%#q: %v", src, err)
		}
		var got = exec.EncodeToCmdArgs(node, exec.MappingFuncFromMap(tc.values))
		if diff := cmp.Diff(tc.want, got); diff != "" {
			tt.Fatalf("case failed src=%#q (-want +got):\n%s", src, diff)
		}
	}
}

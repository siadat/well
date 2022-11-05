package expander_test

import (
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/siadat/well/syntax/strs/expander"
	"github.com/siadat/well/syntax/strs/parser"
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
			// new lines and an empty line
			src: `
			ls  -lash
			    --directory
				-C "\n"
			`,
			want: []string{"ls", "-lash", "--directory", "-C", `\n`},
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
		{
			src:    `echo ${file_1} ${file_2}`,
			want:   []string{"echo", "file_1", "file_2"},
			values: map[string]interface{}{"file_1": "file_1", "file_2": `file_2`},
		},
		{
			src:    `echo ${file_1}${file_2}`,
			want:   []string{"echo", "file", "Afile", "B"},
			values: map[string]interface{}{"file_1": `file A`, "file_2": `file B`},
		},
		{
			src:    `echo ${file_1:%-} ${file_2:%-}`,
			want:   []string{"echo", "file", "A", "file", "B"},
			values: map[string]interface{}{"file_1": `file A`, "file_2": `file B`},
		},
	}

	for _, tc := range testCases {
		var p = parser.NewParser()
		var src = tc.src
		var node, err = p.Parse(strings.NewReader(src))
		if err != nil {
			tt.Fatalf("test case failed src=%#q:\nvalues=%#q\nerror: %v", src, tc.values, err)
		}
		var got, encodeErr = expander.EncodeToCmdArgs(node, expander.MappingFuncFromMap(tc.values))
		if err != nil {
			tt.Fatalf("test case failed src=%#q:\nvalues=%#q\nerror: %v", src, tc.values, encodeErr)
		}
		if diff := cmp.Diff(tc.want, got); diff != "" {
			tt.Fatalf("case failed src=%#q\nvalues=%#q\ndiff (-want +got):\n%s", src, tc.values, diff)
		}
	}
}

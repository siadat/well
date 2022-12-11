package expander_test

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/siadat/well/syntax/strs/expander"
)

func TestGetVariables(tt *testing.T) {
	var testCases = []struct {
		src  string
		want []expander.Variable
		err  string
	}{
		{
			src: `echo "Hello ${your_name}!"`,
			want: []expander.Variable{
				{"your_name", "string"},
			},
		},
	}

	for _, tc := range testCases {
		var src = tc.src
		var got, err = expander.GetVariables(src)
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

package diff

import (
	"reflect"
	"testing"
)

func TestLines2(t *testing.T) {
	type args struct {
		got  string
		want string
	}
	tests := []struct {
		name string
		args args
		want Diff
	}{
		{
			name: "no changes",
			args: args{
				got:  "a\nb\nc\nd",
				want: "a\nb\nc\nd",
			},
			want: Diff{
				lines: []line{
					{number: 1, content: "a", status: match},
					{number: 2, content: "b", status: match},
					{number: 3, content: "c", status: match},
					{number: 4, content: "d", status: match},
				},
			},
		},
		{
			name: "line removed",
			args: args{
				got:  "a\nb\nd",
				want: "a\nb\nc\nd",
			},
			want: Diff{
				lines: []line{
					{number: 1, content: "a", status: match},
					{number: 2, content: "b", status: match},
					{number: 3, content: "c", status: removed},
					{number: 3, content: "d", status: match},
				},
			},
		},
		{
			name: "lines removed",
			args: args{
				got:  "a\ne",
				want: "a\nb\nc\nd\ne",
			},
			want: Diff{
				lines: []line{
					{number: 1, content: "a", status: match},
					{number: 2, content: "b", status: removed},
					{number: 3, content: "c", status: removed},
					{number: 4, content: "d", status: removed},
					{number: 2, content: "e", status: match},
				},
			},
		},
		{
			name: "end lines removed",
			args: args{
				got:  "a",
				want: "a\nb\nc\nd",
			},
			want: Diff{
				lines: []line{
					{number: 1, content: "a", status: match},
					{number: 2, content: "b", status: removed},
					{number: 3, content: "c", status: removed},
					{number: 4, content: "d", status: removed},
				},
			},
		},
		{
			name: "line added",
			args: args{
				got:  "a\nb\nf\nc\nd",
				want: "a\nb\nc\nd",
			},
			want: Diff{
				lines: []line{
					{number: 1, content: "a", status: match},
					{number: 2, content: "b", status: match},
					{number: 3, content: "f", status: added},
					{number: 4, content: "c", status: match},
					{number: 5, content: "d", status: match},
				},
			},
		},
		{
			name: "lines added",
			args: args{
				got:  "a\nb\nf\ng\nh\nc\nd",
				want: "a\nb\nc\nd",
			},
			want: Diff{
				lines: []line{
					{number: 1, content: "a", status: match},
					{number: 2, content: "b", status: match},
					{number: 3, content: "f", status: added},
					{number: 4, content: "g", status: added},
					{number: 5, content: "h", status: added},
					{number: 6, content: "c", status: match},
					{number: 7, content: "d", status: match},
				},
			},
		},
		{
			name: "end lines added",
			args: args{
				got:  "a\nb\nc\nd\ne\nf",
				want: "a\nb\nc\nd",
			},
			want: Diff{
				lines: []line{
					{number: 1, content: "a", status: match},
					{number: 2, content: "b", status: match},
					{number: 3, content: "c", status: match},
					{number: 4, content: "d", status: match},
					{number: 5, content: "e", status: added},
					{number: 6, content: "f", status: added},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewDiff(tt.args.got, tt.args.want)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Lines2() = %v, want %v", got, tt.want)
			}
		})
	}
}

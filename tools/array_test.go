package tools

import (
	"reflect"
	"testing"
)

func TestUnion(t *testing.T) {
	type args struct {
		s1 []string
		s2 []string
	}
	tests := []struct {
		name string
		args args
		want []string
	}{
		// TODO: Add test cases.
		{
			name: "empty",
			args: args{
				s1: []string{},
				s2: []string{},
			},
			want: []string{},
		},
		{
			name: "one",
			args: args{
				s1: []string{"a"},
				s2: []string{"b"},
			},
			want: []string{"a", "b"},
		},
		{
			name: "two",
			args: args{
				s1: []string{"a"},
				s2: []string{"b", "c"},
			},
			want: []string{"a", "b", "c"},
		},
		{
			name: "three",
			args: args{
				s1: []string{},
				s2: []string{"b", "c"},
			},

			want: []string{"b", "c"},
		},
		{
			name: "four",
			args: args{
				s1: []string{"a"},
				s2: []string{},
			},
			want: []string{"a"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Union(tt.args.s1, tt.args.s2); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Union() = %v, want %v", got, tt.want)
			}
		})
	}
}

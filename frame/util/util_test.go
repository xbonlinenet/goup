package util

import (
	"testing"
)

func TestVersionCompare(t *testing.T) {
	type args struct {
		v1 string
		v2 string
	}
	tests := []struct {
		name string
		args args
		want int
	}{
		{
			name: "default",
			args: args{
				v1: "3.1.9",
				v2: "3.1.2",
			},
			want: 1,
		},
		{
			name: "not valid version",
			args: args{
				v1: "3.1.9.8",
				v2: "3.1.2",
			},
			want: 1,
		},
		{
			name: "equal",
			args: args{
				v1: "3.1.2",
				v2: "3.1.2",
			},
			want: 0,
		},
		{
			name: "less",
			args: args{
				v1: "2.1.2",
				v2: "3.1.2",
			},
			want: -1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := VersionCompare(tt.args.v1, tt.args.v2); got != tt.want {
				t.Errorf("VersionCompare() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsStringArrayEqual(t *testing.T) {
	type args struct {
		sorteda []string
		sortedb []string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "equeal",
			args: args{
				sorteda: []string{"a", "b"},
				sortedb: []string{"a", "b"},
			},
			want: true,
		},
		{
			name: "element size not equal",
			args: args{
				sorteda: []string{"a", "b"},
				sortedb: []string{"a"},
			},
			want: false,
		},
		{
			name: "element not equal",
			args: args{
				sorteda: []string{"a", "b"},
				sortedb: []string{"a", "c"},
			},
			want: false,
		},
		{
			name: "empty array",
			args: args{
				sorteda: []string{"a", "b"},
				sortedb: []string{},
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsStringArrayEqual(tt.args.sorteda, tt.args.sortedb); got != tt.want {
				t.Errorf("IsStringArrayEqual() = %v, want %v", got, tt.want)
			}
		})
	}
}

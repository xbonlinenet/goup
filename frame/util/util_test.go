package util

import "testing"

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

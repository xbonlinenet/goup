package util

import (
	"testing"
)

func TestGetDayOfMonth(t *testing.T) {
	tests := []struct {
		name string
		want string
	}{
		// TODO: Add test cases.
		{
			name: "default",
			want: "13",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetDayOfMonth(); got != tt.want {
				t.Errorf("GetDayOfMonth() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetVNDateFromUTC(t *testing.T) {
	type args struct {
		unixSeconds int64
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "default",
			args: args{
				unixSeconds: 1531298043,
			},
			want: "20180711",
		},
		{
			name: "miss",
			args: args{
				unixSeconds: 1531238400,
			},
			want: "20180710",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetVNDateFromUTC(tt.args.unixSeconds); got != tt.want {
				t.Errorf("GetVNDateFromUTC() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFor(t *testing.T) {
	println(GetSecondOfDay())
}

package util

import "testing"

func TestIsSystemSupportEmoji(t *testing.T) {
	type args struct {
		osVersion string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "android_7.0",
			args: args{
				osVersion: "android_7.0",
			},
			want: true,
		},
		{
			name: "android_5.1.1",
			args: args{
				osVersion: "android_5.1.1",
			},
			want: false,
		},
		{
			name: "android_5.1",
			args: args{
				osVersion: "android_5.1",
			},
			want: false,
		},
		{
			name: "android_6.0.1",
			args: args{
				osVersion: "android_6.0.1",
			},
			want: true,
		},
		{
			name: "android_7.1.1",
			args: args{
				osVersion: "android_7.1.1",
			},
			want: true,
		},
		{
			name: "android_4.4.2",
			args: args{
				osVersion: "android_4.4.2",
			},
			want: false,
		},
		{
			name: "android_6.0",
			args: args{
				osVersion: "android_6.0",
			},
			want: false,
		},
		{
			name: "ios_10.3.3.0",
			args: args{
				osVersion: "ios_10.3.3",
			},
			want: true,
		},
		{
			name: "ios",
			args: args{
				osVersion: "ios",
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsSystemSupportEmoji(tt.args.osVersion); got != tt.want {
				t.Errorf("IsSystemSupportEmoji() = %v, want %v", got, tt.want)
			}
		})
	}
}

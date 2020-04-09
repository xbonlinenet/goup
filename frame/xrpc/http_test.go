package xrpc

import (
	"reflect"
	"testing"
	"time"
)

func TestHttpPostWithJson(t *testing.T) {
	type args struct {
		url     string
		data    interface{}
		timeout time.Duration
	}
	tests := []struct {
		name    string
		args    args
		want    []byte
		wantErr bool
	}{
		// TODO: Add test cases.
		{
			name: "default",
			args: args{
				url: "http://localhost:13360/api/demo/sleep",
				data: map[string]interface{}{
					"seconds": 1,
				},
				timeout: time.Second * 4,
			},
			want:    []byte("{\"code\":0}\x0a"),
			wantErr: false,
		},
		{
			name: "timeout",
			args: args{
				url: "http://localhost:13360/api/demo/sleep",
				data: map[string]interface{}{
					"seconds": 3,
				},
				timeout: time.Second * 1,
			},
			want:    []byte{},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := HttpPostWithJson(tt.args.url, tt.args.data, tt.args.timeout)
			if (err != nil) != tt.wantErr {
				t.Errorf("HttpPostWithJson() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("HttpPostWithJson() = %v, want %v", got, tt.want)
			}
		})
	}
}

package ab

import (
	"reflect"
	"testing"
)

func Test_parseRange(t *testing.T) {
	type args struct {
		str string
	}
	tests := []struct {
		name    string
		args    args
		want    []Range
		wantErr bool
	}{
		// TODO: Add test cases.
		{
			name: "default",
			args: args{
				str: "10-20,50-70",
			},
			want: []Range{
				Range{
					Start: 10,
					End:   20,
				},
				Range{
					Start: 50,
					End:   70,
				},
			},
			wantErr: false,
		},
		{
			name: "error",
			args: args{
				str: "10-20,50-",
			},
			want:    []Range{},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseRange(tt.args.str)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseRange() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("parseRange() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestUserIDModGroup_In(t *testing.T) {
	type fields struct {
		Ranges []Range
		Config map[string]interface{}
	}
	type args struct {
		userID int64
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   bool
	}{
		// TODO: Add test cases.
		{
			name: "default_hit",
			fields: fields{
				Ranges: []Range{
					Range{
						Start: 10,
						End:   30,
					},
				},
			},
			args: args{
				userID: 100523,
			},
			want: true,
		},
		{
			name: "default_miss",
			fields: fields{
				Ranges: []Range{
					Range{
						Start: 10,
						End:   30,
					},
				},
			},
			args: args{
				userID: 100530,
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			group := UserIDModGroup{
				Ranges: tt.fields.Ranges,
				Config: tt.fields.Config,
			}
			if got := group.In(tt.args.userID); got != tt.want {
				t.Errorf("UserIDModGroup.In() = %v, want %v", got, tt.want)
			}
		})
	}
}

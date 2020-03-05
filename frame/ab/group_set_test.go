package ab

import (
	"reflect"
	"testing"
)

func TestGroupSet_FindGroup(t *testing.T) {
	type fields struct {
		groups []*UserIDModGroup
	}
	type args struct {
		userID int64
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   *UserIDModGroup
	}{
		{
			name: "default",
			fields: fields{
				groups: []*UserIDModGroup{
					&UserIDModGroup{
						Name: "1",
						Ranges: []Range{
							Range{
								Start: 0,
								End:   10,
							},
						},
					},
					&UserIDModGroup{
						Name: "2",
						Ranges: []Range{
							Range{
								Start: 20,
								End:   60,
							},
						},
					},
				},
			},
			args: args{
				userID: 11111109,
			},
			want: &UserIDModGroup{
				Name: "1",
				Ranges: []Range{
					Range{
						Start: 0,
						End:   10,
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := GroupSet{
				groups: tt.fields.groups,
			}
			if got := s.FindGroup(tt.args.userID); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GroupSet.FindGroup() = %v, want %v", got, tt.want)
			}
		})
	}
}

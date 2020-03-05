package dyncfg

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/deckarep/golang-set"
)

func TestMustGetStringSlice(t *testing.T) {
	type args struct {
		item string
	}
	config = map[string]string{
		"test":    "[\"array\", \"test\"]",
		"testint": "[1, 3, 5]",
	}
	tests := []struct {
		name string
		args args
		want []string
	}{
		// TODO: Add test cases.
		{
			name: "deafult",
			args: args{
				item: "test",
			},
			want: []string{"array", "test"},
		},
		{
			name: "testint",
			args: args{
				item: "testint",
			},
			want: []string{"1", "3", "5"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := MustGetStringSlice(tt.args.item); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("MustGetStringSlice() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMustGetSet(t *testing.T) {
	type args struct {
		item string
	}
	config = map[string]string{
		"test":    "[\"array\", \"test\"]",
		"testint": "[1, 3, 5]",
	}
	tests := []struct {
		name string
		args args
		want mapset.Set
	}{
		// TODO: Add test cases.
		{
			name: "deafult",
			args: args{
				item: "test",
			},
			want: mapset.NewSetFromSlice([]interface{}{"array", "test"}),
		},
		{
			name: "testint",
			args: args{
				item: "testint",
			},
			want: mapset.NewSetFromSlice([]interface{}{json.Number("1"), json.Number("3"), json.Number("5")}),
		}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := MustGetSet(tt.args.item); !got.Equal(tt.want) {
				t.Errorf("MustGetSet() = %v, want %v", got, tt.want)
			}
		})
	}
}

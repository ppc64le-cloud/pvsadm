package client

import (
	"reflect"
	"testing"
)

func TestListEnvironments(t *testing.T) {
	tests := []struct {
		name     string
		wantKeys []string
	}{
		{
			"valid environments",
			[]string{"test", "prod"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if gotKeys := ListEnvironments(); !reflect.DeepEqual(gotKeys, tt.wantKeys) {
				t.Errorf("ListEnvironments() = %v, want %v", gotKeys, tt.wantKeys)
			}
		})
	}
}

func TestGetEnvironment(t *testing.T) {
	type args struct {
		env string
	}
	tests := []struct {
		name  string
		args  args
		want1 map[string]string
		want  error
	}{
		{
			"valid environment - test",
			args{"test"},
			Environments["test"],
			nil,
		},
		{
			"valid environment - prod",
			args{"prod"},
			Environments["prod"],
			nil,
		},
		{
			"valid environment - prod",
			args{"fake"},
			nil,
			EnvironmentNotFound,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got1, got := GetEnvironment(tt.args.env)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetEnvironment() got = %v, want %v", got, tt.want)
			}
			if !reflect.DeepEqual(got1, tt.want1) {
				t.Errorf("GetEnvironment() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

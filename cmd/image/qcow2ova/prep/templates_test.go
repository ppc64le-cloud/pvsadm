package prep

import (
	"strings"
	"testing"
)

func TestRender(t *testing.T) {
	type args struct {
		dist       string
		rhnuser    string
		rhnpasswd  string
		rootpasswd string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			"rhel image",
			args{"rhel", "rhn", "rhnpassword", "some-password"},
			"subscription-manager",
			false,
		},
		{
			"centos image",
			args{dist: "centos", rootpasswd: "some-password"},
			"some-password",
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Render(tt.args.dist, tt.args.rhnuser, tt.args.rhnpasswd, tt.args.rootpasswd)
			if (err != nil) != tt.wantErr {
				t.Errorf("Render() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !strings.Contains(got, tt.want) {
				t.Errorf("Render() %s does not contain the %s", got, tt.want)
			}
		})
	}
}

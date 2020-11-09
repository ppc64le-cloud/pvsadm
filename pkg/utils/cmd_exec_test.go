package utils

import (
	"strings"
	"testing"
)

func TestRunCMD(t *testing.T) {
	type args struct {
		cmd  string
		args []string
	}
	tests := []struct {
		name  string
		args  args
		want  int
		want1 string
		want2 string
	}{
		{
			"echo test",
			args{"echo", []string{"hello world"}},
			0,
			"hello world",
			"",
		},
		{
			"Command not Found",
			args{"some-command", []string{"ls"}},
			1,
			"",
			"exec: \"some-command\": executable file not found in $PATH",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1, got2 := RunCMD(tt.args.cmd, tt.args.args...)
			if got != tt.want {
				t.Errorf("RunCMD() got = %v, want %v", got, tt.want)
			}
			if strings.TrimSuffix(got1, "\n") != tt.want1 {
				t.Errorf("RunCMD() got1 = %v, want %v", got1, tt.want1)
			}
			if strings.TrimSuffix(got2, "\n") != tt.want2 {
				t.Errorf("RunCMD() got2 = %v, want %v", got2, tt.want2)
			}
		})
	}
}

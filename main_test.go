package main

import "testing"

func TestShouldRunCLIRoutesManagementAndDirectLaunchCommands(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want bool
	}{
		{
			name: "no arguments opens app",
			args: nil,
			want: false,
		},
		{
			name: "plain project path opens app",
			args: []string{"."},
			want: false,
		},
		{
			name: "context command",
			args: []string{"context", "list"},
			want: true,
		},
		{
			name: "project command",
			args: []string{"project", "show"},
			want: true,
		},
		{
			name: "generic direct launch",
			args: []string{"--context", "personal", "."},
			want: true,
		},
		{
			name: "personal alias direct launch",
			args: []string{"--personal", "."},
			want: true,
		},
		{
			name: "company alias direct launch",
			args: []string{"--company", "."},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := shouldRunCLI(tt.args); got != tt.want {
				t.Fatalf("shouldRunCLI(%#v) = %v, want %v", tt.args, got, tt.want)
			}
		})
	}
}

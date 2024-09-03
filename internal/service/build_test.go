package service

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPrintBuildInfo(t *testing.T) {
	tests := []struct{
		name string
		version string
		date string
		commit string
		want string
	}{
		{
			name: "HasInfo",
			version: "v1.0.0",
			date: "2024-08-30",
			commit: "abcdefg",
			want: "Build version: v1.0.0\nBuild date: 2024-08-30\nBuild commit: abcdefg\n",
		},
		{
			name: "NoInfo",
			want: "Build version: N/A\nBuild date: N/A\nBuild commit: N/A\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			PrintBuildInfo(tt.version, tt.date, tt.commit, &buf)
			assert.Equal(t, tt.want, buf.String())		
		})
	}
}
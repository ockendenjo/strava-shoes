package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_getInputDirectory(t *testing.T) {
	mainFile := "./cmd/workflows/BP005/data-processing/main.go"
	inputDir := getInputDirectory(mainFile)
	assert.Equal(t, "./cmd/workflows/BP005/data-processing", inputDir)
}

func Test_getOutputPath(t *testing.T) {
	testcases := []struct {
		name     string
		mainFile string
		expPath  string
	}{
		{
			name:     "non-nested directory",
			mainFile: "./cmd/send-pub-to-dip/main.go",
			expPath:  "build/send-pub-to-dip/bootstrap",
		},
		{
			name:     "nested directory",
			mainFile: "./cmd/workflows/BP003/PUB-036/main.go",
			expPath:  "build/workflows-BP003-PUB-036/bootstrap",
		},
	}
	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			outPath := getOutputPath(tc.mainFile)
			assert.Equal(t, tc.expPath, outPath)
		})
	}
}

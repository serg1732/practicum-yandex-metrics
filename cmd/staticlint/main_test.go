package main

import (
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"
)

func TestForbiddenCallsOutsideMainAnalyzer(t *testing.T) {
	testdata := analysistest.TestData()

	analysistest.Run(
		t,
		testdata,
		forbiddenCallsOutsideMainAnalyzer,
		"source",
		"lib",
		"generated",
	)
}

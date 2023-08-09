package analyzer_test

import (
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"

	"github.com/kmirzavaziri/goimportgroups/pkg/analyzer"
)

func TestAnalyzerWithoutConfig(t *testing.T) {
	a := analyzer.NewAnalyzer()

	analysistest.Run(
		t,
		analysistest.TestData(), a,
		"single_group_no_config",
	)
}

func TestAnalyzerWithConfig(t *testing.T) {
	a := analyzer.NewAnalyzer()

	err := a.Flags.Lookup("groups").Value.Set("fmt:os;time;strings;regexp")
	if err != nil {
		t.Fail()
	}

	analysistest.Run(
		t,
		analysistest.TestData(), a,
		"correct",
		"single_group",
		"swapped_groups",
		"no_imports",
		"multiple_sections",
	)
}

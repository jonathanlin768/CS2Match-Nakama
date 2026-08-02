package matchengine

import (
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestProductionCallGraphContainsNoReverseCausalLegacyPath(t *testing.T) {
	forbidden := []string{
		"resolveRoundWinner", "resolveRoundEvents", "pickKillSide", "survivorTarget",
		"inferWinReason", "target_survivors", "TargetWinner", "ForcedRoundWinners",
	}
	files, err := filepath.Glob("*.go")
	if err != nil {
		t.Fatal(err)
	}
	fset := token.NewFileSet()
	for _, file := range files {
		if strings.HasSuffix(file, "_test.go") {
			continue
		}
		source, err := os.ReadFile(file)
		if err != nil {
			t.Fatal(err)
		}
		for _, symbol := range forbidden {
			if strings.Contains(string(source), symbol) {
				t.Fatalf("production source %s still contains reverse-causal symbol %s", file, symbol)
			}
		}
		parsed, err := parser.ParseFile(fset, file, source, parser.ImportsOnly)
		if err != nil {
			t.Fatal(err)
		}
		for _, spec := range parsed.Imports {
			if strings.Contains(spec.Path.Value, "windypath.com/cs2match/config") {
				t.Fatalf("matchengine production source %s imports generated config package", file)
			}
		}
	}
}

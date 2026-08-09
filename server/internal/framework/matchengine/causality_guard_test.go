package matchengine

import (
	"go/ast"
	"go/parser"
	"go/token"
	"path/filepath"
	"strings"
	"testing"
)

// This guard prevents the new causal subsystems from accepting reverse-causal
// inputs while the explicitly baselined legacy engine is being replaced.
func TestCausalSubsystemAPIsRejectPreselectedOutcomeInputs(t *testing.T) {
	files, err := filepath.Glob("*.go")
	if err != nil {
		t.Fatalf("glob production sources: %v", err)
	}
	fset := token.NewFileSet()
	for _, file := range files {
		if strings.HasSuffix(file, "_test.go") {
			continue
		}
		parsed, err := parser.ParseFile(fset, file, nil, 0)
		if err != nil {
			t.Fatalf("parse %s: %v", file, err)
		}
		for _, decl := range parsed.Decls {
			fn, ok := decl.(*ast.FuncDecl)
			if !ok || !isCausalSubsystemName(fn.Name.Name) {
				continue
			}
			assertNoForbiddenFields(t, file, "function "+fn.Name.Name, fn.Type.Params)
		}
		for _, decl := range parsed.Decls {
			gen, ok := decl.(*ast.GenDecl)
			if !ok || gen.Tok != token.TYPE {
				continue
			}
			for _, spec := range gen.Specs {
				typeSpec := spec.(*ast.TypeSpec)
				structType, ok := typeSpec.Type.(*ast.StructType)
				if !ok || !isCausalInputType(typeSpec.Name.Name) {
					continue
				}
				assertNoForbiddenFields(t, file, "type "+typeSpec.Name.Name, structType.Fields)
			}
		}
	}
}

func isCausalSubsystemName(name string) bool {
	name = strings.ToLower(name)
	for _, subsystem := range []string{"encounter", "decision", "bomb", "scheduler", "terminal"} {
		if strings.Contains(name, subsystem) {
			return true
		}
	}
	return false
}

func isCausalInputType(name string) bool {
	lower := strings.ToLower(name)
	if !isCausalSubsystemName(lower) {
		return false
	}
	for _, suffix := range []string{"input", "context", "request", "params"} {
		if strings.HasSuffix(lower, suffix) {
			return true
		}
	}
	return false
}

func assertNoForbiddenFields(t *testing.T, file, owner string, fields *ast.FieldList) {
	t.Helper()
	if fields == nil {
		return
	}
	for _, field := range fields.List {
		for _, name := range field.Names {
			compact := strings.NewReplacer("_", "", "-", "").Replace(strings.ToLower(name.Name))
			for _, forbidden := range []string{"winner", "winreason", "targetwinner", "targetsurvivor", "forcedwinner"} {
				if strings.Contains(compact, forbidden) {
					t.Errorf("%s: %s accepts forbidden reverse-causal field %s", file, owner, name.Name)
				}
			}
		}
	}
}

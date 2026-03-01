package architecture

import (
	"bytes"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

type listedPackage struct {
	ImportPath   string   `json:"ImportPath"`
	Imports      []string `json:"Imports"`
	TestImports  []string `json:"TestImports"`
	XTestImports []string `json:"XTestImports"`
}

func TestDependencyDirectionRules(t *testing.T) {
	root := repoRoot(t)

	domainPkgs := goListPackages(t, root, "./internal/domain/...")
	for _, pkg := range domainPkgs {
		checkNoPrefixImports(t, pkg, []string{
			"snake/internal/app/",
			"snake/internal/infra/",
			"snake/internal/ui/",
		})
		checkNoThirdPartyImports(t, pkg)
	}

	appPkgs := goListPackages(t, root, "./internal/app/...")
	for _, pkg := range appPkgs {
		checkNoPrefixImports(t, pkg, []string{
			"snake/internal/infra/",
			"snake/internal/ui/",
		})
	}

	uiPkgs := goListPackages(t, root, "./internal/ui/...")
	for _, pkg := range uiPkgs {
		checkNoPrefixImports(t, pkg, []string{
			"snake/internal/domain/",
			"snake/internal/infra/",
		})
	}
}

func repoRoot(t *testing.T) string {
	t.Helper()

	dir, err := os.Getwd()
	if err != nil {
		t.Fatalf("get working directory: %v", err)
	}

	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatalf("could not locate repository root from %q", dir)
		}
		dir = parent
	}
}

func goListPackages(t *testing.T, root, pattern string) []listedPackage {
	t.Helper()

	cmd := exec.Command("go", "list", "-json", pattern)
	cmd.Dir = root
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("go list failed for pattern %q: %v\n%s", pattern, err, string(output))
	}

	dec := json.NewDecoder(bytes.NewReader(output))
	var out []listedPackage
	for dec.More() {
		var p listedPackage
		if err := dec.Decode(&p); err != nil {
			t.Fatalf("decode go list output for %q: %v", pattern, err)
		}
		if p.ImportPath != "" {
			out = append(out, p)
		}
	}
	if len(out) == 0 {
		t.Fatalf("go list returned no packages for %q", pattern)
	}
	return out
}

func allImports(p listedPackage) []string {
	var imports []string
	imports = append(imports, p.Imports...)
	imports = append(imports, p.TestImports...)
	imports = append(imports, p.XTestImports...)
	return imports
}

func checkNoPrefixImports(t *testing.T, p listedPackage, forbiddenPrefixes []string) {
	t.Helper()

	for _, imp := range allImports(p) {
		for _, prefix := range forbiddenPrefixes {
			if strings.HasPrefix(imp, prefix) {
				t.Fatalf("package %q imports forbidden dependency %q", p.ImportPath, imp)
			}
		}
	}
}

func checkNoThirdPartyImports(t *testing.T, p listedPackage) {
	t.Helper()

	for _, imp := range allImports(p) {
		if strings.HasPrefix(imp, "snake/") {
			continue
		}
		first := imp
		if idx := strings.IndexByte(imp, '/'); idx >= 0 {
			first = imp[:idx]
		}
		if strings.Contains(first, ".") {
			t.Fatalf("domain package %q imports third-party dependency %q", p.ImportPath, imp)
		}
	}
}

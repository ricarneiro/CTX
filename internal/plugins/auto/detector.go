package auto

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// Stack represents a detected technology stack in a project directory.
type Stack struct {
	Name       string   // "csharp", "react", "go", "python"
	Confidence string   // "high", "medium", "low"
	Evidence   []string // human-readable items explaining the detection
}

// Detect scans rootDir up to 3 levels deep and returns detected stacks,
// ordered by confidence (high first). Build artifacts and dependency
// directories (node_modules, bin, obj, etc.) are skipped.
func Detect(rootDir string) ([]Stack, error) {
	f, err := scanDir(rootDir, 3)
	if err != nil {
		return nil, err
	}
	return buildStacks(f), nil
}

// skipDirs lists directory names that are never descended into during scanning.
var skipDirs = map[string]bool{
	"node_modules": true,
	"bin":          true,
	"obj":          true,
	"dist":         true,
	"build":        true,
	"out":          true,
	".git":         true,
	".vs":          true,
	".vscode":      true,
	".idea":        true,
	"target":       true,
	"vendor":       true,
}

// findings holds raw evidence gathered during the directory walk.
type findings struct {
	slnFiles      []string // forward-slash relative paths to .sln files
	csprojFiles   []string // forward-slash relative paths to .csproj/.fsproj/.vbproj files
	hasGlobalJSON bool
	packageJSONs  []pkgJSON
	hasGoModRoot  bool // go.mod is a direct child of rootDir (depth 1)
	hasPyProject  bool
	hasReqTxt     bool
	hasSetupPy    bool
	hasTSXorJSX   bool
}

type pkgJSON struct {
	relPath       string
	hasReact      bool
	hasTypeScript bool
}

func scanDir(rootDir string, maxDepth int) (findings, error) {
	var f findings

	err := filepath.WalkDir(rootDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil //nolint:nilerr // skip unreadable entries; returning err would abort the walk
		}
		if path == rootDir {
			return nil // skip root itself
		}

		rel, _ := filepath.Rel(rootDir, path)
		depth := len(strings.Split(rel, string(filepath.Separator)))
		relFwd := filepath.ToSlash(rel)

		if d.IsDir() {
			if skipDirs[d.Name()] {
				return filepath.SkipDir
			}
			if depth >= maxDepth {
				return filepath.SkipDir
			}
			return nil
		}

		// File: classify by extension then by base name.
		name := d.Name()
		base := strings.ToLower(name)
		ext := strings.ToLower(filepath.Ext(name))

		switch ext {
		case ".sln":
			f.slnFiles = append(f.slnFiles, relFwd)
		case ".csproj", ".fsproj", ".vbproj":
			f.csprojFiles = append(f.csprojFiles, relFwd)
		case ".tsx", ".jsx":
			f.hasTSXorJSX = true
		}

		switch base {
		case "global.json":
			f.hasGlobalJSON = true
		case "package.json":
			f.packageJSONs = append(f.packageJSONs, parsePackageJSON(path, relFwd))
		case "go.mod":
			if depth == 1 {
				f.hasGoModRoot = true
			}
		case "pyproject.toml":
			f.hasPyProject = true
		case "requirements.txt":
			f.hasReqTxt = true
		case "setup.py":
			f.hasSetupPy = true
		}

		return nil
	})

	return f, err
}

// pkgJSONSchema is the minimal structure needed to check for specific deps.
type pkgJSONSchema struct {
	Dependencies    map[string]json.RawMessage `json:"dependencies"`
	DevDependencies map[string]json.RawMessage `json:"devDependencies"`
}

func parsePackageJSON(path, relPath string) pkgJSON {
	result := pkgJSON{relPath: relPath}
	b, err := os.ReadFile(path)
	if err != nil {
		return result
	}
	var raw pkgJSONSchema
	if err := json.Unmarshal(b, &raw); err != nil {
		return result
	}
	allDeps := make(map[string]bool, len(raw.Dependencies)+len(raw.DevDependencies))
	for k := range raw.Dependencies {
		allDeps[k] = true
	}
	for k := range raw.DevDependencies {
		allDeps[k] = true
	}
	result.hasReact = allDeps["react"]
	result.hasTypeScript = allDeps["typescript"] || allDeps["@types/react"]
	return result
}

// buildStacks applies heuristics to the findings and returns stacks sorted
// by confidence descending.
func buildStacks(f findings) []Stack {
	var stacks []Stack

	// C# — high if .sln or .csproj present; medium if only global.json.
	if len(f.slnFiles) > 0 || len(f.csprojFiles) > 0 {
		var evidence []string
		for _, s := range f.slnFiles {
			evidence = append(evidence, evidenceItem(s))
		}
		for i, s := range f.csprojFiles {
			if i >= 3 {
				evidence = append(evidence, fmt.Sprintf("...and %d more .csproj files", len(f.csprojFiles)-3))
				break
			}
			evidence = append(evidence, evidenceItem(s))
		}
		stacks = append(stacks, Stack{Name: "csharp", Confidence: "high", Evidence: evidence})
	} else if f.hasGlobalJSON {
		stacks = append(stacks, Stack{
			Name: "csharp", Confidence: "medium",
			Evidence: []string{"found: `global.json`"},
		})
	}

	// React — high if package.json lists react; medium if TypeScript + tsx/jsx files.
	reactFound := false
	for _, pkg := range f.packageJSONs {
		if pkg.hasReact {
			stacks = append(stacks, Stack{
				Name:       "react",
				Confidence: "high",
				Evidence:   []string{fmt.Sprintf("found: `%s` with `react` in dependencies", pkg.relPath)},
			})
			reactFound = true
			break
		}
	}
	if !reactFound {
		for _, pkg := range f.packageJSONs {
			if pkg.hasTypeScript && f.hasTSXorJSX {
				stacks = append(stacks, Stack{
					Name:       "react",
					Confidence: "medium",
					Evidence:   []string{fmt.Sprintf("found: `%s` with TypeScript + `.tsx`/`.jsx` files", pkg.relPath)},
				})
				break
			}
		}
	}

	// Go — high if go.mod is at the root (not nested).
	if f.hasGoModRoot {
		stacks = append(stacks, Stack{
			Name: "go", Confidence: "high",
			Evidence: []string{"found: `go.mod`"},
		})
	}

	// Python — medium if any standard config file is present.
	if f.hasPyProject || f.hasReqTxt || f.hasSetupPy {
		var evidence []string
		if f.hasPyProject {
			evidence = append(evidence, "found: `pyproject.toml`")
		}
		if f.hasReqTxt {
			evidence = append(evidence, "found: `requirements.txt`")
		}
		if f.hasSetupPy {
			evidence = append(evidence, "found: `setup.py`")
		}
		stacks = append(stacks, Stack{Name: "python", Confidence: "medium", Evidence: evidence})
	}

	sort.SliceStable(stacks, func(i, j int) bool {
		return confidenceRank(stacks[i].Confidence) > confidenceRank(stacks[j].Confidence)
	})

	return stacks
}

// evidenceItem formats a relative path for evidence display.
// Root-level files get an "at root" suffix.
func evidenceItem(relFwd string) string {
	if !strings.Contains(relFwd, "/") {
		return fmt.Sprintf("found: `%s` at root", relFwd)
	}
	return fmt.Sprintf("found: `%s`", relFwd)
}

func confidenceRank(c string) int {
	switch c {
	case "high":
		return 3
	case "medium":
		return 2
	case "low":
		return 1
	default:
		return 0
	}
}

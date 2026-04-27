package csharp

import (
	"fmt"
	"io"
	"strings"

	"github.com/ricarneiro/ctx/internal/output"
	"github.com/ricarneiro/ctx/internal/plugins/csharp/helper"
)

// WriteSummary formats a ProjectSummary as dense markdown.
func WriteSummary(w io.Writer, s *helper.ProjectSummary) error {
	output.H1(w, "Solution: "+s.SolutionName)
	writeOverview(w, s)
	output.H2(w, "Projects")
	writeProjects(w, s.Projects)
	writeReferenceGraph(w, s.Projects)
	writeMultiTargeting(w, s.Projects)
	return nil
}

func writeOverview(w io.Writer, s *helper.ProjectSummary) {
	exeCount, libCount := 0, 0
	totalDocs := 0
	for _, p := range s.Projects {
		if p.Type == "exe" {
			exeCount++
		} else {
			libCount++
		}
		totalDocs += p.DocumentCount
	}

	projectSummary := fmt.Sprintf("%d", len(s.Projects))
	parts := []string{}
	if exeCount > 0 {
		parts = append(parts, fmt.Sprintf("%d exe", exeCount))
	}
	if libCount > 0 {
		parts = append(parts, fmt.Sprintf("%d lib", libCount))
	}
	if len(parts) > 0 {
		projectSummary += " (" + strings.Join(parts, ", ") + ")"
	}

	output.KeyValue(w, "path", "`"+s.SolutionPath+"`")
	output.KeyValue(w, "projects", projectSummary)
	output.KeyValue(w, "documents", fmt.Sprintf("%d", totalDocs))
	fmt.Fprintln(w)
}

func writeProjects(w io.Writer, projects []helper.ProjectInfo) {
	for _, p := range projects {
		output.H3(w, fmt.Sprintf("%s (%s)", p.Name, p.Type))

		fmt.Fprintf(w, "- **path:** `%s`\n", p.Path)

		targets := strings.Join(p.TargetFrameworks, ", ")
		if targets == "" {
			targets = "_(unknown)_"
		}
		fmt.Fprintf(w, "- **target:** %s\n", targets)
		fmt.Fprintf(w, "- **namespace:** `%s`\n", p.RootNamespace)
		fmt.Fprintf(w, "- **documents:** %d\n", p.DocumentCount)

		// Project references
		if len(p.ProjectReferences) == 0 {
			fmt.Fprintf(w, "- **references:** _(none)_\n")
		} else {
			refs := make([]string, len(p.ProjectReferences))
			for i, r := range p.ProjectReferences {
				refs[i] = "`" + r + "`"
			}
			fmt.Fprintf(w, "- **references:**\n")
			for _, r := range refs {
				fmt.Fprintf(w, "  - %s\n", r)
			}
		}

		// Package references
		if len(p.PackageReferences) == 0 {
			fmt.Fprintf(w, "- **packages:** _(none)_\n")
		} else {
			pkgs := make([]string, len(p.PackageReferences))
			for i, pkg := range p.PackageReferences {
				pkgs[i] = pkg.Name + " " + pkg.Version
			}
			fmt.Fprintf(w, "- **packages:** %s\n", strings.Join(pkgs, ", "))
		}

		fmt.Fprintln(w)
	}
}

func writeReferenceGraph(w io.Writer, projects []helper.ProjectInfo) {
	// Build reverse dependency map: who depends on each project
	dependents := map[string][]string{}
	refMap := map[string][]string{}
	nameSet := map[string]bool{}

	for _, p := range projects {
		nameSet[p.Name] = true
		refMap[p.Name] = p.ProjectReferences
		for _, ref := range p.ProjectReferences {
			dependents[ref] = append(dependents[ref], p.Name)
		}
	}

	// Check if there are any references at all
	hasRefs := false
	for _, p := range projects {
		if len(p.ProjectReferences) > 0 {
			hasRefs = true
			break
		}
	}

	output.H2(w, "Reference graph")

	if !hasRefs {
		fmt.Fprintf(w, "No inter-project references.\n\n")
		return
	}

	// For complex graphs (>5 projects with references), use flat list
	refsCount := 0
	for _, p := range projects {
		if len(p.ProjectReferences) > 0 {
			refsCount++
		}
	}

	if len(projects) > 5 && refsCount > 3 {
		for _, p := range projects {
			if len(p.ProjectReferences) > 0 {
				refs := make([]string, len(p.ProjectReferences))
				for i, r := range p.ProjectReferences {
					refs[i] = "`" + r + "`"
				}
				fmt.Fprintf(w, "- `%s` → %s\n", p.Name, strings.Join(refs, ", "))
			}
		}
		fmt.Fprintln(w)
		return
	}

	// DFS from roots (projects with no dependents)
	roots := []string{}
	for _, p := range projects {
		if len(dependents[p.Name]) == 0 {
			roots = append(roots, p.Name)
		}
	}

	if len(roots) == 0 {
		// Cycle or all have dependents — fall back to flat list
		for _, p := range projects {
			if len(p.ProjectReferences) > 0 {
				refs := make([]string, len(p.ProjectReferences))
				for i, r := range p.ProjectReferences {
					refs[i] = "`" + r + "`"
				}
				fmt.Fprintf(w, "- `%s` → %s\n", p.Name, strings.Join(refs, ", "))
			}
		}
		fmt.Fprintln(w)
		return
	}

	var sb strings.Builder
	visited := map[string]bool{}
	for _, root := range roots {
		dfsRender(&sb, root, refMap, visited, 0)
	}
	output.CodeBlock(w, "", strings.TrimRight(sb.String(), "\n"))
}

func dfsRender(sb *strings.Builder, name string, refMap map[string][]string, visited map[string]bool, depth int) {
	indent := strings.Repeat("  ", depth)
	refs := refMap[name]
	if len(refs) == 0 {
		if depth == 0 {
			fmt.Fprintf(sb, "%s%s (no deps)\n", indent, name)
		} else {
			fmt.Fprintf(sb, "%s%s\n", indent, name)
		}
		return
	}

	if visited[name] {
		fmt.Fprintf(sb, "%s%s (see above)\n", indent, name)
		return
	}
	visited[name] = true

	for i, ref := range refs {
		if i == 0 {
			fmt.Fprintf(sb, "%s%s → %s\n", indent, name, ref)
		} else {
			fmt.Fprintf(sb, "%s%s   └──→ %s\n", indent, strings.Repeat(" ", len(name)), ref)
		}
		dfsRender(sb, ref, refMap, visited, depth+1)
	}
}

func writeMultiTargeting(w io.Writer, projects []helper.ProjectInfo) {
	output.H2(w, "Multi-targeting")

	// Collect all unique frameworks
	frameworkSets := map[string][]string{} // project name → frameworks
	allFrameworks := map[string]bool{}

	for _, p := range projects {
		if len(p.TargetFrameworks) > 1 {
			frameworkSets[p.Name] = p.TargetFrameworks
		}
		for _, tf := range p.TargetFrameworks {
			allFrameworks[tf] = true
		}
	}

	if len(frameworkSets) == 0 {
		// All same framework, or single framework each
		if len(allFrameworks) == 1 {
			for tf := range allFrameworks {
				fmt.Fprintf(w, "None — all projects target `%s`.\n\n", tf)
			}
		} else if len(allFrameworks) == 0 {
			fmt.Fprintf(w, "No target frameworks detected.\n\n")
		} else {
			// Multiple different single targets
			for _, p := range projects {
				if len(p.TargetFrameworks) > 0 {
					fmt.Fprintf(w, "- `%s` targets: `%s`\n", p.Name, strings.Join(p.TargetFrameworks, "`, `"))
				}
			}
			fmt.Fprintln(w)
		}
		return
	}

	for _, p := range projects {
		if tfs, ok := frameworkSets[p.Name]; ok {
			fmt.Fprintf(w, "- `%s` targets: `%s`\n", p.Name, strings.Join(tfs, "`, `"))
		}
	}
	fmt.Fprintln(w)
}

package git

import (
	"fmt"
	"io"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// formatOutput writes the full markdown report to w.
func formatOutput(w io.Writer, data *gitData) {
	fmt.Fprintln(w, "# Git Context")
	fmt.Fprintln(w)

	writeMeta(w, data)
	fmt.Fprintln(w)

	writeCommits(w, data.commits)
	writeWorkingTree(w, data.tree)
	writeDiffSummary(w, data.tree)
}

func writeMeta(w io.Writer, data *gitData) {
	// Branch + upstream
	branch := data.repo.branch
	if data.repo.upstream != "" {
		branch += fmt.Sprintf(" (ahead %d, behind %d vs %s)",
			data.repo.ahead, data.repo.behind, data.repo.upstream)
	}
	fmt.Fprintf(w, "**branch:** %s\n", branch)

	// Status summary
	m, s, u := len(data.tree.modified), len(data.tree.staged), len(data.tree.untracked)
	if m == 0 && s == 0 && u == 0 {
		fmt.Fprintln(w, "**status:** clean")
	} else {
		var parts []string
		if m > 0 {
			parts = append(parts, fmt.Sprintf("%d modified", m))
		}
		if s > 0 {
			parts = append(parts, fmt.Sprintf("%d staged", s))
		}
		if u > 0 {
			parts = append(parts, fmt.Sprintf("%d untracked", u))
		}
		fmt.Fprintf(w, "**status:** %s\n", strings.Join(parts, ", "))
	}

	// Last fetched
	if data.repo.hasFetch {
		fmt.Fprintf(w, "**last fetched:** %s\n", relativeTime(data.repo.lastFetch))
	}
}

func writeCommits(w io.Writer, commits []commit) {
	if len(commits) == 0 {
		return
	}
	fmt.Fprintf(w, "## Recent commits (last %d)\n", len(commits))
	for _, c := range commits {
		t := time.Unix(c.unixTs, 0)
		fmt.Fprintf(w, "- `%s` (%s, %s) %s\n", c.hash, relativeTime(t), c.author, c.subject)
	}
	fmt.Fprintln(w)
}

func writeWorkingTree(w io.Writer, tree workingTree) {
	if len(tree.modified) == 0 && len(tree.staged) == 0 && len(tree.untracked) == 0 {
		return
	}
	fmt.Fprintln(w, "## Working tree")

	if len(tree.modified) > 0 {
		fmt.Fprintf(w, "### Modified (%d)\n", len(tree.modified))
		for _, f := range tree.modified {
			fmt.Fprintf(w, "- `%s` (+%d -%d)\n", f.path, f.added, f.removed)
		}
		fmt.Fprintln(w)
	}

	if len(tree.staged) > 0 {
		fmt.Fprintf(w, "### Staged (%d)\n", len(tree.staged))
		for _, f := range tree.staged {
			fmt.Fprintf(w, "- `%s` (+%d -%d)\n", f.path, f.added, f.removed)
		}
		fmt.Fprintln(w)
	}

	if len(tree.untracked) > 0 {
		fmt.Fprintf(w, "### Untracked (%d)\n", len(tree.untracked))
		for _, u := range tree.untracked {
			fmt.Fprintf(w, "- `%s`\n", u)
		}
		fmt.Fprintln(w)
	}
}

func writeDiffSummary(w io.Writer, tree workingTree) {
	all := mergeChanges(tree.modified, tree.staged)
	if len(all) == 0 {
		return
	}

	totalAdded, totalRemoved := 0, 0
	for _, f := range all {
		totalAdded += f.added
		totalRemoved += f.removed
	}

	fmt.Fprintln(w, "## Diff summary (unstaged + staged)")
	fmt.Fprintf(w, "**Total:** +%d -%d across %d file%s\n\n",
		totalAdded, totalRemoved, len(all), plural(len(all)))

	// By top-level directory (only if files span more than one).
	byDir := groupByTopDir(all)
	if len(byDir) > 1 {
		fmt.Fprintln(w, "### By directory")
		dirs := make([]string, 0, len(byDir))
		for d := range byDir {
			dirs = append(dirs, d)
		}
		sort.Strings(dirs)
		for _, d := range dirs {
			s := byDir[d]
			label := d + "/"
			if d == "." {
				label = "root"
			}
			fmt.Fprintf(w, "- `%s`: +%d -%d\n", label, s.added, s.removed)
		}
		fmt.Fprintln(w)
	}

	// Notable changes — top 5 by total lines changed.
	notable := topBySize(all, 5)
	fmt.Fprintln(w, "### Notable changes")
	for _, f := range notable {
		kind := classify(f)
		fmt.Fprintf(w, "- `%s`: %s (+%d lines) — %s\n", f.path, kind, f.added, reason(kind))
	}
	if len(all) > 5 {
		more := len(all) - 5
		fmt.Fprintf(w, "- ...and %d more file%s\n", more, plural(more))
	}
}

// --- Helpers ---

type dirStat struct{ added, removed int }

func groupByTopDir(changes []fileChange) map[string]dirStat {
	result := make(map[string]dirStat)
	for _, f := range changes {
		d := topDir(f.path)
		s := result[d]
		s.added += f.added
		s.removed += f.removed
		result[d] = s
	}
	return result
}

// topDir returns the first path component (directory), or "." for root files.
func topDir(path string) string {
	// Normalize to forward slashes.
	clean := filepath.ToSlash(path)
	idx := strings.Index(clean, "/")
	if idx == -1 {
		return "."
	}
	return clean[:idx]
}

// mergeChanges deduplicates by path, summing stats for files that appear in both lists.
func mergeChanges(a, b []fileChange) []fileChange {
	merged := make(map[string]fileChange)
	for _, f := range a {
		merged[f.path] = f
	}
	for _, f := range b {
		if existing, ok := merged[f.path]; ok {
			existing.added += f.added
			existing.removed += f.removed
			merged[f.path] = existing
		} else {
			merged[f.path] = f
		}
	}
	result := make([]fileChange, 0, len(merged))
	for _, f := range merged {
		result = append(result, f)
	}
	sort.Slice(result, func(i, j int) bool { return result[i].path < result[j].path })
	return result
}

// topBySize returns up to n files sorted by (added+removed) descending.
func topBySize(changes []fileChange, n int) []fileChange {
	sorted := make([]fileChange, len(changes))
	copy(sorted, changes)
	sort.Slice(sorted, func(i, j int) bool {
		ti := sorted[i].added + sorted[i].removed
		tj := sorted[j].added + sorted[j].removed
		return ti > tj
	})
	if n > len(sorted) {
		n = len(sorted)
	}
	return sorted[:n]
}

func classify(f fileChange) string {
	a, r := f.added, f.removed
	switch {
	case a > 30:
		return "large change"
	case a+r <= 10:
		return "small change"
	case r == 0:
		return "pure addition"
	case a == 0:
		return "pure deletion"
	case a > 20 && r > 20:
		return "rewrite"
	default:
		if r > 0 {
			ratio := float64(a) / float64(r)
			if ratio >= 0.5 && ratio <= 2.0 {
				return "refactor"
			}
		}
		return "mixed change"
	}
}

func reason(kind string) string {
	switch kind {
	case "large change":
		return "likely new feature"
	case "small change":
		return "likely refactor or tweak"
	case "pure addition":
		return "new file or additions only"
	case "pure deletion":
		return "deletions only"
	case "rewrite":
		return "significant rewrite"
	case "refactor":
		return "likely refactor"
	default:
		return "mixed additions and removals"
	}
}

// relativeTime returns a human-readable relative duration string.
func relativeTime(t time.Time) string {
	d := time.Since(t)
	if d < 0 {
		d = -d
	}
	switch {
	case d < time.Minute:
		return "just now"
	case d < time.Hour:
		n := int(d.Minutes())
		return fmt.Sprintf("%d minute%s ago", n, plural(n))
	case d < 24*time.Hour:
		n := int(d.Hours())
		return fmt.Sprintf("%d hour%s ago", n, plural(n))
	case d < 7*24*time.Hour:
		n := int(d.Hours() / 24)
		return fmt.Sprintf("%d day%s ago", n, plural(n))
	case d < 30*24*time.Hour:
		n := int(d.Hours() / (24 * 7))
		return fmt.Sprintf("%d week%s ago", n, plural(n))
	case d < 365*24*time.Hour:
		n := int(d.Hours() / (24 * 30))
		return fmt.Sprintf("%d month%s ago", n, plural(n))
	default:
		n := int(d.Hours() / (24 * 365))
		return fmt.Sprintf("%d year%s ago", n, plural(n))
	}
}

func plural(n int) string {
	if n == 1 {
		return ""
	}
	return "s"
}

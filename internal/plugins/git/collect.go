package git

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"
)

// --- Data types ---

type repoInfo struct {
	branch    string
	upstream  string // "origin/main" or "" if no upstream
	ahead     int
	behind    int
	lastFetch time.Time
	hasFetch  bool
}

type commit struct {
	hash    string
	unixTs  int64
	author  string
	subject string
}

type fileChange struct {
	path    string
	added   int
	removed int
}

type workingTree struct {
	modified  []fileChange // unstaged (git diff --numstat)
	staged    []fileChange // staged (git diff --cached --numstat)
	untracked []string
}

type gitData struct {
	repo    repoInfo
	commits []commit
	tree    workingTree
}

// --- Git runner ---

// gitCmd runs a git command in dir and returns trimmed stdout.
// Returns an error if git exits non-zero.
func gitCmd(dir string, args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	var out, errBuf bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &errBuf
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("git %s: %w", strings.Join(args, " "), err)
	}
	return strings.TrimRight(out.String(), "\r\n"), nil
}

// --- Top-level collector ---

// collect gathers all git data in parallel. Returns a friendly error if dir
// is not inside a git repository.
func collect(dir string) (*gitData, error) {
	// Verify we're in a repo before spawning goroutines.
	if _, err := gitCmd(dir, "rev-parse", "--is-inside-work-tree"); err != nil {
		return nil, fmt.Errorf("not a git repository (run from inside a git repo)")
	}

	var (
		data gitData
		mu   sync.Mutex
		wg   sync.WaitGroup
		errs []error
	)

	record := func(err error) {
		if err != nil {
			mu.Lock()
			errs = append(errs, err)
			mu.Unlock()
		}
	}

	wg.Add(3)

	go func() {
		defer wg.Done()
		info, err := collectRepoInfo(dir)
		record(err)
		if err == nil {
			mu.Lock()
			data.repo = info
			mu.Unlock()
		}
	}()

	go func() {
		defer wg.Done()
		commits, err := collectCommits(dir, 5)
		record(err)
		if err == nil {
			mu.Lock()
			data.commits = commits
			mu.Unlock()
		}
	}()

	go func() {
		defer wg.Done()
		tree, err := collectWorkingTree(dir)
		record(err)
		if err == nil {
			mu.Lock()
			data.tree = tree
			mu.Unlock()
		}
	}()

	wg.Wait()
	if len(errs) > 0 {
		return nil, errs[0]
	}
	return &data, nil
}

// --- Individual collectors ---

func collectRepoInfo(dir string) (repoInfo, error) {
	var info repoInfo

	branch, err := gitCmd(dir, "rev-parse", "--abbrev-ref", "HEAD")
	if err != nil {
		return info, err
	}
	info.branch = branch

	// Ahead/behind vs upstream (fails gracefully if no upstream).
	if ab, err := gitCmd(dir, "rev-list", "--left-right", "--count", "HEAD...@{u}"); err == nil {
		parts := strings.Fields(ab)
		if len(parts) == 2 {
			info.ahead, _ = strconv.Atoi(parts[0])
			info.behind, _ = strconv.Atoi(parts[1])
			if u, err := gitCmd(dir, "rev-parse", "--abbrev-ref", "--symbolic-full-name", "@{u}"); err == nil {
				info.upstream = u
			}
		}
	}

	// FETCH_HEAD mtime — find git dir first (handles worktrees, submodules).
	if gitDir, err := gitCmd(dir, "rev-parse", "--git-dir"); err == nil {
		if !filepath.IsAbs(gitDir) {
			gitDir = filepath.Join(dir, gitDir)
		}
		if fi, err := os.Stat(filepath.Join(gitDir, "FETCH_HEAD")); err == nil {
			info.lastFetch = fi.ModTime()
			info.hasFetch = true
		}
	}

	return info, nil
}

func collectCommits(dir string, n int) ([]commit, error) {
	out, err := gitCmd(dir, "log", fmt.Sprintf("-n%d", n), "--pretty=format:%h|%at|%an|%s")
	if err != nil || out == "" {
		// Empty repo or no commits — not a hard error.
		return nil, nil
	}
	var commits []commit
	for _, line := range strings.Split(out, "\n") {
		if line == "" {
			continue
		}
		// SplitN 4 so subject can contain "|".
		parts := strings.SplitN(line, "|", 4)
		if len(parts) != 4 {
			continue
		}
		ts, _ := strconv.ParseInt(parts[1], 10, 64)
		commits = append(commits, commit{
			hash:    parts[0],
			unixTs:  ts,
			author:  parts[2],
			subject: parts[3],
		})
	}
	return commits, nil
}

func collectWorkingTree(dir string) (workingTree, error) {
	var (
		tree workingTree
		wg   sync.WaitGroup
		mu   sync.Mutex
		errs []error
	)

	record := func(err error) {
		if err != nil {
			mu.Lock()
			errs = append(errs, err)
			mu.Unlock()
		}
	}

	wg.Add(3)

	go func() {
		defer wg.Done()
		changes, err := parseNumstat(dir, false)
		record(err)
		mu.Lock()
		tree.modified = changes
		mu.Unlock()
	}()

	go func() {
		defer wg.Done()
		changes, err := parseNumstat(dir, true)
		record(err)
		mu.Lock()
		tree.staged = changes
		mu.Unlock()
	}()

	go func() {
		defer wg.Done()
		untracked, err := parseUntracked(dir)
		record(err)
		mu.Lock()
		tree.untracked = untracked
		mu.Unlock()
	}()

	wg.Wait()
	if len(errs) > 0 {
		return tree, errs[0]
	}
	return tree, nil
}

func parseNumstat(dir string, cached bool) ([]fileChange, error) {
	args := []string{"diff", "--numstat"}
	if cached {
		args = append(args, "--cached")
	}
	out, err := gitCmd(dir, args...)
	if err != nil || out == "" {
		return nil, nil
	}
	var changes []fileChange
	for _, line := range strings.Split(out, "\n") {
		if line == "" {
			continue
		}
		parts := strings.SplitN(line, "\t", 3)
		if len(parts) != 3 {
			continue
		}
		// Binary files show "-" — treat as 0.
		added, _ := strconv.Atoi(parts[0])
		removed, _ := strconv.Atoi(parts[1])
		changes = append(changes, fileChange{
			path:    parts[2],
			added:   added,
			removed: removed,
		})
	}
	return changes, nil
}

func parseUntracked(dir string) ([]string, error) {
	out, err := gitCmd(dir, "status", "--porcelain=v1")
	if err != nil || out == "" {
		return nil, nil
	}
	var untracked []string
	for _, line := range strings.Split(out, "\n") {
		if len(line) >= 3 && line[0] == '?' && line[1] == '?' {
			untracked = append(untracked, strings.TrimSpace(line[3:]))
		}
	}
	return untracked, nil
}

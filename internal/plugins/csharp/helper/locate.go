package helper

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
)

const helperBinary = "ctx-roslyn-helper"

// ErrHelperNotFound is returned when the Roslyn helper binary cannot be located.
var ErrHelperNotFound = errors.New("roslyn helper not found")

// LocateHelper searches for the ctx-roslyn-helper binary using four strategies:
//  1. CTX_ROSLYN_HELPER environment variable
//  2. Same directory as the running ctx binary
//  3. PATH
//  4. tools/roslyn-helper/publish/ relative to working directory
//
// Returns the absolute path or ErrHelperNotFound with a diagnostic message.
func LocateHelper() (string, error) {
	name := helperBinary
	if runtime.GOOS == "windows" {
		name += ".exe"
	}

	tried := []string{}

	// 1. Environment variable
	if env := os.Getenv("CTX_ROSLYN_HELPER"); env != "" {
		if fileExists(env) {
			return env, nil
		}
		tried = append(tried, fmt.Sprintf("$CTX_ROSLYN_HELPER = %s (not found)", env))
	} else {
		tried = append(tried, "$CTX_ROSLYN_HELPER (not set)")
	}

	// 2. Same directory as ctx binary
	if exePath, err := os.Executable(); err == nil {
		exeDir := filepath.Dir(exePath)
		candidate := filepath.Join(exeDir, name)
		if fileExists(candidate) {
			return candidate, nil
		}
		tried = append(tried, fmt.Sprintf("%s (not found)", exeDir))
	}

	// 3. PATH
	if found, err := exec.LookPath(name); err == nil {
		return found, nil
	}
	tried = append(tried, "PATH (not found)")

	// 4. Dev fallback: tools/roslyn-helper/publish/ relative to working dir
	if cwd, err := os.Getwd(); err == nil {
		candidate := filepath.Join(cwd, "tools", "roslyn-helper", "publish", name)
		if fileExists(candidate) {
			return candidate, nil
		}
		tried = append(tried, fmt.Sprintf("tools/roslyn-helper/publish/ (not found)"))
	}

	return "", fmt.Errorf("%w\n\nLooked in:\n  - %s\n  - %s\n  - %s\n  - %s\n\nTo build the helper, run:\n  cd tools\\roslyn-helper\n  dotnet publish src/RoslynHelper -c Release -r win-x64 --self-contained false -o publish/",
		ErrHelperNotFound,
		safeIdx(tried, 0, "$CTX_ROSLYN_HELPER (not set)"),
		safeIdx(tried, 1, "ctx.exe directory"),
		safeIdx(tried, 2, "PATH"),
		safeIdx(tried, 3, "tools/roslyn-helper/publish/"),
	)
}

func fileExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}

func safeIdx(s []string, i int, fallback string) string {
	if i < len(s) {
		return s[i]
	}
	return fallback
}

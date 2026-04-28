# ctx pre-commit hook (PowerShell)
# To activate: git config core.hooksPath .githooks
$ErrorActionPreference = "Stop"

Write-Host "Running gofmt..."
$unformatted = gofmt -l .
if ($unformatted) {
    Write-Host "Files not formatted with gofmt:"
    Write-Host $unformatted
    Write-Host ""
    Write-Host "Run: gofmt -w ."
    exit 1
}

Write-Host "Running go vet..."
go vet ./...

if (Get-Command golangci-lint -ErrorAction SilentlyContinue) {
    Write-Host "Running golangci-lint..."
    golangci-lint run ./...
} else {
    Write-Host "warning: golangci-lint not installed, skipping"
}

Write-Host "pre-commit checks passed"

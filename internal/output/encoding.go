// Package output provides helpers for writing consistent markdown to stdout.
package output

// UTF-8 and BOM notes:
//
// Go's string type is UTF-8 by default, and os.Stdout writes raw bytes.
// On Windows, some programs write a UTF-8 BOM (0xEF 0xBB 0xBF) to signal
// encoding, but Claude and most Unix tools do not expect or want a BOM.
//
// ctx never writes a BOM. All output is plain UTF-8. If a future caller
// needs a file with BOM (e.g. for Excel compatibility), that's a caller
// responsibility — not this package's job.
//
// Git Bash on Windows already uses UTF-8 for stdout when piped, so no
// runtime encoding conversion is needed.

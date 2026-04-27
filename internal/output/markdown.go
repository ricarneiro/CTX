package output

import (
	"fmt"
	"io"
	"strings"
)

// H1 writes a level-1 markdown heading followed by a blank line.
func H1(w io.Writer, text string) {
	fmt.Fprintf(w, "# %s\n\n", text)
}

// H2 writes a level-2 markdown heading followed by a blank line.
func H2(w io.Writer, text string) {
	fmt.Fprintf(w, "## %s\n\n", text)
}

// H3 writes a level-3 markdown heading followed by a blank line.
func H3(w io.Writer, text string) {
	fmt.Fprintf(w, "### %s\n\n", text)
}

// CodeBlock writes a fenced code block with an optional language identifier.
// If language is empty, the fence has no language tag.
func CodeBlock(w io.Writer, language, content string) {
	fmt.Fprintf(w, "```%s\n%s\n```\n\n", language, strings.TrimRight(content, "\n"))
}

// KeyValue writes a single "**key:** value" line followed by a newline.
func KeyValue(w io.Writer, key, value string) {
	fmt.Fprintf(w, "**%s:** %s\n", key, value)
}

// BulletList writes a markdown bullet list. Each item gets its own line.
// A blank line is written after the list.
func BulletList(w io.Writer, items []string) {
	for _, item := range items {
		fmt.Fprintf(w, "- %s\n", item)
	}
	fmt.Fprintln(w)
}

// Section writes an H2 heading then calls body(w), then writes a blank line.
// Use this to group related output under a named section.
func Section(w io.Writer, title string, body func(w io.Writer)) {
	H2(w, title)
	body(w)
	fmt.Fprintln(w)
}

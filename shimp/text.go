// text.go -- wrapping a refusal so it is readable in a console.
//
// ⚠️ PLATFORM-INDEPENDENT, and it did not start that way. These lived in
// refusal_windows.go, which left `GOOS=linux go vet` failing on a test that
// referenced them: nothing about breaking text at word boundaries is
// Windows-specific, and putting it behind a build tag made the whole file
// look like it was.
//
// Caught by running vet for another platform, which is worth keeping in the
// loop for exactly this: a build tag is easy to inherit by accident from the
// file something was written next to.
package main

// wrap breaks text at word boundaries so a long refusal is readable in a
// console that may be any width. 72 is narrow enough for the default window.
func wrap(s string, width int) string {
	var out, line []rune
	for _, word := range splitWords(s) {
		if len(line) > 0 && len(line)+1+len([]rune(word)) > width {
			out = append(out, line...)
			out = append(out, '\n', ' ', ' ')
			line = line[:0]
		}
		if len(line) > 0 {
			line = append(line, ' ')
		}
		line = append(line, []rune(word)...)
	}
	return string(append(out, line...))
}

func splitWords(s string) []string {
	var out []string
	cur := make([]rune, 0, 16)
	for _, r := range s {
		if r == ' ' || r == '\n' || r == '\t' {
			if len(cur) > 0 {
				out = append(out, string(cur))
				cur = cur[:0]
			}
			continue
		}
		cur = append(cur, r)
	}
	if len(cur) > 0 {
		out = append(out, string(cur))
	}
	return out
}

package report

import "strings"

// CommentMarker is the hidden line that lets a later run find the comment it wrote.
const CommentMarker = "<!-- glassops-report -->"

// commentLimit is the longest comment body sent. GitHub rejects bodies over 65536 characters.
const commentLimit = 60000

// Comment renders the pull request comment: the marker, then the same redacted markdown as the
// step summary. A body over commentLimit runes is cut at a line end with a note.
func Comment(b *Bundle) string {
	body := CommentMarker + "\n" + Summary(b)
	r := []rune(body)
	if len(r) <= commentLimit {
		return body
	}
	cut := string(r[:commentLimit])
	if i := strings.LastIndex(cut, "\n"); i > 0 {
		cut = cut[:i]
	}
	return cut + "\n\n_The report was cut to fit the comment size limit. The run summary has the rest._\n"
}

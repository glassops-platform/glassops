package report

import (
	"strings"
	"testing"

	"github.com/glassops-platform/glassops/internal/core/run"
)

func TestCommentStartsWithMarkerAndRedacts(t *testing.T) {
	b := &Bundle{Results: []run.Result{{
		Step: "validate", Status: run.StatusFail,
		Errors: []string{"deploy as jane.doe@example.com failed"},
	}}}
	got := Comment(b)
	if !strings.HasPrefix(got, CommentMarker+"\n") {
		t.Fatalf("no marker first:\n%s", got)
	}
	if strings.Contains(got, "jane.doe@example.com") {
		t.Fatalf("email reached the comment:\n%s", got)
	}
	if !strings.Contains(got, Summary(b)) {
		t.Fatal("comment does not hold the step summary")
	}
}

func TestCommentIsCutToTheSizeLimit(t *testing.T) {
	b := &Bundle{}
	for i := 0; i < 3000; i++ {
		b.Results = append(b.Results, run.Result{Step: "step", Status: run.StatusPass,
			Errors: []string{strings.Repeat("x", 80)}})
	}
	got := Comment(b)
	if n := len([]rune(got)); n > 65536 {
		t.Fatalf("comment has %d characters", n)
	}
	if !strings.Contains(got, "cut to fit") || !strings.HasPrefix(got, CommentMarker) {
		t.Fatal("cut note or marker missing")
	}
}

package policy

import (
	"testing"
	"time"
)

// window.go windowEnd: `local.Hour()*60+local.Minute() >= from`. The evening part of an
// overnight window ends tomorrow. 22:30 kills `Hour*60 - Minute`, and the start minute kills `>`.
func TestMutWindowEndEveningPart(t *testing.T) {
	cfg := windowCfg(win([]string{"fri"}, "22:00", "06:00", "UTC"))
	// 2027-01-08 is a Friday.
	for _, now := range []string{"2027-01-08T22:00:00Z", "2027-01-08T22:30:00Z", "2027-01-08T23:15:00Z"} {
		end, ok, err := WindowEnd(cfg, "prod", at(t, now))
		if err != nil || !ok {
			t.Fatalf("%s: ok %v err %v", now, ok, err)
		}
		if want := at(t, "2027-01-09T06:00:00Z"); !end.Equal(want) {
			t.Fatalf("%s: end %s, want %s", now, end, want)
		}
	}
	// The morning part belongs to the day the window started on and ends the same day.
	end, ok, err := WindowEnd(cfg, "prod", at(t, "2027-01-09T01:00:00Z"))
	if err != nil || !ok || !end.Equal(at(t, "2027-01-09T06:00:00Z")) {
		t.Fatalf("morning part: %s %v %v", end, ok, err)
	}
}

// window.go WindowEnd: `i <= 8*len(windows)`. Two windows that touch around the clock, every
// day, never stop extending, so the bound decides where the chain ends. With one window the
// bound is 8 passes of 12 hours past the 9th, and the loop runs 17 passes in all for two windows.
func TestMutWindowEndChainBound(t *testing.T) {
	cfg := windowCfg(
		win(allDays, "00:00", "12:00", "UTC"),
		win(allDays, "12:00", "00:00", "UTC"),
	)
	now := at(t, "2027-01-04T01:00:00Z")
	end, ok, err := WindowEnd(cfg, "prod", now)
	if err != nil || !ok {
		t.Fatalf("ok %v err %v", ok, err)
	}
	// The first pass ends the window holding now (12:00). Each of the 16 further passes adds 12
	// hours, 17 passes in all.
	want := at(t, "2027-01-04T12:00:00Z").Add(16 * 12 * time.Hour)
	if !end.Equal(want) {
		t.Fatalf("end %s, want %s", end, want)
	}
}

package policy

import (
	"fmt"
	"strings"
	"time"
	// Embed the time zone database so windows do not depend on the host.
	_ "time/tzdata"

	"github.com/glassops-platform/glassops/internal/core/config"
)

// Window allows changes to a tier only inside its configured weekly windows.
type Window struct{}

// Name returns the rule name. Results are named window.<tier>.
func (Window) Name() string { return "window" }

var dayNames = [...]string{"sun", "mon", "tue", "wed", "thu", "fri", "sat"}

func dayName(d time.Weekday) string { return dayNames[d] }

// clockMinutes parses HH:MM into minutes after midnight.
func clockMinutes(s string) (int, error) {
	t, err := time.Parse("15:04", s)
	if err != nil {
		return 0, fmt.Errorf("clock %q is not HH:MM", s)
	}
	return t.Hour()*60 + t.Minute(), nil
}

// inWindow reports whether now falls inside w. Times are wall clock times in the window's
// zone, so a daylight saving change moves the window with the clock on the wall. A window
// whose to is earlier than its from runs past midnight, and belongs to the day it starts on.
func inWindow(w config.Window, now time.Time) (bool, error) {
	loc, err := time.LoadLocation(w.TZ)
	if err != nil {
		return false, fmt.Errorf("unknown time zone %q", w.TZ)
	}
	from, err := clockMinutes(w.From)
	if err != nil {
		return false, err
	}
	to, err := clockMinutes(w.To)
	if err != nil {
		return false, err
	}
	local := now.In(loc)
	mins := local.Hour()*60 + local.Minute()
	today := dayName(local.Weekday())
	yesterday := dayName((local.Weekday() + 6) % 7)
	has := func(day string) bool {
		for _, d := range w.Days {
			if strings.EqualFold(d, day) {
				return true
			}
		}
		return false
	}
	if from < to {
		return has(today) && mins >= from && mins < to, nil
	}
	if from == to {
		return false, nil
	}
	// Crosses midnight: the evening part belongs to today, the morning part to yesterday.
	return (has(today) && mins >= from) || (has(yesterday) && mins < to), nil
}

// Evaluate checks the windows of the target's tier.
func (Window) Evaluate(in Input) []Result {
	tier, ok := tierOf(in)
	if !ok {
		if in.Config != nil && len(in.Config.Policy.Windows) > 0 {
			return one("window", Fail, "the facts do not name the target tier, so its window cannot be checked")
		}
		return one("window", Skipped, "no windows are configured")
	}
	rule := "window." + tier
	var windows []config.Window
	if in.Config != nil {
		windows = in.Config.Policy.Windows[tier]
	}
	if len(windows) == 0 {
		return one(rule, Skipped, "no windows are configured for tier %s", tier)
	}
	for _, w := range windows {
		inside, err := inWindow(w, in.Now)
		if err != nil {
			return one(rule, Fail, "window for tier %s is not usable: %v", tier, err)
		}
		if inside {
			return one(rule, Pass, "inside the %s-%s %s window", w.From, w.To, w.TZ)
		}
	}
	return one(rule, Fail, "outside every window for tier %s at %s", tier, in.Now.UTC().Format(time.RFC3339))
}

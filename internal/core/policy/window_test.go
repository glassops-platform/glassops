package policy

import (
	"strings"
	"testing"
	"time"

	"github.com/glassops-platform/glassops/internal/core/config"
)

func windowCfg(ws ...config.Window) *config.Config {
	return cfgWith(func(c *config.Config) {
		c.Policy.Windows = map[string][]config.Window{"prod": ws}
	})
}

var allDays = []string{"mon", "tue", "wed", "thu", "fri", "sat", "sun"}

func win(days []string, from, to, tz string) config.Window {
	return config.Window{Days: days, From: from, To: to, TZ: tz}
}

func TestWindow(t *testing.T) {
	weekdays := win([]string{"mon", "tue", "wed", "thu"}, "09:00", "17:00", "America/New_York")
	overnight := win([]string{"fri"}, "22:00", "06:00", "UTC")
	nySpring := win([]string{"sun"}, "02:30", "04:00", "America/New_York")
	nyFall := win([]string{"sun"}, "01:00", "02:00", "America/New_York")
	lonSpring := win([]string{"sun"}, "00:30", "03:00", "Europe/London")
	lonFall := win([]string{"sun"}, "01:00", "02:00", "Europe/London")
	tests := []struct {
		name string
		cfg  *config.Config
		tier string
		now  string
		want Outcome
		rule string
	}{
		// 2027-01-04 is a Monday. 14:00Z is 09:00 EST.
		{"inside, on the start minute", windowCfg(weekdays), "prod", "2027-01-04T14:00:00Z", Pass, "window.prod"},
		{"inside, mid day", windowCfg(weekdays), "prod", "2027-01-04T18:00:00Z", Pass, "window.prod"},
		{"before the window", windowCfg(weekdays), "prod", "2027-01-04T13:59:00Z", Fail, "window.prod"},
		{"the end minute is outside", windowCfg(weekdays), "prod", "2027-01-04T22:00:00Z", Fail, "window.prod"},
		{"last minute inside", windowCfg(weekdays), "prod", "2027-01-04T21:59:00Z", Pass, "window.prod"},
		{"wrong day", windowCfg(weekdays), "prod", "2027-01-08T18:00:00Z", Fail, "window.prod"},
		{"zone decides the day, not UTC", windowCfg(weekdays), "prod", "2027-01-05T02:00:00Z", Fail, "window.prod"},
		{"zone decides the day, late Thursday", windowCfg(weekdays), "prod", "2027-01-08T00:00:00Z", Fail, "window.prod"},
		{"no windows for the tier", windowCfg(weekdays), "qa", "2027-01-08T18:00:00Z", Skipped, "window.qa"},
		{"no windows at all", &config.Config{}, "prod", "2027-01-08T18:00:00Z", Skipped, "window.prod"},
		{"second window matches", windowCfg(win([]string{"sat"}, "09:00", "10:00", "UTC"), weekdays), "prod", "2027-01-04T18:00:00Z", Pass, "window.prod"},
		// 2027-01-08 is a Friday, so the window runs Fri 22:00 to Sat 06:00 UTC.
		{"overnight, evening part", windowCfg(overnight), "prod", "2027-01-08T23:00:00Z", Pass, "window.prod"},
		{"overnight, morning part", windowCfg(overnight), "prod", "2027-01-09T05:59:00Z", Pass, "window.prod"},
		{"overnight, morning end", windowCfg(overnight), "prod", "2027-01-09T06:00:00Z", Fail, "window.prod"},
		{"overnight, before start", windowCfg(overnight), "prod", "2027-01-08T21:59:00Z", Fail, "window.prod"},
		{"overnight, Sunday morning belongs to Saturday", windowCfg(overnight), "prod", "2027-01-10T03:00:00Z", Fail, "window.prod"},
		{"overnight, Friday morning belongs to Thursday", windowCfg(overnight), "prod", "2027-01-08T03:00:00Z", Fail, "window.prod"},
		{"every day window", windowCfg(win(allDays, "00:00", "23:59", "UTC")), "prod", "2027-01-08T12:00:00Z", Pass, "window.prod"},
		// America/New_York spring forward: 2027-03-14 (Sunday), 02:00 EST jumps to 03:00 EDT at 07:00Z.
		{"NY spring forward, after the jump", windowCfg(nySpring), "prod", "2027-03-14T07:30:00Z", Pass, "window.prod"},
		{"NY spring forward, before the jump", windowCfg(nySpring), "prod", "2027-03-14T06:59:00Z", Fail, "window.prod"},
		{"NY spring forward, after the window", windowCfg(nySpring), "prod", "2027-03-14T08:00:00Z", Fail, "window.prod"},
		// Fall back: 2027-11-07, 02:00 EDT returns to 01:00 EST at 06:00Z, so 01:30 happens twice.
		{"NY fall back, first 01:30", windowCfg(nyFall), "prod", "2027-11-07T05:30:00Z", Pass, "window.prod"},
		{"NY fall back, second 01:30", windowCfg(nyFall), "prod", "2027-11-07T06:30:00Z", Pass, "window.prod"},
		{"NY fall back, 02:00 EST is out", windowCfg(nyFall), "prod", "2027-11-07T07:00:00Z", Fail, "window.prod"},
		// Europe/London: spring forward 2027-03-28 at 01:00Z, fall back 2027-10-31 at 01:00Z.
		{"London spring forward, GMT side", windowCfg(lonSpring), "prod", "2027-03-28T00:45:00Z", Pass, "window.prod"},
		{"London spring forward, BST side", windowCfg(lonSpring), "prod", "2027-03-28T01:30:00Z", Pass, "window.prod"},
		{"London spring forward, end is 03:00 BST", windowCfg(lonSpring), "prod", "2027-03-28T02:00:00Z", Fail, "window.prod"},
		{"London fall back, BST 01:30", windowCfg(lonFall), "prod", "2027-10-31T00:30:00Z", Pass, "window.prod"},
		{"London fall back, GMT 01:30", windowCfg(lonFall), "prod", "2027-10-31T01:30:00Z", Pass, "window.prod"},
		{"London fall back, 02:00 GMT is out", windowCfg(lonFall), "prod", "2027-10-31T02:00:00Z", Fail, "window.prod"},
		{"overnight across the London change", windowCfg(win([]string{"sat"}, "23:00", "03:00", "Europe/London")), "prod", "2027-03-28T01:30:00Z", Pass, "window.prod"},
		{"bad zone fails", windowCfg(win(allDays, "09:00", "17:00", "Not/AZone")), "prod", "2027-01-04T18:00:00Z", Fail, "window.prod"},
		{"bad clock fails", windowCfg(win(allDays, "9am", "17:00", "UTC")), "prod", "2027-01-04T18:00:00Z", Fail, "window.prod"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			in := Input{Facts: doc(map[string]any{"target.tier": tc.tier}), Config: tc.cfg, Now: at(t, tc.now)}
			r := one1(t, Window{}.Evaluate(in))
			if r.Outcome != tc.want || r.Rule != tc.rule {
				t.Fatalf("got %+v, want %s %s", r, tc.want, tc.rule)
			}
		})
	}
}

func TestWindowWeekdayCheck(t *testing.T) {
	// Guards the dates the table above relies on.
	for s, want := range map[string]time.Weekday{
		"2027-01-04": time.Monday, "2027-01-08": time.Friday, "2027-03-14": time.Sunday,
		"2027-11-07": time.Sunday, "2027-03-28": time.Sunday, "2027-10-31": time.Sunday,
	} {
		d, _ := time.Parse(time.DateOnly, s)
		if d.Weekday() != want {
			t.Errorf("%s is %s, want %s", s, d.Weekday(), want)
		}
	}
}

func TestWindowMissingTier(t *testing.T) {
	cfg := windowCfg(win(allDays, "00:00", "23:59", "UTC"))
	r := one1(t, Window{}.Evaluate(Input{Facts: doc(nil), Config: cfg, Now: at(t, "2027-01-04T12:00:00Z")}))
	if r.Outcome != Fail || !strings.Contains(r.Detail, "tier") {
		t.Fatalf("got %+v", r)
	}
	r = one1(t, Window{}.Evaluate(Input{Facts: doc(nil), Config: &config.Config{}, Now: at(t, "2027-01-04T12:00:00Z")}))
	if r.Outcome != Skipped {
		t.Fatalf("no windows configured and no tier: %+v", r)
	}
}

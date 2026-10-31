package cli

import (
	"context"
	"errors"
	"sync"

	"github.com/glassops-platform/glassops/internal/providers"
)

// fakeProvider is a provider whose behaviour each test sets. Unset funcs fail. It records
// sessions so tests can check every one was closed.
type fakeProvider struct {
	name       string
	detect     bool
	auth       func(providers.Target) error
	describe   func(providers.Target) (providers.TargetInfo, error)
	plan       func(providers.PlanRequest) (providers.ChangeSet, error)
	validateFn func(providers.Target, providers.ChangeSet, providers.ValidateOptions) (providers.ValidateResult, error)
	deploy     func(providers.Session, providers.ChangeSet, providers.DeployOptions) (providers.DeployResult, error)

	mu          sync.Mutex
	sessions    []*fakeSession
	deployCalls int
}

type fakeSession struct {
	target providers.Target
	closed int
}

func (s *fakeSession) Target() providers.Target { return s.target }
func (s *fakeSession) Close() error             { s.closed++; return nil }

func (f *fakeProvider) Name() string {
	if f.name == "" {
		return "salesforce"
	}
	return f.name
}
func (f *fakeProvider) Detect(string) (bool, error) { return f.detect, nil }
func (f *fakeProvider) Tools() []providers.Tool     { return nil }
func (f *fakeProvider) CanonicalID(id string) string {
	if len(id) > 15 {
		return id[:15]
	}
	return id
}

func (f *fakeProvider) Authenticate(_ context.Context, t providers.Target) (providers.Session, error) {
	if f.auth == nil {
		return nil, errors.New("fake: no auth set")
	}
	if err := f.auth(t); err != nil {
		return nil, err
	}
	s := &fakeSession{target: t}
	f.mu.Lock()
	f.sessions = append(f.sessions, s)
	f.mu.Unlock()
	return s, nil
}

func (f *fakeProvider) Describe(_ context.Context, s providers.Session) (providers.TargetInfo, error) {
	if f.describe == nil {
		return providers.TargetInfo{}, errors.New("fake: no describe set")
	}
	return f.describe(s.Target())
}

func (f *fakeProvider) Plan(_ context.Context, req providers.PlanRequest) (providers.ChangeSet, error) {
	if f.plan == nil {
		return providers.ChangeSet{}, errors.New("fake: no plan set")
	}
	return f.plan(req)
}

func (f *fakeProvider) Validate(_ context.Context, s providers.Session, cs providers.ChangeSet, o providers.ValidateOptions) (providers.ValidateResult, error) {
	if f.validateFn == nil {
		return providers.ValidateResult{}, errors.New("fake: no validate set")
	}
	return f.validateFn(s.Target(), cs, o)
}

func (f *fakeProvider) Deploy(_ context.Context, s providers.Session, cs providers.ChangeSet, o providers.DeployOptions) (providers.DeployResult, error) {
	f.mu.Lock()
	f.deployCalls++
	f.mu.Unlock()
	if f.deploy == nil {
		return providers.DeployResult{}, errors.New("fake: no deploy set")
	}
	return f.deploy(s, cs, o)
}

// allClosed reports whether every session the fake opened was closed at least once.
func (f *fakeProvider) allClosed() bool {
	f.mu.Lock()
	defer f.mu.Unlock()
	for _, s := range f.sessions {
		if s.closed == 0 {
			return false
		}
	}
	return true
}

// reportsAs returns a describe func that reports the target's own configured Id and kind k,
// so the guard passes when k suits the tier.
func reportsAs(k providers.Kind) func(providers.Target) (providers.TargetInfo, error) {
	return func(t providers.Target) (providers.TargetInfo, error) {
		return providers.TargetInfo{ID: t.ID, Kind: k, Facts: map[string]any{}}, nil
	}
}

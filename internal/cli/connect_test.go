package cli

import (
	"context"
	"errors"
	"testing"

	"github.com/glassops-platform/glassops/internal/core/guard"
	"github.com/glassops-platform/glassops/internal/providers"
)

func TestConnect(t *testing.T) {
	h := newHarness(t)
	target := providers.Target{Name: "devhub", ID: "00D000000000001AAA", Tier: "devhub"}

	ok := &fakeProvider{auth: func(providers.Target) error { return nil }, describe: reportsAs(providers.KindProduction)}
	s, info, err := h.app.connect(context.Background(), ok, target)
	if err != nil || s == nil || info.ID != target.ID {
		t.Fatalf("connect: %v", err)
	}
	_ = s.Close()

	wrongKind := &fakeProvider{auth: func(providers.Target) error { return nil }, describe: reportsAs(providers.KindSandbox)}
	_, _, err = h.app.connect(context.Background(), wrongKind, target)
	var m *guard.MismatchError
	if !errors.As(err, &m) || !wrongKind.allClosed() {
		t.Fatalf("mismatch: err=%v closed=%v", err, wrongKind.allClosed())
	}

	noLogin := &fakeProvider{auth: func(providers.Target) error { return errors.New("invalid_grant") }}
	if _, _, err = h.app.connect(context.Background(), noLogin, target); !errors.Is(err, ErrLogin) {
		t.Fatalf("login failure: %v", err)
	}
}

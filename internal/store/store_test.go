package store

import (
	"context"
	"errors"
	"path/filepath"
	"testing"
	"time"

	"sherd-proof/internal/domain"
)

func TestRevisionEventsAndIdempotency(t *testing.T) {
	s, err := Open(filepath.Join(t.TempDir(), "store.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	ctx := context.Background()
	now := time.Now().UTC()
	c, err := domain.NewCase("C1", "T1", "灰陶", "owner", now)
	if err != nil {
		t.Fatal(err)
	}
	if err = s.Create(ctx, c, EventInput{Kind: "case.created", ActorID: "owner", OccurredAt: now, Payload: map[string]string{"case": "C1"}}); err != nil {
		t.Fatal(err)
	}
	if err = s.Save(ctx, c, 0, EventInput{Kind: "bad", ActorID: "owner", OccurredAt: now, Payload: nil}); !errors.Is(err, ErrRevisionConflict) {
		t.Fatalf("应为修订冲突，得到 %v", err)
	}
	if err = s.Save(ctx, c, 1, EventInput{Kind: "case.touched", ActorID: "owner", OccurredAt: now, Payload: map[string]bool{"ok": true}}); err != nil {
		t.Fatal(err)
	}
	head, err := s.VerifyEventChain(ctx, "C1")
	if err != nil || head == "" {
		t.Fatalf("摘要链失败: %s %v", head, err)
	}
	body := []byte(`{"value":1}`)
	result := []byte(`{"ok":true}`)
	if err := s.Remember(ctx, "req-1", body, 200, result); err != nil {
		t.Fatal(err)
	}
	replayed, err := s.Replay(ctx, "req-1", body)
	if err != nil || string(replayed.Body) != string(result) {
		t.Fatalf("重放失败: %#v %v", replayed, err)
	}
	if _, err := s.Replay(ctx, "req-1", []byte(`{"value":2}`)); !errors.Is(err, ErrIdempotencyConflict) {
		t.Fatalf("载荷冲突未识别: %v", err)
	}
}

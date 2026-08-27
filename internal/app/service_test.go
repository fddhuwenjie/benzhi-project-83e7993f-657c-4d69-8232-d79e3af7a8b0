package app

import (
	"context"
	"path/filepath"
	"testing"

	"sherd-proof/internal/domain"
	"sherd-proof/internal/store"
)

func TestCreateReplayAndStaleRevision(t *testing.T) {
	repository, err := store.Open(filepath.Join(t.TempDir(), "app.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer repository.Close()
	service := NewService(repository)
	ctx := context.Background()
	command := CreateCaseCommand{RequestMeta: RequestMeta{RequestID: "r1"}, CaseID: "C1", SiteUnit: "T1", VesselClass: "灰陶", OwnerID: "owner"}
	first, err := service.CreateCase(ctx, command)
	if err != nil {
		t.Fatal(err)
	}
	replayed, err := service.CreateCase(ctx, command)
	if err != nil {
		t.Fatal(err)
	}
	if first.Case.Revision != replayed.Case.Revision || replayed.Case.Revision != 1 {
		t.Fatalf("幂等重放修订号错误")
	}
	sherd := domain.SherdRecord{SherdID: "S1", ContextCode: "T1H1", FabricCode: "F1", RimProfile: "完整", Dimensions: domain.Dimensions{Height: 1, Width: 2, Depth: 3}, ImageRef: "archive://S1", ImageDigest: "7c222fb2927d828af22f592134e8932480637c0d1c89b118e45e509fa60f4322"}
	_, err = service.AddSherd(ctx, AddSherdCommand{RequestMeta: RequestMeta{RequestID: "r2", ExpectedRevision: 99}, CaseID: "C1", ActorID: "owner", Sherd: sherd})
	if !domain.IsCode(err, domain.CodeConflict) {
		t.Fatalf("陈旧修订应冲突，得到 %v", err)
	}
	conflict := command
	conflict.SiteUnit = "T2"
	if _, err = service.CreateCase(ctx, conflict); !domain.IsCode(err, domain.CodeConflict) {
		t.Fatalf("幂等载荷冲突未识别: %v", err)
	}
}

package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"sherd-proof/internal/app"
	"sherd-proof/internal/store"
	webapi "sherd-proof/internal/web"
)

type selftestClient struct {
	baseURL  string
	client   *http.Client
	revision int64
}

func runSelftest(address string) error {
	temporary, err := os.MkdirTemp("", "sherd-proof-selftest-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(temporary)
	repository, err := store.Open(filepath.Join(temporary, "selftest.db"))
	if err != nil {
		return err
	}
	defer repository.Close()
	service := app.NewService(repository)
	server := newHTTPServer(address, webapi.New(service))
	listener, err := net.Listen("tcp", address)
	if err != nil {
		return fmt.Errorf("自检监听 %s: %w", address, err)
	}
	serveErrors := make(chan error, 1)
	go func() { serveErrors <- server.Serve(listener) }()
	client := &selftestClient{baseURL: "http://" + listener.Addr().String(), client: &http.Client{Timeout: 5 * time.Second}}
	flowErr := client.runFlow()
	shutdownContext, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	shutdownErr := server.Shutdown(shutdownContext)
	cancel()
	serveErr := <-serveErrors
	if flowErr != nil {
		return flowErr
	}
	if shutdownErr != nil {
		return shutdownErr
	}
	if serveErr != nil && serveErr != http.ErrServerClosed {
		return serveErr
	}
	fmt.Println("自检通过：真实 HTTP 流程已完成建档、评议、独立复核、定稿与校验")
	return nil
}

func (c *selftestClient) runFlow() error {
	if _, err := c.get("/healthz"); err != nil {
		return err
	}
	view, err := c.post("/api/cases", map[string]any{"request_id": "self-01", "expected_revision": 0, "case_id": "SELFTEST-CASE", "site_unit": "T1-H3", "vessel_class": "夹砂灰陶罐", "owner_id": "owner-a"})
	if err != nil {
		return err
	}
	c.captureRevision(view)
	digestA := "7c222fb2927d828af22f592134e8932480637c0d1c89b118e45e509fa60f4322"
	digestB := "2f222fb2927d828af22f592134e8932480637c0d1c89b118e45e509fa60f4399"
	for index, sherd := range []map[string]any{
		{"sherd_id": "S-001", "context_code": "T1H3-01", "fabric_code": "F-GRIT-2", "rim_profile": "外侈口沿残片", "dimensions_mm": map[string]float64{"height": 43.2, "width": 58.4, "depth": 7.1}, "image_ref": "archive://self/S-001", "image_digest": digestA},
		{"sherd_id": "S-002", "context_code": "T1H3-02", "fabric_code": "F-GRIT-2", "rim_profile": "外侈口沿残片", "dimensions_mm": map[string]float64{"height": 39.8, "width": 51.6, "depth": 7.0}, "image_ref": "archive://self/S-002", "image_digest": digestB},
	} {
		view, err = c.post("/api/cases/SELFTEST-CASE/sherds", map[string]any{"request_id": fmt.Sprintf("self-1%d", index), "expected_revision": c.revision, "actor_id": "owner-a", "sherd": sherd})
		if err != nil {
			return err
		}
		c.captureRevision(view)
	}
	view, err = c.post("/api/cases/SELFTEST-CASE/freeze", c.command("self-20", "owner-a"))
	if err != nil {
		return err
	}
	c.captureRevision(view)
	view, err = c.post("/api/cases/SELFTEST-CASE/hypotheses", map[string]any{"request_id": "self-21", "expected_revision": c.revision, "actor_id": "editor-b", "hypothesis_id": "H-001", "sherd_ids": []string{"S-001", "S-002"}, "evidence": map[string]any{"edge_match": "断口三处凸凹点连续对应", "fabric_match": "胎色与石英砂粒级一致", "decoration_continuity": "两道弦纹跨断口连续", "scale_measurements": map[string]float64{"chord_mm": 31.7, "gap_mm": 0.4}, "image_refs": []string{"archive://self/join-overlay"}}})
	if err != nil {
		return err
	}
	c.captureRevision(view)
	view, err = c.post("/api/cases/SELFTEST-CASE/hypotheses/H-001/submit", c.command("self-22", "editor-b"))
	if err != nil {
		return err
	}
	c.captureRevision(view)
	view, err = c.post("/api/cases/SELFTEST-CASE/challenges", map[string]any{"request_id": "self-23", "expected_revision": c.revision, "actor_id": "critic-c", "challenge_id": "C-001", "hypothesis_id": "H-001", "evidence_key": "edge_match", "statement": "第二对应点可能来自近似磨损，需要补充微距观察"})
	if err != nil {
		return err
	}
	c.captureRevision(view)
	view, err = c.post("/api/cases/SELFTEST-CASE/challenges/C-001/resolve", map[string]any{"request_id": "self-24", "expected_revision": c.revision, "actor_id": "editor-b", "resolution_kind": "supplement", "resolution_note": "补入微距观察，确认断口新鲜面颗粒逐点咬合", "replacement": "三处宏观点及五处微观颗粒逐点对应"})
	if err != nil {
		return err
	}
	c.captureRevision(view)
	view, err = c.post("/api/cases/SELFTEST-CASE/request-review", c.command("self-25", "owner-a"))
	if err != nil {
		return err
	}
	c.captureRevision(view)
	view, err = c.post("/api/cases/SELFTEST-CASE/review", map[string]any{"request_id": "self-26", "expected_revision": c.revision, "reviewer_id": "reviewer-z", "decision": "approve", "reason": "证据链完整，异议补证可复核，支持该拼合假说"})
	if err != nil {
		return err
	}
	c.captureRevision(view)
	dossier, err := c.post("/api/cases/SELFTEST-CASE/finalize", map[string]any{"request_id": "self-27", "expected_revision": c.revision, "actor_id": "archivist-q", "dossier_id": "D-SELFTEST-001"})
	if err != nil {
		return err
	}
	if valid, _ := dossier["valid"].(bool); !valid {
		return fmt.Errorf("定稿响应未通过校验")
	}
	verified, err := c.get("/api/cases/SELFTEST-CASE/dossier/verify")
	if err != nil {
		return err
	}
	if valid, _ := verified["valid"].(bool); !valid {
		return fmt.Errorf("档案验证端点报告无效")
	}
	return nil
}

func (c *selftestClient) command(requestID, actorID string) map[string]any {
	return map[string]any{"request_id": requestID, "expected_revision": c.revision, "actor_id": actorID}
}

func (c *selftestClient) captureRevision(payload map[string]any) {
	caseValue, _ := payload["case"].(map[string]any)
	revision, _ := caseValue["revision"].(float64)
	c.revision = int64(revision)
}

func (c *selftestClient) post(path string, payload any) (map[string]any, error) {
	data, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	request, err := http.NewRequest(http.MethodPost, c.baseURL+path, bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	request.Header.Set("Content-Type", "application/json")
	return c.do(request)
}

func (c *selftestClient) get(path string) (map[string]any, error) {
	request, err := http.NewRequest(http.MethodGet, c.baseURL+path, nil)
	if err != nil {
		return nil, err
	}
	return c.do(request)
}

func (c *selftestClient) do(request *http.Request) (map[string]any, error) {
	response, err := c.client.Do(request)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	var payload map[string]any
	if err := json.NewDecoder(response.Body).Decode(&payload); err != nil {
		return nil, err
	}
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return nil, fmt.Errorf("%s %s 返回 %d: %v", request.Method, request.URL.Path, response.StatusCode, payload)
	}
	return payload, nil
}

package web

import (
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	"sherd-proof/internal/app"
	"sherd-proof/internal/store"
)

func TestWorkbenchHealthAndValidation(t *testing.T) {
	repository, err := store.Open(filepath.Join(t.TempDir(), "web.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer repository.Close()
	server := httptest.NewServer(New(app.NewService(repository)))
	defer server.Close()
	response, err := http.Get(server.URL + "/")
	if err != nil {
		t.Fatal(err)
	}
	if response.StatusCode != http.StatusOK || !strings.HasPrefix(response.Header.Get("Content-Type"), "text/html") {
		t.Fatalf("工作台响应不正确")
	}
	response.Body.Close()
	response, err = http.Get(server.URL + "/healthz")
	if err != nil {
		t.Fatal(err)
	}
	if response.StatusCode != http.StatusOK {
		t.Fatalf("健康检查=%d", response.StatusCode)
	}
	response.Body.Close()
	response, err = http.Post(server.URL+"/api/cases", "application/json", strings.NewReader(`{"request_id":"bad","unknown":1}`))
	if err != nil {
		t.Fatal(err)
	}
	if response.StatusCode != http.StatusBadRequest {
		t.Fatalf("未知字段应返回 400，得到 %d", response.StatusCode)
	}
	response.Body.Close()
}

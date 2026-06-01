package main

import (
	"bufio"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	actionsRun "github.com/chetas1208/ActR.AI/apps/api/api/actions_run"
	health "github.com/chetas1208/ActR.AI/apps/api/api/health"
	providersStatus "github.com/chetas1208/ActR.AI/apps/api/api/providers_status"
	ready "github.com/chetas1208/ActR.AI/apps/api/api/ready"
	rtrvrRun "github.com/chetas1208/ActR.AI/apps/api/api/rtrvr_run"
	transcriptUploadPresign "github.com/chetas1208/ActR.AI/apps/api/api/transcript_upload_presign"
	uploadsPresign "github.com/chetas1208/ActR.AI/apps/api/api/uploads_presign"
	webhooksTigris "github.com/chetas1208/ActR.AI/apps/api/api/webhooks_tigris"
	workflowsContinue "github.com/chetas1208/ActR.AI/apps/api/api/workflows_continue"
	workflowsDirectFileStart "github.com/chetas1208/ActR.AI/apps/api/api/workflows_direct_file_start"
	workflowsGet "github.com/chetas1208/ActR.AI/apps/api/api/workflows_get"
	workflowsUploadStart "github.com/chetas1208/ActR.AI/apps/api/api/workflows_upload_start"
	workflowsYoutubeStart "github.com/chetas1208/ActR.AI/apps/api/api/workflows_youtube_start"
)

func main() {
	loadDotEnv(".env")
	loadDotEnv(filepath.Join("..", "..", ".env"))

	mux := http.NewServeMux()

	mux.HandleFunc("/", health.Handler)
	mux.HandleFunc("/api/health", health.Handler)
	mux.HandleFunc("/api/ready", ready.Handler)
	mux.HandleFunc("/api/providers/status", providersStatus.Handler)
	mux.HandleFunc("/api/uploads/presign", uploadsPresign.Handler)
	mux.HandleFunc("/api/uploads/transcript/presign", transcriptUploadPresign.Handler)
	mux.HandleFunc("/api/rtrvr/run", rtrvrRun.Handler)
	mux.HandleFunc("/api/webhooks/tigris", webhooksTigris.Handler)
	mux.HandleFunc("/api/workflows/youtube/start", workflowsYoutubeStart.Handler)
	mux.HandleFunc("/api/workflows/upload/start", workflowsUploadStart.Handler)
	mux.HandleFunc("/api/workflows/direct-file/start", workflowsDirectFileStart.Handler)

	mux.HandleFunc("/api/actions/", actionsRun.Handler)
	mux.HandleFunc("/api/workflows/", workflowMux)

	addr := ":" + resolvePort()
	log.Printf("[actr-ai-api] listening on http://localhost%s", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatal(err)
	}
}

func workflowMux(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(strings.Trim(r.URL.Path, "/"), "api/workflows/")
	if path == "" {
		http.NotFound(w, r)
		return
	}
	if strings.HasSuffix(path, "/continue") {
		workflowsContinue.Handler(w, r)
		return
	}
	workflowsGet.Handler(w, r)
}

func resolvePort() string {
	if p := strings.TrimSpace(os.Getenv("PORT")); p != "" {
		return p
	}

	base := strings.TrimSpace(os.Getenv("APP_BASE_URL"))
	if base != "" {
		if u, err := url.Parse(base); err == nil {
			if p := u.Port(); p != "" {
				return p
			}
		}
	}
	return "8081"
}

func loadDotEnv(path string) {
	f, err := os.Open(path)
	if err != nil {
		return
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		idx := strings.Index(line, "=")
		if idx <= 0 {
			continue
		}
		key := strings.TrimSpace(line[:idx])
		value := strings.TrimSpace(line[idx+1:])
		value = strings.Trim(value, `"'`)
		if os.Getenv(key) == "" {
			_ = os.Setenv(key, value)
		}
	}
}


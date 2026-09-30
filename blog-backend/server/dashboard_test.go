package server_test

import (
	"bytes"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// dashboardClient drives the dashboard API of a fresh fixture as the admin.
type dashboardClient struct {
	t     *testing.T
	f     *fixture
	h     http.Handler
	token string
}

func newDashboardClient(t *testing.T) *dashboardClient {
	f := newFixture(t, nil)
	d := &dashboardClient{t: t, f: f, h: newTestHandler(t, f.cfg, testUser, testPassword)}
	var login map[string]string
	d.do(jsonReq("POST", "/api/login", `{"username":"admin","password":"correct-horse"}`, ""), http.StatusOK, &login)
	d.token = login["token"]
	return d
}

// do serves req, checks the status and decodes a JSON body into out if given.
func (d *dashboardClient) do(req *http.Request, wantStatus int, out any) string {
	d.t.Helper()
	rec := httptest.NewRecorder()
	d.h.ServeHTTP(rec, req)
	if rec.Code != wantStatus {
		d.t.Fatalf("%s %s: status %d, want %d (body %q)", req.Method, req.URL, rec.Code, wantStatus, rec.Body.String())
	}
	if out != nil {
		if err := json.Unmarshal(rec.Body.Bytes(), out); err != nil {
			d.t.Fatalf("%s %s: decode %q: %v", req.Method, req.URL, rec.Body.String(), err)
		}
	}
	return rec.Body.String()
}

func (d *dashboardClient) authed(req *http.Request) *http.Request {
	req.Header.Set("Authorization", "Bearer "+d.token)
	return req
}

func (d *dashboardClient) replaceReq(name, content string) *http.Request {
	var buf bytes.Buffer
	mw := multipart.NewWriter(&buf)
	mw.WriteField("filename", name)
	fw, _ := mw.CreateFormFile("file", "whatever-the-browser-calls-it.bin")
	fw.Write([]byte(content))
	mw.Close()
	req := httptest.NewRequest(http.MethodPost, "/api/replace-file", &buf)
	req.Header.Set("Content-Type", mw.FormDataContentType())
	return req
}

func TestSettingsAPI(t *testing.T) {
	d := newDashboardClient(t)

	if body := d.do(httptest.NewRequest("GET", "/api/settings", nil), http.StatusOK, nil); body != "{\"highlight\":null}\n" {
		t.Fatalf("empty settings = %q", body)
	}

	colors := map[string]string{}
	for _, k := range []string{"00", "01", "02", "03", "04", "05", "06", "07", "08", "09", "0A", "0B", "0C", "0D", "0E", "0F"} {
		colors["base"+k] = "#123456"
	}
	valid, _ := json.Marshal(map[string]any{"highlight": map[string]any{"preset": "monokai", "colors": colors}})

	d.do(jsonReq("POST", "/api/settings", string(valid), ""), http.StatusUnauthorized, nil)
	d.do(jsonReq("POST", "/api/settings", `{"highlight":{"preset":"x","colors":{"base00":"red"}}}`, d.token), http.StatusBadRequest, nil)
	d.do(jsonReq("POST", "/api/settings", string(valid), d.token), http.StatusOK, nil)

	var got struct {
		Highlight struct {
			Preset string            `json:"preset"`
			Colors map[string]string `json:"colors"`
		} `json:"highlight"`
	}
	d.do(httptest.NewRequest("GET", "/api/settings", nil), http.StatusOK, &got)
	if got.Highlight.Preset != "monokai" || got.Highlight.Colors["base0E"] != "#123456" {
		t.Fatalf("saved settings = %+v", got.Highlight)
	}
}

func TestFileUsageReplaceAndRename(t *testing.T) {
	d := newDashboardClient(t)
	postPath := filepath.Join(d.f.cfg.PostsDir, "gallery.md")
	writeFile(t, postPath, []byte("# Gallery\n\n![a](https://chatter.pw/assets/free%20additional%20features.jpg)\n![b](/assets/Tux.png)\n"))

	var usage map[string][]string
	d.do(httptest.NewRequest("GET", "/api/files/usage", nil), http.StatusUnauthorized, nil)
	d.do(d.authed(httptest.NewRequest("GET", "/api/files/usage", nil)), http.StatusOK, &usage)
	if len(usage) != 2 || usage["Tux.png"][0] != "gallery" || usage["free additional features.jpg"][0] != "gallery" {
		t.Fatalf("usage = %v", usage)
	}

	// Replace keeps the name and swaps the content.
	d.do(d.replaceReq("Tux.png", "new"), http.StatusUnauthorized, nil)
	d.do(d.authed(d.replaceReq("missing.png", "new")), http.StatusNotFound, nil)
	d.do(d.authed(d.replaceReq("../gallery.md", "new")), http.StatusBadRequest, nil)
	var replaced struct {
		File struct {
			Name string `json:"name"`
			Size int64  `json:"size"`
		} `json:"file"`
	}
	d.do(d.authed(d.replaceReq("Tux.png", "new content")), http.StatusOK, &replaced)
	if replaced.File.Name != "Tux.png" || replaced.File.Size != int64(len("new content")) {
		t.Fatalf("replace result = %+v", replaced.File)
	}
	if body := d.do(httptest.NewRequest("GET", "/assets/Tux.png", nil), http.StatusOK, nil); body != "new content" {
		t.Fatalf("served content after replace = %q", body)
	}
	if entries, _ := os.ReadDir(d.f.cfg.AssetsDir); len(entries) != 2 {
		t.Fatalf("replace left temporary files: %v", entries)
	}

	// Rename refuses to overwrite and validates names.
	d.do(jsonReq("POST", "/api/rename-file", `{"from":"Tux.png","to":"free additional features.jpg"}`, d.token), http.StatusConflict, nil)
	d.do(jsonReq("POST", "/api/rename-file", `{"from":"Tux.png","to":"../x.png"}`, d.token), http.StatusBadRequest, nil)
	d.do(jsonReq("POST", "/api/rename-file", `{"from":"nope.png","to":"x.png"}`, d.token), http.StatusNotFound, nil)

	var renamed struct {
		UpdatedPosts []string `json:"updatedPosts"`
	}
	d.do(jsonReq("POST", "/api/rename-file",
		`{"from":"free additional features.jpg","to":"features 2.jpg","updateReferences":true}`, d.token), http.StatusOK, &renamed)
	if len(renamed.UpdatedPosts) != 1 || renamed.UpdatedPosts[0] != "gallery" {
		t.Fatalf("updatedPosts = %v", renamed.UpdatedPosts)
	}
	post, _ := os.ReadFile(postPath)
	if !strings.Contains(string(post), "https://chatter.pw/assets/features%202.jpg)") || !strings.Contains(string(post), "/assets/Tux.png") {
		t.Fatalf("post after rename:\n%s", post)
	}
	d.do(httptest.NewRequest("GET", "/assets/features%202.jpg", nil), http.StatusOK, nil)
	d.do(httptest.NewRequest("GET", "/assets/free%20additional%20features.jpg", nil), http.StatusNotFound, nil)

	// Without updateReferences, posts are left alone.
	d.do(jsonReq("POST", "/api/rename-file", `{"from":"Tux.png","to":"tux.png"}`, d.token), http.StatusOK, &renamed)
	if len(renamed.UpdatedPosts) != 0 {
		t.Fatalf("updatedPosts without updateReferences = %v", renamed.UpdatedPosts)
	}
	if post, _ := os.ReadFile(postPath); !strings.Contains(string(post), "/assets/Tux.png") {
		t.Fatal("post changed although updateReferences was false")
	}
}

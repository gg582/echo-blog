package server_test

// These tests pin the observable HTTP behavior of the backend (status codes,
// selected headers and bodies) in testdata/http_golden.json, so a router or
// handler rewrite can be checked against the previous implementation.
//
// Regenerate the golden file with: go test ./server -run TestHTTPCompat -update

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"flag"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/gg582/echo-blog/blog-backend/config"
)

var update = flag.Bool("update", false, "rewrite testdata/http_golden.json")

// fixedTime is applied to every fixture file so timestamps in responses are stable.
var fixedTime = time.Date(2025, 7, 19, 12, 0, 0, 0, time.UTC)

const (
	testUser     = "admin"
	testPassword = "correct-horse"
	testSecret   = "compat-test-secret"
	koreanSlug   = "리눅스-데스크톱-사용-시-im의-문제"
)

type fixture struct {
	cfg   *config.Config
	token string
}

func writeFile(t *testing.T, path string, content []byte) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, content, 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Chtimes(path, fixedTime, fixedTime); err != nil {
		t.Fatal(err)
	}
}

func newFixture(t *testing.T, origins []string) *fixture {
	t.Helper()
	root := t.TempDir()
	cfg := &config.Config{
		DBPath:         filepath.Join(root, "auth.db"),
		StaticDir:      filepath.Join(root, "build"),
		PostsDir:       filepath.Join(root, "posts"),
		AssetsDir:      filepath.Join(root, "posts", "assets"),
		AboutMD:        filepath.Join(root, "about", "about.md"),
		ContactMD:      filepath.Join(root, "contact", "contact.md"),
		SettingsPath:   filepath.Join(root, "settings.json"),
		AllowedOrigins: origins,
		AuthSecret:     testSecret,
	}

	writeFile(t, filepath.Join(cfg.PostsDir, "hello-world.md"),
		[]byte("---\nauthor: tester\n---\n\n# Hello World\n\nSome <b>text</b> & more.\n"))
	writeFile(t, filepath.Join(cfg.PostsDir, koreanSlug+".md"),
		[]byte("# 리눅스 데스크톱\n\n본문입니다.\n"))
	writeFile(t, filepath.Join(cfg.PostsDir, "plain.md"),
		[]byte("First line without heading\n\nbody\n"))
	writeFile(t, filepath.Join(cfg.PostsDir, "notes.txt"), []byte("not a post"))
	writeFile(t, filepath.Join(cfg.AssetsDir, "Tux.png"), bytes.Repeat([]byte("PNGDATA-"), 64))
	writeFile(t, filepath.Join(cfg.AssetsDir, "free additional features.jpg"), []byte("jpeg bytes"))
	writeFile(t, cfg.AboutMD, []byte("---\nauthor: me\n---\n# About Me\n\nhello\n"))
	writeFile(t, cfg.ContactMD, []byte("mail: me@example.com\n"))
	writeFile(t, filepath.Join(cfg.StaticDir, "index.html"), []byte("<!doctype html><title>SPA</title>\n"))
	writeFile(t, filepath.Join(cfg.StaticDir, "static", "js", "main.js"), []byte("console.log(1)\n"))
	writeFile(t, filepath.Join(cfg.StaticDir, "manifest.json"), []byte(`{"name":"blog"}`))

	f := &fixture{cfg: cfg}
	return f
}

type exchange struct {
	Name    string            `json:"name"`
	Method  string            `json:"method"`
	Target  string            `json:"target"`
	Status  int               `json:"status"`
	Headers map[string]string `json:"headers"`
	Body    string            `json:"body"`
}

// pinnedHeaders are compared in addition to the status and body.
var pinnedHeaders = []string{
	"Content-Type", "Content-Range", "Accept-Ranges", "Last-Modified", "Location",
	"X-Content-Type-Options", "Allow", "Vary",
	"Access-Control-Allow-Origin", "Access-Control-Allow-Credentials",
	"Access-Control-Allow-Methods", "Access-Control-Allow-Headers", "Access-Control-Max-Age",
}

var (
	tokenRe = regexp.MustCompile(`"token":"[^"]*"`)
	timeRe  = regexp.MustCompile(`"(createdAt|modifiedAt)":"([^"]*)"`)
)

// normalize masks values that legitimately differ between runs.
func normalize(body string) string {
	body = tokenRe.ReplaceAllString(body, `"token":"<token>"`)
	return timeRe.ReplaceAllStringFunc(body, func(m string) string {
		sub := timeRe.FindStringSubmatch(m)
		ts, err := time.Parse(time.RFC3339Nano, sub[2])
		if err == nil && ts.Equal(fixedTime) {
			return m
		}
		return `"` + sub[1] + `":"<now>"`
	})
}

func record(name string, h http.Handler, req *http.Request) exchange {
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	res := rec.Result()

	headers := map[string]string{}
	for _, k := range pinnedHeaders {
		if v := res.Header.Values(k); len(v) > 0 {
			headers[k] = strings.Join(v, ", ")
		}
	}
	if lm, ok := headers["Last-Modified"]; ok && lm != fixedTime.Format(http.TimeFormat) {
		headers["Last-Modified"] = "<now>"
	}

	body := rec.Body.String()
	ct := res.Header.Get("Content-Type")
	if !strings.HasPrefix(ct, "text/") && !strings.HasPrefix(ct, "application/json") && body != "" {
		sum := sha256.Sum256(rec.Body.Bytes())
		body = "sha256:" + hex.EncodeToString(sum[:])
	}

	return exchange{
		Name:    name,
		Method:  req.Method,
		Target:  req.URL.RequestURI(),
		Status:  res.StatusCode,
		Headers: headers,
		Body:    normalize(body),
	}
}

func jsonReq(method, target, body, token string) *http.Request {
	req := httptest.NewRequest(method, target, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	return req
}

func uploadReq(t *testing.T, files map[string]string) *http.Request {
	t.Helper()
	var buf bytes.Buffer
	mw := multipart.NewWriter(&buf)
	mw.SetBoundary("compat-boundary")
	for name, content := range files {
		fw, err := mw.CreateFormFile("file", name)
		if err != nil {
			t.Fatal(err)
		}
		fw.Write([]byte(content))
	}
	mw.Close()
	req := httptest.NewRequest(http.MethodPost, "http://blog.example/api/upload-file", &buf)
	req.Header.Set("Content-Type", mw.FormDataContentType())
	return req
}

func withHeader(req *http.Request, k, v string) *http.Request {
	req.Header.Set(k, v)
	return req
}

func TestHTTPCompat(t *testing.T) {
	var got []exchange

	// --- Default deployment: no CORS configured. ---
	f := newFixture(t, nil)
	h := newTestHandler(t, f.cfg, testUser, testPassword)
	enc := "/api/posts/" + koreanSlug

	get := func(target string) *http.Request { return httptest.NewRequest(http.MethodGet, target, nil) }
	post := func(target string) *http.Request { return httptest.NewRequest(http.MethodPost, target, nil) }
	run := func(name string, req *http.Request) exchange {
		ex := record(name, h, req)
		got = append(got, ex)
		return ex
	}

	// Login first so the token can be used below.
	run("login bad body", jsonReq("POST", "/api/login", "{", ""))
	run("login unknown user", jsonReq("POST", "/api/login", `{"username":"nobody","password":"x"}`, ""))
	run("login wrong password", jsonReq("POST", "/api/login", `{"username":"admin","password":"x"}`, ""))
	{
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, jsonReq("POST", "/api/login", `{"username":"admin","password":"correct-horse"}`, ""))
		var resp map[string]string
		if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
			t.Fatalf("login response: %v: %s", err, rec.Body.String())
		}
		f.token = resp["token"]
		got = append(got, exchange{Name: "login ok", Method: "POST", Target: "/api/login",
			Status: rec.Code, Headers: map[string]string{"Content-Type": rec.Header().Get("Content-Type")},
			Body: normalize(rec.Body.String())})
	}
	tok := f.token

	run("list posts", post("/api/posts"))
	run("list posts GET", get("/api/posts"))
	run("post by id", post("/api/posts/hello-world"))
	run("post by korean id", post(enc))
	run("post by percent-encoded korean id", post("/api/posts/%EB%A6%AC%EB%88%85%EC%8A%A4-%EB%8D%B0%EC%8A%A4%ED%81%AC%ED%86%B1-%EC%82%AC%EC%9A%A9-%EC%8B%9C-im%EC%9D%98-%EB%AC%B8%EC%A0%9C"))
	run("post without heading", post("/api/posts/plain"))
	run("post missing", post("/api/posts/missing"))
	run("post encoded slash", post("/api/posts/..%2Fplain"))
	run("raw encoded slash", get("/api/posts/..%2F..%2Fabout%2Fabout/raw"))
	run("raw post", get("/api/posts/hello-world/raw"))
	run("raw korean post", get(enc+"/raw"))
	run("raw missing", get("/api/posts/missing/raw"))
	run("about", get("/api/about"))
	run("contact", get("/api/contact"))

	run("files no auth", get("/api/files"))
	run("files bad token", withHeader(get("/api/files"), "Authorization", "Bearer nope"))
	run("files raw token", withHeader(get("/api/files"), "Authorization", tok))
	run("files", withHeader(get("/api/files"), "Authorization", "Bearer "+tok))

	run("edit no auth", jsonReq("POST", "/api/edit-post/hello-world", `{"content":"x"}`, ""))
	run("edit bad body", jsonReq("POST", "/api/edit-post/hello-world", `nope`, tok))
	run("edit empty", jsonReq("POST", "/api/edit-post/hello-world", `{"content":""}`, tok))
	run("edit ok", jsonReq("POST", "/api/edit-post/hello-world", `{"content":"# Edited\n\nnew body"}`, tok))
	run("raw after edit", get("/api/posts/hello-world/raw"))
	run("edit creates", jsonReq("POST", "/api/edit-post/created-by-edit", `{"content":"# Created"}`, tok))

	run("new post missing fields", jsonReq("POST", "/api/new-post/fresh", `{"title":"T"}`, ""))
	run("new post bad body", jsonReq("POST", "/api/new-post/fresh", `[`, ""))
	run("new post ok", jsonReq("POST", "/api/new-post/fresh", `{"title":"Fresh","author":"me","content":"body"}`, ""))
	run("raw new post", get("/api/posts/fresh/raw"))
	run("new post no slug", jsonReq("POST", "/api/new-post/", `{"title":"Fresh","author":"me","content":"body"}`, ""))

	run("delete no auth", post("/api/delete-post/fresh"))
	run("delete ok", withHeader(post("/api/delete-post/fresh"), "Authorization", "Bearer "+tok))
	run("delete missing", withHeader(post("/api/delete-post/fresh"), "Authorization", "Bearer "+tok))

	run("upload no files", uploadReq(t, nil))
	run("upload not multipart", jsonReq("POST", "/api/upload-file", `{}`, ""))
	run("upload ok", uploadReq(t, map[string]string{"new file.txt": "uploaded"}))
	run("upload over https proxy", withHeader(uploadReq(t, map[string]string{"second.txt": "2"}), "X-Forwarded-Proto", "https"))
	run("uploaded asset", get("/assets/new%20file.txt"))

	run("delete-file no auth", jsonReq("POST", "/api/delete-file", `{"filename":"second.txt"}`, ""))
	run("delete-file traversal", jsonReq("POST", "/api/delete-file", `{"filename":"../plain.md"}`, tok))
	run("delete-file empty", jsonReq("POST", "/api/delete-file", `{"filename":""}`, tok))
	run("delete-file bad body", jsonReq("POST", "/api/delete-file", `x`, tok))
	run("delete-file ok", jsonReq("POST", "/api/delete-file", `{"filename":"second.txt"}`, tok))
	run("delete-file missing", jsonReq("POST", "/api/delete-file", `{"filename":"second.txt"}`, tok))
	run("list posts after writes", post("/api/posts"))

	run("asset", get("/assets/Tux.png"))
	run("asset with space", get("/assets/free%20additional%20features.jpg"))
	run("asset range", withHeader(get("/assets/Tux.png"), "Range", "bytes=0-9"))
	run("asset not modified", withHeader(get("/assets/Tux.png"), "If-Modified-Since", fixedTime.Add(time.Hour).Format(http.TimeFormat)))
	run("asset HEAD", httptest.NewRequest(http.MethodHead, "/assets/Tux.png", nil))
	run("asset POST", post("/assets/Tux.png"))
	run("asset missing", get("/assets/missing.png"))
	run("asset dir listing", get("/assets/"))
	run("assets no slash", get("/assets"))
	run("asset traversal", get("/assets/../plain.md"))

	run("spa root", get("/"))
	run("spa index.html", get("/index.html"))
	run("spa static file", get("/static/js/main.js"))
	run("spa manifest", get("/manifest.json"))
	run("spa client route", get("/posts/hello-world"))
	run("spa korean client route", get("/posts/"+koreanSlug))
	run("spa directory", get("/static/js"))
	run("spa HEAD", httptest.NewRequest(http.MethodHead, "/", nil))

	run("unknown api POST", post("/api/unknown"))
	run("unknown api GET", get("/api/unknown"))
	run("PUT route", httptest.NewRequest(http.MethodPut, "/api/posts", nil))
	run("OPTIONS without CORS", httptest.NewRequest(http.MethodOptions, "/api/posts", nil))

	// --- CORS enabled. ---
	fc := newFixture(t, []string{"https://chatter.pw", "http://localhost:3000"})
	hc := newTestHandler(t, fc.cfg, testUser, testPassword)
	runC := func(name string, req *http.Request) {
		got = append(got, record(name, hc, req))
	}
	pre := httptest.NewRequest(http.MethodOptions, "/api/posts", nil)
	pre.Header.Set("Origin", "https://chatter.pw")
	pre.Header.Set("Access-Control-Request-Method", "POST")
	pre.Header.Set("Access-Control-Request-Headers", "content-type,authorization")
	runC("cors preflight", pre)
	okPre := httptest.NewRequest(http.MethodOptions, "/api/edit-post/x", nil)
	okPre.Header.Set("Origin", "https://chatter.pw")
	okPre.Header.Set("Access-Control-Request-Method", "POST")
	okPre.Header.Set("Access-Control-Request-Headers", "authorization,content-type")
	runC("cors preflight sorted headers", okPre)
	badPre := httptest.NewRequest(http.MethodOptions, "/api/posts", nil)
	badPre.Header.Set("Origin", "https://evil.example")
	badPre.Header.Set("Access-Control-Request-Method", "POST")
	runC("cors preflight bad origin", badPre)
	runC("cors simple", withHeader(post("/api/posts"), "Origin", "http://localhost:3000"))
	runC("cors simple bad origin", withHeader(post("/api/posts"), "Origin", "https://evil.example"))
	runC("cors asset", withHeader(get("/assets/Tux.png"), "Origin", "https://chatter.pw"))

	path := filepath.Join("testdata", "http_golden.json")
	if *update {
		data, err := json.MarshalIndent(got, "", "  ")
		if err != nil {
			t.Fatal(err)
		}
		if err := os.MkdirAll("testdata", 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, append(data, '\n'), 0o644); err != nil {
			t.Fatal(err)
		}
		return
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read golden file (run with -update to create it): %v", err)
	}
	var want []exchange
	if err := json.Unmarshal(data, &want); err != nil {
		t.Fatal(err)
	}
	if len(want) != len(got) {
		t.Fatalf("golden has %d exchanges, got %d", len(want), len(got))
	}
	for i := range want {
		w, g := want[i], got[i]
		if w.Name != g.Name {
			t.Fatalf("exchange %d: golden %q, got %q", i, w.Name, g.Name)
		}
		if w.Status != g.Status {
			t.Errorf("%s: status %d, want %d (body %q)", w.Name, g.Status, w.Status, g.Body)
		}
		for _, k := range pinnedHeaders {
			if w.Headers[k] != g.Headers[k] {
				t.Errorf("%s: header %s = %q, want %q", w.Name, k, g.Headers[k], w.Headers[k])
			}
		}
		if w.Body != g.Body {
			t.Errorf("%s: body\n got: %q\nwant: %q", w.Name, g.Body, w.Body)
		}
	}
}

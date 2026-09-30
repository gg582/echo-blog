package content

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSplitAuthor(t *testing.T) {
	author, body := SplitAuthor([]byte("---\nauthor: yjlee\n---\n\n# Title\n"))
	if author != "yjlee" || string(body) != "# Title\n" {
		t.Fatalf("got author %q body %q", author, body)
	}

	author, body = SplitAuthor([]byte("# No front matter\n"))
	if author != DefaultAuthor || string(body) != "# No front matter\n" {
		t.Fatalf("got author %q body %q", author, body)
	}
}

func TestTitle(t *testing.T) {
	cases := map[string]string{
		"\n\n# Heading\ntext": "Heading",
		"  plain line  \n":    "plain line",
		"\n  \n":              "fallback",
	}
	for body, want := range cases {
		if got := Title([]byte(body), "fallback"); got != want {
			t.Errorf("Title(%q) = %q, want %q", body, got, want)
		}
	}
}

func TestPostsLifecycle(t *testing.T) {
	dir := t.TempDir()
	posts := NewPosts(dir)

	if _, err := posts.Get("missing"); err != ErrNotFound {
		t.Fatalf("Get missing: got %v, want ErrNotFound", err)
	}
	if err := posts.Delete("missing"); err != ErrNotFound {
		t.Fatalf("Delete missing: got %v, want ErrNotFound", err)
	}
	if err := posts.Write("hello", []byte("# Hello\n\nbody")); err != nil {
		t.Fatal(err)
	}
	os.WriteFile(filepath.Join(dir, "notes.txt"), []byte("ignored"), 0o644)

	list, err := posts.List(t.Context())
	if err != nil || len(list) != 1 || list[0].Title != "Hello" || list[0].FileName != "hello.md" {
		t.Fatalf("List = %+v, %v", list, err)
	}
	if err := posts.Delete("hello"); err != nil {
		t.Fatal(err)
	}
	if _, err := posts.Raw("hello"); err != ErrNotFound {
		t.Fatalf("Raw after delete: got %v, want ErrNotFound", err)
	}
}

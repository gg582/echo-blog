package assets

import "testing"

func TestMentions(t *testing.T) {
	md := []byte("![a](https://chatter.pw/assets/my%20shot.png)\nsee /assets/Tux.png.\n")
	cases := map[string]bool{
		"my shot.png": true, // escaped spelling
		"Tux.png":     true, // sentence-ending period
		"shot.png":    false,
		"my":          false, // prefix of a longer name
	}
	for name, want := range cases {
		if got := Mentions(md, name); got != want {
			t.Errorf("Mentions(%q) = %v, want %v", name, got, want)
		}
	}
	if !Mentions([]byte("<img src=\"/assets/Tux.png\">"), "Tux.png") {
		t.Error("expected quoted link to match")
	}
}

func TestRewriteLinks(t *testing.T) {
	md := []byte("![](/assets/a b.png) [x](https://h/assets/a%20b.png?raw=1) /assets/a b.png.bak")
	got, changed := RewriteLinks(md, "a b.png", "new name.png")
	want := "![](/assets/new name.png) [x](https://h/assets/new%20name.png?raw=1) /assets/a b.png.bak"
	if !changed || string(got) != want {
		t.Fatalf("RewriteLinks = %q, %v\nwant %q", got, changed, want)
	}
	if _, changed := RewriteLinks([]byte("nothing here"), "a b.png", "c.png"); changed {
		t.Error("expected no change")
	}
}

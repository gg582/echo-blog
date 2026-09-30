package assets

import (
	"bytes"
	"net/url"
	"strings"
)

// Posts link to uploaded files with URLs ending in /assets/<name>, where the
// name is written raw ("my file.png") or escaped ("my%20file.png").

// linkForms returns the spellings of a link to name, always in the same
// order: raw, path-escaped (Go) and component-escaped (JavaScript's
// encodeURIComponent). Spellings may repeat.
func linkForms(name string) [3]string {
	return [3]string{
		"/assets/" + name,
		"/assets/" + url.PathEscape(name),
		"/assets/" + strings.ReplaceAll(url.QueryEscape(name), "+", "%20"),
	}
}

// endsLink reports whether b[i] can follow a complete link, i.e. the link is
// not just a prefix of a longer file name.
func endsLink(b []byte, i int) bool {
	if i >= len(b) || isLinkEnd(b[i]) {
		return true
	}
	// Sentence punctuation right after a link: "see /assets/a.png." but not
	// "/assets/a.png.bak".
	return (b[i] == '.' || b[i] == ',') && (i+1 >= len(b) || isLinkEnd(b[i+1]))
}

// isLinkEnd reports whether c cannot be part of a link in markdown or HTML.
func isLinkEnd(c byte) bool {
	return strings.IndexByte(" \t\r\n)]>\"'?#", c) >= 0
}

// indexLink returns the index of the first complete occurrence of form in b
// at or after start, or -1.
func indexLink(b []byte, form string, start int) int {
	for start <= len(b) {
		i := bytes.Index(b[start:], []byte(form))
		if i < 0 {
			return -1
		}
		i += start
		if endsLink(b, i+len(form)) {
			return i
		}
		start = i + 1
	}
	return -1
}

// Mentions reports whether markdown links to the uploaded file name.
func Mentions(markdown []byte, name string) bool {
	for _, form := range linkForms(name) {
		if indexLink(markdown, form, 0) >= 0 {
			return true
		}
	}
	return false
}

// RewriteLinks replaces links to from with links to to, keeping each link's
// escaping style. It reports whether anything changed.
func RewriteLinks(markdown []byte, from, to string) ([]byte, bool) {
	fromForms, toForms := linkForms(from), linkForms(to)
	out := markdown
	changed := false
	for k, form := range fromForms {
		if k > 0 && (form == fromForms[0] || form == fromForms[k-1]) {
			continue // same spelling already handled
		}
		var buf bytes.Buffer
		start := 0
		for {
			i := indexLink(out, form, start)
			if i < 0 {
				break
			}
			buf.Write(out[start:i])
			buf.WriteString(toForms[k])
			start = i + len(form)
			changed = true
		}
		if start > 0 {
			buf.Write(out[start:])
			out = buf.Bytes()
		}
	}
	return out, changed
}

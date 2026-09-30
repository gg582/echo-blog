package workerpool

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// blockingReader blocks reads until release is closed.
type blockingReader struct{ release chan struct{} }

func (r blockingReader) Read(p []byte) (int, error) {
	<-r.release
	return 0, errors.New("released")
}

func TestSubmitSavesFile(t *testing.T) {
	p := New(1, 1)
	defer p.Close()
	dir := t.TempDir()

	res, err := p.Submit(context.Background(), UploadJob{File: strings.NewReader("data"), FileName: "a.txt", Dir: dir})
	if err != nil || res.Error != nil || res.SavedFileName != "a.txt" {
		t.Fatalf("Submit = %+v, %v", res, err)
	}
	if got, _ := os.ReadFile(filepath.Join(dir, "a.txt")); string(got) != "data" {
		t.Fatalf("saved content %q", got)
	}
}

func TestSubmitBusyAndCancel(t *testing.T) {
	p := New(1, 1)
	dir := t.TempDir()
	release := make(chan struct{})
	block := UploadJob{File: blockingReader{release}, FileName: "block", Dir: dir}

	// The first job occupies the worker; wait until it has been picked up.
	ctx, cancel := context.WithCancel(context.Background())
	started := make(chan error, 1)
	go func() { _, err := p.Submit(ctx, block); started <- err }()
	waitFor(func() bool { return len(p.jobs) == 0 && fileExists(filepath.Join(dir, "block")) })

	// The second fills the queue, the third finds it full.
	ctx2, cancel2 := context.WithCancel(context.Background())
	queued := make(chan error, 1)
	go func() { _, err := p.Submit(ctx2, block); queued <- err }()
	waitFor(func() bool { return len(p.jobs) == 1 })
	if _, err := p.Submit(context.Background(), block); !errors.Is(err, ErrBusy) {
		t.Fatalf("third Submit: got %v, want ErrBusy", err)
	}

	cancel()
	if err := <-started; !errors.Is(err, context.Canceled) {
		t.Fatalf("cancelled Submit: got %v, want context.Canceled", err)
	}
	cancel2()
	<-queued
	close(release)
	p.Close()
}

func waitFor(cond func() bool) {
	for !cond() {
		runtime.Gosched()
	}
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

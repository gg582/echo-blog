// Package workerpool saves uploaded files on a fixed number of goroutines so
// a burst of uploads cannot open an unbounded number of files at once.
package workerpool

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"sync"
)

// ErrBusy is returned by Submit when the job queue is full.
var ErrBusy = errors.New("worker pool is busy")

// UploadJob asks a worker to copy File into Dir/FileName.
type UploadJob struct {
	File     io.Reader
	FileName string
	Dir      string

	result chan UploadResult
}

// UploadResult reports the outcome of an UploadJob.
type UploadResult struct {
	SavedFileName    string
	OriginalFileName string
	Error            error
}

// Pool runs upload jobs on a fixed set of workers.
type Pool struct {
	jobs chan UploadJob
	wg   sync.WaitGroup
}

// New starts numWorkers workers reading from a queue of queueSize jobs.
func New(numWorkers, queueSize int) *Pool {
	p := &Pool{jobs: make(chan UploadJob, queueSize)}
	for i := range numWorkers {
		p.wg.Add(1)
		go p.work(i + 1)
	}
	return p
}

// Submit queues job without blocking and waits for its result. It returns
// ErrBusy if the queue is full, or ctx.Err() if ctx ends first; in that case
// the worker still finishes the job but the result is discarded.
func (p *Pool) Submit(ctx context.Context, job UploadJob) (UploadResult, error) {
	job.result = make(chan UploadResult, 1)
	select {
	case p.jobs <- job:
	default:
		return UploadResult{}, ErrBusy
	}

	select {
	case res := <-job.result:
		return res, nil
	case <-ctx.Done():
		return UploadResult{}, ctx.Err()
	}
}

// Close stops accepting jobs and waits for queued jobs to finish.
func (p *Pool) Close() {
	close(p.jobs)
	p.wg.Wait()
}

func (p *Pool) work(id int) {
	defer p.wg.Done()
	log.Printf("Worker %d started.", id)
	for job := range p.jobs {
		job.result <- process(id, job)
	}
	log.Printf("Worker %d stopped.", id)
}

func process(id int, job UploadJob) UploadResult {
	result := UploadResult{OriginalFileName: job.FileName}
	log.Printf("Worker %d: Processing file (%s)", id, job.FileName)

	path := filepath.Join(job.Dir, job.FileName)
	if err := copyToFile(path, job.File); err != nil {
		result.Error = fmt.Errorf("worker %d: saving %s: %w", id, path, err)
		log.Printf("Error in worker %d: %v", id, result.Error)
		return result
	}

	result.SavedFileName = job.FileName
	log.Printf("Worker %d: successfully processed file %s. Saved as: %s", id, job.FileName, job.FileName)
	return result
}

func copyToFile(path string, r io.Reader) error {
	dst, err := os.Create(path)
	if err != nil {
		return err
	}
	if _, err := io.Copy(dst, r); err != nil {
		dst.Close()
		return err
	}
	return dst.Close()
}

package searcher

import "sync"

type JobFunc func(line []byte, index int, resCh chan<- JobResult, errCh chan<- error) error

type Job struct {
	executeFunc JobFunc

	// The file in which the search will be performed
	line []byte

	// The passed index of the file in the array of files, used for identifying
	// the file from the caller's side when the word is found
	index int

	wg    *sync.WaitGroup
	resCh chan<- JobResult
	errCh chan<- error
}

type JobResult struct {
	Word  string
	Index int
}

func NewJob(executeFunc JobFunc, wg *sync.WaitGroup, resCh chan<- JobResult, errCh chan<- error, line []byte, index int) *Job {
	return &Job{
		executeFunc: executeFunc,
		line:        line,
		index:       index,
		wg:          wg,
		resCh:       resCh,
		errCh:       errCh,
	}
}

func (j *Job) Execute() error {
	if j.wg != nil {
		defer j.wg.Done()
	}

	if j.executeFunc != nil {
		return j.executeFunc(j.line, j.index, j.resCh, j.errCh)
	}

	return nil
}

func (t *Job) OnFailure(e error) {
}

package searcher

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"io/fs"
	"os"
	"strings"
	"sync"
	"time"
	"word-search-in-files/pkg/pool"
)

const (
	workersNum            = 100
	workerTaskChannelSize = 100
	// Maximum line length
	scanBufferSize = 5 * 1024 * 1024
)

type Searcher struct {
	fs     fs.FS
	absDir string

	// Mutex for the struct while scaning in process
	muGlobal sync.Mutex

	Files  []FileInfo
	Errors []error
	Words  map[string]map[int]struct{}
}

type FileInfo struct {
	Path     string
	Modified time.Time
}

type SearcherSync struct {
	pool   pool.Pool
	wg     *sync.WaitGroup
	resCh  chan JobResult
	errCh  chan error
	doneCh chan struct{}
}

func NewSearcher(dir string) (*Searcher, error) {
	if dir == "" {
		dir = "."
	}

	// absDir, err := filepath.Abs(dir)
	// if err != nil {
	// return nil, err
	// }

	absDir := dir

	return &Searcher{
		fs:     os.DirFS(dir),
		absDir: absDir,

		Words: make(map[string]map[int]struct{}),
	}, nil
}

func newSearcherSync() (*SearcherSync, error) {

	pool, err := pool.NewPool(workersNum, workerTaskChannelSize)
	if err != nil {
		return nil, err
	}

	return &SearcherSync{
		pool:   pool,
		wg:     &sync.WaitGroup{},
		resCh:  make(chan JobResult),
		errCh:  make(chan error),
		doneCh: make(chan struct{}),
	}, nil
}

func (s *Searcher) cleanBeforeScan() {
	s.Words = make(map[string]map[int]struct{})
	s.Files = nil
	s.Errors = nil
}

func (s *Searcher) Scan() error {
	s.muGlobal.Lock()
	defer s.muGlobal.Unlock()

	s.cleanBeforeScan()

	snc, e := newSearcherSync()
	if e != nil {
		return e
	}

	snc.pool.Start()

	snc.wg.Add(1)
	go s.predictWalkDir(snc)

	go func() {
		snc.wg.Wait()
		close(snc.resCh)
		close(snc.errCh)
		snc.pool.Stop()
		snc.doneCh <- struct{}{}
		snc.doneCh <- struct{}{}
	}()

	for i := 2; i > 0; {
		select {
		case w, ok := <-snc.resCh:
			if ok {
				addWordToMap(s.Words, w.Word, w.Index)
			}
		case e, ok := <-snc.errCh:
			if ok {
				s.Errors = append(s.Errors, e)
			}
		case <-snc.doneCh:
			i--
		}
	}

	return nil
}

func (s *Searcher) predictWalkDir(snc *SearcherSync) {
	defer snc.wg.Done()

	e := fs.WalkDir(s.fs, ".", func(path string, di fs.DirEntry, e error) error {
		if e != nil {
			return e
		}

		// For every file
		if !di.IsDir() {

			// Create fullpath to the file
			// fullpath := filepath.Join(s.absDir, path)
			fullpath := path

			// Get file info
			fileInfo, e := di.Info()
			if e != nil {
				return e
			}

			// Add the file to the slice of files
			s.Files = append(s.Files, FileInfo{Path: fullpath, Modified: fileInfo.ModTime()})

			// The index of the added file will be used later to identify the words in the map
			index := len(s.Files) - 1

			e = s.readByLineSimple(fullpath, snc, index)
			if e != nil {
				return e
			}
		}

		return nil
	})

	if e != nil {
		snc.errCh <- e
	}
}

func (s *Searcher) readByLineSimple(path string, snc *SearcherSync, index int) error {

	content, e := fs.ReadFile(s.fs, path)
	if e != nil {
		return e
	}

	scanner := bufio.NewScanner(bytes.NewReader(content))

	for scanner.Scan() {
		snc.wg.Add(1)
		job := NewJob(readByWord, snc.wg, snc.resCh, snc.errCh, scanner.Bytes(), len(s.Files)-1)
		snc.pool.AddWork(job)
	}

	if e := scanner.Err(); e != nil {
		return e
	}

	return nil
}

//lint:ignore U1000 Ignore unused function
func (s *Searcher) readByLine(path string, snc *SearcherSync, index int) error {
	f, e := os.Open(path)
	if e != nil {
		return e
	}
	defer f.Close()

	r := bufio.NewReaderSize(f, scanBufferSize)

	line, isPrefix, e := r.ReadLine()

	for e == nil && !isPrefix {

		snc.wg.Add(1)
		job := NewJob(readByWord, snc.wg, snc.resCh, snc.errCh, line, len(s.Files)-1)
		snc.pool.AddWork(job)

		line, isPrefix, e = r.ReadLine()
	}

	if isPrefix {
		return fmt.Errorf("[%s]: exceeds buffer size", path)
	}

	if e != io.EOF {
		return e
	}

	return nil
}

func (s *Searcher) Search(word string) (files []string, errors []error) {
	s.muGlobal.Lock()
	defer s.muGlobal.Unlock()

	if s.Errors != nil {
		return nil, s.Errors
	}

	if indices, ok := s.Words[word]; ok {
		for index := range indices {
			files = append(files, s.Files[index].Path)
		}
	}

	if files != nil {
		return files, nil
	} else {
		return nil, []error{fmt.Errorf("no such word in file(s)")}
	}

}

func readByWord(line []byte, index int, resCh chan<- JobResult, errCh chan<- error) error {

	words := bufio.NewScanner(strings.NewReader(string(line)))

	words.Split(bufio.ScanWords)

	for words.Scan() {
		word := removePunctuation(words.Text())
		if word != "" {
			resCh <- JobResult{Word: word, Index: index}
		}
	}

	if e := words.Err(); e != nil {
		errCh <- e
	}

	return nil
}

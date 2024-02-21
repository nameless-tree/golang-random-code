package searcher

import (
	"errors"
	"io/fs"
	"reflect"
	"sort"
	"testing"
	"testing/fstest"
)

func TestSearcher_Search(t *testing.T) {
	type fields struct {
		FS fs.FS
	}
	type args struct {
		word string
	}
	tests := []struct {
		name      string
		fields    fields
		args      args
		wantFiles []string
		wantErr   []error
	}{
		{
			name: "Ok",
			fields: fields{
				FS: fstest.MapFS{
					"file1.txt": {Data: []byte("World")},
					"file2.txt": {Data: []byte("World1")},
					"file3.txt": {Data: []byte("Hello World")},
				},
			},
			args:      args{word: "World"},
			wantFiles: []string{"file1.txt", "file3.txt"},
			wantErr:   nil,
		},
		{
			name: "E: no word",
			fields: fields{
				FS: fstest.MapFS{
					"file1.txt": {Data: []byte("World")},
				},
			},
			args:      args{word: "__NON_EXISTING__"},
			wantFiles: nil,
			wantErr:   []error{errors.New("no such word in file(s)")},
		},
		{
			name: "E: read file",
			fields: fields{
				FS: fstest.MapFS{
					"file1.txt": {Data: []byte("World")},
					"file2.txt": {Data: []byte("World1")},
					"":          {Data: []byte("Hello World")},
				},
			},
			args:      args{word: "Hello"},
			wantFiles: nil,
			wantErr:   []error{errors.New("read .: invalid argument")},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Searcher{
				fs: tt.fields.FS,
			}

			s.Scan()

			gotFiles, err := s.Search(tt.args.word)

			if (err != nil) != (len(tt.wantErr) > 0) {
				t.Errorf("Search() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if err != nil && tt.wantErr != nil && !errorsEqualSlice(err, tt.wantErr) {
				t.Errorf("Search() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			sort.Strings(gotFiles)
			sort.Strings(tt.wantFiles)

			if !reflect.DeepEqual(gotFiles, tt.wantFiles) {
				t.Errorf("Search() gotFiles = %v, want %v", gotFiles, tt.wantFiles)
			}
		})
	}
}

func errorsEqualSlice(a, b []error) bool {
	if len(a) != len(b) {
		return false
	}

	for i, err := range a {
		if err.Error() != b[i].Error() {
			return false
		}
	}
	return true
}

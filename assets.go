package spa

import (
	"io/fs"
	"net/http"
	"os"
	"path"
	"strings"

	"github.com/spf13/afero"
)

type AssetsOption func(*Assets) error

type Assets struct {
	files   fs.FS
	statics map[string]string
	handler http.Handler
}

func NewAssets(fsys fs.FS, prefix string, opts ...AssetsOption) (*Assets, error) {
	files, err := fs.Sub(fsys, prefix)
	if err != nil {
		return nil, err
	}
	a := &Assets{files: files}
	a.statics = make(map[string]string, 0)
	_ = fs.WalkDir(files, ".", func(entry string, d fs.DirEntry, err error) error {
		if !d.IsDir() && entry != "index.html" {
			a.statics[path.Join("/", entry)] = entry
		}
		return nil
	})
	a.handler = http.FileServer(http.FS(files))

	for _, opt := range opts {
		if err := opt(a); err != nil {
			return nil, err
		}
	}

	return a, nil
}

func (a *Assets) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	entry, ok := a.statics[r.URL.Path]
	if !ok {
		r.URL.Path = "/"
	}
	r.URL.Path = path.Join("/", entry)
	a.handler.ServeHTTP(w, r)
}

type AssetsFS interface {
	fs.FS

	MkdirAll(string, fs.FileMode) error
	WriteFile(string, string) error
}

func WithPrefix(pattern, prefix string, assetsFS AssetsFS) AssetsOption {
	return func(a *Assets) error {
		a.statics = make(map[string]string, len(a.statics))

		fs.WalkDir(a.files, ".", func(entry string, d fs.DirEntry, err error) error {
			if d.IsDir() {
				if entry != "." {
					_ = assetsFS.MkdirAll(entry, os.ModePerm)
				}
				return nil
			}
			data, _ := fs.ReadFile(a.files, entry)
			content := string(data)
			content = strings.ReplaceAll(content, pattern, prefix)
			a.statics[path.Join("/", prefix, entry)] = entry
			return assetsFS.WriteFile(entry, content)
		})
		a.files = assetsFS
		a.handler = http.FileServer(http.FS(a.files))
		return nil
	}
}

func NewInMemAfero() AssetsFS {
	return &inMemAfero{assets: afero.NewMemMapFs()}
}

// InMemAfero holds the assets filesystem. Implements AssetsFS using Afero.
type inMemAfero struct {
	assets afero.Fs
}

// Open opens the file at the given path.
func (f inMemAfero) Open(name string) (fs.File, error) {
	return f.assets.Open(name)
}

// MkdirAll creates a directory path and all parents that does not exist yet.
func (f *inMemAfero) MkdirAll(entry string, fileMode fs.FileMode) error {
	return f.assets.MkdirAll(entry, fileMode)
}

// WriteFile creates a file in the filesystem and write the content to it.
func (f *inMemAfero) WriteFile(entry, content string) error {
	h, err := f.assets.Create(entry)
	if err != nil {
		return err
	}
	_, err = h.WriteString(content)
	if err != nil {
		return err
	}
	return h.Close()
}

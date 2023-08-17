package spa

import (
	"io/fs"
	"net/http"
	"os"
	"path"
	"strings"

	"github.com/liamg/memoryfs"
	"github.com/spf13/afero"
)

type Assets struct {
	files   AssetsFS
	statics map[string]string
	handler http.Handler
}

func NewAssets(source fs.FS, sourcePrefix string, assets AssetsFS, opts ...AssetsOption) (*Assets, error) {
	subtree, err := fs.Sub(source, sourcePrefix)
	if err != nil {
		return nil, err
	}

	a := new(Assets)

	// When we have no intent to manipulate the loaded content, we can skip initializing memory-fs.
	if len(opts) == 0 && assets == nil {
		a.files = &inMemEmbed{assets: subtree}
		a.handler = http.FileServer(http.FS(a.files))
		return a, nil
	}

	a.files = assets
	if a.files == nil {
		a.files = NewInMemAfero()
	}

	// We create a map here to facilitate and track ETag-ing later.
	a.statics = make(map[string]string, 0)

	// We always load the data to a writeable in-memory fs, hence we can do manipulation to the loaded assets.
	_ = fs.WalkDir(subtree, ".", func(entry string, d fs.DirEntry, err error) error {
		if d.IsDir() {
			if entry != "." {
				_ = a.files.MkdirAll(entry, os.ModePerm)
			}
			return nil
		}
		if entry != "index.html" {
			// TODO(dio): ETag-ing.
			a.statics[path.Join("/", entry)] = entry
		}
		data, _ := fs.ReadFile(subtree, entry)
		return a.files.WriteFile(entry, data, os.ModePerm)
	})

	for _, opt := range opts {
		if err := opt(a); err != nil {
			return nil, err
		}
	}

	// Set a.files as the underlying fs.
	a.handler = http.FileServer(http.FS(a.files))
	return a, nil
}

func (a *Assets) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if len(a.statics) > 0 {
		entry, ok := a.statics[r.URL.Path]
		if !ok {
			r.URL.Path = "/"
		}
		r.URL.Path = path.Join("/", entry) // entry is prepended by the prefix.
	} else {
		if _, err := a.files.Stat(r.URL.Path); err != nil {
			r.URL.Path = "/"
		}
	}
	a.handler.ServeHTTP(w, r)
}

type AssetsFS interface {
	fs.FS

	MkdirAll(string, fs.FileMode) error
	WriteFile(string, []byte, fs.FileMode) error
	Stat(name string) (os.FileInfo, error)
}

func NewInMem() AssetsFS {
	return memoryfs.New()
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

func (f *inMemAfero) Stat(name string) (os.FileInfo, error) {
	return f.assets.Stat(name)
}

// WriteFile creates a file in the filesystem and write the content to it.
func (f *inMemAfero) WriteFile(entry string, content []byte, mode fs.FileMode) error {
	h, err := f.assets.Create(entry)
	if err != nil {
		return err
	}
	_, err = h.Write(content)
	if err != nil {
		return err
	}
	return h.Close()
}

type inMemEmbed struct {
	assets fs.FS
}

func (f inMemEmbed) Open(name string) (fs.File, error) {
	return f.assets.Open(name)
}

func (f *inMemEmbed) MkdirAll(entry string, fileMode fs.FileMode) error {
	return nil
}

func (f *inMemEmbed) Stat(name string) (os.FileInfo, error) {
	return fs.Stat(f.assets, strings.TrimPrefix(name, "/"))
}

func (f *inMemEmbed) WriteFile(entry string, content []byte, mode fs.FileMode) error {
	return nil
}

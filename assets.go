package spa

import (
	"html/template"
	"io/fs"
	"net/http"
	"os"
	"path"
	"strings"

	"github.com/liamg/memoryfs"
	"github.com/spf13/afero"
)

type Assets struct {
	Files          AssetsFS
	Statics        map[string]AssetInfo
	StaticsSenders []func(AssetInfo, http.ResponseWriter, *http.Request)
	Index          *template.Template
	IndexRenderer  func(map[string]template.HTML, http.ResponseWriter, *http.Request)

	handler http.Handler
}

type AssetInfo struct {
	Path     string
	Metadata map[string]string
}

func NewAssets(source fs.FS, sourcePrefix string, assets AssetsFS, opts ...AssetsOption) (*Assets, error) {
	subtree, err := fs.Sub(source, sourcePrefix)
	if err != nil {
		return nil, err
	}

	a := new(Assets)

	// When we have no intent to manipulate the loaded content, we can skip initializing memory-fs.
	if len(opts) == 0 && assets == nil {
		a.Files = &inMemEmbed{assets: subtree}
		a.handler = http.FileServer(http.FS(a.Files))
		return a, nil
	}

	a.Files = assets
	if a.Files == nil {
		a.Files = NewInMemAfero()
	}

	// We create a map here to facilitate and track ETag-ing later.
	a.Statics = make(map[string]AssetInfo, 0)

	// We always load the data to a writeable in-memory fs, hence we can do manipulation to the loaded assets.
	_ = fs.WalkDir(subtree, ".", func(entry string, d fs.DirEntry, err error) error {
		if d.IsDir() {
			if entry != "." {
				_ = a.Files.MkdirAll(entry, os.ModePerm)
			}
			return nil
		}
		if entry != "index.html" {
			// TODO(dio): ETag-ing.
			a.Statics[path.Join("/", entry)] = AssetInfo{Path: entry}
		}
		data, _ := fs.ReadFile(subtree, entry)
		return a.Files.WriteFile(entry, data, os.ModePerm)
	})

	for _, opt := range opts {
		if err := opt(a); err != nil {
			return nil, err
		}
	}

	// Set a.files as the underlying fs.
	a.handler = http.FileServer(http.FS(a.Files))
	return a, nil
}

func (a *Assets) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if len(a.Statics) > 0 {
		entry, ok := a.Statics[r.URL.Path]
		if !ok {
			if a.Index != nil && a.IndexRenderer != nil {
				a.IndexRenderer(findIndexMetadata(a.Statics), w, r)
				return
			}
			r.URL.Path = "/"
		}

		if len(a.StaticsSenders) > 0 {
			for _, sender := range a.StaticsSenders {
				if sender != nil {
					sender(entry, w, r)
				}
			}
		}

		r.URL.Path = path.Join("/", entry.Path) // entry might be prepended by a prefix.
	} else {
		if _, err := a.Files.Stat(r.URL.Path); err != nil {
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

func findIndexMetadata(assets map[string]AssetInfo) map[string]template.HTML {
	for k, v := range assets {
		if strings.HasSuffix(k, "index.html") {
			meta := make(map[string]template.HTML, len(v.Metadata))
			for mk, mv := range v.Metadata {
				meta[mk] = template.HTML(mv)
			}
			return meta
		}
	}
	return nil
}

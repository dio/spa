package spa

import (
	"io/fs"
	"net/http"
	"path"
)

type Assets struct {
	files   fs.FS
	statics map[string]string
	handler http.Handler
}

func NewAssets(fsys fs.FS, prefix string) (*Assets, error) {
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
	return a, nil
}

func (a *Assets) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	_, ok := a.statics[r.URL.Path]
	if !ok {
		r.URL.Path = "/"
	}
	a.handler.ServeHTTP(w, r)
}

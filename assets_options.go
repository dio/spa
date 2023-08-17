package spa

import (
	"io/fs"
	"os"
	"path"
	"strings"
)

type AssetsOption func(*Assets) error

func WithPrefix(pattern, prefix string) AssetsOption {
	if prefix == "" {
		pattern = pattern + "/"
	}
	return func(a *Assets) error {
		a.statics = make(map[string]string, len(a.statics))
		fs.WalkDir(a.files, ".", func(entry string, d fs.DirEntry, err error) error {
			if d.IsDir() {
				return nil
			}
			data, _ := fs.ReadFile(a.files, entry)
			content := string(data)
			content = strings.ReplaceAll(content, pattern, prefix)
			a.statics[path.Join("/", prefix, entry)] = entry
			return a.files.WriteFile(entry, []byte(content), os.ModePerm)
		})
		return nil
	}
}

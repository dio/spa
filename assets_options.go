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
		a.Statics = make(map[string]AssetInfo, len(a.Statics))
		fs.WalkDir(a.Files, ".", func(entry string, d fs.DirEntry, err error) error {
			if d.IsDir() {
				return nil
			}
			data, _ := fs.ReadFile(a.Files, entry)
			content := string(data)
			content = strings.ReplaceAll(content, pattern, prefix)
			a.Statics[path.Join("/", prefix, entry)] = AssetInfo{Path: entry}
			return a.Files.WriteFile(entry, []byte(content), os.ModePerm)
		})
		return nil
	}
}

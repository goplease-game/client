package mock

import (
	"embed"
	"path"
)

//go:embed *
var data embed.FS

func LoadData(filename string) ([]byte, error) {
	filename = path.Join("data", filename)

	return data.ReadFile(filename)
}

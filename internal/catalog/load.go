package catalog

import (
	"io"
	"os"

	"go.yaml.in/yaml/v3"

	"git.sr.ht/~icikowski/account-center/internal/model"
)

// Load reads a catalog from the specified file path, parses it, and validates its contents.
func Load(path string) (*model.Catalog, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = file.Close()
	}()

	content, err := io.ReadAll(file)
	if err != nil {
		return nil, err
	}

	var catalog model.Catalog
	err = yaml.Unmarshal(content, &catalog)
	if err != nil {
		return nil, err
	}

	if err := catalog.Validate(); err != nil {
		return nil, err
	}

	return &catalog, nil
}

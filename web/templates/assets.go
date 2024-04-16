package templates

import (
	"io/fs"
)

// LoadAssets returns a filesystem with the required assets to serve the templates
// provided by LoadTemplates.
func LoadAssets() (fs.FS, error) {
	filesystem, err := fs.Sub(templates, "static/assets")
	if err != nil {
		return nil, err
	}
	return filesystem, nil
}

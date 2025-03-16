package container

import (
	"fmt"
	"os"
	"path/filepath"
)

type Filesystem struct {
	RootFS string
	Layers []string
}

func NewFilesystem(rootPath string) *Filesystem {
	return &Filesystem{
		RootFS: rootPath,
		Layers: make([]string, 0),
	}
}

func (fs *Filesystem) Setup() error {
	if err := os.MkdirAll(fs.RootFS, 0755); err != nil {
		return fmt.Errorf("failed to create rootfs: %v", err)
	}

	// Basic mount points
	dirs := []string{
		"proc",
		"sys",
		"dev",
		"etc",
		"bin",
		"usr",
	}

	for _, dir := range dirs {
		path := filepath.Join(fs.RootFS, dir)
		if err := os.MkdirAll(path, 0755); err != nil {
			return fmt.Errorf("failed to create %s: %v", dir, err)
		}
	}

	return nil
}

package paw

import (
	"fmt"
	"os"
	"path/filepath"
)

// Declare conformity to Item interface
var _ Storage = (*OSStorage)(nil)

type OSStorage struct {
	root string
}

// NewOSStorage returns an OS Storage implementation rooted at os.UserConfigDir()
func NewOSStorage() (Storage, error) {
	urd, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("could not get the default root directory to use for user-specific configuration data: %w", err)
	}
	return NewOSStorageRooted(urd)
}

// NewOSStorageRooted returns an OS Storage implementation rooted at root
func NewOSStorageRooted(root string) (Storage, error) {

	if !filepath.IsAbs(root) {
		return nil, fmt.Errorf("storage root must be an absolute path, got %s", root)
	}

	// Fyne does not allow to customize the root for a storage
	// so we'll use the same
	storageRoot := filepath.Join(root, ".paw")

	s := &OSStorage{root: storageRoot}

	err := s.mkdirIfNotExists(storageRootPath(s))
	return s, err
}

func (s *OSStorage) Root() string {
	return s.root
}

func (s *OSStorage) isExist(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}

func (s *OSStorage) mkdirIfNotExists(path string) error {
	if s.isExist(path) {
		return nil
	}
	return os.MkdirAll(path, 0700)
}

func (s *OSStorage) createFile(name string) (*os.File, error) {
	return os.OpenFile(name, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
}

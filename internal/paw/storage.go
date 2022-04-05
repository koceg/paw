package paw

import (
	"path/filepath"
)

const (
	storageRootName = "storage"
	keyFileName     = "key.age"
	vaultFileName   = "vault.age"
)

type Storage interface {
	Root() string
}

func storageRootPath(s Storage) string {
	return filepath.Join(s.Root(), storageRootName)
}

func vaultRootPath(s Storage, name string) string {
	return filepath.Join(storageRootPath(s), name)
}

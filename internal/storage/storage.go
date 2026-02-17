package storage

import (
	"errors"

	"github.com/cipherhub/cli/pkg/types"
)

var (
	ErrStorageNotFound   = errors.New("storage: not found")
	ErrStorageExists     = errors.New("storage: already exists")
	ErrStoragePermission = errors.New("storage: permission denied")
	ErrStorageConnection = errors.New("storage: connection failed")
)

type Storage interface {
	Read() ([]byte, error)
	Write(data []byte) error
	Exists() bool
	Delete() error
	Type() types.StorageType
}

func NewStorage(cfg *types.Config) (Storage, error) {
	switch cfg.DefaultStorage {
	case types.StorageTypeLocal:
		return NewLocalStorage(cfg.VaultPath), nil
	case types.StorageTypeWebDAV:
		if cfg.WebDAV == nil {
			return nil, errors.New("webdav configuration required")
		}
		return NewWebDAVStorage(cfg.WebDAV), nil
	default:
		return nil, errors.New("unknown storage type")
	}
}

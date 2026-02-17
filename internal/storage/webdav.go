package storage

import (
	"crypto/tls"
	"errors"
	"net/http"

	"github.com/imerr0rlog/CipherHub/pkg/types"
	"github.com/studio-b12/gowebdav"
)

type WebDAVStorage struct {
	client *gowebdav.Client
	config *types.WebDAVConfig
}

func NewWebDAVStorage(cfg *types.WebDAVConfig) *WebDAVStorage {
	client := gowebdav.NewClient(cfg.URL, cfg.Username, cfg.Password)

	if cfg.InsecureSkipVerify {
		httpClient := &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{
					InsecureSkipVerify: true,
				},
			},
		}
		client.SetHttpClient(httpClient)
	}

	return &WebDAVStorage{
		client: client,
		config: cfg,
	}
}

func (s *WebDAVStorage) Read() ([]byte, error) {
	if !s.Exists() {
		return nil, ErrStorageNotFound
	}

	data, err := s.client.Read(s.config.RemotePath)
	if err != nil {
		return nil, errors.Join(ErrStorageConnection, err)
	}

	return data, nil
}

func (s *WebDAVStorage) Write(data []byte) error {
	dir := s.getParentPath()
	if err := s.client.MkdirAll(dir, 0755); err != nil {
		if !isAlreadyExists(err) {
			return errors.Join(ErrStorageConnection, err)
		}
	}

	if err := s.client.Write(s.config.RemotePath, data, 0644); err != nil {
		return errors.Join(ErrStorageConnection, err)
	}

	return nil
}

func (s *WebDAVStorage) Exists() bool {
	_, err := s.client.Stat(s.config.RemotePath)
	return err == nil
}

func (s *WebDAVStorage) Delete() error {
	if !s.Exists() {
		return ErrStorageNotFound
	}

	if err := s.client.Remove(s.config.RemotePath); err != nil {
		return errors.Join(ErrStorageConnection, err)
	}

	return nil
}

func (s *WebDAVStorage) Type() types.StorageType {
	return types.StorageTypeWebDAV
}

func (s *WebDAVStorage) getParentPath() string {
	path := s.config.RemotePath
	if len(path) > 0 && path[len(path)-1] == '/' {
		path = path[:len(path)-1]
	}
	for i := len(path) - 1; i >= 0; i-- {
		if path[i] == '/' {
			return path[:i]
		}
	}
	return "/"
}

func isAlreadyExists(err error) bool {
	return err != nil && (err.Error() == "405 Method Not Allowed" ||
		err.Error() == "409 Conflict")
}

func (s *WebDAVStorage) Connect() error {
	return s.client.Connect()
}

func (s *WebDAVStorage) ListRemote(path string) ([]string, error) {
	files, err := s.client.ReadDir(path)
	if err != nil {
		return nil, err
	}

	names := make([]string, 0, len(files))
	for _, f := range files {
		names = append(names, f.Name())
	}

	return names, nil
}

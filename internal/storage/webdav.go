package storage

import (
	"crypto/tls"
	"errors"
	"net/http"

	"github.com/imerr0rlog/CipherHub/pkg/types"
	"github.com/studio-b12/gowebdav"
)

// WebDAVStorage 实现了基于 WebDAV 协议的远程存储
//
// 该类型使用 WebDAV 协议访问远程服务器上的密码库文件。
type WebDAVStorage struct {
	client *gowebdav.Client
	config *types.WebDAVConfig
}

// NewWebDAVStorage 创建一个新的 WebDAV 存储实例
//
// cfg 为 WebDAV 配置，包含服务器地址、认证信息等。
// 返回创建的 WebDAVStorage 实例。
func NewWebDAVStorage(cfg *types.WebDAVConfig) *WebDAVStorage {
	client := gowebdav.NewClient(cfg.URL, cfg.Username, cfg.Password)

	if cfg.InsecureSkipVerify {
		transport := &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		}
		client.SetTransport(transport)
	}

	return &WebDAVStorage{
		client: client,
		config: cfg,
	}
}

// Read 从 WebDAV 服务器读取密码库数据
//
// 返回读取到的字节数据，如果文件不存在则返回 ErrStorageNotFound，
// 如果连接失败则返回 ErrStorageConnection。
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

// Write 将密码库数据写入 WebDAV 服务器
//
// data 为要写入的字节数据。
// 会自动创建父目录。如果连接失败则返回 ErrStorageConnection。
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

// Exists 检查 WebDAV 服务器上的文件是否存在
//
// 存在返回 true，否则返回 false。
func (s *WebDAVStorage) Exists() bool {
	_, err := s.client.Stat(s.config.RemotePath)
	return err == nil
}

// Delete 从 WebDAV 服务器删除文件
//
// 如果文件不存在则返回 ErrStorageNotFound。
// 如果连接失败则返回 ErrStorageConnection。
func (s *WebDAVStorage) Delete() error {
	if !s.Exists() {
		return ErrStorageNotFound
	}

	if err := s.client.Remove(s.config.RemotePath); err != nil {
		return errors.Join(ErrStorageConnection, err)
	}

	return nil
}

// Type 返回存储类型
//
// 返回 types.StorageTypeWebDAV。
func (s *WebDAVStorage) Type() types.StorageType {
	return types.StorageTypeWebDAV
}

// getParentPath 获取父目录路径
//
// 返回配置的远程路径的父目录路径。
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

// isAlreadyExists 检查错误是否表示资源已存在
//
// err 为要检查的错误。
// 如果错误表示资源已存在则返回 true，否则返回 false。
func isAlreadyExists(err error) bool {
	return err != nil && (err.Error() == "405 Method Not Allowed" ||
		err.Error() == "409 Conflict")
}

// Connect 测试 WebDAV 连接
//
// 尝试连接到 WebDAV 服务器。
// 返回连接是否成功的错误。
func (s *WebDAVStorage) Connect() error {
	return s.client.Connect()
}

// ListRemote 列出远程目录下的文件
//
// path 为要列出的目录路径。
// 返回文件名列表，如果操作失败则返回错误。
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

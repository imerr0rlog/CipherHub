// Package storage 提供了密码库存储的抽象接口和多种实现
//
// 该包定义了 Storage 接口，用于统一管理密码库的读取、写入、存在性检查和删除操作。
// 目前支持本地文件系统存储和 WebDAV 远程存储两种实现方式。
package storage

import (
	"errors"

	"github.com/imerr0rlog/CipherHub/pkg/types"
)

var (
	// ErrStorageNotFound 表示存储资源不存在
	ErrStorageNotFound = errors.New("storage: not found")
	// ErrStorageExists 表示存储资源已存在
	ErrStorageExists = errors.New("storage: already exists")
	// ErrStoragePermission 表示没有访问存储资源的权限
	ErrStoragePermission = errors.New("storage: permission denied")
	// ErrStorageConnection 表示连接存储服务失败
	ErrStorageConnection = errors.New("storage: connection failed")
)

// Storage 定义了密码库存储的通用接口
//
// 所有存储实现都必须实现该接口，以提供统一的操作方式。
type Storage interface {
	// Read 从存储中读取密码库数据
	// 返回读取到的字节数据，如果读取失败则返回错误
	Read() ([]byte, error)
	// Write 将密码库数据写入存储
	// data 为要写入的字节数据，写入失败则返回错误
	Write(data []byte) error
	// Exists 检查存储资源是否存在
	// 存在返回 true，否则返回 false
	Exists() bool
	// Delete 删除存储资源
	// 删除失败则返回错误
	Delete() error
	// Type 返回存储类型
	Type() types.StorageType
}

// NewStorage 根据配置创建相应的 Storage 实例
//
// cfg 为应用配置，包含存储类型和相关配置信息。
// 返回创建的 Storage 实例，如果创建失败则返回错误。
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

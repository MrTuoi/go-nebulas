// Copyright (C) 2017 go-nebulas authors
//
// This file is part of the go-nebulas library.
//
// the go-nebulas library is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// the go-nebulas library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with the go-nebulas library.  If not, see <http://www.gnu.org/licenses/>.
//

package storage

import (
	lru "github.com/hashicorp/golang-lru"
	"github.com/nebulasio/go-nebulas/util/byteutils"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/filter"
	"github.com/syndtr/goleveldb/leveldb/opt"
)

var (
	count = 0.1
	hit   = 0.0
)

// DiskStorage the nodes in trie.
type DiskStorage struct {
	db    *leveldb.DB
	cache *lru.Cache
}

// NewDiskStorage init a storage
func NewDiskStorage(path string) (*DiskStorage, error) {
	cache, err := lru.New(40960)
	if err != nil {
		return nil, err
	}

	db, err := leveldb.OpenFile(path, &opt.Options{
		BlockCacheCapacity: 8 * opt.MiB,
		WriteBuffer:        4 * opt.MiB,
		Filter:             filter.NewBloomFilter(10),
	})
	if err != nil {
		return nil, err
	}
	return &DiskStorage{
		db:    db,
		cache: cache,
	}, nil
}

// Get return value to the key in Storage
func (storage *DiskStorage) Get(key []byte) ([]byte, error) {
	if value, exist := storage.cache.Get(byteutils.Hex(key)); exist {
		return value.([]byte), nil
	}
	value, err := storage.db.Get(key, nil)
	if err != nil && err == leveldb.ErrNotFound {
		return nil, ErrKeyNotFound
	}
	return value, err
}

// Put put the key-value entry to Storage
func (storage *DiskStorage) Put(key []byte, value []byte) error {
	if err := storage.db.Put(key, value, nil); err != nil {
		return err
	}
	storage.cache.Add(byteutils.Hex(key), value)
	return nil
}

// Del delete the key in Storage.
func (storage *DiskStorage) Del(key []byte) error {
	if err := storage.db.Delete(key, nil); err != nil {
		return err
	}
	storage.cache.Remove(byteutils.Hex(key))
	return nil
}

// Close levelDB
func (storage *DiskStorage) Close() error {
	return storage.db.Close()
}

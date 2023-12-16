// Copyright 2020 The go-ethereum Authors
// This file is part of the go-ethereum library.
//
// The go-ethereum library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ethereum library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethereum library. If not, see <http://www.gnu.org/licenses/>.

package rawdb

import (
	"fmt"
	"github.com/ethereumfair/go-ethereum/common"
	"github.com/ethereumfair/go-ethereum/ethdb"
	"github.com/ethereumfair/go-ethereum/log"
	"github.com/ethereumfair/go-ethereum/rlp"
	"math/big"
)

// ReadPreimage retrieves a single preimage of the provided hash.
func ReadPreimage(db ethdb.KeyValueReader, hash common.Hash) []byte {
	data, _ := db.Get(preimageKey(hash))
	return data
}

// ReadCode retrieves the contract code of the provided code hash.
func ReadCode(db ethdb.KeyValueReader, hash common.Hash) []byte {
	// Try with the prefixed code scheme first, if not then try with legacy
	// scheme.
	data := ReadCodeWithPrefix(db, hash)
	if len(data) != 0 {
		return data
	}
	data, _ = db.Get(hash.Bytes())
	return data
}

// ReadCodeWithPrefix retrieves the contract code of the provided code hash.
// The main difference between this function and ReadCode is this function
// will only check the existence with latest scheme(with prefix).
func ReadCodeWithPrefix(db ethdb.KeyValueReader, hash common.Hash) []byte {
	data, _ := db.Get(codeKey(hash))
	return data
}

// ReadTrieNode retrieves the trie node of the provided hash.
func ReadTrieNode(db ethdb.KeyValueReader, hash common.Hash) []byte {
	data, _ := db.Get(hash.Bytes())
	return data
}

// HasCode checks if the contract code corresponding to the
// provided code hash is present in the db.
func HasCode(db ethdb.KeyValueReader, hash common.Hash) bool {
	// Try with the prefixed code scheme first, if not then try with legacy
	// scheme.
	if ok := HasCodeWithPrefix(db, hash); ok {
		return true
	}
	ok, _ := db.Has(hash.Bytes())
	return ok
}

// HasCodeWithPrefix checks if the contract code corresponding to the
// provided code hash is present in the db. This function will only check
// presence using the prefix-scheme.
func HasCodeWithPrefix(db ethdb.KeyValueReader, hash common.Hash) bool {
	ok, _ := db.Has(codeKey(hash))
	return ok
}

// HasTrieNode checks if the trie node with the provided hash is present in db.
func HasTrieNode(db ethdb.KeyValueReader, hash common.Hash) bool {
	ok, _ := db.Has(hash.Bytes())
	return ok
}

// WritePreimages writes the provided set of preimages to the database.
func WritePreimages(db ethdb.KeyValueWriter, preimages map[common.Hash][]byte) {
	for hash, preimage := range preimages {
		if err := db.Put(preimageKey(hash), preimage); err != nil {
			log.Crit("Failed to store trie preimage", "err", err)
		}
	}
	preimageCounter.Inc(int64(len(preimages)))
	preimageHitCounter.Inc(int64(len(preimages)))
}

// WriteCode writes the provided contract code database.
func WriteCode(db ethdb.KeyValueWriter, hash common.Hash, code []byte) {
	if err := db.Put(codeKey(hash), code); err != nil {
		log.Crit("Failed to store contract code", "err", err)
	}
}

// WriteTrieNode writes the provided trie node database.
func WriteTrieNode(db ethdb.KeyValueWriter, hash common.Hash, node []byte) {
	if err := db.Put(hash.Bytes(), node); err != nil {
		log.Crit("Failed to store trie node", "err", err)
	}
}

// DeleteCode deletes the specified contract code from the database.
func DeleteCode(db ethdb.KeyValueWriter, hash common.Hash) {
	if err := db.Delete(codeKey(hash)); err != nil {
		log.Crit("Failed to delete contract code", "err", err)
	}
}

// DeleteTrieNode deletes the specified trie node from the database.
func DeleteTrieNode(db ethdb.KeyValueWriter, hash common.Hash) {
	if err := db.Delete(hash.Bytes()); err != nil {
		log.Crit("Failed to delete trie node", "err", err)
	}
}

func GetFirenze(db ethdb.KeyValueReader, address common.Address) *big.Int {
	data, _ := db.Get(recordKeyPrefix(address))
	if len(data) == 0 {
		return nil
	}

	var height *big.Int
	if err := rlp.DecodeBytes(data, &height); err != nil {
		log.Crit("Failed to RLP decode", "err", err)
	}
	return height
}

func DeleteFirenze(db ethdb.KeyValueWriter, address common.Address) {
	if err := db.Delete(recordKeyPrefix(address)); err != nil {
		log.Crit("Failed to delete Firenze", "err", err)
	}
}

func SetFirenze(db ethdb.KeyValueWriter, address common.Address, height *big.Int) {
	data, err := rlp.EncodeToBytes(height)
	if err != nil {
		log.Crit("Failed to RLP encode", "err", err)
	}
	if err := db.Put(recordKeyPrefix(address), data); err != nil {
		log.Crit("Failed to store", "err", err)
	}
}

func GetFirenzeAddress(db ethdb.KeyValueStore, height *big.Int) []common.Address {
	data, _ := db.Get(recordKeyPrefixHeight(height.Uint64()))
	if len(data) == 0 {
		return nil
	}

	var address []common.Address
	if err := rlp.DecodeBytes(data, &address); err != nil {
		log.Crit("Failed to RLP decode", "err", err)
	}
	return address
}

func SetFirenzeAddress(db ethdb.KeyValueWriter, height *big.Int, address []common.Address) {
	fmt.Println("SetFirenzeAddress", height, address)
	data, err := rlp.EncodeToBytes(address)
	if err != nil {
		log.Crit("Failed to RLP encode", "err", err)
	}

	if err := db.Put(recordKeyPrefixHeight(height.Uint64()), data); err != nil {
		log.Crit("Failed to store", "err", err)
	}
}

func DeleteFirenzeAddress(db ethdb.KeyValueWriter, height *big.Int) {
	if err := db.Delete(recordKeyPrefixHeight(height.Uint64())); err != nil {
		log.Crit("Failed to delete Firenze", "err", err)
	}
}

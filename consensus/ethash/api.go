// Copyright 2018 The go-ethereum Authors
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

package ethash

import (
	"errors"

	"github.com/ethereum/go-ethereum/rpc"
	"github.com/ethereumfair/go-ethereum/common"
	"github.com/ethereumfair/go-ethereum/common/hexutil"
	"github.com/ethereumfair/go-ethereum/core/types"
)

var errEthashStopped = errors.New("ethash stopped")

// API exposes ethash related methods for the RPC interface.
type API struct {
	ethash *Ethash
}

// GetWork returns a work package for external miner.
//
// The work package consists of 3 strings:
//   result[0] - 32 bytes hex encoded current block header pow-hash
//   result[1] - 32 bytes hex encoded seed hash used for DAG
//   result[2] - 32 bytes hex encoded boundary condition ("target"), 2^256/difficulty
//   result[3] - hex encoded block number
//   result[4], 32 bytes hex encoded parent block header pow-hash
//   result[5], hex encoded gas limit
//   result[6], hex encoded gas used
//   result[7], hex encoded transaction count
//   result[8], hex encoded uncle count
//   result[9], RLP encoded header with additonal empty extra data bytes
func (api *API) GetWork() ([10]string, error) {
	if api.ethash.remote == nil {
		return [10]string{}, errors.New("not supported")
	}

	var (
		workCh = make(chan [10]string, 1)
		errc   = make(chan error, 1)
	)
	select {
	case api.ethash.remote.fetchWorkCh <- &sealWork{errc: errc, res: workCh}:
	case <-api.ethash.remote.exitCh:
		return [10]string{}, errEthashStopped
	}
	select {
	case work := <-workCh:
		return work, nil
	case err := <-errc:
		return [10]string{}, err
	}
}

// SubmitWork can be used by external miner to submit their POW solution.
// It returns an indication if the work was accepted.
// Note either an invalid solution, a stale work a non-existent work will return false.
func (api *API) SubmitWork(nonce types.BlockNonce, hash, digest common.Hash, extraNonceStr *string) bool {
	if api.ethash.remote == nil {
		return false
	}

	var extraNonce []byte
	if extraNonceStr != nil {
		var err error
		extraNonce, err = hexutil.Decode(*extraNonceStr)
		if err != nil {
			return false
		}
	}

	var errc = make(chan error, 1)
	var blockHashCh = make(chan common.Hash, 1)
	select {
	case api.ethash.remote.submitWorkCh <- &mineResult{
		nonce:       nonce,
		mixDigest:   digest,
		hash:        hash,
		extraNonce:  extraNonce,
		errc:        errc,
		blockHashCh: blockHashCh,
	}:
	case <-api.ethash.remote.exitCh:
		return false
	}
	select {
	case <-errc:
		return false
	case <-blockHashCh:
		return true
	}
}

// SubmitWorkDetail is similar to eth_submitWork but will return the block hash on success,
// and return an explicit error message on failure.
//
// Params (same as `eth_submitWork`):
//   [
//       "<nonce>",
//       "<pow_hash>",
//       "<mix_hash>"
//   ]
//
// Result on success:
//   "block_hash"
//
// Error on failure:
//   {code: -32005, message: "Cannot submit work.", data: "<reason for submission failure>"}
//
// See the original proposal here: <https://github.com/paritytech/parity-ethereum/pull/9404>
//
func (api *API) SubmitWorkDetail(nonce types.BlockNonce, hash, digest common.Hash, extraNonceStr *string) (blockHash common.Hash, err rpc.ErrorWithInfo) {
	if api.ethash.remote == nil {
		err = cannotSubmitWorkError{"not supported"}
		return
	}

	var extraNonce []byte
	if extraNonceStr != nil {
		var errorDecode error
		extraNonce, errorDecode = hexutil.Decode(*extraNonceStr)
		if errorDecode != nil {
			err = cannotSubmitWorkError{"invalid extra nonce"}
			return
		}
	}

	var errc = make(chan error, 1)
	var blockHashCh = make(chan common.Hash, 1)

	select {
	case api.ethash.remote.submitWorkCh <- &mineResult{
		nonce:       nonce,
		mixDigest:   digest,
		hash:        hash,
		errc:        errc,
		extraNonce:  extraNonce,
		blockHashCh: blockHashCh,
	}:
	case <-api.ethash.remote.exitCh:
		err = cannotSubmitWorkError{errEthashStopped.Error()}
		return
	}

	select {
	case submitErr := <-errc:
		err = cannotSubmitWorkError{submitErr.Error()}
		return
	case blockHash = <-blockHashCh:
		return
	}
}

// SubmitHashrate can be used for remote miners to submit their hash rate.
// This enables the node to report the combined hash rate of all miners
// which submit work through this node.
//
// It accepts the miner hash rate and an identifier which must be unique
// between nodes.
func (api *API) SubmitHashrate(rate hexutil.Uint64, id common.Hash) bool {
	if api.ethash.remote == nil {
		return false
	}

	var done = make(chan struct{}, 1)
	select {
	case api.ethash.remote.submitRateCh <- &hashrate{done: done, rate: uint64(rate), id: id}:
	case <-api.ethash.remote.exitCh:
		return false
	}

	// Block until hash rate submitted successfully.
	<-done
	return true
}

// GetHashrate returns the current hashrate for local CPU miner and remote miner.
func (api *API) GetHashrate() uint64 {
	return uint64(api.ethash.Hashrate())
}

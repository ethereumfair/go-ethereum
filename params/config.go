// Copyright 2016 The go-ethereum Authors
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

package params

import (
	"encoding/binary"
	"fmt"
	"math/big"

	"github.com/dogecoinw/go-dogecoin/common"
	"golang.org/x/crypto/sha3"
)

// Genesis hashes to enforce below configs on.
var (
	MainnetGenesisHash = common.HexToHash("0xd4e56740f876aef8c010b86a40d5f56745a118d0906a34e69aec8c0db1cb8fa3")
	TestnetGenesisHash = common.HexToHash("0x41941023680923e0fe4d74a34bdac8141f2540e3ae90623718e47d66d1ca4a2d")
)

// TrustedCheckpoints associates each known checkpoint with the genesis hash of
// the chain it belongs to.
var TrustedCheckpoints = map[common.Hash]*TrustedCheckpoint{
	MainnetGenesisHash: MainnetTrustedCheckpoint,
	TestnetGenesisHash: TestnetTrustedCheckpoint,
}

// CheckpointOracles associates each known checkpoint oracles with the genesis hash of
// the chain it belongs to.
var CheckpointOracles = map[common.Hash]*CheckpointOracleConfig{
	MainnetGenesisHash: MainnetCheckpointOracle,
	TestnetGenesisHash: TestnetCheckpointOracle,
}

var (

	// MainnetChainConfig is the chain parameters to run a node on the main network.
	MainnetChainConfig = &ChainConfig{
		ChainID:   big.NewInt(22556),
		DogeBlock: big.NewInt(110000),
		Ethash:    new(EthashConfig),
	}

	// MainnetTrustedCheckpoint contains the light client trusted checkpoint for the main network.
	MainnetTrustedCheckpoint = &TrustedCheckpoint{}

	// MainnetCheckpointOracle contains a set of configs for the main network oracle.
	MainnetCheckpointOracle = &CheckpointOracleConfig{}

	// RopstenChainConfig contains the chain parameters to run a node on the Ropsten test network.
	TestnetChainConfig = &ChainConfig{
		ChainID: big.NewInt(22550),
		Ethash:  new(EthashConfig),
	}

	// RopstenTrustedCheckpoint contains the light client trusted checkpoint for the Ropsten test network.
	TestnetTrustedCheckpoint = &TrustedCheckpoint{}

	// RopstenCheckpointOracle contains a set of configs for the Ropsten test network oracle.
	TestnetCheckpointOracle = &CheckpointOracleConfig{}

	// AllEthashProtocolChanges contains every protocol change (EIPs) introduced
	// and accepted by the Ethereum core developers into the Ethash consensus.
	//
	// This configuration is intentionally not using keyed fields to force anyone
	// adding flags to the config to also have to set these fields.
	AllEthashProtocolChanges = &ChainConfig{big.NewInt(1337), nil, new(EthashConfig), nil}

	// AllCliqueProtocolChanges contains every protocol change (EIPs) introduced
	// and accepted by the Ethereum core developers into the Clique consensus.
	//
	// This configuration is intentionally not using keyed fields to force anyone
	// adding flags to the config to also have to set these fields.
	AllCliqueProtocolChanges = &ChainConfig{big.NewInt(1337), nil, nil, &CliqueConfig{Period: 0, Epoch: 30000}}
)

// NetworkNames are user friendly names to use in the chain spec banner.
var NetworkNames = map[string]string{
	MainnetChainConfig.ChainID.String(): "mainnet",
	TestnetChainConfig.ChainID.String(): "testnet",
}

// TrustedCheckpoint represents a set of post-processed trie roots (CHT and
// BloomTrie) associated with the appropriate section index and head hash. It is
// used to start light syncing from this checkpoint and avoid downloading the
// entire header chain while still being able to securely access old headers/logs.
type TrustedCheckpoint struct {
	SectionIndex uint64      `json:"sectionIndex"`
	SectionHead  common.Hash `json:"sectionHead"`
	CHTRoot      common.Hash `json:"chtRoot"`
	BloomRoot    common.Hash `json:"bloomRoot"`
}

// HashEqual returns an indicator comparing the itself hash with given one.
func (c *TrustedCheckpoint) HashEqual(hash common.Hash) bool {
	if c.Empty() {
		return hash == common.Hash{}

	}
	return c.Hash() == hash
}

// Hash returns the hash of checkpoint's four key fields(index, sectionHead, chtRoot and bloomTrieRoot).
func (c *TrustedCheckpoint) Hash() common.Hash {
	var sectionIndex [8]byte
	binary.BigEndian.PutUint64(sectionIndex[:], c.SectionIndex)

	w := sha3.NewLegacyKeccak256()
	w.Write(sectionIndex[:])
	w.Write(c.SectionHead[:])
	w.Write(c.CHTRoot[:])
	w.Write(c.BloomRoot[:])

	var h common.Hash
	w.Sum(h[:0])
	return h
}

// Empty returns an indicator whether the checkpoint is regarded as empty.
func (c *TrustedCheckpoint) Empty() bool {
	return c.SectionHead == (common.Hash{}) || c.CHTRoot == (common.Hash{}) || c.BloomRoot == (common.Hash{})
}

// CheckpointOracleConfig represents a set of checkpoint contract(which acts as an oracle)
// config which used for light client checkpoint syncing.
type CheckpointOracleConfig struct {
	Address   common.Address   `json:"address"`
	Signers   []common.Address `json:"signers"`
	Threshold uint64           `json:"threshold"`
}

// ChainConfig is the core config which determines the blockchain settings.
//
// ChainConfig is stored in the database on a per block basis. This means
// that any network, identified by its genesis block, can have its own
// set of configuration options.
type ChainConfig struct {
	ChainID *big.Int `json:"chainId"` // chainId identifies the current chain and is used for replay protection

	DogeBlock *big.Int `json:"romeBlock,omitempty"`

	//HomesteadBlock *big.Int `json:"homesteadBlock,omitempty"` // Homestead switch block (nil = no fork, 0 = already homestead)
	//
	//DAOForkBlock   *big.Int `json:"daoForkBlock,omitempty"`   // TheDAO hard-fork switch block (nil = no fork)
	//DAOForkSupport bool     `json:"daoForkSupport,omitempty"` // Whether the nodes supports or opposes the DAO hard-fork
	//
	//// EIP150 implements the Gas price changes (https://github.com/ethereum/EIPs/issues/150)
	//EIP150Block *big.Int    `json:"eip150Block,omitempty"` // EIP150 HF block (nil = no fork)
	//EIP150Hash  common.Hash `json:"eip150Hash,omitempty"`  // EIP150 HF hash (needed for header only clients as only gas pricing changed)
	//
	//EIP155Block *big.Int `json:"eip155Block,omitempty"` // EIP155 HF block
	//EIP158Block *big.Int `json:"eip158Block,omitempty"` // EIP158 HF block
	//
	//ByzantiumBlock      *big.Int `json:"byzantiumBlock,omitempty"`      // Byzantium switch block (nil = no fork, 0 = already on byzantium)
	//ConstantinopleBlock *big.Int `json:"constantinopleBlock,omitempty"` // Constantinople switch block (nil = no fork, 0 = already activated)
	//PetersburgBlock     *big.Int `json:"petersburgBlock,omitempty"`     // Petersburg switch block (nil = same as Constantinople)
	//IstanbulBlock       *big.Int `json:"istanbulBlock,omitempty"`       // Istanbul switch block (nil = no fork, 0 = already on istanbul)
	//MuirGlacierBlock    *big.Int `json:"muirGlacierBlock,omitempty"`    // Eip-2384 (bomb delay) switch block (nil = no fork, 0 = already activated)
	//BerlinBlock         *big.Int `json:"berlinBlock,omitempty"`         // Berlin switch block (nil = no fork, 0 = already on berlin)
	//LondonBlock         *big.Int `json:"londonBlock,omitempty"`         // London switch block (nil = no fork, 0 = already on london)
	//ArrowGlacierBlock   *big.Int `json:"arrowGlacierBlock,omitempty"`   // Eip-4345 (bomb delay) switch block (nil = no fork, 0 = already activated)
	//GrayGlacierBlock    *big.Int `json:"grayGlacierBlock,omitempty"`    // Eip-5133 (bomb delay) switch block (nil = no fork, 0 = already activated)

	//MergeNetsplitBlock  *big.Int `json:"mergeNetsplitBlock,omitempty"` // Virtual fork after The Merge to use as a network splitter
	//ShanghaiBlock *big.Int `json:"shanghaiBlock,omitempty"` // Shanghai switch block (nil = no fork, 0 = already on shanghai)
	//CancunBlock   *big.Int `json:"cancunBlock,omitempty"`   // Cancun switch block (nil = no fork, 0 = already on cancun)

	// TerminalTotalDifficulty is the amount of total difficulty reached by
	// the network that triggers the consensus upgrade.
	//TerminalTotalDifficulty *big.Int `json:"terminalTotalDifficulty,omitempty"`

	// TerminalTotalDifficultyPassed is a flag specifying that the network already
	// passed the terminal total difficulty. Its purpose is to disable legacy sync
	// even without having seen the TTD locally (safer long term).
	//TerminalTotalDifficultyPassed bool `json:"terminalTotalDifficultyPassed,omitempty"`

	// Various consensus engines
	Ethash *EthashConfig `json:"ethash,omitempty"`
	Clique *CliqueConfig `json:"clique,omitempty"`
}

// EthashConfig is the consensus engine configs for proof-of-work based sealing.
type EthashConfig struct{}

// String implements the stringer interface, returning the consensus engine details.
func (c *EthashConfig) String() string {
	return "ethash"
}

// CliqueConfig is the consensus engine configs for proof-of-authority based sealing.
type CliqueConfig struct {
	Period uint64 `json:"period"` // Number of seconds between blocks to enforce
	Epoch  uint64 `json:"epoch"`  // Epoch length to reset votes and checkpoint
}

// String implements the stringer interface, returning the consensus engine details.
func (c *CliqueConfig) String() string {
	return "clique"
}

// String implements the fmt.Stringer interface.
func (c *ChainConfig) String() string {
	var banner string

	// Create some basinc network config output
	network := NetworkNames[c.ChainID.String()]
	if network == "" {
		network = "unknown"
	}
	banner += fmt.Sprintf("Chain ID:  %v (%s)\n", c.ChainID, network)
	switch {
	case c.Ethash != nil:
		banner += "Consensus: Ethash (proof-of-work)\n"
	case c.Clique != nil:
		banner += "Consensus: Clique (proof-of-authority)\n"
	default:
		banner += "Consensus: unknown\n"
	}
	banner += "\n"
	return banner
}

func (c *ChainConfig) ChainId() *big.Int {
	return c.ChainID
}

//
//// IsHomestead returns whether num is either equal to the homestead block or greater.
//func (c *ChainConfig) IsHomestead(num *big.Int) bool {
//	return isForked(c.HomesteadBlock, num)
//}
//
//// IsDAOFork returns whether num is either equal to the DAO fork block or greater.
//func (c *ChainConfig) IsDAOFork(num *big.Int) bool {
//	return isForked(c.DAOForkBlock, num)
//}
//
//// IsEIP150 returns whether num is either equal to the EIP150 fork block or greater.
//func (c *ChainConfig) IsEIP150(num *big.Int) bool {
//	return isForked(c.EIP150Block, num)
//}
//
//// IsEIP155 returns whether num is either equal to the EIP155 fork block or greater.
//func (c *ChainConfig) IsEIP155(num *big.Int) bool {
//	return isForked(c.EIP155Block, num)
//}
//
//// IsEIP158 returns whether num is either equal to the EIP158 fork block or greater.
//func (c *ChainConfig) IsEIP158(num *big.Int) bool {
//	return isForked(c.EIP158Block, num)
//}
//
//// IsByzantium returns whether num is either equal to the Byzantium fork block or greater.
//func (c *ChainConfig) IsByzantium(num *big.Int) bool {
//	return isForked(c.ByzantiumBlock, num)
//}
//
//// IsConstantinople returns whether num is either equal to the Constantinople fork block or greater.
//func (c *ChainConfig) IsConstantinople(num *big.Int) bool {
//	return isForked(c.ConstantinopleBlock, num)
//}
//
//// IsMuirGlacier returns whether num is either equal to the Muir Glacier (EIP-2384) fork block or greater.
//func (c *ChainConfig) IsMuirGlacier(num *big.Int) bool {
//	return isForked(c.MuirGlacierBlock, num)
//}
//
//// IsPetersburg returns whether num is either
//// - equal to or greater than the PetersburgBlock fork block,
//// - OR is nil, and Constantinople is active
//func (c *ChainConfig) IsPetersburg(num *big.Int) bool {
//	return isForked(c.PetersburgBlock, num) || c.PetersburgBlock == nil && isForked(c.ConstantinopleBlock, num)
//}
//
//// IsIstanbul returns whether num is either equal to the Istanbul fork block or greater.
//func (c *ChainConfig) IsIstanbul(num *big.Int) bool {
//	return isForked(c.IstanbulBlock, num)
//}
//
//// IsBerlin returns whether num is either equal to the Berlin fork block or greater.
//func (c *ChainConfig) IsBerlin(num *big.Int) bool {
//	return isForked(c.BerlinBlock, num)
//}
//
//// IsLondon returns whether num is either equal to the London fork block or greater.
//func (c *ChainConfig) IsLondon(num *big.Int) bool {
//	return isForked(c.LondonBlock, num)
//}
//
//// IsArrowGlacier returns whether num is either equal to the Arrow Glacier (EIP-4345) fork block or greater.
//func (c *ChainConfig) IsArrowGlacier(num *big.Int) bool {
//	return isForked(c.ArrowGlacierBlock, num)
//}
//
//// IsGrayGlacier returns whether num is either equal to the Gray Glacier (EIP-5133) fork block or greater.
//func (c *ChainConfig) IsGrayGlacier(num *big.Int) bool {
//	return isForked(c.GrayGlacierBlock, num)
//}
//
//func (c *ChainConfig) IsRome(num *big.Int) bool {
//	return isForked(c.RomeBlock, num)
//}
//
//// IsTerminalPoWBlock returns whether the given block is the last block of PoW stage.
//func (c *ChainConfig) IsTerminalPoWBlock(parentTotalDiff *big.Int, totalDiff *big.Int) bool {
//	if c.TerminalTotalDifficulty == nil {
//		return false
//	}
//	return parentTotalDiff.Cmp(c.TerminalTotalDifficulty) < 0 && totalDiff.Cmp(c.TerminalTotalDifficulty) >= 0
//}

// CheckCompatible checks whether scheduled fork transitions have been imported
// with a mismatching chain configuration.
func (c *ChainConfig) CheckCompatible(newcfg *ChainConfig, height uint64) *ConfigCompatError {
	bhead := new(big.Int).SetUint64(height)

	// Iterate checkCompatible to find the lowest conflict.
	var lasterr *ConfigCompatError
	for {
		err := c.checkCompatible(newcfg, bhead)
		if err == nil || (lasterr != nil && err.RewindTo == lasterr.RewindTo) {
			break
		}
		lasterr = err
		bhead.SetUint64(err.RewindTo)
	}
	return lasterr
}

// CheckConfigForkOrder checks that we don't "skip" any forks, geth isn't pluggable enough
// to guarantee that forks can be implemented in a different order than on official networks
//func (c *ChainConfig) CheckConfigForkOrder() error {
//	type fork struct {
//		name     string
//		block    *big.Int
//		optional bool // if true, the fork may be nil and next fork is still allowed
//	}
//	var lastFork fork
//	for _, cur := range []fork{
//		{name: "homesteadBlock", block: c.HomesteadBlock},
//		{name: "daoForkBlock", block: c.DAOForkBlock, optional: true},
//		{name: "eip150Block", block: c.EIP150Block},
//		{name: "eip155Block", block: c.EIP155Block},
//		{name: "eip158Block", block: c.EIP158Block},
//		{name: "byzantiumBlock", block: c.ByzantiumBlock},
//		{name: "constantinopleBlock", block: c.ConstantinopleBlock},
//		{name: "petersburgBlock", block: c.PetersburgBlock},
//		{name: "istanbulBlock", block: c.IstanbulBlock},
//		{name: "muirGlacierBlock", block: c.MuirGlacierBlock, optional: true},
//		{name: "berlinBlock", block: c.BerlinBlock},
//		{name: "londonBlock", block: c.LondonBlock},
//		{name: "arrowGlacierBlock", block: c.ArrowGlacierBlock, optional: true},
//		{name: "grayGlacierBlock", block: c.GrayGlacierBlock, optional: true},
//		{name: "RomeBlock", block: c.RomeBlock, optional: true},
//	} {
//		if lastFork.name != "" {
//			// Next one must be higher number
//			if lastFork.block == nil && cur.block != nil {
//				return fmt.Errorf("unsupported fork ordering: %v not enabled, but %v enabled at %v",
//					lastFork.name, cur.name, cur.block)
//			}
//			if lastFork.block != nil && cur.block != nil {
//				if lastFork.block.Cmp(cur.block) > 0 {
//					return fmt.Errorf("unsupported fork ordering: %v enabled at %v, but %v enabled at %v",
//						lastFork.name, lastFork.block, cur.name, cur.block)
//				}
//			}
//		}
//		// If it was optional and not set, then ignore it
//		if !cur.optional || cur.block != nil {
//			lastFork = cur
//		}
//	}
//	return nil
//}

func (c *ChainConfig) checkCompatible(newcfg *ChainConfig, head *big.Int) *ConfigCompatError {
	//if isForkIncompatible(c.HomesteadBlock, newcfg.HomesteadBlock, head) {
	//	return newCompatError("Homestead fork block", c.HomesteadBlock, newcfg.HomesteadBlock)
	//}
	//if isForkIncompatible(c.DAOForkBlock, newcfg.DAOForkBlock, head) {
	//	return newCompatError("DAO fork block", c.DAOForkBlock, newcfg.DAOForkBlock)
	//}
	//if c.IsDAOFork(head) && c.DAOForkSupport != newcfg.DAOForkSupport {
	//	return newCompatError("DAO fork support flag", c.DAOForkBlock, newcfg.DAOForkBlock)
	//}
	//if isForkIncompatible(c.EIP150Block, newcfg.EIP150Block, head) {
	//	return newCompatError("EIP150 fork block", c.EIP150Block, newcfg.EIP150Block)
	//}
	//if isForkIncompatible(c.EIP155Block, newcfg.EIP155Block, head) {
	//	return newCompatError("EIP155 fork block", c.EIP155Block, newcfg.EIP155Block)
	//}
	//if isForkIncompatible(c.EIP158Block, newcfg.EIP158Block, head) {
	//	return newCompatError("EIP158 fork block", c.EIP158Block, newcfg.EIP158Block)
	//}
	//if c.IsEIP158(head) && !configNumEqual(c.ChainID, newcfg.ChainID) {
	//	return newCompatError("EIP158 chain ID", c.EIP158Block, newcfg.EIP158Block)
	//}
	//if isForkIncompatible(c.ByzantiumBlock, newcfg.ByzantiumBlock, head) {
	//	return newCompatError("Byzantium fork block", c.ByzantiumBlock, newcfg.ByzantiumBlock)
	//}
	//if isForkIncompatible(c.ConstantinopleBlock, newcfg.ConstantinopleBlock, head) {
	//	return newCompatError("Constantinople fork block", c.ConstantinopleBlock, newcfg.ConstantinopleBlock)
	//}
	//if isForkIncompatible(c.PetersburgBlock, newcfg.PetersburgBlock, head) {
	//	// the only case where we allow Petersburg to be set in the past is if it is equal to Constantinople
	//	// mainly to satisfy fork ordering requirements which state that Petersburg fork be set if Constantinople fork is set
	//	if isForkIncompatible(c.ConstantinopleBlock, newcfg.PetersburgBlock, head) {
	//		return newCompatError("Petersburg fork block", c.PetersburgBlock, newcfg.PetersburgBlock)
	//	}
	//}
	//if isForkIncompatible(c.IstanbulBlock, newcfg.IstanbulBlock, head) {
	//	return newCompatError("Istanbul fork block", c.IstanbulBlock, newcfg.IstanbulBlock)
	//}
	//if isForkIncompatible(c.MuirGlacierBlock, newcfg.MuirGlacierBlock, head) {
	//	return newCompatError("Muir Glacier fork block", c.MuirGlacierBlock, newcfg.MuirGlacierBlock)
	//}
	//if isForkIncompatible(c.BerlinBlock, newcfg.BerlinBlock, head) {
	//	return newCompatError("Berlin fork block", c.BerlinBlock, newcfg.BerlinBlock)
	//}
	//if isForkIncompatible(c.LondonBlock, newcfg.LondonBlock, head) {
	//	return newCompatError("London fork block", c.LondonBlock, newcfg.LondonBlock)
	//}
	//if isForkIncompatible(c.ArrowGlacierBlock, newcfg.ArrowGlacierBlock, head) {
	//	return newCompatError("Arrow Glacier fork block", c.ArrowGlacierBlock, newcfg.ArrowGlacierBlock)
	//}
	//if isForkIncompatible(c.GrayGlacierBlock, newcfg.GrayGlacierBlock, head) {
	//	return newCompatError("Gray Glacier fork block", c.GrayGlacierBlock, newcfg.GrayGlacierBlock)
	//}
	return nil
}

// isForkIncompatible returns true if a fork scheduled at s1 cannot be rescheduled to
// block s2 because head is already past the fork.
func isForkIncompatible(s1, s2, head *big.Int) bool {
	return (isForked(s1, head) || isForked(s2, head)) && !configNumEqual(s1, s2)
}

// isForked returns whether a fork scheduled at block s is active at the given head block.
func isForked(s, head *big.Int) bool {
	if s == nil || head == nil {
		return false
	}
	return s.Cmp(head) <= 0
}

func configNumEqual(x, y *big.Int) bool {
	if x == nil {
		return y == nil
	}
	if y == nil {
		return x == nil
	}
	return x.Cmp(y) == 0
}

// ConfigCompatError is raised if the locally-stored blockchain is initialised with a
// ChainConfig that would alter the past.
type ConfigCompatError struct {
	What string
	// block numbers of the stored and new configurations
	StoredConfig, NewConfig *big.Int
	// the block number to which the local chain must be rewound to correct the error
	RewindTo uint64
}

func newCompatError(what string, storedblock, newblock *big.Int) *ConfigCompatError {
	var rew *big.Int
	switch {
	case storedblock == nil:
		rew = newblock
	case newblock == nil || storedblock.Cmp(newblock) < 0:
		rew = storedblock
	default:
		rew = newblock
	}
	err := &ConfigCompatError{what, storedblock, newblock, 0}
	if rew != nil && rew.Sign() > 0 {
		err.RewindTo = rew.Uint64() - 1
	}
	return err
}

func (err *ConfigCompatError) Error() string {
	return fmt.Sprintf("mismatching %s in database (have %d, want %d, rewindto %d)", err.What, err.StoredConfig, err.NewConfig, err.RewindTo)
}

// Rules wraps ChainConfig and is merely syntactic sugar or can be used for functions
// that do not have or require information about the block.
//
// Rules is a one time interface meaning that it shouldn't be used in between transition
// phases.
type Rules struct {
	ChainID                                                 *big.Int
	IsHomestead, IsEIP150, IsEIP155, IsEIP158               bool
	IsByzantium, IsConstantinople, IsPetersburg, IsIstanbul bool
	IsBerlin, IsLondon                                      bool
	IsMerge, IsShanghai, isCancun                           bool
}

// Rules ensures c's ChainID is not nil.
func (c *ChainConfig) Rules(num *big.Int, isMerge bool) Rules {
	chainID := c.ChainID
	if chainID == nil {
		chainID = new(big.Int)
	}
	return Rules{
		ChainID: new(big.Int).Set(chainID),
		//IsHomestead:      c.IsHomestead(num),
		//IsEIP150:         c.IsEIP150(num),
		//IsEIP155:         c.IsEIP155(num),
		//IsEIP158:         c.IsEIP158(num),
		//IsByzantium:      c.IsByzantium(num),
		//IsConstantinople: c.IsConstantinople(num),
		//IsPetersburg:     c.IsPetersburg(num),
		//IsIstanbul:       c.IsIstanbul(num),
		//IsBerlin:         c.IsBerlin(num),
		//IsLondon:         c.IsLondon(num),
		IsMerge: isMerge,
		//IsShanghai:       c.IsShanghai(num),
		//isCancun:         c.IsCancun(num),
	}
}

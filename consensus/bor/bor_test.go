package bor

import (
	"math/big"
	"testing"

	"github.com/StevenBarnett1/bor/common"
	"github.com/StevenBarnett1/bor/common/hexutil"
	"github.com/StevenBarnett1/bor/core"
	"github.com/StevenBarnett1/bor/core/rawdb"
	"github.com/StevenBarnett1/bor/core/state"
	"github.com/StevenBarnett1/bor/core/types"
	"github.com/StevenBarnett1/bor/core/vm"
	"github.com/StevenBarnett1/bor/params"
	"github.com/stretchr/testify/assert"
)

func TestGenesisContractChange(t *testing.T) {
	addr0 := common.Address{0x1}

	b := &Bor{
		config: &params.BorConfig{
			Sprint: 10, // skip sprint transactions in sprint
			BlockAlloc: map[string]interface{}{
				// write as interface since that is how it is decoded in genesis
				"2": map[string]interface{}{
					addr0.Hex(): map[string]interface{}{
						"code":    hexutil.Bytes{0x1, 0x2},
						"balance": "0",
					},
				},
				"4": map[string]interface{}{
					addr0.Hex(): map[string]interface{}{
						"code":    hexutil.Bytes{0x1, 0x3},
						"balance": "0x1000",
					},
				},
			},
		},
	}

	genspec := &core.Genesis{
		Alloc: map[common.Address]core.GenesisAccount{
			addr0: {
				Balance: big.NewInt(0),
				Code:    []byte{0x1, 0x1},
			},
		},
	}

	db := rawdb.NewMemoryDatabase()
	genesis := genspec.MustCommit(db)

	statedb, err := state.New(genesis.Root(), state.NewDatabase(db), nil)
	assert.NoError(t, err)

	config := params.ChainConfig{}
	chain, err := core.NewBlockChain(db, nil, &config, b, vm.Config{}, nil, nil)
	assert.NoError(t, err)

	addBlock := func(root common.Hash, num int64) (common.Hash, *state.StateDB) {
		h := &types.Header{
			ParentHash: root,
			Number:     big.NewInt(num),
		}
		b.Finalize(chain, h, statedb, nil, nil)

		// write state to database
		root, err := statedb.Commit(false)
		assert.NoError(t, err)
		assert.NoError(t, statedb.Database().TrieDB().Commit(root, true, nil))

		statedb, err := state.New(h.Root, state.NewDatabase(db), nil)
		assert.NoError(t, err)

		return root, statedb
	}

	assert.Equal(t, statedb.GetCode(addr0), []byte{0x1, 0x1})

	root := genesis.Root()

	// code does not change
	root, statedb = addBlock(root, 1)
	assert.Equal(t, statedb.GetCode(addr0), []byte{0x1, 0x1})

	// code changes 1st time
	root, statedb = addBlock(root, 2)
	assert.Equal(t, statedb.GetCode(addr0), []byte{0x1, 0x2})

	// code same as 1st change
	root, statedb = addBlock(root, 3)
	assert.Equal(t, statedb.GetCode(addr0), []byte{0x1, 0x2})

	// code changes 2nd time
	_, statedb = addBlock(root, 4)
	assert.Equal(t, statedb.GetCode(addr0), []byte{0x1, 0x3})

	// make sure balance change DOES NOT take effect
	assert.Equal(t, statedb.GetBalance(addr0), big.NewInt(0))
}

func TestEncodeSigHeaderJaipur(t *testing.T) {
	// As part of the EIP-1559 fork in mumbai, an incorrect seal hash
	// was used for Bor that did not included the BaseFee. The Jaipur
	// block is a hard fork to fix that.
	h := &types.Header{
		Difficulty: new(big.Int),
		Number:     big.NewInt(1),
		Extra:      make([]byte, 32+65),
	}

	var (
		// hash for the block without the BaseFee
		hashWithoutBaseFee = common.HexToHash("0x1be13e83939b3c4701ee57a34e10c9290ce07b0e53af0fe90b812c6881826e36")
		// hash for the block with the baseFee
		hashWithBaseFee = common.HexToHash("0xc55b0cac99161f71bde1423a091426b1b5b4d7598e5981ad802cce712771965b")
	)

	// Jaipur NOT enabled and BaseFee not set
	hash := SealHash(h, &params.BorConfig{JaipurBlock: 10})
	assert.Equal(t, hash, hashWithoutBaseFee)

	// Jaipur enabled (Jaipur=0) and BaseFee not set
	hash = SealHash(h, &params.BorConfig{JaipurBlock: 0})
	assert.Equal(t, hash, hashWithoutBaseFee)

	h.BaseFee = big.NewInt(2)

	// Jaipur enabled (Jaipur=Header block) and BaseFee set
	hash = SealHash(h, &params.BorConfig{JaipurBlock: 1})
	assert.Equal(t, hash, hashWithBaseFee)

	// Jaipur NOT enabled and BaseFee set
	hash = SealHash(h, &params.BorConfig{JaipurBlock: 10})
	assert.Equal(t, hash, hashWithoutBaseFee)
}

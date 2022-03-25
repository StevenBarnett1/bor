package filters

import (
	"context"

	"github.com/StevenBarnett1/bor/common"
	"github.com/StevenBarnett1/bor/core/rawdb"
	"github.com/StevenBarnett1/bor/core/types"
)

func (b *testBackend) GetBorBlockReceipt(ctx context.Context, hash common.Hash) (*types.Receipt, error) {
	number := rawdb.ReadHeaderNumber(b.db, hash)
	if number == nil {
		return nil, nil
	}

	receipt := rawdb.ReadBorReceipt(b.db, hash, *number)
	if receipt == nil {
		return nil, nil
	}
	return receipt, nil
}

func (b *testBackend) GetBorBlockLogs(ctx context.Context, hash common.Hash) ([]*types.Log, error) {
	receipt, err := b.GetBorBlockReceipt(ctx, hash)
	if receipt == nil || err != nil {
		return nil, nil
	}
	return receipt.Logs, nil
}

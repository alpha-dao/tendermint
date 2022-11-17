package querier

import (
	"context"
	"fmt"
	sm "github.com/tendermint/tendermint/state"
	"github.com/tendermint/tendermint/store"
	timestamppb "google.golang.org/protobuf/types/known/timestamppb"
	"log"
	"os"
	"time"
)

var (
	errBlockNotFound = fmt.Errorf("block_not_found")
)

type CosmosQuerier struct {
	blockStore *store.BlockStore
	stateStore sm.Store
}

func (q *CosmosQuerier) GetBlock(ctx context.Context, request *GetBlockRequest) (*GetBlockResponse, error) {
	b, err := q.getBlockFromLocal(request.Height)
	if err != nil {
		return nil, err
	}

	return &GetBlockResponse{
		Block: b,
	}, nil
}

func (q *CosmosQuerier) GetBlockStreamFrom(request *GetBlockRequest, stream CosmosIndexer_GetBlockStreamFromServer) error {
	//spawn multiplexing end
	end1, end2 := make(chan struct{}), make(chan struct{})
	go func(input <-chan struct{}, o1 chan<- struct{}, o2 chan<- struct{}) {
		select {
		case <-input:
			o1 <- struct{}{}
			o2 <- struct{}{}
			close(o1)
			close(o2)
		}
	}(stream.Context().Done(), end1, end2)

	//spawn worker
	go q.spawnBlockStreamPushWorker(make(chan int64), stream, end1)

	<-end2
	fmt.Println("Querier grpc stream closed!")
	return nil
}

func (q *CosmosQuerier) spawnBlockStreamPushWorker(pushChannel chan int64, stream CosmosIndexer_GetBlockStreamFromServer, end <-chan struct{}) {
	select {
	case heightToStream := <-pushChannel:

		block, err := q.getBlockFromLocal(heightToStream)
		if err == errBlockNotFound {
			time.Sleep(6 * time.Second)
			pushChannel <- heightToStream
		} else if err != nil {
			log.Fatal(err)
			//TODO: Might need to handle later. not fatal crash
		}

		err = stream.Send(&GetBlockResponse{
			Block: block,
		})

		if err != nil {
			log.New(os.Stderr, "", 0).Println(err)
			log.New(os.Stderr, "", 0).Printf("failed query block height %d\n", heightToStream)
		}
		pushChannel <- heightToStream + 1
	case <-end:
		return
	}
}

func (q *CosmosQuerier) getBlockFromLocal(height int64) (*Block, error) {
	tmBlock := q.blockStore.LoadBlock(height)
	if tmBlock == nil {
		return nil, errBlockNotFound
	}
	abciResponse, err := q.stateStore.LoadABCIResponses(height)
	if err != nil {
		return nil, err
	}

	var txs []*Tx

	for i, tx := range tmBlock.Data.Txs {
		txr := abciResponse.DeliverTxs[i]
		txHash := fmt.Sprintf("%X", tx.Hash())
		pbTx := Tx{
			TxHash:    txHash,
			Code:      txr.Code,
			Log:       txr.Log,
			Info:      txr.Info,
			GasWanted: txr.GasWanted,
			GasUsed:   txr.GasUsed,
			Codespace: txr.Codespace,
			TxBytes:   tx,
		}
		txs = append(txs, &pbTx)
	}

	return &Block{Height: height, Txs: txs, Time: timestamppb.New(tmBlock.Time)}, nil
}

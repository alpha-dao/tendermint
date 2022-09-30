// Package proxy implements an proxy event sink
package proxy

import (
	"context"
	"errors"

	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/libs/pubsub/query"
	"github.com/tendermint/tendermint/types"
)

// EventSink is an indexer backend providing the tx/block index services.  This
// implementation stores records in a PostgreSQL database using the schema
// defined in state/indexer/sink/psql/schema.sql.
var proxiedEventSink ProxiedEventSink

type ProxiedEventSink interface {
	IndexBlockEvents(h types.EventDataNewBlockHeader) error
	IndexTxEvents(txrs []*abci.TxResult) error
}

type EventSink struct {
	chainID    string
	pEventSink ProxiedEventSink
}

// InjectProxy injects Proxy event sink
func InjectProxy(p ProxiedEventSink) {
	proxiedEventSink = p
}

// NewEventSink constructs an event sink associated with the PostgreSQL
// database specified by connStr. Events written to the sink are attributed to
// the specified chainID.
func NewEventSink(chainID string) (*EventSink, error) {
	return &EventSink{
		chainID,
		proxiedEventSink,
	}, nil
}

// IndexBlockEvents indexes the specified block header, part of the
// indexer.EventSink interface.
func (es *EventSink) IndexBlockEvents(h types.EventDataNewBlockHeader) error {
	return es.pEventSink.IndexBlockEvents(h)
}

func (es *EventSink) IndexTxEvents(txrs []*abci.TxResult) error {
	return es.pEventSink.IndexTxEvents(txrs)
}

// SearchBlockEvents is not implemented by this sink, and reports an error for all queries.
func (es *EventSink) SearchBlockEvents(ctx context.Context, q *query.Query) ([]int64, error) {
	return nil, errors.New("block search is not supported via the postgres event sink")
}

// SearchTxEvents is not implemented by this sink, and reports an error for all queries.
func (es *EventSink) SearchTxEvents(ctx context.Context, q *query.Query) ([]*abci.TxResult, error) {
	return nil, errors.New("tx search is not supported via the postgres event sink")
}

// GetTxByHash is not implemented by this sink, and reports an error for all queries.
func (es *EventSink) GetTxByHash(hash []byte) (*abci.TxResult, error) {
	return nil, errors.New("getTxByHash is not supported via the postgres event sink")
}

// HasBlock is not implemented by this sink, and reports an error for all queries.
func (es *EventSink) HasBlock(h int64) (bool, error) {
	return false, errors.New("hasBlock is not supported via the postgres event sink")
}

// Stop closes the underlying PostgreSQL database.
func (es *EventSink) Stop() error { return nil }

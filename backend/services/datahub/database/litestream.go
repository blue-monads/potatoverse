package database

import (
	"context"
	"fmt"
	"time"

	"github.com/ncruces/litestream"
	"github.com/ncruces/litestream/webdav"
)

func (db *DB) StartLitestream() error {
	return StartLitestream("todo")
}

func StartLitestream(filepath string) error {

	ldb := litestream.NewDB(filepath)
	wdav := webdav.NewReplicaClient()
	replica := litestream.NewReplicaWithClient(ldb, wdav)

	ldb.Replica = replica

	levels := litestream.CompactionLevels{
		{Level: 0},
		{Level: 1, Interval: 10 * time.Second},
		{Level: litestream.SnapshotLevel, Interval: 24 * time.Hour},
	}

	store := litestream.NewStore([]*litestream.DB{ldb}, levels)

	if err := store.Open(context.Background()); err != nil {
		return fmt.Errorf("open store: %w", err)
	}

	<-make(chan struct{})

	return nil
}

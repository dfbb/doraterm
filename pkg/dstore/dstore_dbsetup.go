// Copyright 2025, Command Line Inc.
// SPDX-License-Identifier: Apache-2.0

package dstore

import (
	"context"
	"fmt"
	"log"
	"path/filepath"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/sawka/txwrap"
	"github.com/dfbb/doraterm/pkg/util/migrateutil"
	"github.com/dfbb/doraterm/pkg/dorabase"
	"github.com/dfbb/doraterm/pkg/doraobj"

	dbfs "github.com/dfbb/doraterm/db"
)

const WStoreDBName = "waveterm.db"

type TxWrap = txwrap.TxWrap

var globalDB *sqlx.DB

func InitWStore() error {
	ctx, cancelFn := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancelFn()
	var err error
	globalDB, err = MakeDB(ctx)
	if err != nil {
		return err
	}
	err = migrateutil.Migrate("wstore", globalDB.DB, dbfs.WStoreMigrationFS, "migrations-wstore")
	if err != nil {
		return err
	}
	log.Printf("wstore initialized\n")
	return nil
}

func GetDBName() string {
	waveHome := dorabase.GetDoraDataDir()
	return filepath.Join(waveHome, dorabase.WaveDBDir, WStoreDBName)
}

func MakeDB(ctx context.Context) (*sqlx.DB, error) {
	dbName := GetDBName()
	rtn, err := sqlx.Open("sqlite3", fmt.Sprintf("file:%s?mode=rwc&_journal_mode=WAL&_busy_timeout=5000", dbName))
	if err != nil {
		return nil, err
	}
	rtn.DB.SetMaxOpenConns(1)
	return rtn, nil
}

func WithTx(ctx context.Context, fn func(tx *TxWrap) error) (rtnErr error) {
	doraobj.ContextUpdatesBeginTx(ctx)
	defer func() {
		if rtnErr != nil {
			doraobj.ContextUpdatesRollbackTx(ctx)
		} else {
			doraobj.ContextUpdatesCommitTx(ctx)
		}
	}()
	return txwrap.WithTx(ctx, globalDB, fn)
}

func WithTxRtn[RT any](ctx context.Context, fn func(tx *TxWrap) (RT, error)) (rtnVal RT, rtnErr error) {
	doraobj.ContextUpdatesBeginTx(ctx)
	defer func() {
		if rtnErr != nil {
			doraobj.ContextUpdatesRollbackTx(ctx)
		} else {
			doraobj.ContextUpdatesCommitTx(ctx)
		}
	}()
	return txwrap.WithTxRtn(ctx, globalDB, fn)
}

// Copyright 2025, Command Line Inc.
// SPDX-License-Identifier: Apache-2.0

package filestore

// setup for filestore db
// includes migration support and txwrap setup

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/dfbb/doraterm/pkg/util/migrateutil"
	"github.com/dfbb/doraterm/pkg/dorabase"

	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
	"github.com/sawka/txwrap"

	dbfs "github.com/dfbb/doraterm/db"
)

const FilestoreDBName = "filestore.db"

type TxWrap = txwrap.TxWrap

var globalDB *sqlx.DB
var useTestingDb bool // just for testing (forces GetDB() to return an in-memory db)

func InitFilestore() error {
	ctx, cancelFn := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancelFn()
	var err error
	globalDB, err = MakeDB(ctx)
	if err != nil {
		return err
	}
	err = migrateutil.Migrate("filestore", globalDB.DB, dbfs.FilestoreMigrationFS, "migrations-filestore")
	if err != nil {
		return err
	}
	// Self-heal: if the old db_wave_file table still exists (migration ran before
	// the wave→dora rename was applied), nuke the filestore DB and start fresh.
	var oldTableExists bool
	_ = globalDB.Get(&oldTableExists, "SELECT COUNT(*) > 0 FROM sqlite_master WHERE type='table' AND name='db_wave_file'")
	if oldTableExists {
		globalDB.Close()
		dbPath := GetDBName()
		if err := os.Remove(dbPath); err != nil {
			return fmt.Errorf("removing stale filestore db (db_wave_file exists): %w", err)
		}
		log.Printf("filestore: removed stale db with db_wave_file table, recreating\n")
		globalDB, err = MakeDB(ctx)
		if err != nil {
			return err
		}
		err = migrateutil.Migrate("filestore", globalDB.DB, dbfs.FilestoreMigrationFS, "migrations-filestore")
		if err != nil {
			return err
		}
	}
	if !stopFlush.Load() {
		go WFS.runFlusher()
	}
	log.Printf("filestore initialized\n")
	return nil
}

func GetDBName() string {
	doraHome := dorabase.GetDoraDataDir()
	return filepath.Join(doraHome, dorabase.DoraDBDir, FilestoreDBName)
}

func MakeDB(ctx context.Context) (*sqlx.DB, error) {
	var rtn *sqlx.DB
	var err error
	if useTestingDb {
		dbName := ":memory:"
		log.Printf("[db] using in-memory db\n")
		rtn, err = sqlx.Open("sqlite3", dbName)
	} else {
		dbName := GetDBName()
		log.Printf("[db] opening db %s\n", dbName)
		rtn, err = sqlx.Open("sqlite3", fmt.Sprintf("file:%s?mode=rwc&_journal_mode=WAL&_busy_timeout=5000", dbName))
	}
	if err != nil {
		return nil, fmt.Errorf("opening db: %w", err)
	}
	rtn.DB.SetMaxOpenConns(1)
	return rtn, nil
}

func WithTx(ctx context.Context, fn func(tx *TxWrap) error) error {
	return txwrap.WithTx(ctx, globalDB, fn)
}

func WithTxRtn[RT any](ctx context.Context, fn func(tx *TxWrap) (RT, error)) (RT, error) {
	return txwrap.WithTxRtn(ctx, globalDB, fn)
}

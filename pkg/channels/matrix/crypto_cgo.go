//go:build cgo && matrix_crypto

package matrix

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	"go.mau.fi/util/dbutil"
	"maunium.net/go/mautrix/crypto/cryptohelper"

	_ "modernc.org/sqlite"
)

const (
	sqliteDriver = "sqlite"
	dbName       = "store.db"
)

func (c *MatrixChannel) maybeInitCrypto(ctx context.Context) error {
	logger.InfoC("matrix", "Initializing crypto helper")

	// Ensure the crypto database directory exists
	if err := os.MkdirAll(c.cryptoDbPath, 0o700); err != nil {
		return fmt.Errorf("create crypto database directory: %w", err)
	}

	// Create database with sqlite driver (modernc.org/sqlite)
	dbPath := filepath.Join(c.cryptoDbPath, dbName)
	connStr := "file:" + dbPath + "?_foreign_keys=on"

	db, err := sql.Open(sqliteDriver, connStr)
	if err != nil {
		return fmt.Errorf("open crypto database: %w", err)
	}
	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)

	// Execute PRAGMA statements
	// This is equivalent to the "sqlite3-fk-wal" dialect used by cryptohelper
	pragmaStmts := []string{
		"PRAGMA foreign_keys = ON",
		"PRAGMA journal_mode = WAL",
		"PRAGMA synchronous = NORMAL",
		"PRAGMA busy_timeout = 5000",
	}
	for _, pragma := range pragmaStmts {
		if _, err = db.ExecContext(ctx, pragma); err != nil {
			_ = db.Close()
			return fmt.Errorf("execute %s: %w", pragma, err)
		}
	}

	// Wrap with dbutil for dialect support
	wrappedDB, err := dbutil.NewWithDB(db, sqliteDriver)
	if err != nil {
		_ = db.Close()
		return fmt.Errorf("wrap database: %w", err)
	}

	cryptoHelper, err := cryptohelper.NewCryptoHelper(c.client, []byte(c.config.CryptoPassphrase), wrappedDB)
	if err != nil {
		return fmt.Errorf("create crypto helper: %w", err)
	}

	if c.client.DeviceID == "" {
		resp, whoamiErr := c.client.Whoami(ctx)
		if whoamiErr != nil {
			_ = db.Close()
			return fmt.Errorf("get device ID via whoami: %w", whoamiErr)
		}
		c.client.DeviceID = resp.DeviceID
	}

	if err = cryptoHelper.Init(ctx); err != nil {
		cryptoHelper.Close()
		return fmt.Errorf("init crypto helper: %w", err)
	}

	c.client.Crypto = cryptoHelper
	c.cryptoCloser = cryptoHelper

	logger.InfoC("matrix", "Crypto helper initialized successfully")
	return nil
}

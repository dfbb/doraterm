// Copyright 2026, Command Line Inc.
// SPDX-License-Identifier: Apache-2.0

//go:build !windows && !(linux && (mips || mips64))

package doraattach

import (
	"crypto/ed25519"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	_ "modernc.org/sqlite"
	"github.com/dfbb/doraterm/pkg/dorajwt"
	"github.com/dfbb/doraterm/pkg/dshrpc"
	"github.com/dfbb/doraterm/pkg/dshrpc/dshclient"
	"github.com/dfbb/doraterm/pkg/dshutil"
)

const (
	dbSubdir       = "db"
	dbFileName     = "doraterm.db"
	socketFileName = "dora.sock"
)

func ResolveDataDir() (string, error) {
	if v := os.Getenv("DORATERM_DATA_HOME"); v != "" {
		return v, nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("cannot resolve home dir: %w", err)
	}
	var candidates []string
	switch runtime.GOOS {
	case "darwin":
		candidates = []string{
			filepath.Join(home, "Library", "Application Support", "doraterm"),
			filepath.Join(home, "Library", "Application Support", "doraterm-dev"),
			filepath.Join(home, ".doraterm"),
			filepath.Join(home, ".doraterm-dev"),
		}
	case "linux":
		xdgData := os.Getenv("XDG_DATA_HOME")
		if xdgData == "" {
			xdgData = filepath.Join(home, ".local", "share")
		}
		candidates = []string{
			filepath.Join(xdgData, "doraterm"),
			filepath.Join(xdgData, "doraterm-dev"),
			filepath.Join(home, ".doraterm"),
			filepath.Join(home, ".doraterm-dev"),
		}
	default:
		candidates = []string{
			filepath.Join(home, ".doraterm"),
			filepath.Join(home, ".doraterm-dev"),
		}
	}
	for _, candidate := range candidates {
		if st, err := os.Stat(candidate); err == nil && st.IsDir() {
			return candidate, nil
		}
	}
	return "", fmt.Errorf("Dora data directory not found. Is Dora running? (set $DORATERM_DATA_HOME to override)")
}

func loadJwtPrivateKey(dataDir string) (ed25519.PrivateKey, error) {
	dbPath := filepath.Join(dataDir, dbSubdir, dbFileName)
	if _, err := os.Stat(dbPath); err != nil {
		return nil, fmt.Errorf("Dora database not found at %s: %w", dbPath, err)
	}
	dsn := fmt.Sprintf("file:%s?mode=ro&_journal_mode=WAL&_busy_timeout=5000", dbPath)
	db, err := sqlx.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("opening dora db: %w", err)
	}
	defer db.Close()

	var rawJSON string
	if err := db.Get(&rawJSON, "SELECT data FROM db_mainserver LIMIT 1"); err != nil {
		return nil, fmt.Errorf("querying db_mainserver (Dora schema may have changed): %w", err)
	}
	var ms struct {
		JwtPrivateKey string `json:"jwtprivatekey"`
	}
	if err := json.Unmarshal([]byte(rawJSON), &ms); err != nil {
		return nil, fmt.Errorf("parsing mainserver JSON: %w", err)
	}
	if ms.JwtPrivateKey == "" {
		return nil, fmt.Errorf("jwtprivatekey is empty in db_mainserver")
	}
	keyBytes, err := base64.StdEncoding.DecodeString(ms.JwtPrivateKey)
	if err != nil {
		return nil, fmt.Errorf("base64 decoding jwt private key: %w", err)
	}
	if len(keyBytes) != ed25519.PrivateKeySize {
		return nil, fmt.Errorf("jwt private key has wrong length: got %d, want %d", len(keyBytes), ed25519.PrivateKeySize)
	}
	return ed25519.PrivateKey(keyBytes), nil
}

func Connect() (*dshutil.DshRpc, string, error) {
	dataDir, err := ResolveDataDir()
	if err != nil {
		return nil, "", err
	}
	sockPath := filepath.Join(dataDir, socketFileName)
	if _, err := os.Stat(sockPath); err != nil {
		return nil, "", fmt.Errorf("Dora socket not found at %s: %w", sockPath, err)
	}

	privKey, err := loadJwtPrivateKey(dataDir)
	if err != nil {
		return nil, "", err
	}
	if err := dorajwt.SetPrivateKey([]byte(privKey)); err != nil {
		return nil, "", fmt.Errorf("setting jwt private key: %w", err)
	}

	routeId := "doraattach-" + uuid.NewString()
	rpcCtx := dshrpc.RpcContext{
		SockName: sockPath,
		RouteId:  routeId,
	}
	jwtToken, err := dshutil.MakeClientJWTToken(rpcCtx)
	if err != nil {
		return nil, "", fmt.Errorf("creating jwt: %w", err)
	}
	rpcClient, err := dshutil.SetupDomainSocketRpcClient(sockPath, nil, "doraattach")
	if err != nil {
		return nil, "", fmt.Errorf("connecting to %s: %w", sockPath, err)
	}
	authRtn, err := dshclient.AuthenticateCommand(rpcClient, jwtToken, &dshrpc.RpcOpts{Route: dshutil.ControlRoute})
	if err != nil {
		return nil, "", fmt.Errorf("authenticating: %w", err)
	}
	return rpcClient, authRtn.RouteId, nil
}

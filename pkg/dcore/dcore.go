// Copyright 2025, Command Line Inc.
// SPDX-License-Identifier: Apache-2.0

// wave core application coordinator
package dcore

import (
	"context"
	"crypto/ed25519"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/dfbb/doraterm/pkg/dorajwt"
	"github.com/dfbb/doraterm/pkg/doraobj"
	"github.com/dfbb/doraterm/pkg/dps"
	"github.com/dfbb/doraterm/pkg/dstore"
)

// the wcore package coordinates actions across the storage layer
// orchestrating the wave object store, the wave pubsub system, and the wave rpc system

// Ensures that the initial data is present in the store, creates an initial window if needed
func EnsureInitialData() (bool, error) {
	// does not need to run in a transaction since it is called on startup
	ctx, cancelFn := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancelFn()
	client, err := dstore.DBGetSingleton[*doraobj.Client](ctx)
	firstLaunch := false
	if err == dstore.ErrNotFound {
		client, err = CreateClient(ctx)
		if err != nil {
			return false, fmt.Errorf("error creating client: %w", err)
		}
		firstLaunch = true
	}
	if client.TempOID == "" {
		log.Println("client.TempOID is empty")
		client.TempOID = uuid.NewString()
		err = dstore.DBUpdate(ctx, client)
		if err != nil {
			return firstLaunch, fmt.Errorf("error updating client: %w", err)
		}
	}
	if client.InstallId == "" {
		log.Println("client.InstallId is empty")
		client.InstallId = uuid.NewString()
		err = dstore.DBUpdate(ctx, client)
		if err != nil {
			return firstLaunch, fmt.Errorf("error updating client: %w", err)
		}
	}
	log.Printf("clientid: %s\n", client.OID)
	dstore.SetClientId(client.OID)
	if len(client.WindowIds) == 1 {
		log.Println("client has one window")
		CheckAndFixWindow(ctx, client.WindowIds[0])
		return firstLaunch, nil
	}
	if len(client.WindowIds) > 0 {
		log.Println("client has windows")
		return firstLaunch, nil
	}
	wsId := ""
	if firstLaunch {
		log.Println("client has no windows and first launch, creating starter workspace")
		starterWs, err := CreateWorkspace(ctx, "Starter workspace", "custom@wave-logo-solid", "#58C142", false, true)
		if err != nil {
			return firstLaunch, fmt.Errorf("error creating starter workspace: %w", err)
		}
		wsId = starterWs.OID
	}
	_, err = CreateWindow(ctx, nil, wsId)
	if err != nil {
		return firstLaunch, fmt.Errorf("error creating window: %w", err)
	}
	if firstLaunch {
		BootstrapStarterLayout(ctx)
	}
	return firstLaunch, nil
}

func CreateClient(ctx context.Context) (*doraobj.Client, error) {
	client := &doraobj.Client{
		OID:       uuid.NewString(),
		WindowIds: []string{},
	}
	err := dstore.DBInsert(ctx, client)
	if err != nil {
		return nil, fmt.Errorf("error inserting client: %w", err)
	}
	return client, nil
}

func GetClientData(ctx context.Context) (*doraobj.Client, error) {
	clientData, err := dstore.DBGetSingleton[*doraobj.Client](ctx)
	if err != nil {
		return nil, fmt.Errorf("error getting client data: %w", err)
	}
	return clientData, nil
}

func SendDoraObjUpdate(oref doraobj.ORef) {
	ctx, cancelFn := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancelFn()
	// send a waveobj:update event
	waveObj, err := dstore.DBGetORef(ctx, oref)
	if err != nil {
		log.Printf("error getting object for update event: %v", err)
		return
	}
	dps.Broker.Publish(dps.DoraEvent{
		Event:  dps.Event_DoraObjUpdate,
		Scopes: []string{oref.String()},
		Data: doraobj.DoraObjUpdate{
			UpdateType: doraobj.UpdateType_Update,
			OType:      waveObj.GetOType(),
			OID:        doraobj.GetOID(waveObj),
			Obj:        waveObj,
		},
	})
}

func ResolveBlockIdFromPrefix(ctx context.Context, tabId string, blockIdPrefix string) (string, error) {
	if len(blockIdPrefix) != 8 {
		return "", fmt.Errorf("widget_id must be 8 characters")
	}

	tab, err := dstore.DBMustGet[*doraobj.Tab](ctx, tabId)
	if err != nil {
		return "", fmt.Errorf("error getting tab: %w", err)
	}

	for _, blockId := range tab.BlockIds {
		if strings.HasPrefix(blockId, blockIdPrefix) {
			return blockId, nil
		}
	}

	return "", fmt.Errorf("widget_id not found: %q", blockIdPrefix)
}

func InitMainServer() error {
	ctx, cancelFn := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelFn()

	mainServer, err := dstore.DBGetSingleton[*doraobj.MainServer](ctx)
	if err == dstore.ErrNotFound {
		mainServer = &doraobj.MainServer{
			OID: uuid.NewString(),
		}
		err = dstore.DBInsert(ctx, mainServer)
		if err != nil {
			return fmt.Errorf("error inserting mainserver: %w", err)
		}
	} else if err != nil {
		return fmt.Errorf("error getting mainserver: %w", err)
	}

	needsUpdate := false
	if mainServer.JwtPrivateKey == "" || mainServer.JwtPublicKey == "" {
		keyPair, err := dorajwt.GenerateKeyPair()
		if err != nil {
			return fmt.Errorf("error generating jwt keypair: %w", err)
		}
		mainServer.JwtPrivateKey = base64.StdEncoding.EncodeToString(keyPair.PrivateKey)
		mainServer.JwtPublicKey = base64.StdEncoding.EncodeToString(keyPair.PublicKey)
		needsUpdate = true
	}

	if needsUpdate {
		err = dstore.DBUpdate(ctx, mainServer)
		if err != nil {
			return fmt.Errorf("error updating mainserver: %w", err)
		}
	}

	privateKeyBytes, err := base64.StdEncoding.DecodeString(mainServer.JwtPrivateKey)
	if err != nil {
		return fmt.Errorf("error decoding jwt private key: %w", err)
	}
	publicKeyBytes, err := base64.StdEncoding.DecodeString(mainServer.JwtPublicKey)
	if err != nil {
		return fmt.Errorf("error decoding jwt public key: %w", err)
	}

	err = dorajwt.SetPrivateKey(privateKeyBytes)
	if err != nil {
		return fmt.Errorf("error setting jwt private key: %w", err)
	}
	err = dorajwt.SetPublicKey(publicKeyBytes)
	if err != nil {
		return fmt.Errorf("error setting jwt public key: %w", err)
	}

	pubKeyDer, err := x509.MarshalPKIXPublicKey(ed25519.PublicKey(publicKeyBytes))
	if err != nil {
		log.Printf("warning: could not marshal public key for logging: %v", err)
	} else {
		pubKeyPem := pem.EncodeToMemory(&pem.Block{
			Type:  "PUBLIC KEY",
			Bytes: pubKeyDer,
		})
		log.Printf("JWT Public Key:\n%s", string(pubKeyPem))
	}

	return nil
}

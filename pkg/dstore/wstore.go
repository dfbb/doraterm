// Copyright 2025, Command Line Inc.
// SPDX-License-Identifier: Apache-2.0

package dstore

import (
	"context"
	"fmt"
	"sync"

	"github.com/dfbb/doraterm/pkg/dorabase"
	"github.com/dfbb/doraterm/pkg/doraobj"
)

func init() {
	for _, rtype := range doraobj.AllWaveObjTypes() {
		doraobj.RegisterType(rtype)
	}
}

var (
	clientIdLock   sync.Mutex
	cachedClientId string
)

func SetClientId(clientId string) {
	clientIdLock.Lock()
	defer clientIdLock.Unlock()
	cachedClientId = clientId
}

// in the main server, this will not return empty string
// it does return empty in wsh, but all wstore methods are invalid in wsh mode, so that shouldn't be an issue
func GetClientId() string {
	clientIdLock.Lock()
	defer clientIdLock.Unlock()
	if dorabase.IsDevMode() && cachedClientId == "" {
		panic("cachedClientId is empty")
	}
	return cachedClientId
}

func UpdateTabName(ctx context.Context, tabId, name string) error {
	return WithTx(ctx, func(tx *TxWrap) error {
		tab, _ := DBGet[*doraobj.Tab](tx.Context(), tabId)
		if tab == nil {
			return fmt.Errorf("tab not found: %q", tabId)
		}
		if tabId != "" {
			tab.Name = name
			DBUpdate(tx.Context(), tab)
		}
		return nil
	})
}

func UpdateObjectMeta(ctx context.Context, oref doraobj.ORef, meta doraobj.MetaMapType, mergeSpecial bool) error {
	return WithTx(ctx, func(tx *TxWrap) error {
		if oref.IsEmpty() {
			return fmt.Errorf("empty object reference")
		}
		obj, _ := DBGetORef(tx.Context(), oref)
		if obj == nil {
			return ErrNotFound
		}
		objMeta := doraobj.GetMeta(obj)
		if objMeta == nil {
			objMeta = make(map[string]any)
		}
		newMeta := doraobj.MergeMeta(objMeta, meta, mergeSpecial)
		doraobj.SetMeta(obj, newMeta)
		DBUpdate(tx.Context(), obj)
		return nil
	})
}

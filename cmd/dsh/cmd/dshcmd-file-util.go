// Copyright 2026, Command Line Inc.
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"io/fs"
	"strings"

	"github.com/dfbb/doraterm/pkg/remote/connparse"
	"github.com/dfbb/doraterm/pkg/util/fileutil"
	"github.com/dfbb/doraterm/pkg/dshrpc"
	"github.com/dfbb/doraterm/pkg/dshrpc/wshclient"
	"github.com/dfbb/doraterm/pkg/dshutil"
)

func convertNotFoundErr(err error) error {
	if err == nil {
		return nil
	}
	if strings.HasPrefix(err.Error(), "NOTFOUND:") {
		return fs.ErrNotExist
	}
	return err
}

func ensureFile(fileData dshrpc.FileData) (*dshrpc.FileInfo, error) {
	info, err := dshclient.FileInfoCommand(RpcClient, fileData, &dshrpc.RpcOpts{Timeout: fileTimeout})
	err = convertNotFoundErr(err)
	if err == fs.ErrNotExist {
		err = dshclient.FileCreateCommand(RpcClient, fileData, &dshrpc.RpcOpts{Timeout: fileTimeout})
		if err != nil {
			return nil, fmt.Errorf("creating file: %w", err)
		}
		info, err = dshclient.FileInfoCommand(RpcClient, fileData, &dshrpc.RpcOpts{Timeout: fileTimeout})
		if err != nil {
			return nil, fmt.Errorf("getting file info: %w", err)
		}
		return info, err
	}
	if err != nil {
		return nil, fmt.Errorf("getting file info: %w", err)
	}
	return info, nil
}

func streamWriteToFile(fileData dshrpc.FileData, reader io.Reader) error {
	// First truncate the file with an empty write
	emptyWrite := fileData
	emptyWrite.Data64 = ""
	err := dshclient.FileWriteCommand(RpcClient, emptyWrite, &dshrpc.RpcOpts{Timeout: fileTimeout})
	if err != nil {
		return fmt.Errorf("initializing file with empty write: %w", err)
	}

	const chunkSize = dshrpc.FileChunkSize // 32KB chunks
	buf := make([]byte, chunkSize)
	totalWritten := int64(0)

	for {
		n, err := reader.Read(buf)
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("reading input: %w", err)
		}

		// Check total size
		totalWritten += int64(n)
		if totalWritten > MaxFileSize {
			return fmt.Errorf("input exceeds maximum file size of %d bytes", MaxFileSize)
		}

		// Prepare and send chunk
		chunk := buf[:n]
		appendData := fileData
		appendData.Data64 = base64.StdEncoding.EncodeToString(chunk)

		err = dshclient.FileAppendCommand(RpcClient, appendData, &dshrpc.RpcOpts{Timeout: int64(fileTimeout)})
		if err != nil {
			return fmt.Errorf("appending chunk to file: %w", err)
		}
	}

	return nil
}

func streamReadFromFile(ctx context.Context, fileData dshrpc.FileData, writer io.Writer) error {
	broker := RpcClient.StreamBroker
	if broker == nil {
		return fmt.Errorf("stream broker not available")
	}
	if fileData.Info == nil {
		return fmt.Errorf("file info is required")
	}
	readerRouteId := RpcClientRouteId
	if readerRouteId == "" {
		return fmt.Errorf("no route id available")
	}
	conn, err := connparse.ParseURI(fileData.Info.Path)
	if err != nil {
		return fmt.Errorf("parsing file path: %w", err)
	}
	writerRouteId := dshutil.MakeConnectionRouteId(conn.Host)
	reader, streamMeta := broker.CreateStreamReader(readerRouteId, writerRouteId, 256*1024)
	defer reader.Close()
	go func() {
		<-ctx.Done()
		reader.Close()
	}()
	data := dshrpc.CommandFileStreamData{
		Info:       fileData.Info,
		StreamMeta: *streamMeta,
	}
	_, err = dshclient.FileStreamCommand(RpcClient, data, nil)
	if err != nil {
		return fmt.Errorf("starting file stream: %w", err)
	}
	_, err = io.Copy(writer, reader)
	return err
}

func fixRelativePaths(path string) (string, error) {
	conn, err := connparse.ParseURI(path)
	if err != nil {
		return "", err
	}
	if conn.Scheme != connparse.ConnectionTypeWsh || conn.Host != connparse.ConnHostCurrent {
		return "", fmt.Errorf("remote/wsl paths not supported in doraterm: %s", path)
	}
	conn.Host = RpcContext.Conn
	fixedPath, err := fileutil.FixPath(conn.Path)
	if err != nil {
		return "", err
	}
	conn.Path = fixedPath
	return conn.GetFullURI(), nil
}

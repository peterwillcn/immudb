/*
Copyright 2022 CodeNotary, Inc. All rights reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package integration

import (
	"context"
	"net"
	"os"
	"path"
	"testing"

	"github.com/codenotary/immudb/pkg/client"
	"github.com/codenotary/immudb/pkg/server"
	"github.com/codenotary/immudb/pkg/server/servertest"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
)

func testClientNoLogin(t *testing.T) (client.ImmuClient, func()) {
	tmpDir, err := os.MkdirTemp("", "immudb-test")
	require.NoError(t, err)

	srv := servertest.NewBufconnServer(
		server.DefaultOptions().WithDir(path.Join(tmpDir, "srv")),
	)

	cleanup := func() {
		srv.Stop()
		os.RemoveAll(tmpDir)
	}
	initialized := false
	defer func() {
		if !initialized {
			cleanup()
		}
	}()

	err = srv.Start()
	require.NoError(t, err)

	clientOpts := client.DefaultOptions().WithDir(path.Join(tmpDir, "cli"))

	clientOpts.WithDialOptions(append(clientOpts.DialOptions,
		grpc.WithContextDialer(func(ctx context.Context, s string) (net.Conn, error) {
			return srv.Lis.Dial()
		}),
		grpc.WithInsecure(),
	))

	client, err := client.NewImmuClient(clientOpts)
	require.NoError(t, err)

	initialized = true

	return client, cleanup
}

func testClient(t *testing.T) (client.ImmuClient, context.Context, func()) {
	ctx := context.Background()

	cli, cleanup := testClientNoLogin(t)
	err := cli.OpenSession(ctx, []byte("immudb"), []byte("immudb"), "defaultdb")
	if err != nil {
		cleanup()
		require.NoError(t, err)
	}

	return cli, ctx, func() {
		err := cli.CloseSession(ctx)
		require.NoError(t, err)
		cleanup()
	}

}

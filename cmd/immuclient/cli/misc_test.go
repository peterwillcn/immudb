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

package cli

import (
	"os"
	"testing"

	test "github.com/codenotary/immudb/cmd/immuclient/immuclienttest"
	"github.com/codenotary/immudb/pkg/client/tokenservice"
	"github.com/codenotary/immudb/pkg/server"
	"github.com/codenotary/immudb/pkg/server/servertest"
	"github.com/stretchr/testify/require"
)

func testCliNoLogin(t *testing.T) (*cli, func()) {
	options := server.DefaultOptions().WithAuth(true)
	bs := servertest.NewBufconnServer(options)

	err := bs.Start()
	require.NoError(t, err)

	cleanup := func() {
		bs.Stop()
		os.RemoveAll(options.Dir)
		os.Remove(".state-")
	}
	setupFinished := false
	defer func() {
		if !setupFinished {
			cleanup()
		}
	}()

	ts := tokenservice.
		NewFileTokenService().
		WithHds(&test.HomedirServiceMock{}).
		WithTokenFileName("token")
	ic := test.NewClientTest(&test.PasswordReader{
		Pass: []string{"immudb"},
	}, ts)
	ic.Connect(bs.Dialer)

	cli := new(cli)
	cli.immucl = ic.Imc
	cli.immucl.WithFileTokenService(ts)

	setupFinished = true
	return cli, cleanup
}

func testCli(t *testing.T) (*cli, func()) {
	cli, cleanup := testCliNoLogin(t)

	_, err := cli.login([]string{"immudb"})
	if err != nil {
		cleanup()
		require.NoError(t, err)
	}

	return cli, cleanup
}

func testCliWithCommants(t *testing.T) (*cli, func()) {
	cli, cleanup := testCli(t)

	cli.commands = make(map[string]*command, 0)
	cli.commandsList = make([]*command, 0)
	cli.initCommands()
	cli.helpInit()

	return cli, cleanup
}

func TestHealthCheck(t *testing.T) {
	cli, cleanup := testCli(t)
	defer cleanup()

	msg, err := cli.healthCheck([]string{})
	require.NoError(t, err)
	require.Contains(t, msg, "Health check OK")
}

func TestHistory(t *testing.T) {
	cli, cleanup := testCli(t)
	defer cleanup()

	msg, err := cli.history([]string{"key"})
	require.NoError(t, err)
	require.Contains(t, msg, "key not found")

	_, err = cli.set([]string{"key", "value"})
	require.NoError(t, err)

	msg, err = cli.history([]string{"key"})
	require.NoError(t, err)
	require.Contains(t, msg, "value")
}

func TestVersion(t *testing.T) {
	cli, cleanup := testCli(t)
	defer cleanup()

	msg, err := cli.version([]string{"key"})
	require.NoError(t, err)
	require.Contains(t, msg, "no version info available")
}

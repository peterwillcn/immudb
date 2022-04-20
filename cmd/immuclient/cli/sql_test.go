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

func testCli(t *testing.T) (*cli, func()) {
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
	ic.Login("immudb")

	cli := new(cli)
	cli.immucl = ic.Imc

	setupFinished = true
	return cli, cleanup
}

func TestSqlFloat(t *testing.T) {
	cli, cleanup := testCli(t)
	defer cleanup()

	_, err := cli.sqlExec([]string{
		"CREATE TABLE t1(id INTEGER AUTO_INCREMENT, val FLOAT, PRIMARY KEY(id))",
	})
	require.NoError(t, err)

	_, err = cli.sqlExec([]string{
		"INSERT INTO t1(val) VALUES(1.1)",
	})
	require.NoError(t, err)

	s, err := cli.sqlQuery([]string{
		"SELECT id, val FROM t1",
	})
	require.NoError(t, err)
	require.Regexp(t, `(?m)^\|\s+\d+\s+\|\s+1\.1\s+\|$`, s)
}

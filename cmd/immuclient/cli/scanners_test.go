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
	"testing"

	"github.com/stretchr/testify/require"
)

func TestZScan(t *testing.T) {
	cli, cleanup := testCli(t)
	defer cleanup()

	_, err := cli.set([]string{"key", "val"})
	require.NoError(t, err)

	_, err = cli.zAdd([]string{"set", "445.3", "key"})
	require.NoError(t, err)

	msg, err := cli.zScan([]string{"set"})
	require.NoError(t, err)
	require.Contains(t, msg, "value")
}

func TestScan(t *testing.T) {
	cli, cleanup := testCli(t)
	defer cleanup()

	_, err := cli.set([]string{"key", "val"})
	require.NoError(t, err)

	msg, err := cli.scan([]string{"k"})
	require.NoError(t, err)
	require.Contains(t, msg, "value")
}

func _TestCount(t *testing.T) {
	cli, cleanup := testCli(t)
	defer cleanup()

	_, err := cli.set([]string{"key", "val"})
	require.NoError(t, err)

	msg, err := cli.count([]string{"key"})
	require.NoError(t, err)
	require.Contains(t, msg, "1")
}

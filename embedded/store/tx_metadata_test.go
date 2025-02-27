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
package store

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestTxMetadata(t *testing.T) {
	md := &TxMetadata{}

	bs := md.Bytes()
	require.Nil(t, bs)

	err := md.ReadFrom(bs)
	require.NoError(t, err)

	desmd := &TxMetadata{}
	err = desmd.ReadFrom(nil)
	require.NoError(t, err)

	err = desmd.ReadFrom(desmd.Bytes())
	require.NoError(t, err)

	require.True(t, md.Equal(desmd))
}

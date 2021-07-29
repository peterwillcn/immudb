/*
Copyright 2021 CodeNotary, Inc. All rights reserved.

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
package ahtree

import (
	"crypto/sha256"
	"encoding/binary"
	"errors"
	"io/ioutil"
	"os"
	"testing"

	"github.com/codenotary/immudb/embedded/appendable"
	"github.com/codenotary/immudb/embedded/appendable/mocked"
	"github.com/codenotary/immudb/embedded/appendable/multiapp"

	"github.com/stretchr/testify/require"
)

func TestNodeNumberCalculation(t *testing.T) {
	var nodesUptoTests = []struct {
		n        uint64
		expected uint64
	}{
		{1, 1},
		{2, 3},
		{3, 5},
		{4, 8},
		{5, 10},
		{6, 13},
		{7, 16},
		{8, 20},
		{9, 22},
		{10, 25},
		{11, 28},
		{12, 32},
		{13, 35},
		{14, 39},
		{15, 43},
		{16, 48},
	}

	for _, tt := range nodesUptoTests {
		actual := nodesUpto(tt.n)
		require.Equal(t, tt.expected, actual)

		require.Equal(t, tt.expected, nodesUntil(tt.n)+uint64(levelsAt(tt.n))+1)
	}
}

func TestEdgeCases(t *testing.T) {
	_, err := Open("ahtree_test", nil)
	require.Equal(t, ErrIllegalArguments, err)

	_, err = OpenWith(nil, nil, nil, nil)
	require.Equal(t, ErrIllegalArguments, err)

	_, err = OpenWith(nil, nil, nil, DefaultOptions())
	require.Equal(t, ErrIllegalArguments, err)

	dummySetOffset := func(off int64) error {
		return nil
	}

	pLog := &mocked.MockedAppendable{SetOffsetFn: dummySetOffset}
	dLog := &mocked.MockedAppendable{SetOffsetFn: dummySetOffset}
	cLog := &mocked.MockedAppendable{SetOffsetFn: dummySetOffset}

	injectedErr := errors.New("error")

	t.Run("should fail while querying cLog size", func(t *testing.T) {
		cLog.SizeFn = func() (int64, error) {
			return 0, injectedErr
		}

		_, err = OpenWith(pLog, dLog, cLog, DefaultOptions())
		require.ErrorIs(t, err, injectedErr)
	})

	t.Run("should fail while setting cLog offset", func(t *testing.T) {
		cLog.SizeFn = func() (int64, error) {
			return cLogEntrySize - 1, nil
		}
		cLog.SetOffsetFn = func(off int64) error {
			return injectedErr
		}

		_, err = OpenWith(pLog, dLog, cLog, DefaultOptions())
		require.ErrorIs(t, err, injectedErr)
	})

	t.Run("should fail while appending payload", func(t *testing.T) {
		cLog.SetOffsetFn = dummySetOffset
		cLog.SizeFn = func() (int64, error) {
			return cLogEntrySize - 1, nil
		}
		pLog.AppendFn = func(bs []byte) (off int64, n int, err error) {
			return 0, 0, injectedErr
		}

		tree, err := OpenWith(pLog, dLog, cLog, DefaultOptions())
		require.NoError(t, err)

		_, _, err = tree.Append([]byte{1, 2, 3})
		require.ErrorIs(t, err, injectedErr)
	})

	t.Run("should fail while validating plog size", func(t *testing.T) {
		cLog.SizeFn = func() (int64, error) {
			return cLogEntrySize + 1, nil
		}
		cLog.ReadAtFn = func(bs []byte, off int64) (int, error) {
			binary.BigEndian.PutUint64(bs[:], 0)
			binary.BigEndian.PutUint32(bs[offsetSize:], 8)
			return cLogEntrySize, nil
		}
		pLog.SizeFn = func() (int64, error) {
			return 0, nil
		}

		_, err = OpenWith(pLog, dLog, cLog, DefaultOptions())
		require.Equal(t, ErrorCorruptedData, err)
	})

	t.Run("should fail while validating dLog size", func(t *testing.T) {
		cLog.SizeFn = func() (int64, error) {
			return cLogEntrySize + 1, nil
		}
		cLog.ReadAtFn = func(bs []byte, off int64) (int, error) {
			binary.BigEndian.PutUint64(bs[:], 0)
			binary.BigEndian.PutUint32(bs[offsetSize:], 8)
			return cLogEntrySize, nil
		}
		pLog.SizeFn = func() (int64, error) {
			return 8, nil
		}
		dLog.SizeFn = func() (int64, error) {
			return 0, nil
		}

		_, err = OpenWith(pLog, dLog, cLog, DefaultOptions())
		require.Equal(t, ErrorCorruptedDigests, err)
	})

	t.Run("should fail reading dLog size", func(t *testing.T) {
		dLog.SizeFn = func() (int64, error) {
			return 0, injectedErr
		}

		_, err = OpenWith(pLog, dLog, cLog, DefaultOptions())
		require.ErrorIs(t, err, injectedErr)
	})

	t.Run("should fail reading pLog size", func(t *testing.T) {
		pLog.SizeFn = func() (int64, error) {
			return 0, injectedErr
		}

		_, err = OpenWith(pLog, dLog, cLog, DefaultOptions())
		require.ErrorIs(t, err, injectedErr)
	})

	t.Run("should fail reading last cLog entry", func(t *testing.T) {
		cLog.SizeFn = func() (int64, error) {
			return 0, nil
		}
		dLog.SizeFn = func() (int64, error) {
			return 0, nil
		}
		pLog.SizeFn = func() (int64, error) {
			return 0, nil
		}

		metadata := appendable.NewMetadata(nil)
		metadata.PutInt(MetaVersion, Version)

		pLog.MetadataFn = metadata.Bytes
		dLog.MetadataFn = metadata.Bytes
		cLog.MetadataFn = metadata.Bytes

		cLog.SizeFn = func() (int64, error) {
			return cLogEntrySize, nil
		}

		cLog.ReadAtFn = func(bs []byte, off int64) (int, error) {
			return 0, injectedErr
		}

		_, err = OpenWith(pLog, dLog, cLog, DefaultOptions())
		require.ErrorIs(t, err, injectedErr)
	})

	t.Run("should fail reading pLog size", func(t *testing.T) {
		cLog.SizeFn = func() (int64, error) {
			return cLogEntrySize, nil
		}
		pLog.SizeFn = func() (int64, error) {
			return 0, injectedErr
		}

		_, err = OpenWith(pLog, dLog, cLog, DefaultOptions())
		require.ErrorIs(t, err, injectedErr)
	})

	t.Run("should fail flushing pLog", func(t *testing.T) {
		cLog.SizeFn = func() (int64, error) {
			return 0, nil
		}
		pLog.SizeFn = func() (int64, error) {
			return 0, nil
		}
		pLog.SetOffsetFn = func(off int64) error {
			return nil
		}
		pLog.AppendFn = func(bs []byte) (off int64, n int, err error) {
			return 0, 0, nil
		}
		pLog.FlushFn = func() error {
			return injectedErr
		}

		tree, err := OpenWith(pLog, dLog, cLog, DefaultOptions())
		require.NoError(t, err)

		_, _, err = tree.Append([]byte{1, 2, 3})
		require.ErrorIs(t, err, injectedErr)
	})

	t.Run("should fail appending to dLog", func(t *testing.T) {
		pLog.AppendFn = func(bs []byte) (off int64, n int, err error) {
			return 0, 0, nil
		}
		pLog.FlushFn = func() error {
			return nil
		}
		dLog.AppendFn = func(bs []byte) (off int64, n int, err error) {
			return 0, 0, injectedErr
		}

		tree, err := OpenWith(pLog, dLog, cLog, DefaultOptions())
		require.NoError(t, err)

		_, _, err = tree.Append(nil)
		require.ErrorIs(t, err, ErrIllegalArguments)

		_, _, err = tree.Append([]byte{1, 2, 3})
		require.ErrorIs(t, err, injectedErr)
	})

	t.Run("should fail due to invalid path", func(t *testing.T) {
		_, err = Open("options.go", DefaultOptions())
		require.Equal(t, ErrorPathIsNotADirectory, err)
	})

	t.Run("should fail due to invalid cache size", func(t *testing.T) {
		_, err = Open("ahtree_test", DefaultOptions().WithDataCacheSlots(-1))
		require.Equal(t, ErrIllegalArguments, err)
	})

	t.Run("should fail due to invalid digests cache size", func(t *testing.T) {
		_, err = Open("ahtree_test", DefaultOptions().WithDigestsCacheSlots(-1))
		require.Equal(t, ErrIllegalArguments, err)
	})

	tree, err := Open("ahtree_test", DefaultOptions().WithSynced(false))
	require.NoError(t, err)
	defer os.RemoveAll("ahtree_test")

	_, _, err = tree.Root()
	require.Equal(t, ErrEmptyTree, err)

	_, err = tree.rootAt(1)
	require.Equal(t, ErrEmptyTree, err)

	_, err = tree.rootAt(0)
	require.Equal(t, ErrIllegalArguments, err)

	_, err = tree.DataAt(0)
	require.Equal(t, ErrIllegalArguments, err)

	err = tree.Sync()
	require.NoError(t, err)

	_, _, err = tree.Append([]byte{1})
	require.NoError(t, err)

	err = tree.Close()
	require.NoError(t, err)

	_, _, err = tree.Append(nil)
	require.Equal(t, ErrAlreadyClosed, err)

	_, err = tree.InclusionProof(1, 2)
	require.Equal(t, ErrAlreadyClosed, err)

	_, err = tree.ConsistencyProof(1, 2)
	require.Equal(t, ErrAlreadyClosed, err)

	_, _, err = tree.Root()
	require.Equal(t, ErrAlreadyClosed, err)

	_, err = tree.rootAt(1)
	require.Equal(t, ErrAlreadyClosed, err)

	_, err = tree.DataAt(1)
	require.Equal(t, ErrAlreadyClosed, err)

	err = tree.Sync()
	require.Equal(t, ErrAlreadyClosed, err)

	err = tree.Close()
	require.Equal(t, ErrAlreadyClosed, err)
}

func TestReadOnly(t *testing.T) {
	_, err := Open("ahtree_test", DefaultOptions().WithReadOnly(true))
	defer os.RemoveAll("ahtree_test")
	require.Error(t, err)

	tree, err := Open("ahtree_test", DefaultOptions().WithReadOnly(false))
	require.NoError(t, err)
	err = tree.Close()
	require.NoError(t, err)

	tree, err = Open("ahtree_test", DefaultOptions().WithReadOnly(true))
	require.NoError(t, err)

	_, _, err = tree.Append(nil)
	require.Equal(t, ErrReadOnly, err)

	err = tree.Sync()
	require.Equal(t, ErrReadOnly, err)

	err = tree.Close()
	require.NoError(t, err)
}

func TestAppend(t *testing.T) {
	opts := DefaultOptions().WithSynced(false).WithDigestsCacheSlots(100).WithDataCacheSlots(100)
	tree, err := Open("ahtree_test", opts)
	require.NoError(t, err)
	defer os.RemoveAll("ahtree_test")

	N := 1024

	for i := 1; i <= N; i++ {
		p := []byte{byte(i)}

		_, _, err := tree.Append(p)
		require.NoError(t, err)

		ri, err := tree.RootAt(uint64(i))
		require.NoError(t, err)

		n, r, err := tree.Root()
		require.NoError(t, err)
		require.Equal(t, uint64(i), n)
		require.Equal(t, r, ri)

		sz := tree.Size()
		require.Equal(t, uint64(i), sz)

		rp, err := tree.DataAt(uint64(i))
		require.NoError(t, err)
		require.Equal(t, p, rp)

		_, err = tree.RootAt(uint64(i) + 1)
		require.Equal(t, ErrUnexistentData, err)

		_, err = tree.DataAt(uint64(i) + 1)
		require.Equal(t, ErrUnexistentData, err)
	}

	rp, err := tree.DataAt(uint64(1))
	require.NoError(t, err)
	require.Equal(t, []byte{byte(1)}, rp)

	err = tree.Sync()
	require.NoError(t, err)

	err = tree.Close()
	require.NoError(t, err)
}

func TestIntegrity(t *testing.T) {
	tree, err := Open("ahtree_test", DefaultOptions().WithSynced(false))
	require.NoError(t, err)
	defer os.RemoveAll("ahtree_test")

	N := 1024

	for i := 1; i <= N; i++ {
		_, _, err := tree.Append([]byte{byte(i)})
		require.NoError(t, err)
	}

	n, _, err := tree.Root()
	require.NoError(t, err)

	for i := uint64(1); i <= n; i++ {
		r, err := tree.RootAt(i)
		require.NoError(t, err)

		for j := uint64(1); j <= i; j++ {
			iproof, err := tree.InclusionProof(j, i)
			require.NoError(t, err)

			d, err := tree.DataAt(j)
			require.NoError(t, err)

			pd := make([]byte, 1+len(d))
			pd[0] = LeafPrefix
			copy(pd[1:], d)

			verifies := VerifyInclusion(iproof, j, i, sha256.Sum256(pd), r)
			require.True(t, verifies)
		}
	}
}

func TestOpenFail(t *testing.T) {
	_, err := Open("/dev/null", DefaultOptions().WithSynced(false))
	require.Error(t, err)
	os.Mkdir("ro_dir1", 0500)
	defer os.RemoveAll("ro_dir1")
	_, err = Open("ro_dir/bla", DefaultOptions().WithSynced(false))
	require.Error(t, err)
	_, err = Open("wrongdir\000", DefaultOptions().WithSynced(false))
	require.Error(t, err)
	defer os.RemoveAll("tt1")
	_, err = Open("tt1", DefaultOptions().WithSynced(false).WithAppFactory(
		func(rootPath, subPath string, opts *multiapp.Options) (a appendable.Appendable, e error) {
			if subPath == "tree" {
				e = errors.New("simulated error")
			}
			return
		}))
	_, err = Open("tt1", DefaultOptions().WithSynced(false).WithAppFactory(
		func(rootPath, subPath string, opts *multiapp.Options) (a appendable.Appendable, e error) {
			if subPath == "commit" {
				e = errors.New("simulated error")
			}
			return
		}))
	require.Error(t, err)
}

func TestInclusionAndConsistencyProofs(t *testing.T) {
	tree, err := Open("ahtree_test", DefaultOptions().WithSynced(false))
	require.NoError(t, err)
	defer os.RemoveAll("ahtree_test")

	N := 1024

	for i := 1; i <= N; i++ {
		_, r, err := tree.Append([]byte{byte(i)})
		require.NoError(t, err)

		iproof, err := tree.InclusionProof(uint64(i), uint64(i))
		require.NoError(t, err)

		h := sha256.Sum256([]byte{LeafPrefix, byte(i)})

		verifies := VerifyInclusion(iproof, uint64(i), uint64(i), h, r)
		require.True(t, verifies)
	}

	_, err = tree.InclusionProof(2, 1)
	require.Equal(t, ErrIllegalArguments, err)

	_, err = tree.ConsistencyProof(2, 1)
	require.Equal(t, ErrIllegalArguments, err)

	for i := 1; i <= N; i++ {
		for j := i; j <= N; j++ {
			iproof, err := tree.InclusionProof(uint64(i), uint64(j))
			require.NoError(t, err)

			jroot, err := tree.RootAt(uint64(j))
			require.NoError(t, err)

			h := sha256.Sum256([]byte{LeafPrefix, byte(i)})

			verifies := VerifyInclusion(iproof, uint64(i), uint64(j), h, jroot)
			require.True(t, verifies)

			cproof, err := tree.ConsistencyProof(uint64(i), uint64(j))
			require.NoError(t, err)

			iroot, err := tree.RootAt(uint64(i))
			require.NoError(t, err)

			verifies = VerifyConsistency(cproof, uint64(i), uint64(j), iroot, jroot)
			require.True(t, verifies)
		}
	}

	for i := 1; i <= N; i++ {
		iproof, err := tree.InclusionProof(uint64(i), uint64(N))
		require.NoError(t, err)

		h := sha256.Sum256([]byte{LeafPrefix, byte(i)})
		root, err := tree.RootAt(uint64(i))
		require.NoError(t, err)

		verifies := VerifyLastInclusion(iproof, uint64(i), h, root)

		if i < N {
			require.False(t, verifies)
		} else {
			require.True(t, verifies)
		}
	}

	err = tree.Close()
	require.NoError(t, err)
}

func TestReOpenningImmudbStore(t *testing.T) {
	defer os.RemoveAll("ahtree_test")

	ItCount := 5
	ACount := 100

	for it := 0; it < ItCount; it++ {
		tree, err := Open("ahtree_test", DefaultOptions().WithSynced(false))
		require.NoError(t, err)

		for i := 0; i < ACount; i++ {
			p := []byte{byte(i)}

			_, _, err := tree.Append(p)
			require.NoError(t, err)
		}

		err = tree.Close()
		require.NoError(t, err)
	}

	tree, err := Open("ahtree_test", DefaultOptions().WithSynced(false))
	require.NoError(t, err)

	for i := 1; i <= ItCount*ACount; i++ {
		for j := i; j <= ItCount*ACount; j++ {
			proof, err := tree.InclusionProof(uint64(i), uint64(j))
			require.NoError(t, err)

			root, _ := tree.RootAt(uint64(j))

			h := sha256.Sum256([]byte{LeafPrefix, byte((i - 1) % ACount)})

			verifies := VerifyInclusion(proof, uint64(i), uint64(j), h, root)
			require.True(t, verifies)
		}
	}

	err = tree.Close()
	require.NoError(t, err)
}

func TestReset(t *testing.T) {
	path, err := ioutil.TempDir("", "ahtree_test_reset")
	require.NoError(t, err)
	defer os.RemoveAll(path)

	tree, err := Open(path, DefaultOptions())
	require.NoError(t, err)

	N := 32

	for i := 1; i <= N; i++ {
		_, _, err := tree.Append([]byte{byte(i)})
		require.NoError(t, err)
	}

	err = tree.ResetSize(0)
	require.NoError(t, err)
	require.Zero(t, tree.Size())

	N = 1024

	for i := 1; i <= N; i++ {
		_, _, err := tree.Append([]byte{byte(i)})
		require.NoError(t, err)
	}

	err = tree.ResetSize(uint64(N + 1))
	require.ErrorIs(t, err, ErrCannotResetToLargerSize)

	err = tree.ResetSize(uint64(N))
	require.NoError(t, err)
	require.Equal(t, uint64(N), tree.Size())

	N = 512

	err = tree.ResetSize(uint64(N))
	require.NoError(t, err)
	require.Equal(t, uint64(N), tree.Size())

	for i := 1; i <= N; i++ {
		for j := i; j <= N; j++ {
			iproof, err := tree.InclusionProof(uint64(i), uint64(j))
			require.NoError(t, err)

			jroot, err := tree.RootAt(uint64(j))
			require.NoError(t, err)

			h := sha256.Sum256([]byte{LeafPrefix, byte(i)})

			verifies := VerifyInclusion(iproof, uint64(i), uint64(j), h, jroot)
			require.True(t, verifies)

			cproof, err := tree.ConsistencyProof(uint64(i), uint64(j))
			require.NoError(t, err)

			iroot, err := tree.RootAt(uint64(i))
			require.NoError(t, err)

			verifies = VerifyConsistency(cproof, uint64(i), uint64(j), iroot, jroot)
			require.True(t, verifies)
		}
	}

	for i := 1; i <= N; i++ {
		iproof, err := tree.InclusionProof(uint64(i), uint64(N))
		require.NoError(t, err)

		h := sha256.Sum256([]byte{LeafPrefix, byte(i)})
		root, err := tree.RootAt(uint64(i))
		require.NoError(t, err)

		verifies := VerifyLastInclusion(iproof, uint64(i), h, root)

		if i < N {
			require.False(t, verifies)
		} else {
			require.True(t, verifies)
		}
	}

	err = tree.Close()
	require.NoError(t, err)

	err = tree.ResetSize(uint64(N))
	require.ErrorIs(t, err, ErrAlreadyClosed)

	tree, err = Open(path, DefaultOptions().WithReadOnly(true))
	require.NoError(t, err)

	err = tree.ResetSize(1)
	require.ErrorIs(t, err, ErrReadOnly)

	err = tree.Close()
	require.NoError(t, err)
}

func BenchmarkAppend(b *testing.B) {
	tree, _ := Open("ahtree_test", DefaultOptions().WithSynced(false))
	defer os.RemoveAll("ahtree_test")

	for i := 0; i < b.N; i++ {
		_, _, err := tree.Append([]byte{byte(i)})
		if err != nil {
			panic(err)
		}
	}
}

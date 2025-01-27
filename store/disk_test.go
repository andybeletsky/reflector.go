package store

import (
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"testing"

	"github.com/lbryio/lbry.go/v2/extras/errors"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDiskStore_Get(t *testing.T) {
	tmpDir, err := ioutil.TempDir("", "reflector_test_*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)
	d := NewDiskStore(tmpDir, 2)

	hash := "f428b8265d65dad7f8ffa52922bba836404cbd62f3ecfe10adba6b444f8f658938e54f5981ac4de39644d5b93d89a94b"
	data := []byte("oyuntyausntoyaunpdoyruoyduanrstjwfjyuwf")

	expectedPath := path.Join(tmpDir, hash[:2], hash)
	err = os.MkdirAll(filepath.Dir(expectedPath), os.ModePerm)
	require.NoError(t, err)
	err = ioutil.WriteFile(expectedPath, data, os.ModePerm)
	require.NoError(t, err)

	blob, _, err := d.Get(hash)
	assert.NoError(t, err)
	assert.EqualValues(t, data, blob)
}

func TestDiskStore_GetNonexistentBlob(t *testing.T) {
	tmpDir, err := ioutil.TempDir("", "reflector_test_*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)
	d := NewDiskStore(tmpDir, 2)

	blob, _, err := d.Get("nonexistent")
	assert.Nil(t, blob)
	assert.True(t, errors.Is(err, ErrBlobNotFound))
}

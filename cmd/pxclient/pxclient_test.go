package pxclient

import (
	"flag"
	"testing"

	"github.com/libopenstorage/openstorage/api"
	"github.com/stretchr/testify/require"
)

const (
	volName   = "testvol"
	snapName  = "testsnap"
	cloneName = "testclone"
)

var pxhost = flag.String("pxhost", "localhost", "The address on which PX server is running")

func TestVolCreate(t *testing.T) {
	require.NotEmpty(t, *pxhost, "pxhost not given")

	drv, err := getVolDriver(*pxhost)
	require.NoError(t, err, "failed to get vol driver")

	locator := &api.VolumeLocator{
		Name: volName,
	}
	spec := &api.VolumeSpec{
		HaLevel: 1,
		Format:  api.FSType_FS_TYPE_EXT4,
	}

	volID, err := drv.Create(locator, nil, spec)
	require.NoError(t, err, "failed to create vol")
	require.NotEmpty(t, volID, "empty volume ID")
}

func TestSnapCreate(t *testing.T) {
	require.NotEmpty(t, *pxhost, "pxhost not given")

	drv, err := getVolDriver(*pxhost)
	require.NoError(t, err, "failed to get vol driver")

	locator := &api.VolumeLocator{
		Name: snapName,
	}
	snapID, err := drv.Snapshot(volName, true, locator)
	require.NoError(t, err, "failed to create snapshot")
	require.NotEmpty(t, snapID, "empty snap ID")
}

func TestCloneCreate(t *testing.T) {
	require.NotEmpty(t, *pxhost, "pxhost not given")

	drv, err := getVolDriver(*pxhost)
	require.NoError(t, err, "failed to get vol driver")

	locator := &api.VolumeLocator{
		Name: cloneName,
	}
	source := &api.Source{
		Parent: snapName,
	}

	spec := &api.VolumeSpec{
		HaLevel: 1,
		Format:  api.FSType_FS_TYPE_EXT4,
	}

	cloneID, err := drv.Create(locator, source, spec)
	require.NoError(t, err, "failed to create snapshot")
	require.NotEmpty(t, cloneID, "empty clone ID")
}

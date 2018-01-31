package pxclient

import (
	"fmt"
	"net/url"
	"strconv"

	httpclient "github.com/libopenstorage/openstorage/api/client"
	volumeclient "github.com/libopenstorage/openstorage/api/client/volume"
	"github.com/libopenstorage/openstorage/volume"
)

var srcID string

func buildHTTPSEndpoint(host string, port string) string {
	endpoint := &url.URL{}
	endpoint.Scheme = "http"
	endpoint.Host = fmt.Sprintf("%s:%s", host, port)

	return endpoint.String()
}

func getNewVolumeclient(endpoint, port, driverVersion, driverName string) (*httpclient.Client, error) {
	endpoint = buildHTTPSEndpoint(endpoint, port)
	if driverName == "" {
		driverName = "pxd"
	}
	clnt, err := volumeclient.NewDriverClient(endpoint, "pxd", driverVersion, driverName)
	if err != nil {
		return nil, err
	}

	return clnt, nil
}

func getVolDriver(host string) (volume.VolumeDriver, error) {
	var driverVersion string
	clnt, err := getNewVolumeclient(host, strconv.FormatInt(9001, 10), "", "")
	if err != nil {
		return nil, err
	}

	endpoint := buildHTTPSEndpoint(host, strconv.FormatInt(9001, 10))
	versions, err := clnt.Versions(endpoint)

	if err != nil {
		// Default to whatever OSD gives us
		// We are masking an error here for now until
		// we see the need to add another version
		driverVersion = ""
	} else {
		// Add logic to select appropriate version
		driverVersion = versions[0]
	}

	clnt, err = getNewVolumeclient(host, strconv.FormatInt(9001, 10), driverVersion, "")
	if err != nil {
		return nil, err
	}

	volDriver := volumeclient.VolumeDriver(clnt)
	return volDriver, nil
}

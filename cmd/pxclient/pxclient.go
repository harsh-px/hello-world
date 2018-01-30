package main

import (
	"flag"
	"fmt"
	"net/url"
	"strconv"

	"github.com/libopenstorage/openstorage/api"
	httpclient "github.com/libopenstorage/openstorage/api/client"
	volumeclient "github.com/libopenstorage/openstorage/api/client/volume"
	"github.com/libopenstorage/openstorage/volume"
	"github.com/sirupsen/logrus"
)

var pxhost string
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

func main() {
	fmt.Printf("Hello from px client. Host is: %s\n", pxhost)
	volDriver, err := getVolDriver(pxhost)
	if err != nil {
		logrus.Fatalf("error: failed to get vol driver. Err: %v", err)
	}

	locator := &api.VolumeLocator{}
	vols, err := volDriver.Enumerate(locator, nil)
	if err != nil {
		logrus.Fatalf("error: failed to enumerate volumes. Err: %v", err)
	}

	for _, v := range vols {
		logrus.Infof("vol: %v", v)
	}
}

func init() {
	flag.StringVar(&pxhost, "pxhost", "localhost", "The address on which PX server is running")
	flag.Parse()
}

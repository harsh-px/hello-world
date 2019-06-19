package main

import (
	"flag"
	"time"

	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/opsworks"
	"github.com/libopenstorage/cloudops"
	"github.com/libopenstorage/cloudops/aws"
	"github.com/sirupsen/logrus"
)

var (
	d     cloudops.Ops
	volID string
)

func main() {
	logrus.Infof("hello world")

	flag.StringVar(&volID, "vol-id", "", "volume to attach")
	flag.Parse()

	var err error
	d, err = aws.NewClient()
	if err != nil {
		logrus.Errorf("failed to create AWS client: %v", err)
		return
	}

	instanceID := d.InstanceID()
	if len(instanceID) == 0 {
		logrus.Errorf("failed to get instance ID")
		return
	}
	logrus.Infof("tester running on: %s", instanceID)

	description, err := d.Describe()
	if err != nil {
		logrus.Infof("failed to describe the instance: %v", err)
		return
	}

	logrus.Infof("instance description: %v", description)

	info, err := d.InspectInstance(instanceID)
	if err != nil {
		logrus.Errorf("failed to inspect instance: %v", err)
		return
	}

	logrus.Infof("instance info: %v", info)

	mappings, err := d.DeviceMappings()
	if err != nil {
		logrus.Errorf("failed to get device mappings: %v", err)
		return
	}

	logrus.Infof("device mapping before: %v", mappings)

	if len(volID) == 0 {
		logrus.Infof("User did not provide a volume ID. Creating one...")
		volType := opsworks.VolumeTypeGp2
		volSize := int64(2)
		zone := info.Zone

		ebsVol := &ec2.Volume{
			AvailabilityZone: &zone,
			VolumeType:       &volType,
			Size:             &volSize,
		}

		disk, err := d.Create(ebsVol, nil)
		if err != nil {
			logrus.Errorf("failed to create EBS volume: %v", err)
			return
		}

		logrus.Infof("created disk: %v", disk)

		volID, err = d.GetDeviceID(disk)
		if err != nil {
			logrus.Errorf("failed to get device ID: %v", err)
			return
		}

		logrus.Infof("Disk has ID: %v", volID)
	} else {
		logrus.Infof("Using user provided volume ID: %s", volID)
	}

	devPath, err := d.Attach(volID)
	if err != nil {
		logrus.Errorf("failed to attach: %v", err)
		return
	}

	mappings, err = d.DeviceMappings()
	if err != nil {
		logrus.Errorf("failed to get device mappings: %v", err)
		return
	}

	logrus.Infof("device mapping after: %v", mappings)

	logrus.Infof("attached %s at %s", volID, devPath)

	err = teardown(volID)
	if err != nil {
		return
	}
}

func teardown(diskID string) error {
	err := d.Detach(diskID)
	if err != nil {
		logrus.Errorf("failed to detach: %v", err)
		return err
	}

	time.Sleep(3 * time.Second)

	err = d.Delete(diskID)
	if err != nil {
		logrus.Errorf("failed to delete: %v", err)
		return err
	}

	return nil
}

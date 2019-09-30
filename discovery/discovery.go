package discovery

import (
	"context"
	"fmt"
	"time"

	"github.com/muka/go-bluetooth/api"
	"github.com/muka/go-bluetooth/api/beacon"
	"github.com/muka/go-bluetooth/bluez/profile/adapter"
	"github.com/muka/go-bluetooth/bluez/profile/device"
	log "github.com/sirupsen/logrus"
)

func Run(adapterID string, onlyBeacon bool) ([]*device.Device1Properties, error) {
	//clean up connection on exit
	defer api.Exit()

	a, err := adapter.GetAdapter(adapterID)
	if err != nil {
		return nil, err
	}

	log.Debug("Flush cached devices")
	err = a.FlushDevices()
	if err != nil {
		return nil, err
	}

	log.Debug("Start discovery")
	discovery, cancel, err := Discover(a, nil)
	if err != nil {
		return nil, err
	}
	defer cancel()
	var list = make([]*device.Device1Properties, 10)

	go func() {
		for ev := range discovery {

			if ev.Type == adapter.DeviceRemoved {
				continue
			}

			dev, err := device.NewDevice1(ev.Path)
			if err != nil {
				log.Errorf("%s: %s", ev.Path, err)
				continue
			}

			if dev == nil {
				log.Errorf("%s: not found", ev.Path)
				continue
			}
			log.Infof("name=%s addr=%s addrType=%s rssi=%d",
				dev.Properties.Name, dev.Properties.Address,
				dev.Properties.AddressType, dev.Properties.RSSI)
			list = append(list, dev.Properties)
			//err = handleBeacon(dev)
			//if err != nil {
			//	log.Errorf("%s: %s", ev.Path, err)
			//}
		}

	}()

	select {
	case <-time.After(time.Second * 10):
		fmt.Println("timeout")
		return list, nil
	}
}

// Discover start device discovery
func Discover(a *adapter.Adapter1, filter *adapter.DiscoveryFilter) (chan *adapter.DeviceDiscovered, func(), error) {

	err := a.SetPairable(false)
	if err != nil {
		return nil, nil, err
	}
	err = a.SetDiscoverable(false)
	if err != nil {
		return nil, nil, err
	}
	err = a.SetPowered(true)
	if err != nil {
		return nil, nil, err
	}

	filterMap := make(map[string]interface{})
	if filter != nil {
		filterMap = filter.ToMap()
	}
	err = a.SetDiscoveryFilter(filterMap)
	if err != nil {
		return nil, nil, err
	}

	err = a.StartDiscovery()
	if err != nil {
		return nil, nil, err
	}

	ch, discoveryCancel, err := a.OnDeviceDiscovered()

	cancel := func() {
		err := a.StopDiscovery()
		if err != nil {
			log.Warnf("Error stopping discovery: %s", err)
		}
		discoveryCancel()
	}

	return ch, cancel, nil
}

func handleBeacon(dev *device.Device1) error {

	b, err := beacon.NewBeacon(dev)
	if err != nil {
		return err
	}

	beaconUpdated, err := b.WatchDeviceChanges(context.Background())
	if err != nil {
		return err
	}

	isBeacon := <-beaconUpdated
	if !isBeacon {
		return nil
	}

	name := b.Device.Properties.Alias
	if name == "" {
		name = b.Device.Properties.Name
	}

	log.Debugf("Found beacon %s %s", b.Type, name)

	if b.IsEddystone() {
		eddystone := b.GetEddystone()
		switch eddystone.Frame {
		case beacon.EddystoneFrameUID:
			log.Debugf(
				"Eddystone UID %s instance %s (%ddbi)",
				eddystone.UID,
				eddystone.InstanceUID,
				eddystone.CalibratedTxPower,
			)
			break
		case beacon.EddystoneFrameTLM:
			log.Debugf(
				"Eddystone TLM temp:%.0f batt:%d last reboot:%d advertising pdu:%d (%ddbi)",
				eddystone.TLMTemperature,
				eddystone.TLMBatteryVoltage,
				eddystone.TLMLastRebootedTime,
				eddystone.TLMAdvertisingPDU,
				eddystone.CalibratedTxPower,
			)
			break
		case beacon.EddystoneFrameURL:
			log.Debugf(
				"Eddystone URL %s (%ddbi)",
				eddystone.URL,
				eddystone.CalibratedTxPower,
			)
			break
		}

	}
	if b.IsIBeacon() {
		ibeacon := b.GetIBeacon()
		log.Debugf(
			"IBeacon %s (%ddbi) (major=%d minor=%d)",
			ibeacon.ProximityUUID,
			ibeacon.MeasuredPower,
			ibeacon.Major,
			ibeacon.Minor,
		)
	}

	return nil
}

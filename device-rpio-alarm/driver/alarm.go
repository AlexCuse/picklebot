//
// Copyright (c) 2022 One Track Consulting
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//

package driver

import (
	"fmt"
	"reflect"
	"time"

	"github.com/edgexfoundry/go-mod-core-contracts/v2/clients/logger"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/common"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/models"
	"github.com/stianeikeland/go-rpio/v4"

	"github.com/alexcuse/device-rpio-alert/config"
	sdkModels "github.com/edgexfoundry/device-sdk-go/v2/pkg/models"
	"github.com/edgexfoundry/device-sdk-go/v2/pkg/service"
)

const (
	sos = "... --- ..."
)

type RPIOAlarm struct {
	lc            logger.LoggingClient
	alertPin      rpio.Pin // pin to read for alarm state
	alarmPin      rpio.Pin // pin to activate for alarm on acknowledgement
	asyncCh       chan<- *sdkModels.AsyncValues
	deviceCh      chan<- []sdkModels.DiscoveredDevice
	serviceConfig *config.ServiceConfig
	alertUntil    time.Time
}

// Initialize performs protocol-specific initialization for the device
// service.
func (s *RPIOAlarm) Initialize(lc logger.LoggingClient, asyncCh chan<- *sdkModels.AsyncValues, deviceCh chan<- []sdkModels.DiscoveredDevice) error {
	s.lc = lc
	s.asyncCh = asyncCh
	s.deviceCh = deviceCh
	s.serviceConfig = &config.ServiceConfig{}

	if err := rpio.Open(); err != nil {
		return fmt.Errorf("failed to initialize GPIO %w", err)
	}

	s.lc.Info("RPIO initialized")

	ds := service.RunningService()

	if err := ds.LoadCustomConfig(s.serviceConfig, "Alarm"); err != nil {
		return fmt.Errorf("unable to load 'Alarm' custom configuration: %s", err.Error())
	}

	lc.Infof("Custom config is: %v", s.serviceConfig.Alarm)

	if err := s.serviceConfig.Alarm.Validate(); err != nil {
		return fmt.Errorf("'Alarm' custom configuration validation failed: %s", err.Error())
	}

	if err := ds.ListenForCustomConfigChanges(
		&s.serviceConfig.Alarm.Writable,
		"Alarm/Writable", s.ProcessCustomConfigChanges); err != nil {
		return fmt.Errorf("unable to listen for changes for 'Alarm.Writable' custom configuration: %s", err.Error())
	}

	s.alertPin = rpio.Pin(s.serviceConfig.Alarm.AlertPin)
	s.alertPin.Input()

	s.alarmPin = rpio.Pin(s.serviceConfig.Alarm.AlarmPin)
	s.alarmPin.Output()

	go s.listen()

	return nil
}

func (s *RPIOAlarm) listen() {
	for {
		if alertState := s.alertPin.Read(); alertState == rpio.High {
			cv, err := sdkModels.NewCommandValue("Alert", common.ValueTypeBool, true)

			if err != nil {
				s.lc.Errorf("failed to create command values: %s", err.Error())
			} else {
				s.asyncCh <- &sdkModels.AsyncValues{
					DeviceName: s.serviceConfig.Alarm.Name,
					SourceName: fmt.Sprintf("%s-%v", s.serviceConfig.Alarm.Name, s.alertPin),
					CommandValues: []*sdkModels.CommandValue{
						cv,
					},
				}

				if s.serviceConfig.Alarm.RequireAck {
					go s.triggerAlarm()
				}
				s.alertUntil = time.Now().Add(s.serviceConfig.Alarm.Writable.AlarmDuration)
				time.Sleep(s.serviceConfig.Alarm.Writable.AlarmDuration)
			}
		}
	}
}

func (s *RPIOAlarm) triggerAlarm() {
	sendMorse(sos, s.alarmPin)
}

// ProcessCustomConfigChanges ...
func (s *RPIOAlarm) ProcessCustomConfigChanges(rawWritableConfig interface{}) {
	updated, ok := rawWritableConfig.(*config.AlarmWritable)
	if !ok {
		s.lc.Error("unable to process custom config updates: Can not cast raw config to type 'AlarmWritablee'")
		return
	}

	s.lc.Info("Received configuration updates for 'Alarm.Writable' section")

	previous := s.serviceConfig.Alarm.Writable

	if reflect.DeepEqual(previous, *updated) {
		s.lc.Info("No changes detected")
		return
	}

	// Now check to determine what changed.
	// In this example we only have the one writable setting,
	// so the check is not really need but left here as an example.
	// Since this setting is pulled from configuration each time it is need, no extra processing is required.
	// This may not be true for all settings, such as external host connection info, which
	// may require re-establishing the connection to the external host for example.
	if previous.AlarmDuration != updated.AlarmDuration {
		s.lc.Infof("AlarmDuration changed to: %d", updated.AlarmDuration)
	}

	s.serviceConfig.Alarm.Writable = *updated
}

// HandleReadCommands triggers a protocol Read operation for the specified device.
func (s *RPIOAlarm) HandleReadCommands(deviceName string, protocols map[string]models.ProtocolProperties, reqs []sdkModels.CommandRequest) (res []*sdkModels.CommandValue, err error) {
	s.lc.Debugf("RPIOAlarm.HandleReadCommands: protocols: %v resource: %v attributes: %v", protocols, reqs[0].DeviceResourceName, reqs[0].Attributes)

	if len(reqs) == 1 {
		res = make([]*sdkModels.CommandValue, 1)
		if reqs[0].DeviceResourceName == "Alert" {
			cv, _ := sdkModels.NewCommandValue(reqs[0].DeviceResourceName, common.ValueTypeBool, time.Now().Before(s.alertUntil))
			res[0] = cv
		}
	}

	return
}

// HandleWriteCommands passes a slice of CommandRequest struct each representing
// a ResourceOperation for a specific device resource.
// Since the commands are actuation commands, params provide parameters for the individual
// command.
func (s *RPIOAlarm) HandleWriteCommands(deviceName string, protocols map[string]models.ProtocolProperties, reqs []sdkModels.CommandRequest,
	params []*sdkModels.CommandValue) error {
	var err error

	for i, r := range reqs {
		s.lc.Debugf("RPIOAlarm.HandleWriteCommands: protocols: %v, resource: %v, parameters: %v, attributes: %v", protocols, reqs[i].DeviceResourceName, params[i], reqs[i].Attributes)
		switch r.DeviceResourceName {
		case "Alert":
			if time.Now().Before(s.alertUntil) {
				go s.triggerAlarm()
			}

		}
	}
	return err
}

// Stop the protocol-specific DS code to shutdown gracefully, or
// if the force parameter is 'true', immediately. The driver is responsible
// for closing any in-use channels, including the channel used to send async
// readings (if supported).
func (s *RPIOAlarm) Stop(force bool) error {
	// Then Logging Client might not be initialized
	if s.lc != nil {
		s.lc.Debug("closing RPIO")
	}
	return rpio.Close()
}

// AddDevice is a callback function that is invoked
// when a new Device associated with this Device Service is added
func (s *RPIOAlarm) AddDevice(deviceName string, protocols map[string]models.ProtocolProperties, adminState models.AdminState) error {
	s.lc.Debugf("a new Device is added: %s", deviceName)
	return nil
}

// UpdateDevice is a callback function that is invoked
// when a Device associated with this Device Service is updated
func (s *RPIOAlarm) UpdateDevice(deviceName string, protocols map[string]models.ProtocolProperties, adminState models.AdminState) error {
	s.lc.Debugf("Device %s is updated", deviceName)
	return nil
}

// RemoveDevice is a callback function that is invoked
// when a Device associated with this Device Service is removed
func (s *RPIOAlarm) RemoveDevice(deviceName string, protocols map[string]models.ProtocolProperties) error {
	s.lc.Debugf("Device %s is removed", deviceName)
	return nil
}

// Discover triggers protocol specific device discovery, which is an asynchronous operation.
// Devices found as part of this discovery operation are written to the channel devices.
func (s *RPIOAlarm) Discover() {

}
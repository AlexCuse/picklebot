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
	"context"
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/edgexfoundry/go-mod-core-contracts/v2/clients/logger"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/common"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/models"
	"github.com/warthog618/gpiod"

	"github.com/alexcuse/picklebot/edgex/device-gpio-alarm/config"
	sdkModels "github.com/edgexfoundry/device-sdk-go/v2/pkg/models"
	"github.com/edgexfoundry/device-sdk-go/v2/pkg/service"
)

type Alarm struct {
	lc            logger.LoggingClient
	asyncCh       chan<- *sdkModels.AsyncValues
	deviceCh      chan<- []sdkModels.DiscoveredDevice
	serviceConfig *config.ServiceConfig
	alertUntil    time.Time
	onClose       func() error
	alarms        map[string]*alarmPin
}

type alarmEvent struct {
	level string
	when  time.Time
}

type alarmPin struct {
	ch  chan alarmEvent
	pin int
}

func (s *Alarm) trigger(level string) {
	ap, found := s.alarms[level]

	if found {
		ap.ch <- alarmEvent{level: level, when: time.Now().UTC()}
	}
}

// Initialize performs protocol-specific initialization for the device
// service.
func (s *Alarm) Initialize(lc logger.LoggingClient, asyncCh chan<- *sdkModels.AsyncValues, deviceCh chan<- []sdkModels.DiscoveredDevice) error {
	s.lc = lc
	s.asyncCh = asyncCh
	s.deviceCh = deviceCh
	s.serviceConfig = &config.ServiceConfig{}
	s.alarms = make(map[string]*alarmPin)

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

	for k, v := range s.serviceConfig.Alarm.Alarms {
		lc.Infof("Setting up GPIO alarm at alert: %v level: %v alarm: %v (default message: %q)", s.serviceConfig.Alarm.AlertPin, k, v.Pin, v.DefaultMessage)
		s.alarms[k] = &alarmPin{ch: make(chan alarmEvent, 10), pin: v.Pin}
	}

	s.listen()

	return nil
}

func (s *Alarm) listen() {
	listenerCtx, stopListeners := context.WithCancel(context.Background())

	// start listeners for alarm output
	for _, v := range s.alarms {
		go func(a *alarmPin) {
			for {
				select {
				case <-listenerCtx.Done():
					return
				case l := <-a.ch:
					s.runAlarm(l)
				}
			}
		}(v)
	}

	// listen for GPIO sensor input
	l, err := gpiod.RequestLine(s.serviceConfig.Alarm.Chip, s.serviceConfig.Alarm.AlertPin,
		gpiod.WithPullUp,
		gpiod.LineEdgeRising,
		gpiod.WithDebounce(10*time.Millisecond),
		gpiod.WithEventHandler(func(event gpiod.LineEvent) {
			cv, err := sdkModels.NewCommandValue("Alert", common.ValueTypeBool, true)

			if err != nil {
				s.lc.Errorf("failed to create command values: %s", err.Error())
			} else {
				s.asyncCh <- &sdkModels.AsyncValues{
					DeviceName: s.serviceConfig.Alarm.Name,
					SourceName: fmt.Sprintf("%s-%v", s.serviceConfig.Alarm.Name, s.serviceConfig.Alarm.AlertPin),
					CommandValues: []*sdkModels.CommandValue{
						cv,
					},
				}

				s.lc.Info("sent alert to event channel")

				if !s.serviceConfig.Alarm.RequireAck {
					go s.trigger(s.serviceConfig.Alarm.DefaultLevel)
				}
				s.alertUntil = time.Now().Add(s.serviceConfig.Alarm.Writable.AlarmDuration)
			}
		}))
	if err != nil {
		s.lc.Errorf("failed to read GPIO %d: %s", s.serviceConfig.Alarm.AlertPin, err.Error())
	}

	s.onClose = func() error {
		// close listener context
		stopListeners()
		return l.Close()
	}
}

func (s *Alarm) runAlarm(event alarmEvent) {
	lev := strings.ToLower(event.level)
	alarm, found := s.serviceConfig.Alarm.Alarms[lev]

	if !found {
		if s.serviceConfig.Alarm.DefaultLevel != "" {
			s.lc.Warnf("alarm requested for invalid event %q using system default %q", lev, s.serviceConfig.Alarm.DefaultLevel)
			lev = s.serviceConfig.Alarm.DefaultLevel
		} else {
			s.lc.Errorf("alarm requested for invalid event %q using system default %q", lev, s.serviceConfig.Alarm.DefaultLevel)
			return
		}
	}

	if alarm.Pin == 0 {
		s.lc.Infof("Alarm triggered for event %q but no output pin configured", lev)
		return
	}

	line, err := gpiod.RequestLine(s.serviceConfig.Alarm.Chip, alarm.Pin, gpiod.AsOutput(0))

	if err != nil {
		s.lc.Errorf("failed to initialize open line on GPIO %v for %q: %s", alarm.Pin, event, err.Error())
		return
	}

	defer func() {
		line.Reconfigure(gpiod.AsInput)
		line.Close()
	}()

	if err := sendMorse(alarm.DefaultMessage, line); err != nil {
		s.lc.Errorf("failed to send alarm to %d for %q: %s", alarm.Pin, event, err.Error())
	}

	s.lc.Infof("Alarm wait %q", time.Since(event.when).String())
}

// ProcessCustomConfigChanges ...
func (s *Alarm) ProcessCustomConfigChanges(rawWritableConfig interface{}) {
	updated, ok := rawWritableConfig.(*config.AlarmWritable)
	if !ok {
		s.lc.Error("unable to process custom config updates: Can not cast raw config to type 'AlarmWritable'")
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
func (s *Alarm) HandleReadCommands(deviceName string, protocols map[string]models.ProtocolProperties, reqs []sdkModels.CommandRequest) (res []*sdkModels.CommandValue, err error) {
	s.lc.Debugf("Alarm.HandleReadCommands: protocols: %v resource: %v attributes: %v", protocols, reqs[0].DeviceResourceName, reqs[0].Attributes)

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
func (s *Alarm) HandleWriteCommands(deviceName string, protocols map[string]models.ProtocolProperties, reqs []sdkModels.CommandRequest,
	params []*sdkModels.CommandValue) error {
	for i, r := range reqs {
		s.lc.Debugf("Alarm.HandleWriteCommands: protocols: %v, resource: %v, parameters: %+v, attributes: %+v", protocols, reqs[i].DeviceResourceName, params[i], reqs[i].Attributes)

		switch r.DeviceResourceName {
		case "Level":
			level, err := params[i].StringValue()
			if err != nil {
				return err
			}
			s.lc.Infof("received %s update for %s -> %q", r.DeviceResourceName, deviceName, level)
			if time.Now().Before(s.alertUntil) {
				go s.trigger(level)
			}
		}
	}
	return nil
}

// Stop the protocol-specific DS code to shutdown gracefully, or
// if the force parameter is 'true', immediately. The driver is responsible
// for closing any in-use channels, including the channel used to send async
// readings (if supported).
func (s *Alarm) Stop(force bool) error {
	// Then Logging Client might not be initialized
	if s.lc != nil {
		s.lc.Debug("closing RPIO")
	}
	return s.onClose()
}

// AddDevice is a callback function that is invoked
// when a new Device associated with this Device Service is added
func (s *Alarm) AddDevice(deviceName string, protocols map[string]models.ProtocolProperties, adminState models.AdminState) error {
	s.lc.Debugf("a new Device is added: %s", deviceName)
	return nil
}

// UpdateDevice is a callback function that is invoked
// when a Device associated with this Device Service is updated
func (s *Alarm) UpdateDevice(deviceName string, protocols map[string]models.ProtocolProperties, adminState models.AdminState) error {
	s.lc.Debugf("Device %s is updated", deviceName)
	return nil
}

// RemoveDevice is a callback function that is invoked
// when a Device associated with this Device Service is removed
func (s *Alarm) RemoveDevice(deviceName string, protocols map[string]models.ProtocolProperties) error {
	s.lc.Debugf("Device %s is removed", deviceName)
	return nil
}

// Discover triggers protocol specific device discovery, which is an asynchronous operation.
// Devices found as part of this discovery operation are written to the channel devices.
func (s *Alarm) Discover() {

}

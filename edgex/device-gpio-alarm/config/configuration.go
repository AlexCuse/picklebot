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

package config

import (
	"time"
)

type ServiceConfig struct {
	Alarm AlarmConfig
}

// AlarmConfig is example of service's custom structured configuration that is specified in the service's
// configuration.yaml file and Configuration Provider (aka Consul), if enabled.
type AlarmConfig struct {
	Writable AlarmWritable
	AlertPin int
	// Alarms maps GPIO pins to their default messages
	// which will be sent to the pin in morse code by default
	Alarms       map[string]Alarm
	Chip         string
	Name         string
	RequireAck   bool
	Mode         string
	DefaultLevel string
}

type AlarmWritable struct {
	AlarmDuration time.Duration
}

type Alarm struct {
	Pin            int
	DefaultMessage string
}

// UpdateFromRaw updates the service's full configuration from raw data received from
// the Service Provider.
func (sw *ServiceConfig) UpdateFromRaw(rawConfig interface{}) bool {
	configuration, ok := rawConfig.(*ServiceConfig)
	if !ok {
		return false
	}

	if err := configuration.Alarm.Validate(); err != nil {
		return false
	}
	*sw = *configuration

	return true
}

// Validate ensures your custom configuration has proper values.
// Example of validating the sample custom configuration
func (scc *AlarmConfig) Validate() error {

	if scc.Writable.AlarmDuration == 0 {
		scc.Writable.AlarmDuration = 2 * time.Second
	}

	return nil
}

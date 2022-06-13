// TODO: Change Copyright to your company if open sourcing or remove header
//
// Copyright (c) 2021 Intel Corporation
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

// This file contains example of custom configuration that can be loaded from the service's configuration.toml
// and/or the Configuration Provider, aka Consul (if enabled).
// For more details see https://docs.edgexfoundry.org/2.0/microservices/application/GeneralAppServiceConfig/#custom-configuration
// TODO: Update this configuration as needed for you service's needs and remove this comment
//       or remove this file if not using custom configuration.

// TODO: Define your structured custom configuration types. Must be wrapped with an outer struct with
//       single element that matches the top level custom configuration element in your configuration.toml file,
//       'Surveillance' in this example. Replace this example with your configuration structure or
//       remove this file if not using structured custom configuration.
type ServiceConfig struct {
	Surveillance SurveillanceConfig
}

// SurveillanceConfig is example of service's custom structured configuration that is specified in the service's
// configuration.toml file and Configuration Provider (aka Consul), if enabled.
type SurveillanceConfig struct {
	CameraName          string
	SnapshotCommandName string
}

// HostInfo is example struct for defining connection information for external service
type HostInfo struct {
	Host     string
	Port     int
	Protocol string
}

// TODO: Update using your Custom configuration type.
// UpdateFromRaw updates the service's full configuration from raw data received from
// the Service Provider.
func (c *ServiceConfig) UpdateFromRaw(rawConfig interface{}) bool {
	configuration, ok := rawConfig.(*ServiceConfig)
	if !ok {
		return false //errors.New("unable to cast raw config to type 'ServiceConfig'")
	}

	*c = *configuration

	return true
}

// Validate ensures your custom configuration has proper values.
func (ac *SurveillanceConfig) Validate() error {
	return nil
}

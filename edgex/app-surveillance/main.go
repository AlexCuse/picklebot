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

package main

import (
	"os"
	"reflect"

	"app-surveillance/config"
	"app-surveillance/functions"
	"github.com/edgexfoundry/app-functions-sdk-go/v2/pkg"
	"github.com/edgexfoundry/app-functions-sdk-go/v2/pkg/interfaces"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/clients/logger"
)

const (
	serviceKey = "app-surveillance"
)

type myApp struct {
	service       interfaces.ApplicationService
	lc            logger.LoggingClient
	serviceConfig *config.ServiceConfig
	configChanged chan bool
}

func main() {
	app := myApp{}
	code := app.CreateAndRunAppService(serviceKey, pkg.NewAppService)
	os.Exit(code)
}

// CreateAndRunAppService wraps what would normally be in main() so that it can be unit tested
func (app *myApp) CreateAndRunAppService(serviceKey string, newServiceFactory func(string) (interfaces.ApplicationService, bool)) int {
	var ok bool
	app.service, ok = newServiceFactory(serviceKey)
	if !ok {
		return -1
	}

	app.lc = app.service.LoggingClient()

	// More advance custom structured configuration can be defined and loaded as in this example.
	// For more details see https://docs.edgexfoundry.org/2.0/microservices/application/GeneralAppServiceConfig/#custom-configuration
	app.serviceConfig = &config.ServiceConfig{}
	if err := app.service.LoadCustomConfig(app.serviceConfig, "Surveillance"); err != nil {
		app.lc.Errorf("failed load custom configuration: %s", err.Error())
		return -1
	}

	// Optionally validate the custom configuration after it is loaded.
	// TODO: remove if you don't have custom configuration or don't need to validate it
	if err := app.serviceConfig.Surveillance.Validate(); err != nil {
		app.lc.Errorf("custom configuration failed validation: %s", err.Error())
		return -1
	}

	// Custom configuration can be 'writable' or a section of the configuration can be 'writable' when using
	// the Configuration Provider, aka Consul.
	// For more details see https://docs.edgexfoundry.org/2.0/microservices/application/GeneralAppServiceConfig/#writable-custom-configuration
	// TODO: Remove if not using writable custom configuration
	if err := app.service.ListenForCustomConfigChanges(&app.serviceConfig.Surveillance, "Surveillance", app.ProcessConfigUpdates); err != nil {
		app.lc.Errorf("unable to watch custom writable configuration: %s", err.Error())
		return -1
	}

	sample := functions.NewSample(app.serviceConfig.Surveillance.CameraName, app.serviceConfig.Surveillance.SnapshotCommandName)

	var err error

	err = app.service.AddFunctionsPipelineForTopics("gpio-alarms", []string{"edgex/events/device/gpio-alarm/#"},
		sample.LogEventDetails,
		sample.CaptureSnapshot)
	if err != nil {
		app.lc.Errorf("AddFunctionsPipelineForTopic returned error: %s", err.Error())
		return -1
	}

	if err := app.service.MakeItRun(); err != nil {
		app.lc.Errorf("MakeItRun returned error: %s", err.Error())
		return -1
	}

	return 0
}

// ProcessConfigUpdates processes the updated configuration for the service's writable configuration.
// At a minimum it must copy the updated configuration into the service's current configuration. Then it can
// do any special processing for changes that require more.
func (app *myApp) ProcessConfigUpdates(rawWritableConfig interface{}) {
	updated, ok := rawWritableConfig.(*config.SurveillanceConfig)
	if !ok {
		app.lc.Error("unable to process config updates: Can not cast raw config to type 'SurveillanceConfig'")
		return
	}

	previous := app.serviceConfig.Surveillance
	app.serviceConfig.Surveillance = *updated

	if reflect.DeepEqual(previous, updated) {
		app.lc.Info("No changes detected")
		return
	}

	if previous.CameraName != updated.CameraName {
		app.lc.Infof("Surveillance.CameraName changed to: %d", updated.CameraName)
	}
	if previous.SnapshotCommandName != updated.SnapshotCommandName {
		app.lc.Infof("Surveillance.SnapshotCommandName changed to: %s", updated.SnapshotCommandName)
	}
}

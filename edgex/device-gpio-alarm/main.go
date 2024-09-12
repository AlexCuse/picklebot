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
	"github.com/alexcuse/picklebot/edgex/device-gpio-alarm/driver"
	"github.com/edgexfoundry/device-sdk-go/v3"
	"github.com/edgexfoundry/device-sdk-go/v3/pkg/startup"
)

const (
	serviceName string = "device-gpio-alarm"
)

func main() {
	sd := driver.Alarm{}
	startup.Bootstrap(serviceName, device.Version, &sd)
}

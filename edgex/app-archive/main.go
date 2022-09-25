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
	"bytes"
	"fmt"
	"image/jpeg"
	"image/png"
	"os"
	"strings"

	"github.com/edgexfoundry/app-functions-sdk-go/v2/pkg"
	"github.com/edgexfoundry/app-functions-sdk-go/v2/pkg/interfaces"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/dtos"
)

var (
	serviceKey        = "app-archive"
	basePath          = "/data"
	imageResourceName = "Snapshot"
)

func main() {
	// turn off secure mode for examples. Not recommended for production
	_ = os.Setenv("EDGEX_SECURITY_SECRET_STORE", "false")

	// First thing to do is to create an instance of the EdgeX SDK Service, which also runs the bootstrap initialization.
	service, ok := pkg.NewAppService(serviceKey)
	if !ok {
		os.Exit(-1)
	}

	if configuredPath, err := service.GetAppSetting("BasePath"); err != nil {
		basePath = configuredPath
	}

	if configuredResource, err := service.GetAppSetting("ImageResourceName"); err != nil {
		imageResourceName = configuredResource
	}

	var err error

	//use this to process using default pipeline only
	err = service.SetDefaultFunctionsPipeline(savePNG)
	if err != nil {
		service.LoggingClient().Errorf("SetDefaultFunctionsPipeline returned error: %s", err.Error())
		os.Exit(-1)
	}

	// Lastly, we'll go ahead and tell the SDK to "start" and begin listening for events
	err = service.MakeItRun()
	if err != nil {
		service.LoggingClient().Error("MakeItRun returned error: ", err.Error())
		os.Exit(-1)
	}

	service.LoggingClient().Info("Exiting service")
	// Do any required cleanup here
	os.Exit(0)
}

func savePNG(ctx interfaces.AppFunctionContext, data interface{}) (bool, interface{}) {
	lc := ctx.LoggingClient()

	evt, ok := data.(dtos.Event)

	if !ok {
		return false, fmt.Errorf("expected event got %T", data)
	}

	level := "DEFAULT"
	var (
		imageBytes []byte
		timestamp  int64
		cameraName string
	)

	lowerImageName := strings.ToLower(imageResourceName)

	for _, r := range evt.Readings {
		switch strings.ToLower(r.ResourceName) {
		case lowerImageName:
			imageBytes = r.BinaryValue
			cameraName = r.DeviceName
			timestamp = r.Origin
		case "level":
			level = r.Value
		}
	}

	if imageBytes == nil {
		return false, fmt.Errorf("no snapshot received for archive")
	}

	jpg, err := jpeg.Decode(bytes.NewReader(imageBytes))
	if err != nil {
		return false, fmt.Errorf("failed to decode JPG image %w", err)
	}
	buf := &bytes.Buffer{}
	err = png.Encode(buf, jpg)

	if err != nil {
		return false, fmt.Errorf("failed to decode JPG image %w", err)
	}

	savePath := fmt.Sprintf("%s/%s_%d_%s.png", basePath, cameraName, timestamp, level)

	err = os.WriteFile(savePath, buf.Bytes(), 0644)

	if err != nil {
		return false, fmt.Errorf("failed to write file %q %w", savePath, err)
	}

	lc.Infof("stored snapshot at %s", savePath)

	return true, data
}

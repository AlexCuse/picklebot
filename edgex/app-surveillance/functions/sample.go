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

package functions

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/edgexfoundry/app-functions-sdk-go/v2/pkg/interfaces"

	"github.com/edgexfoundry/go-mod-core-contracts/v2/common"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/dtos"
)

// NewSample ...
func NewSample(cameraName, snapshotCommandName string) Sample {
	return Sample{cameraName: cameraName, snapshotCommandName: snapshotCommandName, ackCommand: "Alert"}
}

// Sample ...
type Sample struct {
	cameraName          string
	snapshotCommandName string
	ackCommand          string
}

// LogEventDetails is example of processing an Event and passing the original Event to next function in the pipeline
// For more details on the Context API got here: https://docs.edgexfoundry.org/1.3/microservices/application/ContextAPI/
func (s *Sample) LogEventDetails(ctx interfaces.AppFunctionContext, data interface{}) (bool, interface{}) {
	lc := ctx.LoggingClient()
	lc.Debugf("LogEventDetails called in pipeline '%s'", ctx.PipelineId())

	if data == nil {
		// Go here for details on Error Handle: https://docs.edgexfoundry.org/1.3/microservices/application/ErrorHandling/
		return false, fmt.Errorf("function LogEventDetails in pipeline '%s': No Data Received", ctx.PipelineId())
	}

	event, ok := data.(dtos.Event)
	if !ok {
		return false, fmt.Errorf("function LogEventDetails in pipeline '%s', type received is not an Event", ctx.PipelineId())
	}

	lc.Infof("Event received in pipeline '%s': ID=%s, Device=%s, and ReadingCount=%d",
		ctx.PipelineId(),
		event.Id,
		event.DeviceName,
		len(event.Readings))
	for index, reading := range event.Readings {
		switch strings.ToLower(reading.ValueType) {
		case strings.ToLower(common.ValueTypeBinary):
			lc.Infof(
				"Reading #%d received in pipeline '%s' with ID=%s, Resource=%s, ValueType=%s, MediaType=%s and BinaryValue of size=`%d`",
				index+1,
				ctx.PipelineId(),
				reading.Id,
				reading.ResourceName,
				reading.ValueType,
				reading.MediaType,
				len(reading.BinaryValue))
		default:
			lc.Infof("Reading #%d received in pipeline '%s' with ID=%s, Resource=%s, ValueType=%s, Value=`%s`",
				index+1,
				ctx.PipelineId(),
				reading.Id,
				reading.ResourceName,
				reading.ValueType,
				reading.Value)
		}
	}

	// Returning true indicates that the pipeline execution should continue with the next function
	// using the event passed as input in this case.
	return true, event
}

func (s *Sample) CaptureSnapshot(cxt interfaces.AppFunctionContext, data interface{}) (bool, interface{}) {
	cc := cxt.CommandClient()

	evt := data.(dtos.Event)

	er, err := cc.IssueGetCommandByName(context.Background(), s.cameraName, s.snapshotCommandName, "no", "yes")

	if err != nil {
		return false, err
	}

	// add readings from snapshot to the event
	evt.Readings = append(evt.Readings, er.Event.Readings...)

	return true, evt
}

func (s *Sample) SendAck(cxt interfaces.AppFunctionContext, data interface{}) (bool, interface{}) {
	cc := cxt.CommandClient()

	evt := data.(dtos.Event)

	er, err := cc.IssueSetCommandByName(context.Background(), evt.DeviceName, s.ackCommand, nil)

	if err != nil {
		return false, err
	}

	if er.StatusCode != http.StatusOK {
		return false, fmt.Errorf("unexpected response from %s for %s", s.ackCommand, evt.DeviceName)
	}

	return true, evt
}

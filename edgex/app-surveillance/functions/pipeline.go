package functions

import (
	"fmt"
	"strings"

	"github.com/edgexfoundry/app-functions-sdk-go/v3/pkg/interfaces"
	"github.com/edgexfoundry/go-mod-core-contracts/v3/common"
	"github.com/edgexfoundry/go-mod-core-contracts/v3/dtos"
)

// NewPipeline ...
func NewPipeline(cameraName, snapshotCommandName string) Pipeline {
	return Pipeline{cameraName: cameraName, snapshotCommandName: snapshotCommandName, ackCommand: "Acknowledge"}
}

// Pipeline ...
type Pipeline struct {
	cameraName          string
	snapshotCommandName string
	ackCommand          string
}

// LogEventDetails is example of processing an Event and passing the original Event to next function in the pipeline
// For more details on the Context API got here: https://docs.edgexfoundry.org/1.3/microservices/application/ContextAPI/
func (s *Pipeline) LogEventDetails(ctx interfaces.AppFunctionContext, data interface{}) (bool, interface{}) {
	lc := ctx.LoggingClient()
	receiveTopic, _ := ctx.GetValue(interfaces.RECEIVEDTOPIC)

	lc.Infof("LogEventDetails called in pipeline '%s' on '%s'", ctx.PipelineId(), receiveTopic)

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

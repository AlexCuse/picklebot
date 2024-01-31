package functions

import (
	"context"

	"github.com/edgexfoundry/app-functions-sdk-go/v3/pkg/interfaces"
	"github.com/edgexfoundry/go-mod-core-contracts/v3/dtos"
)

func (s *Pipeline) Record(ctx interfaces.AppFunctionContext, data interface{}) (bool, interface{}) {
	cc := ctx.CommandClient()

	evt := data.(dtos.Event)

	if s.cameraName != "" {
		er, err := cc.IssueGetCommandByName(context.Background(), s.cameraName, s.snapshotCommandName, false, true)

		if err != nil {
			return false, err
		}

		// add readings from snapshot to the event
		evt.Readings = append(evt.Readings, er.Event.Readings...)
	}

	return true, evt
}

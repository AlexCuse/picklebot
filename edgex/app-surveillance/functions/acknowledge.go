package functions

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/edgexfoundry/app-functions-sdk-go/v3/pkg/interfaces"
	"github.com/edgexfoundry/go-mod-core-contracts/v3/dtos"
)

func (s *Pipeline) Acknowledge(ctx interfaces.AppFunctionContext, data interface{}) (bool, interface{}) {
	cc := ctx.CommandClient()

	evt := data.(dtos.Event)

	for _, r := range evt.Readings {
		if r.ResourceName == "Level" {
			er, err := cc.IssueSetCommandByName(context.Background(), evt.DeviceName, s.ackCommand, map[string]string{"Level": r.SimpleReading.Value})

			if err != nil {
				return false, err
			}

			if er.StatusCode != http.StatusOK {
				return false, fmt.Errorf("unexpected response from %s for %s", s.ackCommand, evt.DeviceName)
			}
		}
	}

	response, _ := json.Marshal(evt)

	ctx.SetResponseData(response)

	return true, evt
}

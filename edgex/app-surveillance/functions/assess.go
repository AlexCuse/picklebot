package functions

import (
	"math/rand"

	"github.com/edgexfoundry/app-functions-sdk-go/v3/pkg/interfaces"
	"github.com/edgexfoundry/go-mod-core-contracts/v3/common"
	"github.com/edgexfoundry/go-mod-core-contracts/v3/dtos"
)

func (s *Pipeline) Assess(ctx interfaces.AppFunctionContext, data interface{}) (bool, interface{}) {
	evt := data.(dtos.Event)

	lvls := []string{"low", "default", "high"}
	lvlIdx := rand.Int31n(3)

	lvlReading, err := dtos.NewSimpleReading(evt.ProfileName, evt.DeviceName, "Level", common.ValueTypeString, lvls[lvlIdx])

	if err != nil {
		return false, err
	}

	evt.Readings = append(evt.Readings, lvlReading)

	return true, evt
}

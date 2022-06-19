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

package functions

import (
	"context"

	"github.com/edgexfoundry/app-functions-sdk-go/v2/pkg/interfaces"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/dtos"
)

func (s *Sample) Record(ctx interfaces.AppFunctionContext, data interface{}) (bool, interface{}) {
	cc := ctx.CommandClient()

	evt := data.(dtos.Event)

	if s.cameraName != "" {
		er, err := cc.IssueGetCommandByName(context.Background(), s.cameraName, s.snapshotCommandName, "no", "yes")

		if err != nil {
			return false, err
		}

		// add readings from snapshot to the event
		evt.Readings = append(evt.Readings, er.Event.Readings...)
	}

	return true, evt
}

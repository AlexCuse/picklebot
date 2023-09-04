package functions

import (
	"testing"

	"github.com/edgexfoundry/app-functions-sdk-go/v3/pkg"
	"github.com/edgexfoundry/app-functions-sdk-go/v3/pkg/interfaces"
	"github.com/edgexfoundry/go-mod-core-contracts/v3/clients/logger"
	"github.com/edgexfoundry/go-mod-core-contracts/v3/common"
	"github.com/edgexfoundry/go-mod-core-contracts/v3/dtos"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// This file contains example of how to unit test pipeline functions
// TODO: Change these sample unit tests to test your custom type and function(s)

var appContext interfaces.AppFunctionContext

func TestMain(m *testing.M) {
	//
	// This can be changed to a real logger when needing more debug information output to the console
	// lc := logger.NewClient("testing", "DEBUG")
	//
	lc := logger.NewMockClient()
	correlationId := uuid.New().String()

	// NewAppFuncContextForTest creates a context with basic dependencies for unit testing with the passed in logger
	// If more additional dependencies (such as mock clients) are required, then use
	// NewAppFuncContext(correlationID string, dic *di.Container) and pass in an initialized DIC (dependency injection container)
	appContext = pkg.NewAppFuncContextForTest(correlationId, lc)
}

func TestPipeline_LogEventDetails(t *testing.T) {
	expectedEvent := createTestEvent(t)
	expectedContinuePipeline := true

	target := NewPipeline("", "")
	actualContinuePipeline, actualEvent := target.LogEventDetails(appContext, expectedEvent)

	assert.Equal(t, expectedContinuePipeline, actualContinuePipeline)
	assert.Equal(t, expectedEvent, actualEvent)
}

func createTestEvent(t *testing.T) dtos.Event {
	profileName := "MyProfile"
	deviceName := "MyDevice"
	sourceName := "MySource"
	resourceName := "MyResource"

	event := dtos.NewEvent(profileName, deviceName, sourceName)
	err := event.AddSimpleReading(resourceName, common.ValueTypeInt32, int32(1234))
	require.NoError(t, err)

	event.Tags = map[string]interface{}{
		"WhereAmI": "NotKansas",
	}

	return event
}

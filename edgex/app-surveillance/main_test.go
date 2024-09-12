package main

import (
	"fmt"
	"testing"

	"github.com/edgexfoundry/go-mod-core-contracts/v3/clients/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/edgexfoundry/app-functions-sdk-go/v3/pkg/interfaces"
	"github.com/edgexfoundry/app-functions-sdk-go/v3/pkg/interfaces/mocks"
)

// This is an example of how to test the code that would typically be in the main() function use mocks
// Not to helpful for a simple main() , but can be if the main() has more complexity that should be unit tested
// TODO: add/update tests for your customized CreateAndRunAppService or remove if your main code doesn't require unit testing.

func TestCreateAndRunService_Success(t *testing.T) {
	app := myApp{}

	mockFactory := func(_ string) (interfaces.ApplicationService, bool) {
		mockAppService := &mocks.ApplicationService{}
		mockAppService.On("LoggingClient").Return(logger.NewMockClient())
		mockAppService.On("GetAppSettingStrings", "DeviceNames").
			Return([]string{"Random-Boolean-Device, Random-Integer-Device"}, nil)
		mockAppService.On("SetDefaultFunctionsPipeline", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
			Return(nil)
		mockAppService.On("LoadCustomConfig", mock.Anything, mock.Anything, mock.Anything).
			Return(nil).Run(func(args mock.Arguments) {
		})
		mockAppService.On("ListenForCustomConfigChanges", mock.Anything, mock.Anything, mock.Anything).
			Return(nil)
		mockAppService.On("Run").Return(nil)

		return mockAppService, true
	}

	expected := 0
	actual := app.CreateAndRunAppService("TestKey", mockFactory)
	assert.Equal(t, expected, actual)
}

func TestCreateAndRunService_NewService_Failed(t *testing.T) {
	app := myApp{}

	mockFactory := func(_ string) (interfaces.ApplicationService, bool) {
		return nil, false
	}
	expected := -1
	actual := app.CreateAndRunAppService("TestKey", mockFactory)
	assert.Equal(t, expected, actual)
}

func TestCreateAndRunService_SetFunctionsPipeline_Failed(t *testing.T) {
	app := myApp{}

	// ensure failure is from SetFunctionsPipeline
	setFunctionsPipelineCalled := false

	mockFactory := func(_ string) (interfaces.ApplicationService, bool) {
		mockAppService := &mocks.ApplicationService{}
		mockAppService.On("LoggingClient").Return(logger.NewMockClient())
		mockAppService.On("LoadCustomConfig", mock.Anything, mock.Anything, mock.Anything).
			Return(nil).Run(func(args mock.Arguments) {
		})
		mockAppService.On("ListenForCustomConfigChanges", mock.Anything, mock.Anything, mock.Anything).
			Return(nil)
		mockAppService.On("SetDefaultFunctionsPipeline", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
			Return(fmt.Errorf("Failed")).Run(func(args mock.Arguments) {
			setFunctionsPipelineCalled = true
		})

		return mockAppService, true
	}

	expected := -1
	actual := app.CreateAndRunAppService("TestKey", mockFactory)
	require.True(t, setFunctionsPipelineCalled, "SetFunctionsPipeline never called")
	assert.Equal(t, expected, actual)
}

func TestCreateAndRunService_Run_Failed(t *testing.T) {
	app := myApp{}

	// ensure failure is from MakeItRun
	runCalled := false

	mockFactory := func(_ string) (interfaces.ApplicationService, bool) {
		mockAppService := &mocks.ApplicationService{}
		mockAppService.On("LoggingClient").Return(logger.NewMockClient())
		mockAppService.On("LoadCustomConfig", mock.Anything, mock.Anything, mock.Anything).
			Return(nil).Run(func(args mock.Arguments) {
		})
		mockAppService.On("ListenForCustomConfigChanges", mock.Anything, mock.Anything, mock.Anything).
			Return(nil)
		mockAppService.On("SetDefaultFunctionsPipeline", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
			Return(nil)
		mockAppService.On("Run").Return(fmt.Errorf("Failed")).Run(func(args mock.Arguments) {
			runCalled = true
		})

		return mockAppService, true
	}

	expected := -1
	actual := app.CreateAndRunAppService("TestKey", mockFactory)
	require.True(t, runCalled, "MakeItRun never called")
	assert.Equal(t, expected, actual)
}

package main

import (
	"strings"
	"testing"

	"github.com/mattermost/mattermost-server/model"

	"github.com/stretchr/testify/assert"

	"github.com/mattermost/mattermost-server/plugin/plugintest"
	"github.com/mattermost/mattermost-server/plugin/plugintest/mock"
)

func TestCallDiceAPI(t *testing.T) {
	resp, err := callDiceAPI("1d5")
	assert.Nil(t, err)
	assert.NotNil(t, resp)
	assert.True(t, resp.Success)
	assert.Equal(t, 1, len(resp.Dice))
	assert.Equal(t, "d5", resp.Dice[0].Type)
}

func TestBadTrigger(t *testing.T) {
	genericWrongInputTestPlugin(t, "/lolzies d20")
}

func TestEmptyRequest(t *testing.T) {
	genericWrongInputTestPlugin(t, "/roll ")
}

func TestBadRequestD0(t *testing.T) {
	genericWrongInputTestPlugin(t, "/roll d0")
}

func TestBadRequestWeirdString(t *testing.T) {
	genericWrongInputTestPlugin(t, "/roll hahaha")
}

func TestBadRequestDiceWithoutNumber(t *testing.T) {
	genericWrongInputTestPlugin(t, "/roll 6d")
}

func TestBadRequest0D5(t *testing.T) {
	genericWrongInputTestPlugin(t, "/roll 0d5")
}

func genericWrongInputTestPlugin(t *testing.T, badInput string) {
	api := initTestPlugin(t)
	p := DiceRollingPlugin{}
	assert.Nil(t, p.OnActivate(api))

	var command *model.CommandArgs
	// Wrong dice requests
	command = &model.CommandArgs{
		Command: badInput,
	}
	response, err := p.ExecuteCommand(command)
	assert.NotNil(t, err)
	assert.Nil(t, response)
}

func TestGoodRequestRoll1(t *testing.T) {
	genericCorrectInputTestPlugin(t, "*rolls a d1:* **1**", "1")
}

func TestGoodRequestRoll5D1(t *testing.T) {
	genericCorrectInputTestPlugin(
		t,
		"*rolls a d1:* **1**\n*rolls a d1:* **1**\n*rolls a d1:* **1**\n*rolls a d1:* **1**\n*rolls a d1:* **1**",
		"5d1")
}

func TestGoodRequestRoll3D1Sum(t *testing.T) {
	genericCorrectInputTestPlugin(t, "*rolls a d1:* **1**\n*rolls a d1:* **1**\n*rolls a d1:* **1**\n**Total = 3**", "3d1 sum")
}

// TODO : how do you test the random result ? by mocking the Dice API I guess

func genericCorrectInputTestPlugin(t *testing.T, expectedText string, inputDiceRequest string) {

	api := initTestPlugin(t)
	p := DiceRollingPlugin{}
	assert.Nil(t, p.OnActivate(api))

	command := &model.CommandArgs{
		Command: "/roll " + inputDiceRequest,
		UserId:  "userid",
	}
	response, err := p.ExecuteCommand(command)
	assert.Nil(t, err)
	assert.NotNil(t, response)
	assert.NotNil(t, response.Attachments)
	assert.Equal(t, 1, len(response.Attachments))
	assert.NotNil(t, response.Attachments[0])
	assert.Equal(t, expectedText, strings.TrimSpace(response.Attachments[0].Text))
}

func initTestPlugin(t *testing.T) *plugintest.API {

	api := &plugintest.API{}
	api.On("RegisterCommand", mock.Anything).Return(nil)
	api.On("UnregisterCommand", mock.Anything, mock.Anything).Return(nil)
	api.On("GetUser", mock.Anything).Return(&model.User{
		Id:       "userid",
		Nickname: "User",
	}, (*model.AppError)(nil))

	return api
}

func TestLifecyclePlugin(t *testing.T) {

	api := initTestPlugin(t)
	p := DiceRollingPlugin{}

	assert.Nil(t, p.OnActivate(api))

	// TODO : test executecommand while deactivated ?

	assert.Nil(t, p.OnDeactivate())

}

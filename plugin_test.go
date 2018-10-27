package main

import (
	"strings"
	"testing"

	"github.com/mattermost/mattermost-server/model"
	"github.com/mattermost/mattermost-server/plugin"

	"github.com/stretchr/testify/assert"

	"github.com/mattermost/mattermost-server/plugin/plugintest"
	"github.com/mattermost/mattermost-server/plugin/plugintest/mock"
)

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
	p := initTestPlugin(t)
	assert.Nil(t, p.OnActivate())

	var command *model.CommandArgs
	// Wrong dice requests
	command = &model.CommandArgs{
		Command: badInput,
	}
	response, err := p.ExecuteCommand(&plugin.Context{}, command)
	assert.NotNil(t, err)
	assert.Nil(t, response)
}

func TestGoodRequestRoll1(t *testing.T) {
	genericCorrectInputTestPlugin(t, "*rolls 1:* **1 **", "1")
}

func TestGoodRequestRoll5D1(t *testing.T) {
	genericCorrectInputTestPlugin(
		t,
		"*rolls 5d1:* **1 1 1 1 1 **",
		"5d1")
}

func TestGoodRequestRoll3D1Sum(t *testing.T) {
	genericCorrectInputTestPlugin(t, "*rolls 3d1:* **1 1 1 **\n**Total = 3**", "3d1 sum")
}

func TestGoodRequestHelp(t *testing.T) {
    p := initTestPlugin(t)
    assert.Nil(t, p.OnActivate())

    command := &model.CommandArgs{
        Command: "/roll help",
        UserId:  "userid",
    }

    response, err := p.ExecuteCommand(&plugin.Context{}, command)
    assert.Nil(t, err)
    assert.NotNil(t, response)
    assert.Nil(t, response.Attachments)
}

// TODO : how do you test the random result ? by mocking the Dice API I guess

func genericCorrectInputTestPlugin(t *testing.T, expectedText string, inputDiceRequest string) {

	p := initTestPlugin(t)
	assert.Nil(t, p.OnActivate())

	command := &model.CommandArgs{
		Command: "/roll " + inputDiceRequest,
		UserId:  "userid",
	}
	response, err := p.ExecuteCommand(&plugin.Context{}, command)
	assert.Nil(t, err)
	assert.NotNil(t, response)
	assert.NotNil(t, response.Attachments)
	assert.Equal(t, 1, len(response.Attachments))
	assert.NotNil(t, response.Attachments[0])
	assert.Equal(t, expectedText, strings.TrimSpace(response.Attachments[0].Text))
}

func initTestPlugin(t *testing.T) *DiceRollingPlugin {

	api := &plugintest.API{}
	api.On("RegisterCommand", mock.Anything).Return(nil)
	api.On("UnregisterCommand", mock.Anything, mock.Anything).Return(nil)
	api.On("GetUser", mock.Anything).Return(&model.User{
		Id:       "userid",
		Nickname: "User",
	}, (*model.AppError)(nil))

	p := DiceRollingPlugin{}
	p.SetAPI(api)

	return &p
}

func TestLifecyclePlugin(t *testing.T) {

	p := initTestPlugin(t)

	assert.Nil(t, p.OnActivate())

	// TODO : test executecommand while deactivated ?

	assert.Nil(t, p.OnDeactivate())

}

package main

import (
	"strings"
	"testing"

	"github.com/mattermost/mattermost-server/v5/model"
	"github.com/mattermost/mattermost-server/v5/plugin"

	"github.com/stretchr/testify/assert"

	"github.com/mattermost/mattermost-server/v5/plugin/plugintest"
	"github.com/mattermost/mattermost-server/v5/plugin/plugintest/mock"
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
	p, _ := initTestPlugin(t)
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
	genericCorrectInputTestPlugin(t, "**User** *rolls 1:* **1 **", "1")
}

func TestGoodRequestRoll5D1(t *testing.T) {
	genericCorrectInputTestPlugin(
		t,
		"**User** *rolls 5d1:* **1 1 1 1 1 **",
		"5d1")
}

func TestGoodRequestRoll3D1Sum(t *testing.T) {
	genericCorrectInputTestPlugin(t, "**User** *rolls 3d1:* **1 1 1 **\n**Total = 3**", "3d1 sum")
}

func TestGoodRequestHelp(t *testing.T) {
	p, _ := initTestPlugin(t)
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

func genericCorrectInputTestPlugin(t *testing.T, expectedText string, inputDiceRequest string) {

	p, api := initTestPlugin(t)
	var post *model.Post
	api.On("CreatePost", mock.AnythingOfType("*model.Post")).Return(nil, nil).Run(func(args mock.Arguments) {
		post = args.Get(0).(*model.Post)
	})
	assert.Nil(t, p.OnActivate())

	command := &model.CommandArgs{
		Command: "/roll " + inputDiceRequest,
		UserId:  "userid",
	}
	response, err := p.ExecuteCommand(&plugin.Context{}, command)
	assert.Nil(t, err)
	assert.NotNil(t, response)
	assert.NotNil(t, post)
	assert.NotNil(t, post.Message)
	assert.Equal(t, expectedText, strings.TrimSpace(post.Message))
}

func initTestPlugin(t *testing.T) (*Plugin, *plugintest.API) {

	api := &plugintest.API{}
	api.On("RegisterCommand", mock.Anything).Return(nil)
	api.On("UnregisterCommand", mock.Anything, mock.Anything).Return(nil)
	api.On("GetUser", mock.Anything).Return(&model.User{
		Id:       "userid",
		Nickname: "User",
	}, (*model.AppError)(nil))

	p := Plugin{}
	p.SetAPI(api)

	return &p, api
}

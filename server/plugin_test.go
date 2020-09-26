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

func TestBadInputs(t *testing.T) {
	p, _ := initTestPlugin(t)
	assert.Nil(t, p.OnActivate())

	testCases := []string{
		"/lolzies d20",
		"/roll ",
		"/roll d0",
		"/roll hahaha",
		"/roll 6d",
		"/roll 0d5",
	}
	for _, testCase := range testCases {
		// Wrong dice requests
		command := &model.CommandArgs{
			Command: testCase,
		}
		response, err := p.ExecuteCommand(&plugin.Context{}, command)

		assert.NotNil(t, err)
		assert.Nil(t, response)
	}
}

func TestGoodInputs(t *testing.T) {
	p, api := initTestPlugin(t)
	var post *model.Post
	api.On("CreatePost", mock.AnythingOfType("*model.Post")).Return(nil, nil).Run(func(args mock.Arguments) {
		post = args.Get(0).(*model.Post)
	})
	assert.Nil(t, p.OnActivate())

	testCases := []struct {
		inputDiceRequest string
		expectedText     string
	}{
		{inputDiceRequest: "3d1 sum", expectedText: "**User** rolls:\n- 3d1: **1 1 1 **\n**Total = 3**"},
		{inputDiceRequest: "5d1", expectedText: "**User** rolls:\n- 5d1: **1 1 1 1 1 **\n**Total = 5**"},
		{inputDiceRequest: "1", expectedText: "**User** rolls:\n- 1: **1 **\n**Total = 1**"},
		{inputDiceRequest: "+42", expectedText: "**User** rolls:\n- **+42 **\n**Total = 42**"},
		{inputDiceRequest: "4d1+3", expectedText: "**User** rolls:\n- 4d1+3: **4 4 4 4 **\n**Total = 16**"},
		{inputDiceRequest: "4d1 +3", expectedText: "**User** rolls:\n- 4d1: **1 1 1 1 **\n- **+3 **\n**Total = 7**"},
	}
	for _, testCase := range testCases {
		command := &model.CommandArgs{
			Command: "/roll " + testCase.inputDiceRequest,
			UserId:  "userid",
		}
		response, err := p.ExecuteCommand(&plugin.Context{}, command)
		testLabel := "Testing " + testCase.inputDiceRequest
		assert.Nil(t, err, testLabel)
		assert.NotNil(t, response, testLabel)
		assert.NotNil(t, post, testLabel)
		assert.NotNil(t, post.Message, testLabel)
		assert.Equal(t, testCase.expectedText, strings.TrimSpace(post.Message), testLabel)
	}
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

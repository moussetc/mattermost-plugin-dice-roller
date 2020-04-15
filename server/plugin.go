package main

import (
	"fmt"
	"math/rand"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/mattermost/mattermost-server/v5/model"
	"github.com/mattermost/mattermost-server/v5/plugin"
)

const (
	trigger      string = "roll"
	diceFilename string = "icon.png"
)

// Plugin implements the interface expected by the Mattermost server to communicate between the server and plugin processes.
type Plugin struct {
	plugin.MattermostPlugin

	// configurationLock synchronizes access to the configuration.
	configurationLock sync.RWMutex

	// configuration is the active plugin configuration. Consult getConfiguration and
	// setConfiguration for usage.
	configuration *configuration

	//BotId of the created bot account for dice rolling
	diceBotId string
}

func (p *Plugin) OnActivate() error {

	rand.Seed(time.Now().UnixNano())

	return p.API.RegisterCommand(&model.Command{
		Trigger:          trigger,
		Description:      "Roll one or more dice",
		DisplayName:      "Dice roller ⚄",
		AutoComplete:     true,
		AutoCompleteDesc: "Roll one or several dice. ⚁ ⚄ Try /roll help for a list of possibilities.",
		AutoCompleteHint: "20 d6 3d4 [sum]",
	})
}

func (p *Plugin) GetHelpMessage() *model.CommandResponse {
	props := map[string]interface{}{
		"from_webhook": "true",
	}

	return &model.CommandResponse{
		ResponseType: model.COMMAND_RESPONSE_TYPE_EPHEMERAL,
		Text: "Here are some examples:\n" +
			"- `/roll 20` to roll a 20-sided die. You can use any number.\n" +
			"- `/roll 5D6` to roll 5 six-sided dice in one go.\n" +
			"- `/roll 5 d8 13D20` to roll different kind of dice all at once.\n" +
			"- Add `sum` at the end to get the sum of all the dice results as well.\n" +
			"- `/roll help` will show this help text.\n\n" +
			" ⚅ ⚂ Let's get rolling! ⚁ ⚄",
		Props: props,
	}
}

// ExecuteCommand returns a post that displays the result of the dice rolls
func (p *Plugin) ExecuteCommand(c *plugin.Context, args *model.CommandArgs) (*model.CommandResponse, *model.AppError) {
	if p.API == nil {
		return nil, appError("Cannot access the plugin API.", nil)
	}

	cmd := "/" + trigger
	if strings.HasPrefix(args.Command, cmd) {
		query := strings.TrimSpace((strings.Replace(args.Command, cmd, "", 1)))

		if query == "help" || query == "--help" || query == "h" || query == "-h" {
			return p.GetHelpMessage(), nil
		}

		post, generatePostError := p.generateDicePost(query, args.UserId, args.ChannelId, args.RootId)
		if generatePostError != nil {
			return nil, generatePostError
		}
		_, createPostError := p.API.CreatePost(post)
		if createPostError != nil {
			return nil, createPostError
		}

		return &model.CommandResponse{}, nil
	}

	return nil, appError("Expected trigger "+cmd+" but got "+args.Command, nil)
}

func (p *Plugin) generateDicePost(query, userId, channelId, rootId string) (*model.Post, *model.AppError) {

	// Get the user to we can display the right name
	user, userErr := p.API.GetUser(userId)
	if userErr != nil {
		return nil, userErr
	}
	displayName := user.Nickname
	if displayName == "" {
		displayName = user.Username
	}

	sumRequest := false
	text := ""
	sum := 0

	rollRequests := strings.Fields(query)
	for _, rollRequest := range rollRequests {
		if rollRequest == "sum" {
			sumRequest = true
		} else {
			result, err := rollDice(rollRequest)
			if err != nil {
				return nil, appError(fmt.Sprintf("%s See `/roll help` for examples.", err.Error()), err)
			}
			formattedResults := ""
			for _, roll := range result.results {
				formattedResults += fmt.Sprintf("%d ", roll)
				sum += roll
			}
			text += fmt.Sprintf("**%s** *rolls %s:* **%s**\n", displayName, rollRequest, formattedResults)
		}
	}

	if len(rollRequests) == 0 || sumRequest && len(rollRequests) == 1 {
		return nil, appError("No roll request arguments found (such as '20', '4d6', etc.).", nil)
	}

	if sumRequest {
		text += fmt.Sprintf("**Total = %d**", sum)
	}

	return &model.Post{
		UserId:    p.diceBotId,
		ChannelId: channelId,
		RootId:    rootId,
		Message:   text,
	}, nil
}

type rollAPIResult struct {
	Success bool `json:"success"`
	Dice    []struct {
		Value int    `json:"value"`
		Type  string `json:"type"`
	} `json:"dice"`
}

func appError(message string, err error) *model.AppError {
	errorMessage := ""
	if err != nil {
		errorMessage = err.Error()
	}
	return model.NewAppError("Dice Roller Plugin", message, nil, errorMessage, http.StatusBadRequest)
}

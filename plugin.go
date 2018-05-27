package main

import (
	"fmt"
	"net/http"
	"strings"
	"sync/atomic"

	"github.com/mattermost/mattermost-server/model"
	"github.com/mattermost/mattermost-server/plugin"
	"github.com/mattermost/mattermost-server/plugin/rpcplugin"
)

// DiceRollingPlugin is a Mattermost plugin that adds a slash command
// to roll dices in-chat
type DiceRollingPlugin struct {
	api           plugin.API
	configuration atomic.Value
	enabled       bool
}

const (
	trigger    string = "roll"
	diceAPIURL string = "http://roll.diceapi.com/json/"
)

// OnActivate register the plugin command
func (p *DiceRollingPlugin) OnActivate(api plugin.API) error {
	p.api = api
	p.enabled = true

	return api.RegisterCommand(&model.Command{
		Trigger:          trigger,
		Description:      "Roll one or more dice",
		DisplayName:      "Dice rolling command",
		AutoComplete:     true,
		AutoCompleteDesc: "Roll one or several dice with the possibility to compute the sum automatically because we are lazy, lazy people",
		AutoCompleteHint: "20 d6 3d4 [sum]",
	})
}

// OnDeactivate handles plugin deactivation
func (p *DiceRollingPlugin) OnDeactivate() error {
	p.enabled = false
	return nil
}

// ExecuteCommand returns a post that displays the result of the dice rolls
func (p *DiceRollingPlugin) ExecuteCommand(args *model.CommandArgs) (*model.CommandResponse, *model.AppError) {
	if !p.enabled {
		return nil, p.error("Cannot execute command while the plugin is disabled.")
	}
	if p.api == nil {
		return nil, p.error("Cannot access the plugin API.")
	}

	cmd := "/" + trigger
	if strings.HasPrefix(args.Command, cmd) {
		query := strings.Replace(args.Command, cmd, "", 1)

		rollRequests := strings.Fields(strings.TrimSpace(query))

		sumRequest := false
		text := ""
		sum := 0
		for _, rollRequest := range rollRequests {
			if rollRequest == "sum" {
				sumRequest = true
			} else {
				result, err := rollDice(rollRequest)
				if err != nil {
					return nil, p.error(fmt.Sprint(err))
				}
				formattedResults := ""
				for _, roll := range result.results {
					formattedResults += fmt.Sprintf("%d ", roll)
					sum += roll
				}
				text += fmt.Sprintf("*rolls %s:* **%s**\n", rollRequest, formattedResults)
			}
		}

		if len(rollRequests) == 0 || sumRequest && len(rollRequests) == 1 {
			return nil, p.error("No roll request arguments found (such as '20', '4d6', etc.).")
		}

		if sumRequest {
			text += fmt.Sprintf("**Total = %d**", sum)
		}

		// Get the user to we can display the right name
		user, userErr := p.api.GetUser(args.UserId)
		if userErr != nil {
			return nil, userErr
		}

		attachments := []*model.SlackAttachment{
			&model.SlackAttachment{
				Text:     text,
				Fallback: fmt.Sprintf("%s rolled some dice!", user.GetFullName()),
				ThumbURL: "http://upload.wikimedia.org/wikipedia/commons/f/f5/Twenty_sided_dice.png",
			},
		}

		props := map[string]interface{}{
			"from_webhook":  "true",
			"use_user_icon": "true",
		}

		return &model.CommandResponse{
			ResponseType: "in_channel",
			// Username:     user.GetFullName(),
			Attachments: attachments,
			Props:       props,
		}, nil
	}

	return nil, p.error("Expected trigger " + cmd + " but got " + args.Command)
}

func (p *DiceRollingPlugin) error(message string) *model.AppError {
	return model.NewAppError("Dice Roller Plugin ExecuteCommand", message, nil, "", http.StatusBadRequest)
}

type rollAPIResult struct {
	Success bool `json:"success"`
	Dice    []struct {
		Value int    `json:"value"`
		Type  string `json:"type"`
	} `json:"dice"`
}

// Install the RCP plugin
func main() {
	rpcplugin.Main(&DiceRollingPlugin{})
}

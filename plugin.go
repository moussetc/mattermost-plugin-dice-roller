package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"regexp"
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
	TeamId        string
}

const (
	trigger    string = "roll"
	diceAPIURL string = "http://roll.diceapi.com/json/"
)

// OnActivate register the plugin command
func (p *DiceRollingPlugin) OnActivate(api plugin.API) error {
	p.api = api
	return api.RegisterCommand(&model.Command{
		Trigger:          trigger,
		TeamId:           p.TeamId,
		Description:      "Roll one or more dice",
		DisplayName:      "Dice rolling command",
		AutoComplete:     true,
		AutoCompleteDesc: "Roll one or several dice with the possibility to compute the sum automatically because we are lazy, lazy people",
		AutoCompleteHint: "20 d6 3d4 [sum]",
	})
}

// OnDeactivate unregisters the plugin command
func (p *DiceRollingPlugin) OnDeactivate() error {
	return p.api.UnregisterCommand(p.TeamId, trigger)
}

// ExecuteCommand returns a post that displays the result of the dice rolls
func (p *DiceRollingPlugin) ExecuteCommand(args *model.CommandArgs) (*model.CommandResponse, *model.AppError) {
	cmd := "/" + trigger
	if strings.HasPrefix(args.Command, cmd) {
		query := strings.Replace(args.Command, cmd, "", 1)

		rollRequests := strings.Fields(strings.TrimSpace(query))

		sumRequest := false
		formatedParams := make([]string, 0)
		for _, rollRequest := range rollRequests {
			if rollRequest == "sum" {
				sumRequest = true
			} else if matched, err := regexp.MatchString("^([1-9]\\d*)?[dD][1-9]\\d*$", rollRequest); err == nil && matched {
				// Correct dice format such ad "4d20"
				formatedParams = append(formatedParams, rollRequest)
			} else if matched, _ := regexp.MatchString("^[1-9]\\d*$", rollRequest); matched {
				// Convert simple number like "20" to "d20"
				formatedParams = append(formatedParams, "d"+rollRequest)
			} else {
				return nil, p.error("The value " + rollRequest + " is not a valid dice roll request")
			}
		}

		resp, err := callDiceAPI(strings.Join(formatedParams, "/"))
		if err != nil {
			return nil, p.error("Could not get rolls from API : " + fmt.Sprint(err))
		}

		text := ""
		if resp.Success {
			var sum int
			for _, die := range resp.Dice {
				text += fmt.Sprintf("*rolls a %s:* **%d**\n", die.Type, die.Value)
				sum += die.Value
			}
			if sumRequest {
				text += fmt.Sprintf("**Total = %d**", sum)
			}
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

func callDiceAPI(rollRequest string) (*rollAPIResult, error) {
	resp := new(rollAPIResult)
	url := diceAPIURL + rollRequest
	r, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer r.Body.Close()
	// fmt.Println(r.Body)
	if r.StatusCode == 200 {
		err = json.NewDecoder(r.Body).Decode(resp)
		if err != nil {
			return nil, err
		}
		return resp, nil
	}
	return nil, errors.New("API returned statusCode: " + string(r.StatusCode))
}

// Install the RCP plugin
func main() {
	rpcplugin.Main(&DiceRollingPlugin{})
}

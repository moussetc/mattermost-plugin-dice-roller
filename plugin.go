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

// DiceRollingPluginConfiguration contains all settings configurable for the DiceRollingPlugin
type DiceRollingPluginConfiguration struct {
	Trigger      string
	AutoComplete bool
}

const (
	defaultTrigger string = "roll"
	diceAPIURL     string = "http://roll.diceapi.com/json/"
)

// OnActivate register the plugin command
func (p *DiceRollingPlugin) OnActivate(api plugin.API) error {
	p.api = api
	err := p.OnConfigurationChange()
	if err != nil {
		return err
	}
	config := p.config()
	if config.Trigger == "" {
		config.Trigger = defaultTrigger
	}
	err = api.RegisterCommand(&model.Command{
		Trigger:          config.Trigger,
		TeamId:           p.TeamId,
		Description:      "Roll one or more dice",
		DisplayName:      "Dice rolling command",
		AutoComplete:     config.AutoComplete,
		AutoCompleteDesc: "Roll one or several dice with the possibility to compute the sum automatically because we are lazy, lazy people",
		AutoCompleteHint: "20 d6 3d4 [sum]",
	})
	return err
}

func (p *DiceRollingPlugin) config() *DiceRollingPluginConfiguration {
	return p.configuration.Load().(*DiceRollingPluginConfiguration)
}

// OnConfigurationChange applies configuration change to the plugin
func (p *DiceRollingPlugin) OnConfigurationChange() error {
	var configuration DiceRollingPluginConfiguration
	err := p.api.LoadPluginConfiguration(&configuration)
	p.configuration.Store(&configuration)
	return err
}

// OnDeactivate unregisters the plugin command
func (p *DiceRollingPlugin) OnDeactivate() error {
	return p.api.UnregisterCommand(p.TeamId, p.config().Trigger)
}

// ExecuteCommand returns a post that displays the result of the dice rolls
func (p *DiceRollingPlugin) ExecuteCommand(args *model.CommandArgs) (*model.CommandResponse, *model.AppError) {
	config := p.config()
	if config.Trigger == "" {
		config.Trigger = defaultTrigger
	}
	cmd := "/" + config.Trigger
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
				text += fmt.Sprintf("%d ", die.Value)
				sum += die.Value
			}
			if sumRequest {
				text += fmt.Sprintf("\n*sum = %d*", sum)
			}
		}

		attachments := []*model.SlackAttachment{
			&model.SlackAttachment{
				Pretext:  fmt.Sprintf("*I rolled: %s*", strings.Join(formatedParams, " ")),
				Text:     text,
				Fallback: "I rolled some dices...",
			},
		}

		return &model.CommandResponse{ResponseType: "in_channel", Username: args.UserId, Attachments: attachments}, nil
	}

	return nil, p.error("Expected trigger " + cmd + " but got " + args.Command)
}

func (p *DiceRollingPlugin) error(message string) *model.AppError {
	return model.NewAppError("Dice Roller Plugin ExecuteCommand", message, nil, "", http.StatusBadRequest)
}

type RollAPIResult struct {
	Success bool `json:"success"`
	Dice    []struct {
		Value int    `json:"value"`
		Type  string `json:"type"`
	} `json:"dice"`
}

func callDiceAPI(rollRequest string) (*RollAPIResult, error) {
	resp := new(RollAPIResult)
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

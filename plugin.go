package main

import (
	"fmt"
	"net/http"
	"strings"
	"sync/atomic"

	"github.com/gorilla/mux"
	"github.com/mattermost/mattermost-server/model"
	"github.com/mattermost/mattermost-server/plugin"
	"github.com/mattermost/mattermost-server/plugin/rpcplugin"
)

// DiceRollingPlugin is a Mattermost plugin that adds a slash command
// to roll dices in-chat
type DiceRollingPlugin struct {
	api           plugin.API
	configuration atomic.Value
	router        *mux.Router
	enabled       bool
}

const (
	trigger    string = "roll"
	pluginPath string = "plugins/com.github.moussetc.mattermost.plugin.diceroller"
	iconPath   string = pluginPath + "/icon.png"
	iconURL    string = "/" + iconPath
)

// OnActivate register the plugin command
func (p *DiceRollingPlugin) OnActivate(api plugin.API) error {
	p.api = api
	p.enabled = true

	p.router = mux.NewRouter()

	p.router.Handle(iconURL, http.StripPrefix("/", http.FileServer(http.Dir(iconPath))))

	return api.RegisterCommand(&model.Command{
		Trigger:          trigger,
		Description:      "Roll one or more dice",
		DisplayName:      "Dice rolling command",
		AutoComplete:     true,
		AutoCompleteDesc: "Roll one or several dice with the possibility to compute the sum automatically because we are lazy, lazy people",
		AutoCompleteHint: "20 d6 3d4 [sum]",
	})
}

func (p *DiceRollingPlugin) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Header.Get("Mattermost-User-Id") == "" {
		http.Error(w, "please log in", http.StatusForbidden)
		return
	}

	p.router.ServeHTTP(w, r)
}

// OnDeactivate handles plugin deactivation
func (p *DiceRollingPlugin) OnDeactivate() error {
	p.enabled = false
	return nil
}

func DisplayHelp() (*model.CommandResponse) {
	return &model.CommandResponse{
		ResponseType: model.COMMAND_RESPONSE_TYPE_EPHEMERAL,
		Text: "- `/roll <integer>` will roll a die with the corresponding number of sides. Example: `/roll 20` rolls a 20-sided die.
		- `/roll <N:integer>d<S:integer>` will roll N S-sided dice. Example: `/roll 5D6`
		- `/roll <roll1> <roll2> <roll3> [...]` will roll all the requested dice. Example: `/roll 5 d8 13D20` will roll one 5-sided die, 1 8-sided die and 13 20-sided dice.
		- `/roll <roll1> <roll2> [...] sum` will roll all the requested dice and compute the sum of all the roll results. Example: `/roll 2d6 8` will roll two 6-sided die, 1 8-sided die and display the sum of all the results.
		- `/roll help` will show this help text.",
	}
}

// ExecuteCommand returns a post that displays the result of the dice rolls
func (p *DiceRollingPlugin) ExecuteCommand(args *model.CommandArgs) (*model.CommandResponse, *model.AppError) {
	if !p.enabled {
		return nil, appError("Cannot execute command while the plugin is disabled.", nil)
	}
	if p.api == nil {
		return nil, appError("Cannot access the plugin API.", nil)
	}

	cmd := "/" + trigger
	if strings.HasPrefix(args.Command, cmd) {
		query := strings.TrimSpace((strings.Replace(args.Command, cmd, "", 1)))

		if query == "help" || query == "--help" || query == "h" || query == "-h" {
			return GetHelpMessage(), nil
		}
		
		rollRequests := strings.Fields(query)

		sumRequest := false
		text := ""
		sum := 0
		for _, rollRequest := range rollRequests {
			if rollRequest == "sum" {
				sumRequest = true
			} else {
				result, err := rollDice(rollRequest)
				if err != nil {
					return nil, appError("Unrecognized dice rolling request", err)
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
			return nil, appError("No roll request arguments found (such as '20', '4d6', etc.).", nil)
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
				ThumbURL: iconURL,
			},
		}

		props := map[string]interface{}{
			"from_webhook":  "true",
			"use_user_icon": "true",
		}

		return &model.CommandResponse{
			ResponseType: model.COMMAND_RESPONSE_TYPE_IN_CHANNEL,
			// Username:     user.GetFullName(),
			Attachments: attachments,
			Props:       props,
		}, nil
	}

	return nil, appError("Expected trigger "+cmd+" but got "+args.Command, nil)
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

// Install the RCP plugin
func main() {
	rpcplugin.Main(&DiceRollingPlugin{})
}

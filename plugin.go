package main

import (
	"fmt"
	"math/rand"
	"net/http"
	"strings"
	"sync/atomic"
	"time"

	"github.com/gorilla/mux"
	"github.com/mattermost/mattermost-server/model"
	"github.com/mattermost/mattermost-server/plugin"
)

// DiceRollingPlugin is a Mattermost plugin that adds a slash command
// to roll dices in-chat
type DiceRollingPlugin struct {
	plugin.MattermostPlugin
	configuration atomic.Value
	router        *mux.Router
	enabled       bool
}

const (
	trigger      string = "roll"
	pluginPath   string = "plugins/com.github.moussetc.mattermost.plugin.diceroller"
	iconFilename string = "icon.png"
	iconPath     string = pluginPath + "/" + iconFilename
	iconURL      string = "/" + iconPath
)

// OnActivate register the plugin command
func (p *DiceRollingPlugin) OnActivate() error {
	p.enabled = true

	// Serve URL for the dice icon displayed in messages
	p.router = mux.NewRouter()
	p.router.HandleFunc("/"+iconFilename, func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, iconPath)
	})

	return p.API.RegisterCommand(&model.Command{
		Trigger:          trigger,
		Description:      "Roll one or more dice",
		DisplayName:      "Dice roller ⚄",
		AutoComplete:     true,
		AutoCompleteDesc: "Roll one or several dice. ⚁ ⚄ Try /roll help for a list of possibilities.",
		AutoCompleteHint: "20 d6 3d4 [sum]",
		IconURL:          iconURL,
	})
}

func (p *DiceRollingPlugin) ServeHTTP(c *plugin.Context, w http.ResponseWriter, r *http.Request) {
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

func GetHelpMessage() *model.CommandResponse {
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
		Props:   props,
		IconURL: iconURL,
	}
}

// ExecuteCommand returns a post that displays the result of the dice rolls
func (p *DiceRollingPlugin) ExecuteCommand(c *plugin.Context, args *model.CommandArgs) (*model.CommandResponse, *model.AppError) {
	if !p.enabled {
		return nil, appError("Cannot execute command while the plugin is disabled.", nil)
	}
	if p.API == nil {
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
					return nil, appError("We didn't understand what to roll. See `/roll help` for examples.", err)
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
		user, userErr := p.API.GetUser(args.UserId)
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
			Attachments:  attachments,
			Props:        props,
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
	rand.Seed(time.Now().UnixNano())
	plugin.ClientMain(&DiceRollingPlugin{})
}

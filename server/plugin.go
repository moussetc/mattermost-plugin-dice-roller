package main

import (
	"fmt"
	"net/http"
	"strings"
	"sync"

	"github.com/mattermost/mattermost/server/public/model"
	"github.com/mattermost/mattermost/server/public/plugin"
)

const (
	trigger string = "roll"
)

// Plugin implements the interface expected by the Mattermost server to communicate between the server and plugin processes.
type Plugin struct {
	plugin.MattermostPlugin

	// configurationLock synchronizes access to the configuration.
	configurationLock sync.RWMutex

	// configuration is the active plugin configuration. Consult getConfiguration and
	// setConfiguration for usage.
	configuration *configuration

	// BotId of the created bot account for dice rolling
	diceBotID string
}

func (p *Plugin) OnActivate() error {
	return p.API.RegisterCommand(&model.Command{
		Trigger:          trigger,
		Description:      "Roll one or more dice",
		DisplayName:      "Dice roller ⚄",
		AutoComplete:     true,
		AutoCompleteDesc: "Roll one or several dice. ⚁ ⚄ Try /roll help for a list of possibilities.",
		AutoCompleteHint: "20 d6+4 3d4 [sum]",
	})
}

func (p *Plugin) GetHelpMessage() *model.CommandResponse {
	props := map[string]interface{}{
		"from_webhook": "true",
	}

	return &model.CommandResponse{
		ResponseType: model.CommandResponseTypeEphemeral,
		Text: "Here are some examples:\n" +
			"- `/roll 20` to roll a 20-sided die. You can use any number.\n" +
			"- `/roll 5D6` to roll five 6-sided dice in one go.\n" +
			"- `/roll 5D6+3` to roll five 6-sided dice and add 3 the result of each die.\n" +
			"- `/roll 5D6 +3` (with a space) to roll five 6-sided dice and add 3 the total.\n" +
			"- `/roll 5 d8 13D20` to roll different dice at the same time.\n" +
			"- `/roll help` will show this help text.\n\n" +
			" ⚅ ⚂ Let's get rolling! ⚁ ⚄",
		Props: props,
	}
}

// ExecuteCommand returns a post that displays the result of the dice rolls
func (p *Plugin) ExecuteCommand(_ *plugin.Context, args *model.CommandArgs) (*model.CommandResponse, *model.AppError) {
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

func (p *Plugin) generateDicePost(query, userID, channelID, rootID string) (*model.Post, *model.AppError) {
	// Get the user to display their name
	user, userErr := p.API.GetUser(userID)
	if userErr != nil {
		return nil, userErr
	}
	displayName := user.Nickname
	if displayName == "" {
		displayName = user.Username
	}

	reason := ""
	queries := strings.SplitN(query, "\n", 2)
	if len(queries) == 2 {
		reason = queries[1]
	}
	query = queries[0]

	text := fmt.Sprintf("**%s** rolls *%s* = ", displayName, query)
	sum := 0
	rollRequests := strings.Fields(query)
	if len(rollRequests) == 0 || query == "sum" {
		return nil, appError("No roll request arguments found (such as '20', '4d6', etc.).", nil)
	}
	singleResultCount := 0
	numericDiceCount := 0
	formattedRollDetails := make([]string, len(rollRequests))
	for i, rollRequest := range rollRequests {
		// Ignore the 'sum' keyword, remnant of a previous version
		// kept for the compatibility
		if rollRequest == "sum" {
			continue
		}
		result, err := rollDice(rollRequest)
		if err != nil {
			return nil, appError(fmt.Sprintf("%s See `/roll help` for examples.", err.Error()), err)
		}
		if result.rollType == numeric {
			numericDiceCount++
			rollDetails := fmt.Sprintf("%s: ", rollRequest)
			singleResultCount += len(result.results)
			for _, roll := range result.results {
				rollDetails += fmt.Sprintf("%d ", roll)
				sum += roll
			}
			formattedRollDetails[i] = strings.TrimSpace(rollDetails)
		} else {
			formattedRollDetails[i] = fmt.Sprintf("%+d", result.sumModifier)
			sum += result.sumModifier
		}
	}

	// Always display the total
	text += fmt.Sprintf("**%d**", sum)

	// Display roll details only of necessary
	if singleResultCount > 1 {
		formattedRollDetails = filterEmptyString(formattedRollDetails)
		text += fmt.Sprintf("\n- %s", strings.Join(formattedRollDetails, "\n- "))
	}

	// Display roll reason
	if reason != "" {
		text += fmt.Sprintf("\n```\n%s\n```", reason)
	}

	return &model.Post{
		UserId:    p.diceBotID,
		ChannelId: channelID,
		RootId:    rootID,
		Message:   text,
	}, nil
}

func filterEmptyString(arr []string) []string {
	result := []string{}
	for _, val := range arr {
		if val != "" {
			result = append(result, val)
		}
	}
	return result
}

func appError(message string, err error) *model.AppError {
	errorMessage := ""
	if err != nil {
		errorMessage = err.Error()
	}
	return model.NewAppError("Dice Roller Plugin", message, nil, errorMessage, http.StatusBadRequest)
}

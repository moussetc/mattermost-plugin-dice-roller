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
	rand.Seed(time.Now().UnixNano())

	return p.API.RegisterCommand(&model.Command{
		Trigger:          trigger,
		Description:      "Roll one or more dice",
		DisplayName:      "Dice roller ⚄",
		AutoComplete:     true,
		AutoCompleteDesc: "Roll one or several dice. ⚁ ⚄ Try /roll help for a list of possibilities.",
		AutoCompleteHint: "(3d20+4)/2",
	})
}

func (p *Plugin) GetHelpMessage() *model.CommandResponse {
	props := map[string]interface{}{
		"from_webhook": "true",
	}

	return &model.CommandResponse{
		ResponseType: model.COMMAND_RESPONSE_TYPE_EPHEMERAL,
		Text: "Here are some examples:\n" +
			"- `/roll 3d20` Roll 3 `d20` dice and add the results.\n" +
			"- `/roll 4d6k3-3d4dl1` k,kh=keep highest; kl=keep lowest; d,dl=drop lowest, dh=drop highest.\n" +
			"- `/roll d20a` a=advantage; d=disadvantage. `d20a`=`2d20kh1`.\n" +
			"- `/roll d%a` d% is a synonym for d100.\n" +
			"- `/roll (5+3-2)*7/3` will use `()+-*/` with their usual meanings, except `/` rounds down.\n" +
			"- `/roll stats` will roll stats for a DnD 5e character (4d6d1 6 times).\n" +
			"- `/roll death save` will roll a death save for DnD 5e.\n" +
			"- `/roll 1d20+5 for insight` You can add a label to the roll.\n" +
			"- `/roll 1d20+4 to hit, (1d6+2 slashing)+(2d8 radiant) damage` You can make several rolls separated by commas, and add labels to nested parentheses.\n" +
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

		lQuery := strings.ToLower(query)
		if lQuery == "help" || lQuery == "--help" || lQuery == "h" || lQuery == "-h" {
			return p.GetHelpMessage(), nil
		}

		// Suppress lint error
		// > G404: Use of weak random number generator (math/rand instead of crypto/rand) (gosec)
		// because dice rolls don't need to be cryptographically secure.
		//#nosec G404
		roller := func(x int) int { return 1 + rand.Intn(x) }
		post, generatePostError := p.generateDicePost(query, args.UserId, args.ChannelId, args.RootId, roller)
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

func (p *Plugin) generateDicePost(query, userID, channelID, rootID string, roller Roller) (*model.Post, *model.AppError) {
	// Get the user to display their name
	user, userErr := p.API.GetUser(userID)
	if userErr != nil {
		return nil, userErr
	}
	displayName := user.Nickname
	if displayName == "" {
		displayName = user.Username
	}

	parsedNode, err := parse(query)
	if err != nil {
		return nil, appError(fmt.Sprintf("%s: See `/roll help` for examples.", err.Error()), err)
	}

	rolledNode := parsedNode.roll(roller)
	renderResult := rolledNode.renderToplevel()

	text := fmt.Sprintf("**%s** rolls %s", displayName, renderResult)

	return &model.Post{
		UserId:    p.diceBotID,
		ChannelId: channelID,
		RootId:    rootID,
		Message:   text,
	}, nil
}

func appError(message string, err error) *model.AppError {
	errorMessage := ""
	if err != nil {
		errorMessage = err.Error()
	}
	return model.NewAppError("Dice Roller Plugin", message, nil, errorMessage, http.StatusBadRequest)
}

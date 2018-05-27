# mattermost-plugin-dice-roller
This Mattermost plugin adds a `/roll` slash command to roll all kinds of virtual dice.

## Requirements
- Mattermost 4.6 (to allow plugins to create slash commands) 

## Installation and configuration
1. Go to the [Releases page](https://github.com/moussetc/mattermost-plugin-dice-roller/releases) and download the package for your OS and architecture.
2. Use the Mattermost `System Console > Plugins Management > Management` page to upload the `.tar.gz` package
3. **Activate the plugin** in the `System Console > Plugins Management > Management` page

## Manual configuration
If you need to enable & configure this plugin directly in the Mattermost configuration file `config.json`, for example if you are doing a [High Availability setup](https://docs.mattermost.com/deployment/cluster.html), you can use the following lines (remember to set the API key!):
```json
 "PluginSettings": {
        // [...]
        "Plugins": {
            "com.github.moussetc.mattermost.plugin.diceroller": {
            },
        },
        "PluginStates": {
            // [...]
            "com.github.moussetc.mattermost.plugin.diceroller": {
                "Enable": true
            },
        }
    }
```

## Usage
- `/roll <integer>` will roll a die with the corresponding number of sides. Example: `/roll 20` rolls a 20-sided die.
- `/roll <N:integer>d<S:integer>` will roll N S-sided dice. Example: `/roll 5D6`
- `/roll <roll1> <roll2> <roll3> [...]` will roll all the requested dice. Example: `/roll 5 d8 13D20` will roll one 5-sided die, 1 8-sided die and 13 20-sided dice.
- `/roll <roll1> <roll2> [...] sum` will roll all the requested dice and compute the sum of all the roll results. Example: `/roll 2d6 8` will roll two 6-sided die, 1 8-sided die and display the sum of all the results.
- `/roll help` will show a reminder of how to use the plugin.

## Development
Run make vendor to install dependencies, then develop like any other Go project, using go test, go build, etc.

If you want to create a fully bundled plugin that will run on a local server, you can use `make mattermost-plugin-diceroller.tar.gz`.

## What's next?
- Better code testing
- Generate roll results as accurate image (number of dice, number of faces...)

## Credits
- This project uses a dice icon provided by [openclipart](https://openclipart.org/detail/94501/twentysided-dice) under the [Creative Commons Zero 1.0 Public Domain License](https://creativecommons.org/publicdomain/zero/1.0/).
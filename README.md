# mattermost-plugin-dice-roller
This Mattermost plugin adds a `/roll` slash command to roll all kinds of virtual dice.

## Requirements
- for Mattermost 5.12 or higher: use the latest v3.x.x release
- for Mattermost 5.2 to 5.11: use the latest v2.x.x release
- for Mattermost 4.6 to 5.1: use the latest v1.x.x release
- for Mattermost below: unsupported versions (plugins can't create slash commands)

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

# Development

To avoid having to manually install your plugin, build and deploy your plugin using one of the following options.

### Deploying with Local Mode

If your Mattermost server is running locally, you can enable [local mode](https://docs.mattermost.com/administration/mmctl-cli-tool.html#local-mode) to streamline deploying your plugin. Edit your server configuration as follows:

```json
{
    "ServiceSettings": {
        ...
        "EnableLocalMode": true,
        "LocalModeSocketLocation": "/var/tmp/mattermost_local.socket"
    }
}
```

and then deploy your plugin:
```
make deploy
```

You may also customize the Unix socket path:
```
export MM_LOCALSOCKETPATH=/var/tmp/alternate_local.socket
make deploy
```

If developing a plugin with a webapp, watch for changes and deploy those automatically:
```
export MM_SERVICESETTINGS_SITEURL=http://localhost:8065
export MM_ADMIN_TOKEN=j44acwd8obn78cdcx7koid4jkr
make watch
```

### Deploying with credentials

Alternatively, you can authenticate with the server's API with credentials:
```
export MM_SERVICESETTINGS_SITEURL=http://localhost:8065
export MM_ADMIN_USERNAME=admin
export MM_ADMIN_PASSWORD=password
make deploy
```

or with a [personal access token](https://docs.mattermost.com/developer/personal-access-tokens.html):
```
export MM_SERVICESETTINGS_SITEURL=http://localhost:8065
export MM_ADMIN_TOKEN=j44acwd8obn78cdcx7koid4jkr
make deploy
```

## What's next?
- Better code testing
- Generate roll results as accurate image (number of dice, number of faces...)

## Credits
- This plugin is based of the [Mattermost plugin starter template](https://github.com/mattermost/mattermost-plugin-starter-template)
- This project uses a dice icon provided by [openclipart](https://openclipart.org/detail/94501/twentysided-dice) under the [Creative Commons Zero 1.0 Public Domain License](https://creativecommons.org/publicdomain/zero/1.0/).

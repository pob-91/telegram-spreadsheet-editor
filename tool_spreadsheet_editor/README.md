# telegram-spreadsheet-editor

A telegram bot integration for talking to your spreadsheets.

### Overview

This project allows you to send simple commands to a finances spreadsheet. The spreadsheet has an expected structure but it is very simple (categories in one comlumn with values against each category in another). This could be extended to other structures.

This project is open source under the [MIT License](LICENSE)

### Features

<p float="left">
    <img src="read_pt1.png" width="200" />
    <img src="read_pt2.png" width="200" />
</p>

**Available Commands**

- **READ** - choose a category and get the value.
- **DETAILS** - choose a category and get the function breakdown e.g. `25+67`.
- **UPDATE** - choose a category and specify how much to add to it.
- **REMOVE** - choose a category and remove the last added element e.g. `25+67+82` becomes `25+67`.
- **HELP** - prints list of available commands.
- **PING** - pong

*and some various easter eggs but what would be the fun in revealing those*

### Deploying

Telegram spreadsheet editor can be deployed from source, as a [compose stack](docker-compose.yml) or with kubernetes. We do not provide a helm chart but see [our tofu example](k3s_example.tf).

If building from source, the project is wrtten entirely in go and uses the mod syntax. Our [Dockerfile](Dockerfile) has an example of the go build command we use that is very portable. The program has only 1 requirement and that is ca-certificates.

Currently this project assumes you have a Telegram Bot set up and also a spreadsheet in an expected format and accessible for download and upload. See [Setting Up A Telegram Bot](#setting-up-a-telegram-bot) and [Spreadsheet API and Expectations](#spredsheet-api-and-expectations) for more info.

#### Setting Up A Telegram Bot

This project currently assumes that you have a Telegram bot set up, this is the primary (and only) inteface with this program. To do this you need to use the BotFather bot in the telegram app. There is a [Telegram tutorial here](https://core.telegram.org/bots/tutorial) that you can follow to get up and running. BotFather will give you a Telegram bot key which is one of the environment variables required.

By default, Telegram allows anyone to interact with your bot, whilst it is unlikely that it will be discovered, it is not ideal to allow it to be left open. To tackle this, you can specify a list of allowed Telegram user IDs. If a message is sent from an unrecognised account it will be rejected. **NOTE**: By default this program responds to all messages, see [Environment Variables](#environment-variables) to see how to filter by user ID.

#### Spreadsheet API and Expectations

**API**

This project modifies a spreadsheet by downloading it, modifying it, and then re-uploading it from your specified source. The list of sources will grow as time allows.

Sources:
- Nextcloud

**Authentication**

This project currently supports unauthenticated and basic auth endpoints.

**Spreadsheet Format**

There is an expected spreadsheet format but it is very basic. This project assumes the following:

- The last tab is the one that is to be edited.
- There is a column that lists all the categories available.
- There is a column where the values for each category are listed.
- The row for category and value are the same.

The assumption here is that the spreadsheet will be financial in nature therefore these columns are exposed in the config as cost and earnings. However, this does not *have* to be true.

The columns in the [example spreadsheet](Example.xlsx) are D and E.

### Running Locally

- Make sure you have a Telegram bot set up and also a spreadsheet URL available (see above).
- Create a .env file from .env.example and fill in the values.

> If you want to test out this project locally (no debug):

Make sure you have docker installed and run:

```shell
make up
```

This will start the stack locally and expose the API via `localhost:8080`. Then you can use a service like [ngrok](https://ngrok.com/) to generate a publicly accessible URL to use with the Telegram bot.

> If you want to debug and make edits:

Make sure you have go installed, install the dependencies (`go mod tidy`) and run either with your debugger. The makefile has a `make valkey` command that will start a valkey instance for you which is required by the project. Use something like [ngrok](https://ngrok.com/) to interact with the Telegram bot.

### Environment Variables

| NAME | Description | Default | Required |
|------|-------------|---------|----------|
| **ENVIRONMENT** | Environment name - development puts the logger into dev mode and changes behaviour of the panic level | "production" | false |
| **LOG_LEVEL** | Which level of logs to include | "Warning" | false |
| **CONFIG_PATH** | The path where the config.yaml can be found. See the config section for more. | "/home/nonroot/config.yaml" | false |
| **VALKEY_HOST** | The valkey URL. If running via docker compose set to valkey:6379. If running locally set to localhost:6379. | | true |

### Config.yaml

The project expects a `config.yaml` file to be present that is read that tells the editor:

- How to access the spreadsheet (source system, location, columns, auth e.t.c)
- What input to expect (e.g. telegram, whatsapp, text e.t.c)

See [the example config](./config.example.yaml) and the [config parser](./model/config.go) for info on eactly which properties are available.

This project uses the `gcr.io/distroless/static-debian12:nonroot` base docker image and publishes it to this repo's repository.
If you use this and do not set the `CONFIG_PATH` explicitly then the expected config path is `/home/nonroot/config.yaml`. 
There is an example of setting this in the [k3s example file](./k3s_example.tf).

### Contributing

If you want to fix a bug or add a feature to this project then just raise a Pull Request. When doing so please bear the following in mind:

- Please look at the norms and patterns in the project and adhere to them.
- Please consider users when adding features.
- It may take me a little while to come and deal with it. Apologies but I am on it!

##### Use Of AI Agents When Contributing

There is nothing inherently wrong in using LLMs to generate code to contribute to this project however, it must follow the style apparent
in the existing code and the author **MUST** be able to understand and explain every line that is being comitted.

There is a danger with generated code that plausible looking code that is actually poorly designed or not performant slips through.

### Future

- Add Google Docs support
- Add text and whatsapp support via Twilio.
- Add support for office online (probably a nightmare...)
- Add voice integration using whisper or something
- Add in currency converson using symbols or codes e.g. USD, GBP, JPY
- Add in a NEW MONTH command that creates a new month's tab and optionally reads some defaults.

### Long Term Vision

The long term vision is that this project expands out to be able to interact with a network of tools that can perform all sorts of different tasks. 
They should be interactable with via a controller or switch that can process natrual language like an LLM.

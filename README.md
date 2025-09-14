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

This project modifies a spreadsheet by downloading it, modifying it, and then re-uploading it from your specified source. This means that there must be a GET endpoint for spreadsheet downloads and a PUT for spreadsheet uploads.

**Authentication**

This project currently supports unauthenticated and basic auth endpoints.

**Spreadsheet Format**

There is an expected spreadsheet format but it is very basic. This project assumes the following:

- The last tab is the one that is to be edited.
- There is a column that lists all the categories available.
- There is a column where the values for each category are listed.
- The row for category and value are the same.

The columns in the [example spreadsheet](Example.xlsx) are D and E.

**Workarounds**

If this project does not meet your requirements due to one or more of the above assumptions you can try the following:

- If your spreadsheet is not available via GET and PUT, you could write a middleware API layer that exposes it via for this program, open an issue requesting support, or even better add support for your source and raise an MR! [See contributing](#contributing).
- If your spreadsheet is not in the expected format, and you think it is a reasonable and common format, then open an issue or add support. If your super awesome amazing unique spreadsheet format is not supported, it probably never will be!

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

Make sure you have go installed, install the dependencies (`go mod tidy`) and run either with your debugger. The makefile has a `make redis` command that will start a redis instance for you which is required by the project. Use something like [ngrok](https://ngrok.com/) to interact with the Telegram bot.

### Environment Variables

| NAME | Description | Default | Required |
|------|-------------|---------|----------|
| **ENVIRONMENT** | Environment name - development puts the logger into dev mode and changes behaviour of the panic level | "production" | true |
| **HOST** | Address on which the API listens | "0.0.0.0" | true |
| **PORT** | Port on which the API listens | "8080" | true |
| **LOG_LEVEL** | Which level of logs to include | "Warning" | true |
| **BASIC_AUTH_USER** | Sets basic auth user for spreadsheet GET and PUT | | false |
| **BASIC_AUTH_PASSWORD** | Sets basic auth password for spreadsheet GET and PUT | | false |
| **SHEET_BASE_URL** | Base URL (not full path) for the spreadsheet e.g. `https://epic-server.com` | | true |
| **XLSX_FILE_PATH** | Path for the spreadsheet e.g. `docs/my-cool-sheet.xlsx`. Joined with `SHEET_BASE_URL` | | true |
| **KEY_COLUMN** | Spreadsheet column where the categories are listed | | true |
| **VALUE_COLUMN** | Spreadsheet column where the values are listed for each category | | true |
| **START_ROW** | The row at which to start looking for categories and values. Use this to skip header rows | | false |
| **TELEGRAM_BOT_TOKEN** | The token for your telegram bot | | true |
| **TELEGRAM_ALLOWED_USERS** | A comma sepaarated list of telegram user IDs. If set then the API will reject unrecognised users. | | false |
| **SERVICE_HOST** | Public URL that Telegram can use to communicate with this API. E.g. `https://my-cool-bot.com`. If using `ngrok` then set this variable to the ngrok url | | true |
| **REDIS_HOST** | The redis URL. If running via docker compose set to redis:6379. If running locally set to localhost:6379. | | true |

### Contributing

If you want to fix a bug or add a feature to this project then just raise a Pull Request. When doing so please bear the following in mind:

- Please look at the norms and patterns in the project and adhere to them.
- Please consider users when adding features.
- It may take me a little while to come and deal with it. Apologies but I am on it!

### Future

- Add text and whatsapp support via Twilio.
- Add voice integration using whisper or something
- Add in currency converson using symbols or codes e.g. USD, GBP, JPY
- Add in a NEW MONTH command that creates a new month's tab and optionally reads some defaults.

### Long Term Vision

The long term vision for this project is to become a fully fledged financial management system where transactions are logged with descriptions and as much info and context as possible. Then, as well as the basic commands like READ and UPDATE, with LLM integration one could have a very natual chat about one's finances. This would require this API to become an MCP server also that the LLM could call upon to get info about financial transactions. This is actually more secure than other ideas e.g. give an LLM access to your OpenBanking API (madness!).

Overall this project could help people take more control over their finances more easily.

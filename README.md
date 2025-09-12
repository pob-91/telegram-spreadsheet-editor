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

#### Setting Up A Telegram Bot

TODO

#### Spreadsheet API and expectations

TODO

### Running Locally

### Environment Variables

TODO

### Contributing

### Future

- Add voice integration using whisper or something
- Add in currency converson using symbols or codes e.g. USD, GBP, JPY
- Add in a NEW MONTH command that creates a new month's tab and optionally reads some defaults.

### Long Term Vision

The long term vision for this project is to become a fully fledged financial management system where transactions are logged with descriptions and as much info and context as possible. Then, as well as the basic commands like READ and UPDATE, with LLM integration one could have a very natual chat about one's finances. This would require this API to become an MCP server also that the LLM could call upon to get info about financial transactions. This is actually more secure than other ideas e.g. give an LLM access to your OpenBanking API (madness!).

Overall this project could help people take more control over their finances more easily.

# nextcloud-spreadsheet-editor
An integration for editing spreadsheets via tools like WhatsApp

### TODO

- Upload docker image to docker hub or GCR
- Docs and env vars detailed & steps to get telegram working

### Future

- Add voice integration using whisper or something
- Add in currency converson using symbols or codes e.g. USD, GBP, JPY
- Add in a NEW MONTH command that creates a new month's tab and optionally reads some defaults.

### Long Term Vision

The long term vision for this project is to become a fully fledged financial management system where transactions are logged with descriptions and as much info and context as possible. Then, as well as the basic commands like READ and UPDATE, with LLM integration one could have a very natual chat about one's finances. This would require this API to become an MCP server also that the LLM could call upon to get info about financial transactions. This is actually more secure than other ideas e.g. give an LLM access to your OpenBanking API (madness!).

Overall this project could help people take more control over their finances more easily.

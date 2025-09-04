# nextcloud-spreadsheet-editor
An integration for editing spreadsheets via tools like WhatsApp


### TODO

- At the moment if a cell does not have a formula just a value setting the formula overwrites it, need to test for empty formula but existing values and add it on. Something crashing around valueToAdd
- Make READ command for category and row number (check there is a category before doing) - can be similar to update but don't look for a number
- Make COMMAND that lists all available commmands, also HELP.

- Dockerfile
- Upload docker image to docker hub
- Compose stack
- Docs and env vars detailed & steps to get telegram working


### Future

- Add voice integration using whisper or something
- Add in currency converson using symbols or codes e.g. USD, GBP, JPY

# nextcloud-spreadsheet-editor
An integration for editing spreadsheets via tools like WhatsApp


### TODO

- Add endpoint that can be hit from a Telegram bot that parses the message and turns it into an instruction - respond with old and new value and name of category updated
- If telegram can send custom tokens in requests make middleware for that
- Make middleware that filters accepted account numbers that can be set in env vars
- Accept numbers in place of categories for rows to update
- Endpoint that lists all row numbers and categories and one that does the same for Telegram on LIST command
- Make READ command for category and row number (check there is a category before doing)

- Dockerfile
- Upload docker image to docker hub
- Compose stack
- Docs and env vars detailed & steps to get telegram working


### Future

- Add voice integration using whisper or something

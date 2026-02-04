# TODO

I know this is bad for agents but fuck off.

- [ ] Add commands for adding earnings
- [ ] Add commands for creating new sheets for new months
- [ ] Migrate from a webhook to the golang chan way of doing things if it works
- [ ] Add a yaml config that can accept multiple users and migrate key settings to that (upgrade k3s tofu to input this as a map)
- [ ] Update to subscribe to multiple users and handle commands for them seperately based on their bot tokens and user ids
- [ ] Add support for voice commands via LLM providers (use some provider agnostic tool like ngrok if free)
- [ ] Migrate to controller -> tool command interpreter -> tool architecture
- [ ] Add in server status tool

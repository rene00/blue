# blue

## experimental project: use at your own risk.

A small terminal tool used to interact with various online services, mostly centered around AI.

To install, check out the releases page or use [task](https://taskfile.dev):
```bash
$ task build
```

Then set `OPENAI_API_KEY` to a valid [OpenAI key](https://platform.openai.com/account/api-keys):
```bash
$ export OPENAI_API_KEY=secret
```

## chat

To chat interactively:
```bash
$ blue chat
┃ Write two tag lines for an ice cream shop:
┃
┃

1. "Indulge in the ultimate scoop of happiness!"
2. "Cool treats to satisfy your sweet cravings!"
```
- press [tab] to submit the prompt.

Or using vim to write your prompt:
```
$ blue chat --editor
```

## commands
`chat` can accept commands in the form of `c:$NAME` (example: `c:print`)

### print
`c:print` will print the current messages. The messages will not be sent.
```bash
┃ why does it rain?
┃ c:print
┃

{"role":"user","content":"why does it rain?\n"}
```

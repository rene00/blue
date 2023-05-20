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
- press `[tab]` to submit the prompt.
- press `[enter]` to continue after receiving the response.

Or using `vim` to write your prompt:
```
$ blue chat --editor
```

## commands
`chat` can accept commands in the form of `c:$NAME` (example: `c:print`).

Commands will run and will not send the prompt to openai. Instead the prompt will be returned to the screen for further editing or sending with `[tab]`.

### echo
`c:echo` echos the prompt to standard out.
```bash
┃ why does it rain?
┃ c:echo
┃

why does it rain?
```

### editor
`c:editor` opens the prompt within your editor.
```bash
┃ why does it rain?
┃ c:editor
```

### printmessages
`c:printmessages` prints the messages stored within the chat completion message
```bash
┃ why does it rain?
┃ c:editor
┃

[]openai.ChatCompletionMessage{openai.ChatCompletionMessage{Role:"user", Content:"why does it rain?", Name:""}}
```

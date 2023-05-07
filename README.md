# blue

## experimental project: use at your own risk.

A small terminal tool used to interact with various online services, mostly centered around AI.

To install, check out the releases page or use [task](https://taskfile.dev):
```bash
$ task
```

Then set `OPENAI_API_KEY` to a valid [OpenAI key](https://platform.openai.com/account/api-keys):
```bash
$ export OPENAI_API_KEY=secret
```

## chat

To chat interactively:
```bash
$ blue chat 
Input text [Press tab to submit] :
Write two taglines for an ice cream shop.
1. "Scoops of happiness in every cone!"
2. "Sweet treats that will cool your soul."
```

Or using vim to write your prompt
```
$ blue chat --editor
```

Or both:
```
$ blue chat
Input text [Press tab to submit] :
c:editor
```
The last line will open up vim for your prompt whilst keeping you in interactive mode.

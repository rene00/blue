# blue

A small terminal tool used to interact with various online services, mostly centered around AI.

To install check out the releases page or use [task](https://taskfile.dev):
```bash
$ task
```

Then set `OPENAI_API_KEY` to a valid [openai key](https://platform.openai.com/account/api-keys):
```bash
$ export OPENAI_API_KEY=secret
```

## Chat

To chat interactively:
```bash
$ blue chat 
> finish this: you can feel  
the warm sun on your skin, the cool breeze in your hair, and the sense of contentment in your heart.
```

Or with vim opening for your prompt
```
$ blue chat --editor
```

Or both:
```
$ blue chat
> name three random countries
Mali, Moldova, Samoa.
> c:editor
```
The last line will open up your editor for your prompt.

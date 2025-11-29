# TODOne
Find the time to get your TODOs TODOne.

TODOne is an agent that lives in your terminal and helps you fill free time with prioritized, time-bound todos. Tasks are aggregated from your environment so you don't have to lift all of your tasks into a todo app first. 

Currently, only code-based tasks are supported. See [#configuration] for configuring TODOne.

## Requirements
* [Go](https://go.dev/doc/install)
* [ripgrep](https://github.com/BurntSushi/ripgrep?tab=readme-ov-file#installation)
* exported OPENAI_API_KEY (e.g. `export OPENAI_API_KEY=<key>`)

## Configuration
> [!NOTE] 
> Configuration is a work in progress, config is only read from the `todone.toml` in the repo root.


## Usage
```bash
git clone https://github.com/mgleason97/todone.git
cd todone
go run ./cmd/todone -config todone.toml
```

This will start an interactive session where you can talk back and forth with your task planning agent.

A sample repo is provided at `/sample`, so todone will be able to plan tasks for you to complete in this repo. 

![Sample todone usage](docs/images/todone_sample.png)

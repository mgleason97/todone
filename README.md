# TODOne
Find the time to get your TODOs TODOne

## Requirements
* [Go](https://go.dev/doc/install)
* [`ripgrep`](https://github.com/BurntSushi/ripgrep?tab=readme-ov-file#installation)
* exported OPENAI_API_KEY (e.g. `export OPENAI_API_KEY=<key>`)

## Usage
`go run ./cmd/todone -config todone.toml -question "I want to schedule an hour of focused work. What should I work on?"`

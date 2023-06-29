## [ironkey.dev](Ironkey.dev) API

<div style="display: flex;">
    <img src="https://img.shields.io/badge/go-%2300ADD8.svg?style=for-the-badge&logo=go&logoColor=white" alt="Go logo">
    &nbsp;
    <img src="https://img.shields.io/badge/docker-%230db7ed.svg?style=for-the-badge&logo=docker&logoColor=white" alt="Docker logo">
</div>


## About

[ironkey.dev](https://ironkey.dev) is an open source user management tool written with Golang and Sveltekit. It is currently under development.

## Docs

TODO (will be at [docs.ironkey.dev](https://docs.ironkey.dev))

## Running

`go run main.go`

Or if you prefer using [air](https://github.com/cosmtrek/air):

Place in root directory:

```toml
root = "."
testdata_dir = "testdata"
tmp_dir = "tmp"

[build]
  args_bin = []
  bin = "tmp\\main.exe"
  cmd = "go build -o ./tmp/main.exe ./src"
  delay = 0
  exclude_dir = ["assets", "tmp", "vendor", "testdata"]
  exclude_file = []
  exclude_regex = ["_test.go"]
  exclude_unchanged = false
  follow_symlink = false
  full_bin = ""
  include_dir = ['']
  include_ext = ["go", "tpl", "tmpl", "html"]
  include_file = []
  kill_delay = "0s"
  log = "build-errors.log"
  poll = false
  poll_interval = 0
  rerun = false
  rerun_delay = 500
  send_interrupt = false
  stop_on_error = false

[color]
  app = ""
  build = "yellow"
  main = "magenta"
  runner = "green"
  watcher = "cyan"

[log]
  main_only = false
  time = false

[misc]
  clean_on_exit = false

[screen]
  clear_on_rebuild = true
  keep_scroll = true

```
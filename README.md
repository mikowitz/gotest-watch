# gotest-watch

`gotest-watch` is an interactive test runner for Go projects

* [Installation](#installation)
* [Usage](#usage)
  * [Interactive Commands](#interactive-commands)
  * [CLI arguments](#cli-arguments)
  * [.gotest-watch.yml](#.gotest-watch.yml)

Inspired by Randy Coulman's [mix test.interactive],
`gotest-watch` allows you to change the parameters of your `go test`
command with just a few keystrokes. It also watches your `*.go` files
and reruns your test suite with the set parameters after any file change.

[mix test.interactive]: https://github.com/randycoulman/mix_test_interactive

## Installation

```bash
go install github.com/mikowitz/gotest-watch
```

## Usage

Start `gotest-watch` by running it from the root directory of your project

```bash
gotest-watch
```

This will run your test suite once, and then begin watching your project's `*.go` files
and wait for your input. By default, this command runs `go test ./...`,
but this can be changed by passing one of the following interactive commands:

### Interactive Commands

| Command | Function | `go test` equivalent |
| ------------- | -------------- | -------------- |
| `v` | toggle verbose mode | `-v` |
| `race` | toggle race mode | `-race` |
| `ff` | toggle failfast mode | `-failfast` |
| `cover` | toggle test coverage mode | `-cover` |
| `count <n>` | how many times to run each test | `-count <n>` |
| `r <pattern>` | only run tests whose names match the given pattern | `-run pattern` |
| `r` | clears the `-run` flag pattern |  |
| `s <pattern>` | skips tests whose names match the given pattern | `-skip pattern` |
| `s` | clears the `-skip` flag pattern |  |
| `p <pattern>` | sets the directory to run tests from (default `./...` all test packages) | package(s) path passed to `go test` |
| `p` | resets the packages under test to `./...` |  |
| `clear` | resets and clears all parameters to `go test` |  |
| `cmd` | sets the base command to run (default `go test`)|  |
| `color` | toggles colorization for the test output | no equivalent |
| `cls` | toggles clearing the screen before each test run | no equivalent |
| `f` | trigger a run of the tests per the current gotest-watch configuration | no equivalent |
| `help` | print out a list of the available commands | no equivalent |

### CLI arguments

Many of the interactive commands can also have their initial values set via flags passed to the initial `gotest-watch` invocation.

The following flags are supported

| Flag   | `gotest-watch` command    |
|--------------- | --------------- |
| `-v`, `--verbose[=false]`   | `v`   |
| `-r PATTERN`, `--run=PATTERN`   | `r`   |
| `-s PATTERN`, `--skip=PATTERN`   | `s`   |
| `-n COUNT`, `--count=COUNT`   | `count`   |
| `-l` `--cls`   | `cls`   |
| `-c` `--color[=false]`   | `color`   |
| `-m CMD`, `--cmd=CMD`   | `cmd`   |
| `-p PATH`, `--path=PATH`   | `-p`   |

### .gotest-watch.yml

Initial configuration can also be set via a file named `.gotest-watch.yml` in the root of your project.
Below is a sample file containing all the valid keys with the default values set.

```yaml
---
# Configures the test command
commandBase:
- go
- test
testPath: ./...
verbose: false
runPattern: ""
skipPattern: ""
race: false
cover: false
failfast: false
count: 0
# Configures gotest-watch
clearScreen: false
color: false
```

# gotest-watch

`gotest-watch` is an interactive test runner for Go projects

* [Installation](#installation)
* [Usage](#usage)
  * [Interactive Commands](#interactive-commands)

Inspired by Randy Coulman's [mix test.interactive],
`gotest-watch` allows you to change the parameters of your `go test`
command with just a few keystrokes. It also watches your `*.go` files
and reruns your test suite with the set parameters after any file change.

[mix test.interactive]: https://github.com/randycoulman/mix_test_interactive

## Installation

```bash
> go install github.com/mikowitz/gotest-watch
```

## Usage

Start `gotest-watch` by running it from the root directory of your project

```bash
> gotest-watch
```

This will run your test suite once, and then begin watching your project's `*.go` files
and wait for your input. By default, this command runs `go test ./...`,
but this can be changed by passing one of the following interactive commands:

### Interactive Commands

* `v` - toggle verbose mode on and off (equivalent to the `-v` flag)
* `r <pattern>` - run tests whose names match the given pattern (equivalent to the `-run` flag)
* `r` - clears any previously set pattern
* `s <pattern>` - skips tests whose names match the given pattern (equivalent to the `-skip` flag)
* `s` - clears any previously set skip pattern
* `p <pattern>` - runs tests under the given path
* `p` - sets the given path to `./...` (runs all tests)
* `clear` - clears all set parameters, returns to running `go test ./...`
* `cls` - clear the terminal window
* `f` - force a run of the tests, using the currently set configuration
* `help` - print out the interactive command options

# Starter

Starter is an additional utility for tracking errors in the logs of the main process. If the given substring is found in the logs more than N times in the given period of time, then the starter will restart the process.

If starter receives a signal to stop, it will stop the main process and exit.

## Build

```go build -o /build/bin/starter cmd/starter/main.go```

## Usage

First 4 arguments are required, the rest are optional.

```starter <error_substring> <max_count> <period> <command> [args...]```

- `error_substring` - substring to search in the logs

- `max_count` - maximum number of occurrences of the substring in the given period

- `period` - period of time in seconds (e.g. 60s, 1m, 1h, 1d)

- `command` - command to run

- `[args...]` - arguments for the command

## Example

This command will restart erigon if the substring 'No block bodies' is found in the logs more than 50 times in 5 minutes:

```starter 'No block bodies' 50 5m erigon --datadir /data --private.api.addr```

It is possible to track multiple errors by using starter multiple times.

```starter 'No block bodies' 50 5m starter 'No peers' 10 60s erigon --datadir /data --private.api.addr```


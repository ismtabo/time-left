Time left
===

## Description

This is a simple countdown timer that can be used to track the time left until the end of the work day.

## Usage

To use this timer, simply run the script and it will display the time left until the end of the work day.

```bash
time-left [options]
```

### Options

The following options are available:

```bash
  -c, --config string   The path to the configuration file (default "$HOME/.config/time-left/config.yaml")
```

## Installation

### Requirements

Readmore about systray requirements [here](https://github.com/getlantern/systray#platform-notes).

### Using go get

To install this script, simply run the following command:

```bash
go install github.com/ismtabo/time-left/cmd/time-left@latest
```

### Source
To install this script, simply clone the repository and build the binary:

```bash
git clone
cd time-left
go build -o time-left cmd/time-left/main.go
```

## Configuration

The script can be configured by editing the `config.yaml` file in your XDG config directory (`$HOME/.config/time-left`). Use the following snippet to configure the script:

```yaml
# The time when the work day starts
start: 09:00:00
# The duration of the work day as a duration string
duration: 7h30m
# The duration of the rest time as a duration string
rest: 30m
```

## Contributing

To contribute to this project, please fork the repository and submit a pull request.

## License

This project is licensed under the MIT License - see the [LICENSE.md](LICENSE.md) file for details

## Acknowledgements

This project was inspired by the need to track the time left until the end of the work day.

- getlantern/systray: https://github.com/getlantern/systray

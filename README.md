# Arma Reforger Mod Syncer

A terminal app that downloads the mods for your favorite Arma Reforger servers, so they are ready before you launch the game.

> [!IMPORTANT]
> This has been developed and tested on Linux only. Windows support is still WIP.

## Requirements

- Arma Reforger installed through Steam
- Steam running and logged in
- Go `1.26+`

## Build

```sh
go build -o armsync ./cmd/armsync
```

## Usage

```sh
armsync           # open the TUI
armsync config    # choose the addons (and game) directory again
armsync --debug   # write logs to ./debug.log
```

On first run `armsync` detects the game directory and asks you to pick the `addons` directory. Pick a server from your favorites list to sync its mods.

`config` flags:

- `-a, --addons DIR` set the addons directory without the picker
- `-g, --game DIR` set the game directory without auto-detection

Config is saved to `<user config dir>/armsync/config.json`.

## Keys

| Key        | Action          |
|------------|-----------------|
| `enter`    | select server   |
| `r`        | refresh servers |
| `esc`      | back            |
| `q`/`ctrl+c` | quit          |

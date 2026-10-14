# gamewrap

A Go wrapper for running games on Linux that starts commands through GameMode when enabled and
manages VRR for the duration of the game. Settings and game-specific variables can be supplied
through profiles or the environment.

> [!NOTE]
> This project is highly tailored to my personal setup and workflow. Expect profiles and behavior
> specific to my environment. I maintain it for my own use and do not plan to add or change
> functionality outside the scope I choose for that use.

## Features

- Per-game profiles selected by executable name or Steam App ID.
- Configurable GameMode and MangoHud integration.
- Automatic VRR management with state restoration and concurrent-game locking.
- Native, Proton, and Steam Linux Runtime command support.
- Executable replacement for skipping launchers or selecting alternate targets.
- Optional Steam identity cleanup for Proton games.
- Winecfg and Registry Editor integration using system Wine or the active Proton prefix.
- Environment overrides through profiles and command-line options.

## Build

Gamewrap targets Linux and is not intended to build or run on other operating systems.

With `mise`:

```sh
mise run build
```

Without `mise`, use Go 1.25 or newer directly:

```sh
go build -o dist/gamewrap ./cmd/gamewrap
```

The binary is created at `dist/gamewrap`. Install or link it somewhere in your `PATH`.

```sh
gamewrap --version
```

## Runtime dependencies

Runtime commands are resolved through `PATH` and are required only when their corresponding features
are used.

<!-- dprint-ignore -->
| Dependency | Required when | Behavior when unavailable |
| --- | --- | --- |
| `gamemoderun` | GameMode is enabled | The game does not start |
| `kscreen-doctor` | The VRR policy is not `"unchanged"` | VRR is skipped, the error is logged, and the game starts normally |
| `MangoHud` | The `"env"` or `"wrapper"` mode is selected | The `"wrapper"` mode prevents startup when its executable is missing. The `"env"` mode is not prevalidated |
| `winecfg` | `--winecfg` runs without a Proton command | The utility does not start |
| `regedit` | `--regedit` runs without a Proton command | The utility does not start |

## Usage

Run any game or command through `gamewrap`:

```sh
gamewrap -- ./game
```

The `--` separator is required before the game command. Arguments after it are passed through
without being parsed by gamewrap.

The wrapper invokes `gamemoderun` internally when GameMode is enabled. MangoHud, Proton options, DLL
overrides, and other environment variables are inherited by the game.

## Profiles

Create a `profiles` directory next to the resolved `gamewrap` binary. For example, a binary at
`dist/gamewrap` reads profiles from `dist/profiles`. Symlinks to the binary are resolved before
locating this directory.

Profiles and environment values are applied in this order:

1. Environment inherited from the launcher or terminal
2. `global.toml`
3. `<executable name>.toml`, such as `nikke_launcher.exe.toml`
4. `<Steam App ID>.toml`, such as `3259600.toml`
5. Values passed with `--env NAME=VALUE`

Later values override earlier ones. Missing profiles are ignored. Invalid profiles are skipped and
the error is logged without preventing the game from starting. Profile environment values replace
variables already supplied by Steam, including compatibility paths.

The executable profile name comes from the actual game target. Gamewrap skips Proton and Steam Linux
Runtime wrappers such as `steam-launch-wrapper`, `reaper`, `_v2-entry-point`, and
`scout-on-soldier-entry-point-v2` before selecting the profile.

Example `profiles/global.toml`:

```toml
[config]
gamemode = true
mangohud = "disabled"
vrr = "unchanged"
vrr_output = "detect"
hide_steam = false

[env]
DXVK_LOG_LEVEL = "none"
```

Example `profiles/3259600.toml`:

```toml
[config]
mangohud = "env"
vrr = "automatic"

[env]
NVPRESENT_ENABLE_SMOOTH_MOTION = 1
```

Values under `[env]` may be strings, integers, floats, or booleans and are converted to environment
strings for the game process. String values may begin with `@home` for the current user's home
directory or `@target-dir` for the effective game executable directory. The same expansion applies
to values passed with `--env`. Inherited environment values are left unchanged. Tokens are
recognized only at the beginning of a value and cannot be used for executable overrides.

```toml
[env]
MANGOHUD_CONFIGFILE = "@home/.config/MangoHud/game.conf"
XDG_DATA_HOME = "@target-dir/data"
```

MangoHud profiles accept only `"disabled"`, `"env"`, and `"wrapper"`. The `"env"` mode is preferred
for Vulkan and games using Proton/DXVK/VKD3D. For a native OpenGL game that requires the MangoHud
executable wrapper, use:

```toml
[config]
mangohud = "wrapper"

[env]
MANGOHUD_DLSYM = 1
```

The `MANGOHUD_DLSYM` setting is optional and only needed by some OpenGL games. Use `GW_MANGOHUD` to
temporarily replace the profile mode. Gamewrap manages `MANGOHUD` itself and logs a warning when it
is supplied directly. Other MangoHud variables remain available normally.

```sh
gamewrap --env GW_GAMEMODE=0 --env GW_MANGOHUD=disabled --env GW_VRR=never -- %command%
```

Each `--env` accepts exactly one `NAME=VALUE` argument. Quote the complete argument when its value
contains spaces or shell characters. If the same name is repeated, the last value wins.

Set `hide_steam` in a game profile when needed. `--hide-steam` forces it on and `--no-hide-steam`
forces it off.

```toml
[config]
hide_steam = true
```

`executable` replaces the original game target after profiles have been selected. This can replace
any executable, including skipping a launcher, while still loading the profile named after the
original target (or the Steam App ID). A value containing only a filename is resolved next to the
original executable, so a target in the same directory does not need a full path. Values containing
`/` or `\` are passed through unchanged. Proton and Steam Linux Runtime wrappers are preserved. Only
the original game executable and its arguments are replaced. Direct native commands behave the same
way.

Use `GW_EXECUTABLE` for an environment override. The usual precedence applies: an inherited value is
replaced by `[config].executable`, `[env].GW_EXECUTABLE`, or `--env GW_EXECUTABLE=...`, in that
order. An empty final value disables the override. For example:

```toml
# profiles/<Steam App ID>.toml
[config]
executable = "game.exe"
```

```sh
gamewrap --env 'GW_EXECUTABLE=C:\Games\Game.exe' -- %command%
```

## Steam

Use `gamewrap` as the final wrapper in the game's Launch Options:

```sh
gamewrap -- %command%
```

For a Proton game that must not inherit Steam identification variables, place `--hide-steam` before
`%command%`:

```sh
gamewrap --hide-steam -- %command%
```

Gamewrap preserves the outer Steam Linux Runtime command and removes Steam identity variables
immediately before Proton starts. The cleanup also applies to values supplied with `--env`. Use
`--no-hide-steam` when those variables must reach Proton. Compatibility variables such as
`STEAM_COMPAT_DATA_PATH` are preserved. This only hides environment-based detection and does not
prevent applications from connecting to the Steam client by other means.

## Wine utilities

Open Winecfg or Registry Editor using the system Wine and its current `WINEPREFIX`:

```sh
gamewrap --winecfg
gamewrap --regedit
```

To use the Proton version and prefix selected for a Steam game, keep `%command%` after the utility
option:

```sh
gamewrap --winecfg -- %command%
gamewrap --regedit -- %command%
```

In this mode, gamewrap preserves the Steam Linux Runtime command but replaces the game executable
with the requested Wine utility. GameMode, VRR, and MangoHud are disabled. Profiles are not loaded.
Other overrides remain available, for example `gamewrap --env WINEPREFIX=/path --winecfg`. Pass
`--hide-steam` explicitly when it is needed with a Proton utility.

## Game and runtime variables

<!-- dprint-ignore -->
| Variable | Values | Default |
| --- | --- | --- |
| `GW_GAMEMODE` | `1` or `0` | Enabled |
| `GW_MANGOHUD` | `disabled`, `env`, or `wrapper` | Disabled |
| `GW_VRR` | `automatic`, `never`, `always`, or `unchanged` | Keep current policy |
| `GW_VRR_OUTPUT` | KScreen output name or `detect` | `detect` |

Invalid game and runtime values are logged without preventing the game from starting.

### GameMode

Set `gamemode = false` under `[config]`, or use `--env GW_GAMEMODE=0`, to run the game directly
without `gamemoderun`. Disabling it also skips the Proton GameMode library preload and removes
GameMode as a runtime dependency.

For Proton commands on Fedora/RHEL-style multilib layouts, GameMode prepends
`/usr/$LIB/libgamemode.so.0` to `LD_PRELOAD` when both the i686 and x86-64 libraries are installed.
This lets Steam Linux Runtime import the matching library architecture while preserving existing
preloads. Other filesystem layouts skip this optimization and still start the game normally through
`gamemoderun`.

### VRR

VRR profiles use the explicit policies `"automatic"`, `"never"`, `"always"`, and `"unchanged"`. Use
`vrr = "unchanged"` to override an inherited profile setting and preserve the current host policy
without invoking `kscreen-doctor`. Use `--env GW_VRR=unchanged` for the same temporary override.

When gamewrap changes VRR, it restores the previous policy after the game exits. VRR management
works only on KDE Plasma because it uses KScreen's `kscreen-doctor` utility.

With `vrr_output = "detect"`, gamewrap selects the only connected and enabled output that reports
VRR support. If no output qualifies, or more than one does, the game starts without changing VRR and
the reason is written to the log. Set an explicit output name when multiple active displays support
VRR.

### Concurrent games

VRR modifies host-wide state. The first game that changes it holds a session lock until it exits.
Other games still start normally with all their environment variables and GameMode, but skip VRR
changes while that lock is held. This prevents one wrapper from restoring state owned by another.

## Logging

Logs are written to `$XDG_STATE_HOME/gamewrap/gamewrap.log`, or
`~/.local/state/gamewrap/gamewrap.log`.

Logging is configured before command options and profiles are processed. Set these wrapper variables
before `gamewrap`. They are not controlled by profile `[env]` values or the `--env` option:

<!-- dprint-ignore -->
| Variable | Values | Default |
| --- | --- | --- |
| `GW_LOG` | `0` to disable logging | On |
| `GW_LOG_FILE` | Custom log path | `$XDG_STATE_HOME/gamewrap/gamewrap.log` |

```sh
GW_LOG=0 gamewrap -- %command%
GW_LOG_FILE=/tmp/gamewrap.log gamewrap -- %command%
```

## Project structure

- `cmd/gamewrap`: CLI composition, options, help, and logging.
- `internal/config`: profile parsing, merging, and runtime configuration.
- `internal/invocation`: target detection and command transforms.
- `internal/host`: VRR host state management and restoration.
- `internal/process`: game process construction, execution, and signal handling.

## References

- [Steam Linux Runtime](https://github.com/ValveSoftware/steam-runtime): container runtime
  generations and `pressure-vessel` behavior.
- [Proton](https://github.com/ValveSoftware/Proton): Proton commands, environment, and Steam
  compatibility behavior.
- [GameMode](https://github.com/FeralInteractive/gamemode): `gamemoderun` usage and library preload
  behavior.
- [MangoHud](https://github.com/flightlessmango/MangoHud): environment, wrapper, and OpenGL modes.
- [KScreen Doctor](https://invent.kde.org/plasma/libkscreen/-/tree/master/src/doctor): output
  discovery and VRR policies.
- [BurntSushi TOML](https://github.com/burntsushi/toml): TOML parser.

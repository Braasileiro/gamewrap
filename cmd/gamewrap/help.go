package main

import (
	"io"
	"os"
)

func printHelp() {
	_, _ = io.WriteString(os.Stdout, `Usage:
  gamewrap [OPTIONS...] -- COMMAND [ARGUMENTS...]
  gamewrap [OPTIONS...] --winecfg [-- STEAM_COMMAND...]
  gamewrap [OPTIONS...] --regedit [-- STEAM_COMMAND...]
  gamewrap (-h | --help | -v | --version)

Options:
  --env NAME=VALUE    Override one game/runtime variable; may be repeated
  --hide-steam        Hide Steam identification from a Proton game process
  --no-hide-steam     Do not hide Steam, overriding the active profile
  --winecfg           Open Winecfg using system Wine or the active Proton prefix
  --regedit           Open Registry Editor using system Wine or the active Proton prefix
  --                   End gamewrap options and begin the command
  -h, --help          Show this help and exit
  -v, --version       Show the version and exit

Game/runtime variables:
  GW_GAMEMODE     GameMode: 1 or 0 (default: 1)
  GW_MANGOHUD     MangoHud: disabled, env, or wrapper (default: disabled)
  GW_VRR          VRR: automatic, never, always, or unchanged (default: unchanged)
  GW_VRR_OUTPUT   KScreen output name or detect (default: detect)
  GW_EXECUTABLE   Replace the game target and discard its arguments

Wrapper environment:
  GW_LOG          Set to 0 to disable logging
  GW_LOG_FILE     Override the log file path

Profiles:
  Loaded from the profiles directory next to the resolved gamewrap binary.
  Order: global.toml, executable-name.toml, Steam-App-ID.toml, then --env.
  @home and @target-dir expand in [env] and --env values except GW_EXECUTABLE.

Examples:
  gamewrap -- ./game
  gamewrap --env GW_VRR=never -- %command%
  gamewrap --hide-steam -- %command%
  gamewrap --winecfg -- %command%

Exit status:
  Commands preserve their exit status. Gamewrap errors return 1.

Default log:
  $XDG_STATE_HOME/gamewrap/gamewrap.log
  or ~/.local/state/gamewrap/gamewrap.log
`)
}

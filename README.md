# ccw (claude-with)

[![CI](https://github.com/glstep/claude-with/actions/workflows/ci.yml/badge.svg)](https://github.com/glstep/claude-with/actions/workflows/ci.yml)
[![Release](https://img.shields.io/github/v/release/glstep/claude-with)](https://github.com/glstep/claude-with/releases/latest)
[![License: MIT](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)

`ccw` is a small CLI wrapper around [`claude`](https://github.com/anthropics/claude-code) (the Claude Code CLI) that lets you point it at local or self-hosted LLM backends instead of Anthropic's API.

You define named **profiles** in a TOML config file — each one sets `ANTHROPIC_BASE_URL`, `ANTHROPIC_MODEL`, and `ANTHROPIC_API_KEY` — and `ccw` launches `claude` with the right environment for whichever profile you pick.

```sh
ccw local "hi"
```

## Installation

### Install script (Linux, macOS)

```sh
curl -fsSL https://raw.githubusercontent.com/glstep/claude-with/main/install.sh | sh
```

Downloads the right binary for your platform from the latest release and installs it to `/usr/local/bin` (using `sudo` only if needed). Set `BINDIR` to install somewhere else:

```sh
curl -fsSL https://raw.githubusercontent.com/glstep/claude-with/main/install.sh | BINDIR=$HOME/.local/bin sh
```

Prefer not to pipe scripts into your shell? Use one of the options below.

### Prebuilt binary (manual)

Download the binary for your platform from the [latest release](https://github.com/glstep/claude-with/releases/latest) (Linux, macOS, and Windows; amd64 and arm64), rename it to `ccw`, and put it somewhere on your `PATH`. For example, on macOS with Apple silicon:

```sh
curl -Lo ccw https://github.com/glstep/claude-with/releases/latest/download/ccw_darwin_arm64
chmod +x ccw
sudo mv ccw /usr/local/bin/
```

On Windows, download `ccw_windows_amd64.exe` (or `ccw_windows_arm64.exe`), rename it to `ccw.exe`, and put it on your `PATH`.

### go install

Requires Go 1.26+.

```sh
go install github.com/glstep/claude-with/cmd/ccw@latest
```

This installs a `ccw` binary to `$(go env GOPATH)/bin` (make sure that's on your `PATH`).

### Build from source

```sh
git clone https://github.com/glstep/claude-with.git
cd claude-with
go build -o ccw ./cmd/ccw
```

## Quick start

```sh
ccw init          # create a starter config file
ccw --list        # see the profiles you have
ccw local "hi"    # run `claude "hi"` using the "local" profile's env
```

## Configuration

`ccw` reads a TOML config file from the first of these that applies:

1. `$CLAUDE_WITH_CONFIG` (an exact file path), if set
2. `$XDG_CONFIG_HOME/claude-with/config.toml`, if `$XDG_CONFIG_HOME` is set
3. `~/.config/claude-with/config.toml` otherwise

Run `ccw init` to create a starter file at the resolved path (see [`ccw init`](#ccw-init) below).

### File format

```toml
default_profile = "local"

[profiles.local]
base_url = "http://localhost:11434/v1"
model = "llama3.1"

[profiles.together]
base_url = "https://api.together.xyz/v1"
model = "meta-llama/Llama-3-70b-chat-hf"
api_key_env = "TOGETHER_API_KEY"

[profiles.together.env]
SOME_OTHER_VAR = "value"
```

#### Top-level keys

| Key | Description |
| --- | --- |
| `default_profile` | Name of the profile to use when none is given on the command line. |
| `[profiles.NAME]` | One table per profile; `NAME` is what you pass on the command line. |

#### Profile keys

| Key | Maps to | Description |
| --- | --- | --- |
| `base_url` | `ANTHROPIC_BASE_URL` | The endpoint `claude` should talk to. |
| `model` | `ANTHROPIC_MODEL` | The model name to request. |
| `api_key` | `ANTHROPIC_API_KEY` | The API key, stored directly in the file. **See the warning below.** |
| `api_key_env` | `ANTHROPIC_API_KEY` | Name of an environment variable to read the API key from at runtime. |
| `env` | *(arbitrary)* | A table of extra `NAME = "value"` env vars to set, for anything else you need. |

If both `api_key` and `api_key_env` are set, `api_key_env` wins. Profile values override variables already set in your shell, but empty values are skipped — a profile can't *unset* an inherited variable.

> **Don't commit real API keys.** If you plan to push your config file anywhere (a dotfiles repo, a gist, etc.), use `api_key_env` instead of `api_key`, and set the actual secret only in your shell environment (e.g. in `~/.bashrc`, `~/.zshrc`, or a local `.env` you don't commit). `ccw init` generates a starter file that shows this pattern.

## Usage

```text
ccw [profile_name] [args...]
```

- `profile_name` — which profile to use. If omitted, `default_profile` from the config is used. Note that the profile can only be omitted when no other arguments follow — the first argument is always read as the profile name.
- `args...` — everything else is forwarded as-is to the `claude` binary (e.g. a prompt, or `claude`'s own flags like `-p`).

### Commands and flags

| Flag / command | Description |
| --- | --- |
| `[profile_name]` | The profile to use. Defaults to `default_profile` in config if omitted. |
| `--help`, `-h` | Show the help message. |
| `--version`, `-v` | Show the `ccw` version. |
| `--list`, `-l` | List all available profiles. |
| `--dry-run` | Print what would happen without doing it. Works for launching `claude` and for `init`. Can appear anywhere in the arguments — which also means it can't be *forwarded* to `claude` itself. |
| `init` | Create a starter config file. See below. |

### Examples

```sh
# One-shot prompt using the "local" profile
ccw local "hi there"

# Use the default profile (a profile name can only be omitted
# when no other arguments are passed)
ccw

# Non-interactive (print) mode, forwarded straight to claude
ccw local -p "hi there"

# See what would happen without actually launching claude
ccw --dry-run local "hi there"

# List configured profiles
ccw --list

# Show the version
ccw --version
```

### `ccw init`

Creates a starter config file at the resolved config path (see [Configuration](#configuration)), including the parent directory if it doesn't exist yet.

```sh
ccw init             # fails if a config file already exists
ccw init --force     # overwrite an existing config file
ccw init --dry-run   # preview the file without writing anything
ccw init --help      # show init-specific help
```

Unrecognized arguments are rejected, so a typo can't silently create or overwrite a config.

### `--dry-run`

Resolves the profile and prints the command and environment variables that would be used, then exits without launching `claude`. Values of variables whose name contains `KEY`, `TOKEN`, `SECRET`, or `PASSWORD` (case-insensitive) are redacted in the output — this includes `ANTHROPIC_API_KEY` as well as anything you set via the `[profiles.NAME.env]` table. Still, give the output a glance before sharing it in case a secret hides behind an unconventional name.

```sh
$ ccw --dry-run local "hi"
Dry run mode. The following command would be executed:
Command: claude hi
Environment variables:
  ANTHROPIC_BASE_URL=http://localhost:11434/v1
  ANTHROPIC_MODEL=llama3.1
  ANTHROPIC_API_KEY=<REDACTED>
```

## Signals and exit codes

While `claude` runs, `ccw` ignores `Ctrl+C` (SIGINT) — `claude` uses it to interrupt the current generation without quitting, so `ccw` stays alive and waits until `claude` actually exits.

`ccw` forwards `claude`'s exit code as its own. If `claude` is terminated by a signal, `ccw` exits with the shell convention `128 + signal number`. If `ccw` itself fails (e.g. bad config, unknown profile), it exits `1` with an error message on stderr.

## License

[MIT](LICENSE)

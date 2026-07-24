# ccw (claude-with)

`ccw` is a small CLI wrapper around [`claude`](https://github.com/anthropics/claude-code) (the Claude Code CLI) that lets you point it at local or self-hosted LLM backends instead of Anthropic's API.

You define named **profiles** in a TOML config file — each one sets `ANTHROPIC_BASE_URL`, `ANTHROPIC_MODEL`, and `ANTHROPIC_API_KEY` — and `ccw` launches `claude` with the right environment for whichever profile you pick.

```sh
ccw local "hi"
```

## Installation

Requires Go 1.26+.

```sh
go install github.com/glstep/claude-with/cmd/ccw@latest
```

This installs a `ccw` binary to `$(go env GOPATH)/bin` (make sure that's on your `PATH`).

Alternatively, clone and build locally:

```sh
git clone https://github.com/glstep/claude-with.git
cd claude-with
go build -o ccw ./cmd/ccw
```

## Quick start

```sh
ccw init          # create a starter config file
ccw --list        # see the profiles you have
ccw local "hi"     # run `claude "hi"` using the "local" profile's env
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

If both `api_key` and `api_key_env` are set, `api_key_env` wins.

> **Don't commit real API keys.** If you plan to push your config file anywhere (a dotfiles repo, a gist, etc.), use `api_key_env` instead of `api_key`, and set the actual secret only in your shell environment (e.g. in `~/.bashrc`, `~/.zshrc`, or a local `.env` you don't commit). `ccw init` generates a starter file that shows this pattern.

## Usage

```text
ccw [profile_name] [args...]
```

- `profile_name` — which profile to use. If omitted, `default_profile` from the config is used.
- `args...` — everything else is forwarded as-is to the `claude` binary (e.g. a prompt, or `claude`'s own flags like `-p`).

### Commands and flags

| Flag / command | Description |
| --- | --- |
| `[profile_name]` | The profile to use. Defaults to `default_profile` in config if omitted. |
| `--help`, `-h` | Show the help message. |
| `--version`, `-v` | Show the `ccw` version. |
| `--list`, `-l` | List all available profiles. |
| `--dry-run` | Print the resolved command and env vars without actually running `claude`. Can appear anywhere in the arguments. |
| `init` | Create a starter config file. See below. |

### Examples

```sh
# Use the default profile, one-shot prompt
ccw local "hi there"

# Use the default profile explicitly
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
ccw init            # fails if a config file already exists
ccw init --force     # overwrite an existing config file
ccw init --help      # show init-specific help
```

### `--dry-run`

Resolves the profile and prints the command and environment variables that would be used, then exits without launching `claude`. Any `ANTHROPIC_API_KEY` value is redacted in the output, so it's safe to share.

```sh
$ ccw --dry-run local "hi"
Dry run mode. The following command would be executed:
Command: claude [hi]
Environment variables:
  ANTHROPIC_BASE_URL=http://localhost:11434/v1
  ANTHROPIC_MODEL=llama3.1
  ANTHROPIC_API_KEY=<REDACTED>
```

## Exit codes

`ccw` forwards `claude`'s exit code as its own. If `ccw` itself fails (e.g. bad config, unknown profile), it exits `1` with an error message on stderr.

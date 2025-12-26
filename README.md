# kakucho

Extend any CLI with custom subcommands — like `gh extension`, but for any command.

## Installation

```bash
go install github.com/qawatake/kakucho/cmd/kakucho@latest
```

## Setup

Add to your `~/.bashrc` or `~/.zshrc`:

```bash
eval "$(kakucho init bash)"  # or zsh
```

Then reload your shell:

```bash
source ~/.bashrc
```

## Usage

### Adding an extension

Create an executable named `{target}-{name}` and place it in your PATH:

```bash
# Example: Create git-hello
echo '#!/bin/bash
echo "Hello, ${1:-World}!"' > ~/bin/git-hello
chmod +x ~/bin/git-hello

# Register it with kakucho
kakucho add git hello -d "Say hello"
```

Now you can use it:

```bash
git hello        # prints "Hello, World!"
git hello Alice  # prints "Hello, Alice!"
```

### Viewing extensions

```bash
kakucho ls       # List all extensions
kakucho ls git   # List extensions for git only
```

### Help integration

Extensions automatically appear in the target CLI's help:

```bash
git help
# ... normal git help ...
#
# EXTENSIONS (via kakucho):
#   hello  Say hello
```

### Removing an extension

```bash
kakucho rm git hello
```

## Configuration

Config file location: `~/.config/kakucho/config.toml`

```toml
[git]
hello = "Say hello"
browse = "Open repository in browser"

[kubectl]
ctx = "Switch context quickly"
```

## Commands

| Command | Description |
|---------|-------------|
| `kakucho init <shell>` | Output shell integration script |
| `kakucho add <target> <name> -d <desc>` | Register an extension |
| `kakucho rm <target> <name>` | Remove an extension |
| `kakucho ls [target]` | List registered extensions |
| `kakucho version` | Show version |

## License

MIT

# shellector

`shellector` is an interactive Linux CLI for browsing and managing SSH hosts.
It reads `~/.ssh/config`, provides live search and keyboard navigation, and
starts the system SSH client for the selected host.

## Features

- Search configured hosts as you type.
- Navigate long host lists with the arrow keys.
- Connect through the system `ssh` command.
- Add password or private-key hosts.
- Edit existing hosts without discarding unrelated directives.
- Follow OpenSSH `Include` directives.
- Atomically update edited config files.

## Build and install

```sh
make
```

The executable is written to `bin/shellector`. To install it into `/bin`:

```sh
sudo make install
```

Run the test suite with `make test` and remove build output with `make clean`.

## Usage

```sh
shellector
```

Use a different SSH config file with:

```sh
shellector -f ./config
```

### Host list

| Key | Action |
| --- | --- |
| Type | Filter hosts |
| <kbd>↑</kbd> / <kbd>↓</kbd> | Change selection |
| <kbd>Enter</kbd> | Connect to the selected host |
| <kbd>Ctrl-A</kbd> | Add a host |
| <kbd>Ctrl-E</kbd> | Edit the selected host |
| <kbd>Esc</kbd> / <kbd>Ctrl-C</kbd> | Quit |

### Add and edit forms

| Key | Action |
| --- | --- |
| <kbd>Tab</kbd> / <kbd>Enter</kbd> | Move to the next field |
| <kbd>Shift-Tab</kbd> | Move to the previous field |
| <kbd>↑</kbd> / <kbd>↓</kbd> | Choose the authentication method |
| <kbd>Enter</kbd> on the final field | Save changes |
| <kbd>Esc</kbd> / <kbd>Ctrl-C</kbd> | Cancel |

Host forms include the alias, address, login user, port, and authentication
method. Private-key authentication also requires an identity-file path.

## SSH config behavior

- Passwords are never stored; `ssh` prompts for them when connecting.
- New config files and directories are created with restrictive permissions.
- Included config files are searched and edited at their original location.
- Wildcard and negated host patterns are hidden because they are matching rules,
  not directly connectable aliases.
- Editing preserves unrelated directives such as `ProxyJump`.
- A shared declaration such as `Host web-1 web-2` cannot be edited because
  changing it could unintentionally modify multiple aliases.

Run `shellector --help` for command-line help.

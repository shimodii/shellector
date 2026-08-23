# shellector

A small interactive SSH host picker. It reads named hosts from `~/.ssh/config`,
lets you filter them as you type, and starts `ssh <host>` when you press Enter.

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

Run `shellector`, type to search, use <kbd>↑</kbd>/<kbd>↓</kbd> to move, and
press <kbd>Enter</kbd> to connect. Press <kbd>Esc</kbd> or <kbd>Ctrl-C</kbd> to
quit. To read another config file, use `shellector -f ./config`.

Wildcard-only SSH patterns are intentionally hidden because they are rules,
not directly connectable host names. `Include` directives are followed.

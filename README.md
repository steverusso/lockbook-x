# lockbook-x

Collection of [Lockbook](https://github.com/lockbook/lockbook) extensions and experiments.
A project such as [`go-lockbook`](./go-lockbook) has the hope of being merged into the
official Lockbook source tree one day. Other ones such as `lbcli` (and other Go projects)
will likely live here forever.

## Projects

* [`go-lockbook`](./go-lockbook) - Go bindings for the core Lockbook library `lb-rs`.
* [`lbcli`](./lbcli) - A CLI implemented in Go using the Go bindings.
* [`lbgui`](./lbcli) - A GUI implemented in Go using the Go bindings and GioUI.

## Adopted Upstream

* [`lbcli-rs`](https://github.com/steverusso/lockbook-x/tree/b8803ebdc0928eafa14d414b842de20fc0573f99/lbcli-rs)
  was a simple and practical CLI that originated here. It was
  [removed](https://github.com/steverusso/lockbook-x/pull/12) after the vast majority of
  its functionality and design principles were [integrated into the official Lockbook
  CLI](https://github.com/lockbook/lockbook/pull/1561).
* [`c_interface_v2`](https://github.com/steverusso/lockbook-x/tree/33317a4329f3ba6795f89f051b27e78550468715/c_interface_v2)
  was an overhauled and improved FFI wrapper for `lockbook_core`. It was
  [removed](https://github.com/steverusso/lockbook-x/commit/da3b484609db8ac607eb337c250eadd1f1c07bc2)
  a couple months after being [adopted in its
  entirety](https://github.com/lockbook/lockbook/pull/1715) upstream.

## Getting Started

This project uses [Task](https://taskfile.dev/) to run various build and maintenance
commands. Once this repo is cloned and [Task is
installed](https://taskfile.dev/installation/), simply run `task init` to pull in the
Lockbook submodule files and install the necessary dependency programs.

## License

This is free and unencumbered software released into the public domain. Please see the
[UNLICENSE](./UNLICENSE) file for more information.

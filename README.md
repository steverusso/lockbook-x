# lockbook-x

Collection of [Lockbook](https://github.com/lockbook/lockbook) extensions and
experiments. Projects such as `c_interface_v2` and `go-lockbook` have the hope
of being merged into the official Lockbook repository one day. Other ones such
as `lbcli` (and other Go projects) will likely live here forever.

## Projects

* [`c_interface_v2`](./c_interface_v2) - C FFI (in Rust) for [`lockook_core`](https://github.com/lockbook/lockbook/tree/master/libs/core).
* [`go-lockbook`](./go-lockbook) - Go bindings for Lockbook core using `c_interface_v2`.
* [`lbcli`](./lbcli) - A CLI written in Go using the Go bindings.

## Adopted Upstream

* [`lbcli-rs`](https://github.com/steverusso/lockbook-x/tree/b8803ebdc0928eafa14d414b842de20fc0573f99/lbcli-rs)
  was a simple and practical CLI that originated here. It was
  [removed](https://github.com/steverusso/lockbook-x/pull/12) after the vast
  majority of its functionality and design principles were [integrated into the
  official Lockbook CLI](https://github.com/lockbook/lockbook/pull/1561).

## License

This is free and unencumbered software released into the public domain. Please
see the [UNLICENSE](./UNLICENSE) file for more information.

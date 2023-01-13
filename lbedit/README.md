# lbedit

I made a vi-like markdown editor using [Gio](https://gioui.org/) called
[`mdedit`](https://github.com/steverusso/mdedit). The bulk of the functionality
is in a pkg that works with an fs-like interface as the base. Therefore,
`lbedit` here just implements that interface for lockbook core (so an `mdedit`
session can use it) and handles the top-level window events.

This is alpha level software and not very useful yet. It only syncs on startup
and when a file is written, and it has almost _no other functionality_ outside
of exploring files and editing & previewing markdown files.

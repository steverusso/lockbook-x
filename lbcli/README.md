# lbcli

A Lockbook CLI written in Go.

## Development

To build, simply run `task lbcli` from anywhere in the repository. That will build
`lockbook-core`, generate the C header, and build the Go bindings before building this.
The `lbcli` binary will appear in this directory.

## Usage

### Get Started

To get started with a new account, use the `init` command:
```shell
lbcli init

# or to restore an account with a private key:

<your_key_to_stdin> | lbcli init --restore
```

### Create some files

Easily create files with the `mkdir` and `mkdoc` commands. They will
auto-adjust the provided paths to correspond to the correct file type and will
not fail if the file path already exists. For example:

```shell
# This will still create a directory even though there is no trailing slash.
lbcli mkdir path/to/dir

# This will still create a document even though there is a trailing slash.
lbcli mkdoc path/to/doc.md/

# This won't fail.
lbcli mkdoc path/to/doc.md/
```

### Listing files

```shell
# This will list files in root (/).
lbcli ls

# This will just list files in 'dir' along with their ID prefixes and shared info.
lbcli ls -l path/to/dir

# This is the current cli list command. There are the optional '--dirs' & '--docs' filters.
lbcli ls -r --paths
```

### Other cool things

* `ls --tree`
* `jot` command
* `write` command

# lbcli

A simpler and more practical CLI for lockbook.

## Why

I wanted to have several useful capabilities, especially around development:

1. Reliably and quickly list files within a certain directory, and optionally have additional information such as an ID and/or the shared status.
2. Export arbitrary files. You could backup root and print documents to stdout, but beyond that (for example, exporting an entire folder), you have to do a good bit of scripting.
3. Make a directory or a document if they don't exist, and not fail if they do exist.
4. Read from stdin and write to file.
5. Aside from the above functional capabilities, this was an opportunity to implement a few random personal preferences:
  * mkdir & mkdoc won't fail if dir or doc already exists.
  * more 'unix-like' names for commands (print -> cat, remove -> rm, list -> ls, move -> mv).
  * more intuitive names for other things (copy -> import, get-usage -> usage, import account -> restore account).
  * lowercased output when possible.

Additionally, the extremely verbose error handling felt like it was getting in
the way of improving the source code in the current CLI and developing out new
features. Getting this out of the way within _this_ codebase is an experiment
in and of itself (going great so far).

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

* print pending shares in a readable table
* read-only shares
* accept a share using a custom name
* nice output for `debug finfo` by default (with option for raw)

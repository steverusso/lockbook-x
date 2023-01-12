# lbcli

A simpler and more practical CLI for Lockbook.

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

Additionally, the extremely verbose error handling felt like it was getting in the way of
improving the source code and developing out new features in the current CLI. Getting this
out of the way within _this_ codebase is an experiment in and of itself (going great so
far).

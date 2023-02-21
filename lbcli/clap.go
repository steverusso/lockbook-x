// This file is generated via 'go generate'; DO NOT EDIT
package main

import (
	"fmt"
	"os"
	"strings"
)

func claperr(format string, a ...any) {
	format = "\033[1;31merror:\033[0m " + format
	fmt.Fprintf(os.Stderr, format, a...)
}

func exitEmptyOpt() {
	claperr("emtpy option ('-') found\n")
	os.Exit(1)
}

type clapUsagePrinter interface {
	printUsage(to *os.File)
}

func exitMissingArg(u clapUsagePrinter, name string) {
	claperr("not enough args: no \033[1;33m%s\033[0m provided\n", name)
	u.printUsage(os.Stderr)
	os.Exit(1)
}

func exitUsgGood(u clapUsagePrinter) {
	u.printUsage(os.Stdout)
	os.Exit(0)
}

func optParts(arg string) (string, string) {
	if arg == "-" {
		exitEmptyOpt()
	}
	if arg[0] == '-' {
		arg = arg[1:]
	}
	if arg[0] == '-' {
		arg = arg[1:]
	}
	name := arg
	eqVal := ""
	if eqIdx := strings.IndexByte(name, '='); eqIdx != -1 {
		name = arg[:eqIdx]
		eqVal = arg[eqIdx+1:]
	}
	return name, eqVal
}

func exitUnknownCmd(u clapUsagePrinter, name string) {
	claperr("unknown command '%s'\n", name)
	u.printUsage(os.Stderr)
	os.Exit(1)
}

func (*acctRestoreCmd) printUsage(to *os.File) {
	fmt.Fprintf(to, `%[1]s acct restore - restore an existing account from its secret account string

usage:
   restore [options]

options:
   --no-sync    don't perform the initial sync
   --help, -h   show this help message
`, os.Args[0])
}

func (c *acctRestoreCmd) parse(args []string) {
	var i int
	for ; i < len(args); i++ {
		if args[i][0] != '-' {
			break
		}
		if args[i] == "--" {
			i++
			break
		}
		k, eqv := optParts(args[i][1:])
		switch k {
		case "no-sync":
			c.noSync = (eqv == "" || eqv == "true")
		case "help", "h":
			exitUsgGood(c)
		}
	}
}

func (*acctPrivKeyCmd) printUsage(to *os.File) {
	fmt.Fprintf(to, `%[1]s acct privkey - print out the private key for this lockbook

usage:
   privkey [options]

options:
   --help, -h   show this help message
`, os.Args[0])
}

func (c *acctPrivKeyCmd) parse(args []string) {
	var i int
	for ; i < len(args); i++ {
		if args[i][0] != '-' {
			break
		}
		if args[i] == "--" {
			i++
			break
		}
		k, _ := optParts(args[i][1:])
		switch k {
		case "help", "h":
			exitUsgGood(c)
		}
	}
}

func (*acctStatusCmd) printUsage(to *os.File) {
	fmt.Fprintf(to, `%[1]s acct status - overview of your account

usage:
   status [options]

options:
   --help, -h   show this help message
`, os.Args[0])
}

func (c *acctStatusCmd) parse(args []string) {
	var i int
	for ; i < len(args); i++ {
		if args[i][0] != '-' {
			break
		}
		if args[i] == "--" {
			i++
			break
		}
		k, _ := optParts(args[i][1:])
		switch k {
		case "help", "h":
			exitUsgGood(c)
		}
	}
}

func (*acctSubscribeCmd) printUsage(to *os.File) {
	fmt.Fprintf(to, `%[1]s acct subscribe - create a new subscription with a credit card

usage:
   subscribe [options]

options:
   --help, -h   show this help message
`, os.Args[0])
}

func (c *acctSubscribeCmd) parse(args []string) {
	var i int
	for ; i < len(args); i++ {
		if args[i][0] != '-' {
			break
		}
		if args[i] == "--" {
			i++
			break
		}
		k, _ := optParts(args[i][1:])
		switch k {
		case "help", "h":
			exitUsgGood(c)
		}
	}
}

func (*acctUnsubscribeCmd) printUsage(to *os.File) {
	fmt.Fprintf(to, `%[1]s acct unsubscribe - cancel an existing subscription

usage:
   unsubscribe [options]

options:
   --help, -h   show this help message
`, os.Args[0])
}

func (c *acctUnsubscribeCmd) parse(args []string) {
	var i int
	for ; i < len(args); i++ {
		if args[i][0] != '-' {
			break
		}
		if args[i] == "--" {
			i++
			break
		}
		k, _ := optParts(args[i][1:])
		switch k {
		case "help", "h":
			exitUsgGood(c)
		}
	}
}

func (*acctCmd) printUsage(to *os.File) {
	fmt.Fprintf(to, `%[1]s acct - account related commands

usage:
   acct <command> [args...]

options:
   --help, -h   show this help message

commands:
   restore       restore an existing account from its secret account string
   privkey       print out the private key for this lockbook
   status        overview of your account
   subscribe     create a new subscription with a credit card
   unsubscribe   cancel an existing subscription
`, os.Args[0])
}

func (c *acctCmd) parse(args []string) {
	var i int
	for ; i < len(args); i++ {
		if args[i][0] != '-' {
			break
		}
		if args[i] == "--" {
			i++
			break
		}
		k, _ := optParts(args[i][1:])
		switch k {
		case "help", "h":
			exitUsgGood(c)
		}
	}
	if i >= len(args) {
		c.printUsage(os.Stderr)
		os.Exit(1)
	}
	switch args[i] {
	case "restore":
		c.restore = new(acctRestoreCmd)
		c.restore.parse(args[i+1:])
	case "privkey":
		c.privkey = new(acctPrivKeyCmd)
		c.privkey.parse(args[i+1:])
	case "status":
		c.status = new(acctStatusCmd)
		c.status.parse(args[i+1:])
	case "subscribe":
		c.subscribe = new(acctSubscribeCmd)
		c.subscribe.parse(args[i+1:])
	case "unsubscribe":
		c.unsubscribe = new(acctUnsubscribeCmd)
		c.unsubscribe.parse(args[i+1:])
	default:
		exitUnknownCmd(c, args[i])
	}
}

func (*catCmd) printUsage(to *os.File) {
	fmt.Fprintf(to, `%[1]s cat - print one or more documents to stdout

usage:
   cat [options] <target>

options:
   --help, -h   show this help message

arguments:
   <target>   lockbook file path or id
`, os.Args[0])
}

func (c *catCmd) parse(args []string) {
	var i int
	for ; i < len(args); i++ {
		if args[i][0] != '-' {
			break
		}
		if args[i] == "--" {
			i++
			break
		}
		k, _ := optParts(args[i][1:])
		switch k {
		case "help", "h":
			exitUsgGood(c)
		}
	}
	args = args[i:]
	if len(args) < 1 {
		exitMissingArg(c, "<target>")
	}
	c.target = args[0]
}

func (*debugFinfoCmd) printUsage(to *os.File) {
	fmt.Fprintf(to, `%[1]s debug finfo - view info about a target file

usage:
   finfo [options] <target>

options:
   --help, -h   show this help message

arguments:
   <target>   the target can be a file path, uuid, or uuid prefix
`, os.Args[0])
}

func (c *debugFinfoCmd) parse(args []string) {
	var i int
	for ; i < len(args); i++ {
		if args[i][0] != '-' {
			break
		}
		if args[i] == "--" {
			i++
			break
		}
		k, _ := optParts(args[i][1:])
		switch k {
		case "help", "h":
			exitUsgGood(c)
		}
	}
	args = args[i:]
	if len(args) < 1 {
		exitMissingArg(c, "<target>")
	}
	c.target = args[0]
}

func (*debugValidateCmd) printUsage(to *os.File) {
	fmt.Fprintf(to, `%[1]s debug validate - find invalid states within your lockbook

usage:
   validate [options]

options:
   --help, -h   show this help message
`, os.Args[0])
}

func (c *debugValidateCmd) parse(args []string) {
	var i int
	for ; i < len(args); i++ {
		if args[i][0] != '-' {
			break
		}
		if args[i] == "--" {
			i++
			break
		}
		k, _ := optParts(args[i][1:])
		switch k {
		case "help", "h":
			exitUsgGood(c)
		}
	}
}

func (*debugCmd) printUsage(to *os.File) {
	fmt.Fprintf(to, `%[1]s debug - investigative commands mainly intended for devs

usage:
   debug [options] <command>

options:
   --help, -h   show this help message

commands:
   finfo      view info about a target file
   validate   find invalid states within your lockbook
`, os.Args[0])
}

func (c *debugCmd) parse(args []string) {
	var i int
	for ; i < len(args); i++ {
		if args[i][0] != '-' {
			break
		}
		if args[i] == "--" {
			i++
			break
		}
		k, _ := optParts(args[i][1:])
		switch k {
		case "help", "h":
			exitUsgGood(c)
		}
	}
	if i >= len(args) {
		c.printUsage(os.Stderr)
		os.Exit(1)
	}
	switch args[i] {
	case "finfo":
		c.finfo = new(debugFinfoCmd)
		c.finfo.parse(args[i+1:])
	case "validate":
		c.validate = new(debugValidateCmd)
		c.validate.parse(args[i+1:])
	default:
		exitUnknownCmd(c, args[i])
	}
}

func (*drawingCmd) printUsage(to *os.File) {
	fmt.Fprintf(to, `%[1]s drawing - export a lockbook drawing as an image written to stdout

usage:
   drawing <target> [png|jpeg|pnm|tga|farbfeld|bmp]

options:
   --help, -h   show this help message

arguments:
   <target>   the drawing to export
   [imgfmt]   the format to convert the drawing into
`, os.Args[0])
}

func (c *drawingCmd) parse(args []string) {
	var i int
	for ; i < len(args); i++ {
		if args[i][0] != '-' {
			break
		}
		if args[i] == "--" {
			i++
			break
		}
		k, _ := optParts(args[i][1:])
		switch k {
		case "help", "h":
			exitUsgGood(c)
		}
	}
	args = args[i:]
	if len(args) < 1 {
		exitMissingArg(c, "<target>")
	}
	c.target = args[0]
	if len(args) < 2 {
		return
	}
	c.imgFmt = args[1]
}

func (*exportCmd) printUsage(to *os.File) {
	fmt.Fprintf(to, `%[1]s export - copy a lockbook file to your file system

usage:
   export <target> [dest-dir]

options:
   --verbose, -v   print out each file as it's being exported
   --help, -h      show this help message

arguments:
   <target>   lockbook file path or id
   [dest]     disk file path (defaults to working dir)
`, os.Args[0])
}

func (c *exportCmd) parse(args []string) {
	var i int
	for ; i < len(args); i++ {
		if args[i][0] != '-' {
			break
		}
		if args[i] == "--" {
			i++
			break
		}
		k, eqv := optParts(args[i][1:])
		switch k {
		case "verbose", "v":
			c.verbose = (eqv == "" || eqv == "true")
		case "help", "h":
			exitUsgGood(c)
		}
	}
	args = args[i:]
	if len(args) < 1 {
		exitMissingArg(c, "<target>")
	}
	c.target = args[0]
	if len(args) < 2 {
		return
	}
	c.dest = args[1]
}

func (*initCmd) printUsage(to *os.File) {
	fmt.Fprintf(to, `%[1]s init - create a lockbook account

usage:
   init [options]

options:
   --welcome    include the welcome document
   --no-sync    don't perform the initial sync
   --help, -h   show this help message
`, os.Args[0])
}

func (c *initCmd) parse(args []string) {
	var i int
	for ; i < len(args); i++ {
		if args[i][0] != '-' {
			break
		}
		if args[i] == "--" {
			i++
			break
		}
		k, eqv := optParts(args[i][1:])
		switch k {
		case "welcome":
			c.welcome = (eqv == "" || eqv == "true")
		case "no-sync":
			c.noSync = (eqv == "" || eqv == "true")
		case "help", "h":
			exitUsgGood(c)
		}
	}
}

func (*lsCmd) printUsage(to *os.File) {
	fmt.Fprintf(to, `%[1]s ls - list files in a directory

usage:
   ls [options] [target]

options:
   --short, -s       just display the name (or file path)
   --recursive, -r   recursively include all children of the target directory
   --paths           show absolute file paths instead of file names
   --dirs            only show folders
   --docs            only show documents
   --ids             show full uuids instead of prefixes
   --help, -h        show this help message

arguments:
   [target]   target directory (defaults to root)
`, os.Args[0])
}

func (c *lsCmd) parse(args []string) {
	var i int
	for ; i < len(args); i++ {
		if args[i][0] != '-' {
			break
		}
		if args[i] == "--" {
			i++
			break
		}
		k, eqv := optParts(args[i][1:])
		switch k {
		case "short", "s":
			c.short = (eqv == "" || eqv == "true")
		case "recursive", "r":
			c.recursive = (eqv == "" || eqv == "true")
		case "paths":
			c.paths = (eqv == "" || eqv == "true")
		case "dirs":
			c.onlyDirs = (eqv == "" || eqv == "true")
		case "docs":
			c.onlyDocs = (eqv == "" || eqv == "true")
		case "ids":
			c.fullIDs = (eqv == "" || eqv == "true")
		case "help", "h":
			exitUsgGood(c)
		}
	}
	args = args[i:]
	if len(args) < 1 {
		return
	}
	c.target = args[0]
}

func (*mkdirCmd) printUsage(to *os.File) {
	fmt.Fprintf(to, `%[1]s mkdir - create a directory or do nothing if it exists

usage:
   mkdir [options] <path>

options:
   --help, -h   show this help message

arguments:
   <path>   a path at which to create the directory
`, os.Args[0])
}

func (c *mkdirCmd) parse(args []string) {
	var i int
	for ; i < len(args); i++ {
		if args[i][0] != '-' {
			break
		}
		if args[i] == "--" {
			i++
			break
		}
		k, _ := optParts(args[i][1:])
		switch k {
		case "help", "h":
			exitUsgGood(c)
		}
	}
	args = args[i:]
	if len(args) < 1 {
		exitMissingArg(c, "<path>")
	}
	c.path = args[0]
}

func (*mkdocCmd) printUsage(to *os.File) {
	fmt.Fprintf(to, `%[1]s mkdoc - create a document or do nothing if it exists

usage:
   mkdoc [options] <path>

options:
   --help, -h   show this help message

arguments:
   <path>   a path at which to create the document
`, os.Args[0])
}

func (c *mkdocCmd) parse(args []string) {
	var i int
	for ; i < len(args); i++ {
		if args[i][0] != '-' {
			break
		}
		if args[i] == "--" {
			i++
			break
		}
		k, _ := optParts(args[i][1:])
		switch k {
		case "help", "h":
			exitUsgGood(c)
		}
	}
	args = args[i:]
	if len(args) < 1 {
		exitMissingArg(c, "<path>")
	}
	c.path = args[0]
}

func (*mvCmd) printUsage(to *os.File) {
	fmt.Fprintf(to, `%[1]s mv - move a file to another parent

usage:
   mv [options] <src> <dest>

options:
   --help, -h   show this help message

arguments:
   <src>    the file to move
   <dest>   the destination directory
`, os.Args[0])
}

func (c *mvCmd) parse(args []string) {
	var i int
	for ; i < len(args); i++ {
		if args[i][0] != '-' {
			break
		}
		if args[i] == "--" {
			i++
			break
		}
		k, _ := optParts(args[i][1:])
		switch k {
		case "help", "h":
			exitUsgGood(c)
		}
	}
	args = args[i:]
	if len(args) < 1 {
		exitMissingArg(c, "<src>")
	}
	if len(args) < 2 {
		exitMissingArg(c, "<dest>")
	}
	c.src = args[0]
	c.dest = args[1]
}

func (*renameCmd) printUsage(to *os.File) {
	fmt.Fprintf(to, `%[1]s rename - rename a file

usage:
   rename [-f] <target> [new-name]

options:
   --force, -f   non-interactive (fail instead of prompting for corrections)
   --help, -h    show this help message

arguments:
   <target>    the file to rename
   <newname>   the desired new name
`, os.Args[0])
}

func (c *renameCmd) parse(args []string) {
	var i int
	for ; i < len(args); i++ {
		if args[i][0] != '-' {
			break
		}
		if args[i] == "--" {
			i++
			break
		}
		k, eqv := optParts(args[i][1:])
		switch k {
		case "force", "f":
			c.force = (eqv == "" || eqv == "true")
		case "help", "h":
			exitUsgGood(c)
		}
	}
	args = args[i:]
	if len(args) < 1 {
		exitMissingArg(c, "<target>")
	}
	if len(args) < 2 {
		exitMissingArg(c, "<newname>")
	}
	c.target = args[0]
	c.newName = args[1]
}

func (*rmCmd) printUsage(to *os.File) {
	fmt.Fprintf(to, `%[1]s rm - delete a file

usage:
   rm [options] <target>

options:
   --force, -f   don't prompt for confirmation
   --help, -h    show this help message

arguments:
   <target>   lockbook path or id to delete
`, os.Args[0])
}

func (c *rmCmd) parse(args []string) {
	var i int
	for ; i < len(args); i++ {
		if args[i][0] != '-' {
			break
		}
		if args[i] == "--" {
			i++
			break
		}
		k, eqv := optParts(args[i][1:])
		switch k {
		case "force", "f":
			c.force = (eqv == "" || eqv == "true")
		case "help", "h":
			exitUsgGood(c)
		}
	}
	args = args[i:]
	if len(args) < 1 {
		exitMissingArg(c, "<target>")
	}
	c.target = args[0]
}

func (*shareCreateCmd) printUsage(to *os.File) {
	fmt.Fprintf(to, `%[1]s share create - share a file with another lockbook user

usage:
   create [options] <target> <username>

options:
   --ro         the other user will not be able to edit the file
   --help, -h   show this help message

arguments:
   <target>     the path or id of the lockbook file you'd like to share
   <username>   the username of the other lockbook user
`, os.Args[0])
}

func (c *shareCreateCmd) parse(args []string) {
	var i int
	for ; i < len(args); i++ {
		if args[i][0] != '-' {
			break
		}
		if args[i] == "--" {
			i++
			break
		}
		k, eqv := optParts(args[i][1:])
		switch k {
		case "ro":
			c.readOnly = (eqv == "" || eqv == "true")
		case "help", "h":
			exitUsgGood(c)
		}
	}
	args = args[i:]
	if len(args) < 1 {
		exitMissingArg(c, "<target>")
	}
	if len(args) < 2 {
		exitMissingArg(c, "<username>")
	}
	c.target = args[0]
	c.username = args[1]
}

func (*sharePendingCmd) printUsage(to *os.File) {
	fmt.Fprintf(to, `%[1]s share pending - list pending shares

usage:
   pending [options]

options:
   --ids        print full uuids instead of prefixes
   --help, -h   show this help message
`, os.Args[0])
}

func (c *sharePendingCmd) parse(args []string) {
	var i int
	for ; i < len(args); i++ {
		if args[i][0] != '-' {
			break
		}
		if args[i] == "--" {
			i++
			break
		}
		k, eqv := optParts(args[i][1:])
		switch k {
		case "ids":
			c.fullIDs = (eqv == "" || eqv == "true")
		case "help", "h":
			exitUsgGood(c)
		}
	}
}

func (*shareAcceptCmd) printUsage(to *os.File) {
	fmt.Fprintf(to, `%[1]s share accept - accept a pending share

usage:
   accept [options] <target> <dest> [newname]

options:
   --help, -h   show this help message

arguments:
   <target>    id or id prefix of the pending share to accept
   <dest>      where to place this in your file tree
   [newname]   name this file something else
`, os.Args[0])
}

func (c *shareAcceptCmd) parse(args []string) {
	var i int
	for ; i < len(args); i++ {
		if args[i][0] != '-' {
			break
		}
		if args[i] == "--" {
			i++
			break
		}
		k, _ := optParts(args[i][1:])
		switch k {
		case "help", "h":
			exitUsgGood(c)
		}
	}
	args = args[i:]
	if len(args) < 1 {
		exitMissingArg(c, "<target>")
	}
	if len(args) < 2 {
		exitMissingArg(c, "<dest>")
	}
	c.target = args[0]
	c.dest = args[1]
	if len(args) < 3 {
		return
	}
	c.newName = args[2]
}

func (*shareRejectCmd) printUsage(to *os.File) {
	fmt.Fprintf(to, `%[1]s share reject - reject a pending share

usage:
   reject [options] <target>

options:
   --help, -h   show this help message

arguments:
   <target>   id or id prefix of a pending share
`, os.Args[0])
}

func (c *shareRejectCmd) parse(args []string) {
	var i int
	for ; i < len(args); i++ {
		if args[i][0] != '-' {
			break
		}
		if args[i] == "--" {
			i++
			break
		}
		k, _ := optParts(args[i][1:])
		switch k {
		case "help", "h":
			exitUsgGood(c)
		}
	}
	args = args[i:]
	if len(args) < 1 {
		exitMissingArg(c, "<target>")
	}
	c.target = args[0]
}

func (*shareCmd) printUsage(to *os.File) {
	fmt.Fprintf(to, `%[1]s share - sharing related commands

usage:
   share [options] <command>

options:
   --help, -h   show this help message

commands:
   create    share a file with another lockbook user
   pending   list pending shares
   accept    accept a pending share
   reject    reject a pending share
`, os.Args[0])
}

func (c *shareCmd) parse(args []string) {
	var i int
	for ; i < len(args); i++ {
		if args[i][0] != '-' {
			break
		}
		if args[i] == "--" {
			i++
			break
		}
		k, _ := optParts(args[i][1:])
		switch k {
		case "help", "h":
			exitUsgGood(c)
		}
	}
	if i >= len(args) {
		c.printUsage(os.Stderr)
		os.Exit(1)
	}
	switch args[i] {
	case "create":
		c.create = new(shareCreateCmd)
		c.create.parse(args[i+1:])
	case "pending":
		c.pending = new(sharePendingCmd)
		c.pending.parse(args[i+1:])
	case "accept":
		c.accept = new(shareAcceptCmd)
		c.accept.parse(args[i+1:])
	case "reject":
		c.reject = new(shareRejectCmd)
		c.reject.parse(args[i+1:])
	default:
		exitUnknownCmd(c, args[i])
	}
}

func (*statusCmd) printUsage(to *os.File) {
	fmt.Fprintf(to, `%[1]s status - which operations a sync would perform

usage:
   status [options]

options:
   --help, -h   show this help message
`, os.Args[0])
}

func (c *statusCmd) parse(args []string) {
	var i int
	for ; i < len(args); i++ {
		if args[i][0] != '-' {
			break
		}
		if args[i] == "--" {
			i++
			break
		}
		k, _ := optParts(args[i][1:])
		switch k {
		case "help", "h":
			exitUsgGood(c)
		}
	}
}

func (*syncCmd) printUsage(to *os.File) {
	fmt.Fprintf(to, `%[1]s sync - get updates from the server and push changes

usage:
   sync [options]

options:
   --verbose, -v   output every sync step and progress
   --help, -h      show this help message
`, os.Args[0])
}

func (c *syncCmd) parse(args []string) {
	var i int
	for ; i < len(args); i++ {
		if args[i][0] != '-' {
			break
		}
		if args[i] == "--" {
			i++
			break
		}
		k, eqv := optParts(args[i][1:])
		switch k {
		case "verbose", "v":
			c.verbose = (eqv == "" || eqv == "true")
		case "help", "h":
			exitUsgGood(c)
		}
	}
}

func (*usageCmd) printUsage(to *os.File) {
	fmt.Fprintf(to, `%[1]s usage - local and server disk utilization (uncompressed and compressed)

usage:
   usage [-e]

options:
   --exact, -e   show amounts in bytes
   --help, -h    show this help message
`, os.Args[0])
}

func (c *usageCmd) parse(args []string) {
	var i int
	for ; i < len(args); i++ {
		if args[i][0] != '-' {
			break
		}
		if args[i] == "--" {
			i++
			break
		}
		k, eqv := optParts(args[i][1:])
		switch k {
		case "exact", "e":
			c.exact = (eqv == "" || eqv == "true")
		case "help", "h":
			exitUsgGood(c)
		}
	}
}

func (*whoamiCmd) printUsage(to *os.File) {
	fmt.Fprintf(to, `%[1]s whoami - print user information for this lockbook

usage:
   whoami [-l]

options:
   --long, -l   prints the data directory and server url as well
   --help, -h   show this help message
`, os.Args[0])
}

func (c *whoamiCmd) parse(args []string) {
	var i int
	for ; i < len(args); i++ {
		if args[i][0] != '-' {
			break
		}
		if args[i] == "--" {
			i++
			break
		}
		k, eqv := optParts(args[i][1:])
		switch k {
		case "long", "l":
			c.long = (eqv == "" || eqv == "true")
		case "help", "h":
			exitUsgGood(c)
		}
	}
}

func (*writeCmd) printUsage(to *os.File) {
	fmt.Fprintf(to, `%[1]s write - write data from stdin to a lockbook document

usage:
   write [--trunc] <target>

options:
   --trunc      truncate the file instead of appending to it
   --help, -h   show this help message

arguments:
   <target>   lockbook path or id to write
`, os.Args[0])
}

func (c *writeCmd) parse(args []string) {
	var i int
	for ; i < len(args); i++ {
		if args[i][0] != '-' {
			break
		}
		if args[i] == "--" {
			i++
			break
		}
		k, eqv := optParts(args[i][1:])
		switch k {
		case "trunc":
			c.trunc = (eqv == "" || eqv == "true")
		case "help", "h":
			exitUsgGood(c)
		}
	}
	args = args[i:]
	if len(args) < 1 {
		exitMissingArg(c, "<target>")
	}
	c.target = args[0]
}

func (*lbcli) printUsage(to *os.File) {
	fmt.Fprintf(to, `%[1]s - an unofficial lockbook cli implemented in go

usage:
   %[1]s [options] <command>

options:
   --help, -h   show this help message

commands:
   acct      account related commands
   cat       print one or more documents to stdout
   debug     investigative commands mainly intended for devs
   drawing   export a lockbook drawing as an image written to stdout
   export    copy a lockbook file to your file system
   init      create a lockbook account
   ls        list files in a directory
   mkdir     create a directory or do nothing if it exists
   mkdoc     create a document or do nothing if it exists
   mv        move a file to another parent
   rename    rename a file
   rm        delete a file
   share     sharing related commands
   status    which operations a sync would perform
   sync      get updates from the server and push changes
   usage     local and server disk utilization (uncompressed and compressed)
   whoami    print user information for this lockbook
   write     write data from stdin to a lockbook document
`, os.Args[0])
}

func (c *lbcli) parse(args []string) {
	if len(args) > 0 && len(args) == len(os.Args) {
		args = args[1:]
	}
	var i int
	for ; i < len(args); i++ {
		if args[i][0] != '-' {
			break
		}
		if args[i] == "--" {
			i++
			break
		}
		k, _ := optParts(args[i][1:])
		switch k {
		case "help", "h":
			exitUsgGood(c)
		}
	}
	if i >= len(args) {
		c.printUsage(os.Stderr)
		os.Exit(1)
	}
	switch args[i] {
	case "acct":
		c.acct = new(acctCmd)
		c.acct.parse(args[i+1:])
	case "cat":
		c.cat = new(catCmd)
		c.cat.parse(args[i+1:])
	case "debug":
		c.debug = new(debugCmd)
		c.debug.parse(args[i+1:])
	case "drawing":
		c.drawing = new(drawingCmd)
		c.drawing.parse(args[i+1:])
	case "export":
		c.export = new(exportCmd)
		c.export.parse(args[i+1:])
	case "init":
		c.init = new(initCmd)
		c.init.parse(args[i+1:])
	case "ls":
		c.ls = new(lsCmd)
		c.ls.parse(args[i+1:])
	case "mkdir":
		c.mkdir = new(mkdirCmd)
		c.mkdir.parse(args[i+1:])
	case "mkdoc":
		c.mkdoc = new(mkdocCmd)
		c.mkdoc.parse(args[i+1:])
	case "mv":
		c.mv = new(mvCmd)
		c.mv.parse(args[i+1:])
	case "rename":
		c.rename = new(renameCmd)
		c.rename.parse(args[i+1:])
	case "rm":
		c.rm = new(rmCmd)
		c.rm.parse(args[i+1:])
	case "share":
		c.share = new(shareCmd)
		c.share.parse(args[i+1:])
	case "status":
		c.status = new(statusCmd)
		c.status.parse(args[i+1:])
	case "sync":
		c.sync = new(syncCmd)
		c.sync.parse(args[i+1:])
	case "usage":
		c.usage = new(usageCmd)
		c.usage.parse(args[i+1:])
	case "whoami":
		c.whoami = new(whoamiCmd)
		c.whoami.parse(args[i+1:])
	case "write":
		c.write = new(writeCmd)
		c.write.parse(args[i+1:])
	default:
		exitUnknownCmd(c, args[i])
	}
}

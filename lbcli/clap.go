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

func exitMissingArg(name, u string) {
	claperr("not enough args: no \033[1;33m%s\033[0m provided\n", name)
	fmt.Fprint(os.Stderr, u)
	os.Exit(1)
}

func exitUnknownCmd(name, u string) {
	claperr("unknown command '%s'\n", name)
	fmt.Fprint(os.Stderr, u)
	os.Exit(1)
}

func exitUsgGood(u string) {
	fmt.Fprint(os.Stdout, u)
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

const acctRestoreCmdUsage = `name:
   lbcli acct restore - restore an existing account from its secret account string

usage:
   restore [options]

options:
   --no-sync    don't perform the initial sync
   --help, -h   show this help message
`

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
		name, eqv := optParts(args[i][1:])
		switch name {
		case "no-sync":
			c.noSync = (eqv == "" || eqv == "true")
		case "help", "h":
			exitUsgGood(acctRestoreCmdUsage)
		}
	}
}

const acctPrivKeyCmdUsage = `name:
   lbcli acct privkey - print out the private key for this lockbook

usage:
   privkey [options]

options:
   --help, -h   show this help message
`

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
		name, _ := optParts(args[i][1:])
		switch name {
		case "help", "h":
			exitUsgGood(acctPrivKeyCmdUsage)
		}
	}
}

const acctStatusCmdUsage = `name:
   lbcli acct status - overview of your account

usage:
   status [options]

options:
   --help, -h   show this help message
`

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
		name, _ := optParts(args[i][1:])
		switch name {
		case "help", "h":
			exitUsgGood(acctStatusCmdUsage)
		}
	}
}

const acctSubscribeCmdUsage = `name:
   lbcli acct subscribe - create a new subscription with a credit card

usage:
   subscribe [options]

options:
   --help, -h   show this help message
`

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
		name, _ := optParts(args[i][1:])
		switch name {
		case "help", "h":
			exitUsgGood(acctSubscribeCmdUsage)
		}
	}
}

const acctUnsubscribeCmdUsage = `name:
   lbcli acct unsubscribe - cancel an existing subscription

usage:
   unsubscribe [options]

options:
   --help, -h   show this help message
`

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
		name, _ := optParts(args[i][1:])
		switch name {
		case "help", "h":
			exitUsgGood(acctUnsubscribeCmdUsage)
		}
	}
}

const acctCmdUsage = `name:
   lbcli acct - account related commands

usage:
   acct <command> [args...]

commands:
   restore      restore an existing account from its secret account string
   privkey      print out the private key for this lockbook
   status       overview of your account
   subscribe    create a new subscription with a credit card
   unsubscribe  cancel an existing subscription

options:
   --help, -h   show this help message
`

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
		name, _ := optParts(args[i][1:])
		switch name {
		case "help", "h":
			exitUsgGood(acctCmdUsage)
		}
	}
	if i >= len(args) {
		fmt.Fprint(os.Stderr, acctCmdUsage)
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
		exitUnknownCmd(args[i], acctCmdUsage)
	}
}

const catCmdUsage = `name:
   lbcli cat - print one or more documents to stdout

usage:
   cat [options] <target>

arguments:
   <target>   lockbook file path or id

options:
   --help, -h   show this help message
`

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
		name, _ := optParts(args[i][1:])
		switch name {
		case "help", "h":
			exitUsgGood(catCmdUsage)
		}
	}
	args = args[i:]
	if len(args) < 1 {
		exitMissingArg("<target>", catCmdUsage)
	}
	c.target = args[0]
}

const debugFinfoCmdUsage = `name:
   lbcli debug finfo - view info about a target file

usage:
   finfo [options] <target>

arguments:
   <target>   the target can be a file path

options:
   --help, -h   show this help message
`

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
		name, _ := optParts(args[i][1:])
		switch name {
		case "help", "h":
			exitUsgGood(debugFinfoCmdUsage)
		}
	}
	args = args[i:]
	if len(args) < 1 {
		exitMissingArg("<target>", debugFinfoCmdUsage)
	}
	c.target = args[0]
}

const debugValidateCmdUsage = `name:
   lbcli debug validate - find invalid states within your lockbook

usage:
   validate [options]

options:
   --help, -h   show this help message
`

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
		name, _ := optParts(args[i][1:])
		switch name {
		case "help", "h":
			exitUsgGood(debugValidateCmdUsage)
		}
	}
}

const debugCmdUsage = `name:
   lbcli debug - investigative commands mainly intended for devs

usage:
   debug [options] <command>

commands:
   finfo     view info about a target file
   validate  find invalid states within your lockbook

options:
   --help, -h   show this help message
`

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
		name, _ := optParts(args[i][1:])
		switch name {
		case "help", "h":
			exitUsgGood(debugCmdUsage)
		}
	}
	if i >= len(args) {
		fmt.Fprint(os.Stderr, debugCmdUsage)
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
		exitUnknownCmd(args[i], debugCmdUsage)
	}
}

const drawingCmdUsage = `name:
   lbcli drawing - export a lockbook drawing as an image written to stdout

usage:
   drawing <target> [png|jpeg|pnm|tga|farbfeld|bmp]

arguments:
   <target>   the drawing to export
   [imgfmt]   the format to convert the drawing into

options:
   --help, -h   show this help message
`

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
		name, _ := optParts(args[i][1:])
		switch name {
		case "help", "h":
			exitUsgGood(drawingCmdUsage)
		}
	}
	args = args[i:]
	if len(args) < 1 {
		exitMissingArg("<target>", drawingCmdUsage)
	}
	c.target = args[0]
	if len(args) < 2 {
		return
	}
	c.imgFmt = args[1]
}

const exportCmdUsage = `name:
   lbcli export - copy a lockbook file to your file system

usage:
   export <target> [dest-dir]

arguments:
   <target>   lockbook file path or id
   [dest]     disk file path (defaults to working dir)

options:
   --verbose, -v   print out each file as it's being exported
   --help, -h      show this help message
`

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
		name, eqv := optParts(args[i][1:])
		switch name {
		case "verbose", "v":
			c.verbose = (eqv == "" || eqv == "true")
		case "help", "h":
			exitUsgGood(exportCmdUsage)
		}
	}
	args = args[i:]
	if len(args) < 1 {
		exitMissingArg("<target>", exportCmdUsage)
	}
	c.target = args[0]
	if len(args) < 2 {
		return
	}
	c.dest = args[1]
}

const initCmdUsage = `name:
   lbcli init - create a lockbook account

usage:
   init [options]

options:
   --welcome    include the welcome document
   --no-sync    don't perform the initial sync
   --help, -h   show this help message
`

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
		name, eqv := optParts(args[i][1:])
		switch name {
		case "welcome":
			c.welcome = (eqv == "" || eqv == "true")
		case "no-sync":
			c.noSync = (eqv == "" || eqv == "true")
		case "help", "h":
			exitUsgGood(initCmdUsage)
		}
	}
}

const lsCmdUsage = `name:
   lbcli ls - list files in a directory

usage:
   ls [options] [target]

arguments:
   [target]   target directory (defaults to root)

options:
   --short, -s       just display the name (or file path)
   --recursive, -r   recursively include all children of the target directory
   --paths           show absolute file paths instead of file names
   --dirs            only show folders
   --docs            only show documents
   --ids             show full uuids instead of prefixes
   --help, -h        show this help message
`

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
		name, eqv := optParts(args[i][1:])
		switch name {
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
			exitUsgGood(lsCmdUsage)
		}
	}
	args = args[i:]
	if len(args) < 1 {
		return
	}
	c.target = args[0]
}

const mkdirCmdUsage = `name:
   lbcli mkdir - create a directory or do nothing if it exists

usage:
   mkdir [options] <path>

arguments:
   <path>   a path at which to create the directory

options:
   --help, -h   show this help message
`

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
		name, _ := optParts(args[i][1:])
		switch name {
		case "help", "h":
			exitUsgGood(mkdirCmdUsage)
		}
	}
	args = args[i:]
	if len(args) < 1 {
		exitMissingArg("<path>", mkdirCmdUsage)
	}
	c.path = args[0]
}

const mkdocCmdUsage = `name:
   lbcli mkdoc - create a document or do nothing if it exists

usage:
   mkdoc [options] <path>

arguments:
   <path>   a path at which to create the document

options:
   --help, -h   show this help message
`

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
		name, _ := optParts(args[i][1:])
		switch name {
		case "help", "h":
			exitUsgGood(mkdocCmdUsage)
		}
	}
	args = args[i:]
	if len(args) < 1 {
		exitMissingArg("<path>", mkdocCmdUsage)
	}
	c.path = args[0]
}

const mvCmdUsage = `name:
   lbcli mv - move a file to another parent

usage:
   mv [options] <src> <dest>

arguments:
   <src>   
   <dest>  

options:
   --help, -h   show this help message
`

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
		name, _ := optParts(args[i][1:])
		switch name {
		case "help", "h":
			exitUsgGood(mvCmdUsage)
		}
	}
	args = args[i:]
	if len(args) < 1 {
		exitMissingArg("<src>", mvCmdUsage)
	}
	if len(args) < 2 {
		exitMissingArg("<dest>", mvCmdUsage)
	}
	c.src = args[0]
	c.dest = args[1]
}

const renameCmdUsage = `name:
   lbcli rename - rename a file

usage:
   rename [-f] <target> [new-name]

arguments:
   <target>   
   <newname>  

options:
   --force, -f   non-interactive (fail instead of prompting for corrections)
   --help, -h    show this help message
`

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
		name, eqv := optParts(args[i][1:])
		switch name {
		case "force", "f":
			c.force = (eqv == "" || eqv == "true")
		case "help", "h":
			exitUsgGood(renameCmdUsage)
		}
	}
	args = args[i:]
	if len(args) < 1 {
		exitMissingArg("<target>", renameCmdUsage)
	}
	if len(args) < 2 {
		exitMissingArg("<newname>", renameCmdUsage)
	}
	c.target = args[0]
	c.newName = args[1]
}

const rmCmdUsage = `name:
   lbcli rm - delete a file

usage:
   rm [options] <target>

arguments:
   <target>   lockbook path or id to delete

options:
   --force, -f   don't prompt for confirmation
   --help, -h    show this help message
`

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
		name, eqv := optParts(args[i][1:])
		switch name {
		case "force", "f":
			c.force = (eqv == "" || eqv == "true")
		case "help", "h":
			exitUsgGood(rmCmdUsage)
		}
	}
	args = args[i:]
	if len(args) < 1 {
		exitMissingArg("<target>", rmCmdUsage)
	}
	c.target = args[0]
}

const shareCreateCmdUsage = `name:
   lbcli share create - share a file with another lockbook user

usage:
   create [options] <target> <username>

arguments:
   <target>   the path or id of the lockbook file you'd like to share
   <username> the username of the other lockbook user

options:
   --ro         the other user will not be able to edit the file
   --help, -h   show this help message
`

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
		name, eqv := optParts(args[i][1:])
		switch name {
		case "ro":
			c.readOnly = (eqv == "" || eqv == "true")
		case "help", "h":
			exitUsgGood(shareCreateCmdUsage)
		}
	}
	args = args[i:]
	if len(args) < 1 {
		exitMissingArg("<target>", shareCreateCmdUsage)
	}
	if len(args) < 2 {
		exitMissingArg("<username>", shareCreateCmdUsage)
	}
	c.target = args[0]
	c.username = args[1]
}

const sharePendingCmdUsage = `name:
   lbcli share pending - list pending shares

usage:
   pending [options]

options:
   --ids        print full uuids instead of prefixes
   --help, -h   show this help message
`

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
		name, eqv := optParts(args[i][1:])
		switch name {
		case "ids":
			c.fullIDs = (eqv == "" || eqv == "true")
		case "help", "h":
			exitUsgGood(sharePendingCmdUsage)
		}
	}
}

const shareAcceptCmdUsage = `name:
   lbcli share accept - accept a pending share

usage:
   accept [options] <target> <dest> [newname]

arguments:
   <target>   id or id prefix of the pending share to accept
   <dest>     where to place this in your file tree
   [newname]  name this file something else

options:
   --help, -h   show this help message
`

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
		name, _ := optParts(args[i][1:])
		switch name {
		case "help", "h":
			exitUsgGood(shareAcceptCmdUsage)
		}
	}
	args = args[i:]
	if len(args) < 1 {
		exitMissingArg("<target>", shareAcceptCmdUsage)
	}
	if len(args) < 2 {
		exitMissingArg("<dest>", shareAcceptCmdUsage)
	}
	c.target = args[0]
	c.dest = args[1]
	if len(args) < 3 {
		return
	}
	c.newName = args[2]
}

const shareRejectCmdUsage = `name:
   lbcli share reject - reject a pending share

usage:
   reject [options] <target>

arguments:
   <target>   id or id prefix of a pending share

options:
   --help, -h   show this help message
`

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
		name, _ := optParts(args[i][1:])
		switch name {
		case "help", "h":
			exitUsgGood(shareRejectCmdUsage)
		}
	}
	args = args[i:]
	if len(args) < 1 {
		exitMissingArg("<target>", shareRejectCmdUsage)
	}
	c.target = args[0]
}

const shareCmdUsage = `name:
   lbcli share - sharing related commands

usage:
   share [options] <command>

commands:
   create   share a file with another lockbook user
   pending  list pending shares
   accept   accept a pending share
   reject   reject a pending share

options:
   --help, -h   show this help message
`

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
		name, _ := optParts(args[i][1:])
		switch name {
		case "help", "h":
			exitUsgGood(shareCmdUsage)
		}
	}
	if i >= len(args) {
		fmt.Fprint(os.Stderr, shareCmdUsage)
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
		exitUnknownCmd(args[i], shareCmdUsage)
	}
}

const statusCmdUsage = `name:
   lbcli status - which operations a sync would perform

usage:
   status [options]

options:
   --help, -h   show this help message
`

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
		name, _ := optParts(args[i][1:])
		switch name {
		case "help", "h":
			exitUsgGood(statusCmdUsage)
		}
	}
}

const syncCmdUsage = `name:
   lbcli sync - get updates from the server and push changes

usage:
   sync [options]

options:
   --verbose, -v   output every sync step and progress
   --help, -h      show this help message
`

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
		name, eqv := optParts(args[i][1:])
		switch name {
		case "verbose", "v":
			c.verbose = (eqv == "" || eqv == "true")
		case "help", "h":
			exitUsgGood(syncCmdUsage)
		}
	}
}

const usageCmdUsage = `name:
   lbcli usage - local and server disk utilization (uncompressed and compressed)

usage:
   usage [-e]

options:
   --exact, -e   show amounts in bytes
   --help, -h    show this help message
`

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
		name, eqv := optParts(args[i][1:])
		switch name {
		case "exact", "e":
			c.exact = (eqv == "" || eqv == "true")
		case "help", "h":
			exitUsgGood(usageCmdUsage)
		}
	}
}

const whoamiCmdUsage = `name:
   lbcli whoami - print user information for this lockbook

usage:
   whoami [-l]

options:
   --long, -l   prints the data directory and server url as well
   --help, -h   show this help message
`

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
		name, eqv := optParts(args[i][1:])
		switch name {
		case "long", "l":
			c.long = (eqv == "" || eqv == "true")
		case "help", "h":
			exitUsgGood(whoamiCmdUsage)
		}
	}
}

const writeCmdUsage = `name:
   lbcli write - write data from stdin to a lockbook document

usage:
   write [--trunc] <target>

arguments:
   <target>   lockbook path or id to write

options:
   --trunc      truncate the file instead of appending to it
   --help, -h   show this help message
`

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
		name, eqv := optParts(args[i][1:])
		switch name {
		case "trunc":
			c.trunc = (eqv == "" || eqv == "true")
		case "help", "h":
			exitUsgGood(writeCmdUsage)
		}
	}
	args = args[i:]
	if len(args) < 1 {
		exitMissingArg("<target>", writeCmdUsage)
	}
	c.target = args[0]
}

const lbcliUsage = `name:
   lbcli - an unofficial lockbook cli implemented in go

usage:
   lbcli [options] <command>

commands:
   acct     account related commands
   cat      print one or more documents to stdout
   debug    investigative commands mainly intended for devs
   drawing  export a lockbook drawing as an image written to stdout
   export   copy a lockbook file to your file system
   init     create a lockbook account
   ls       list files in a directory
   mkdir    create a directory or do nothing if it exists
   mkdoc    create a document or do nothing if it exists
   mv       move a file to another parent
   rename   rename a file
   rm       delete a file
   share    sharing related commands
   status   which operations a sync would perform
   sync     get updates from the server and push changes
   usage    local and server disk utilization (uncompressed and compressed)
   whoami   print user information for this lockbook
   write    write data from stdin to a lockbook document

options:
   --help, -h   show this help message
`

func (c *lbcli) parse(args []string) {
	var i int
	for ; i < len(args); i++ {
		if args[i][0] != '-' {
			break
		}
		if args[i] == "--" {
			i++
			break
		}
		name, _ := optParts(args[i][1:])
		switch name {
		case "help", "h":
			exitUsgGood(lbcliUsage)
		}
	}
	if i >= len(args) {
		fmt.Fprint(os.Stderr, lbcliUsage)
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
		exitUnknownCmd(args[i], lbcliUsage)
	}
}

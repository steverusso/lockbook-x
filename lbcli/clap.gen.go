// generated by goclap; DO NOT EDIT

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

type clapUsagePrinter interface {
	printUsage(to *os.File)
}

func exitMissingArg(u clapUsagePrinter, name string) {
	claperr("not enough args: no \033[1;33m%s\033[0m provided\n", name)
	u.printUsage(os.Stderr)
	os.Exit(1)
}

func exitUnknownCmd(u clapUsagePrinter, name string) {
	claperr("unknown command '%s'\n", name)
	u.printUsage(os.Stderr)
	os.Exit(1)
}

func clapParseBool(s string) bool {
	if s == "" || s == "true" {
		return true
	}
	if s != "false" {
		claperr("invalid boolean value '%s'\n", s)
		os.Exit(1)
	}
	return false
}

func optParts(arg string) (string, string, bool) {
	if arg == "-" {
		claperr("emtpy option ('-') found\n")
		os.Exit(1)
	}
	if arg[0] == '-' {
		arg = arg[1:]
	}
	if arg[0] == '-' {
		arg = arg[1:]
	}
	if eqIdx := strings.IndexByte(arg, '='); eqIdx != -1 {
		name := arg[:eqIdx]
		eqVal := ""
		if eqIdx < len(arg) {
			eqVal = arg[eqIdx+1:]
		}
		return name, eqVal, true
	}
	return arg, "", false
}

type clapOpt struct {
	long  string
	short string
	v     any
}

func parseOpts(args []string, u clapUsagePrinter, data []clapOpt) int {
	var i int
argsLoop:
	for ; i < len(args); i++ {
		if args[i][0] != '-' {
			break
		}
		if args[i] == "--" {
			i++
			break
		}
		k, eqv, hasEq := optParts(args[i][1:])
		for z := range data {
			if k == data[z].long || k == data[z].short {
				switch v := data[z].v.(type) {
				case *bool:
					*v = clapParseBool(eqv)
				case *string:
					if hasEq {
						*v = eqv
					} else if i == len(args)-1 {
						claperr("string option '%s' needs an argument\n", k)
						os.Exit(1)
					} else {
						i++
						*v = args[i]
					}
				}
				continue argsLoop
			}
		}
		if k == "h" || k == "help" {
			u.printUsage(os.Stdout)
			os.Exit(0)
		}
		claperr("unknown option '%s'\n", k)
		os.Exit(1)
	}
	return i
}

func (*acctInitCmd) printUsage(to *os.File) {
	fmt.Fprintf(to, `%[1]s acct init - create a lockbook account

usage:
   init [options]

options:
       --welcome   include the welcome document
       --no-sync   don't perform the initial sync
   -h, --help      show this help message
`, os.Args[0])
}

func (c *acctInitCmd) parse(args []string) {
	parseOpts(args, c, []clapOpt{
		{"welcome", "", &c.welcome},
		{"no-sync", "", &c.noSync},
	})
}

func (*acctRestoreCmd) printUsage(to *os.File) {
	fmt.Fprintf(to, `%[1]s acct restore - restore an existing account from its secret account string

overview:
   The restore command reads the secret account string from standard input (stdin).
   In other words, pipe your account string to this command like:
   'cat lbkey.txt | lbcli restore'.

usage:
   restore [options]

options:
       --no-sync   don't perform the initial sync
   -h, --help      show this help message
`, os.Args[0])
}

func (c *acctRestoreCmd) parse(args []string) {
	parseOpts(args, c, []clapOpt{
		{"no-sync", "", &c.noSync},
	})
}

func (*acctPrivKeyCmd) printUsage(to *os.File) {
	fmt.Fprintf(to, `%[1]s acct privkey - print out the private key for this lockbook

overview:
   Your private key should always be kept secret and should only be displayed when you are
   in a secure location. For that reason, this command will prompt you before printing
   your private key just to double check. However, this prompt can be satisfied by passing
   the '--no-prompt' option.

usage:
   privkey [options]

options:
       --no-prompt   don't require confirmation before displaying the private key
   -h, --help        show this help message
`, os.Args[0])
}

func (c *acctPrivKeyCmd) parse(args []string) {
	parseOpts(args, c, []clapOpt{
		{"no-prompt", "", &c.noPrompt},
	})
}

func (*acctStatusCmd) printUsage(to *os.File) {
	fmt.Fprintf(to, `%[1]s acct status - overview of your account

usage:
   status [options]

options:
   -h, --help   show this help message
`, os.Args[0])
}

func (c *acctStatusCmd) parse(args []string) {
	parseOpts(args, c, nil)
}

func (*acctSubscribeCmd) printUsage(to *os.File) {
	fmt.Fprintf(to, `%[1]s acct subscribe - create a new subscription with a credit card

usage:
   subscribe [options]

options:
   -h, --help   show this help message
`, os.Args[0])
}

func (c *acctSubscribeCmd) parse(args []string) {
	parseOpts(args, c, nil)
}

func (*acctUnsubscribeCmd) printUsage(to *os.File) {
	fmt.Fprintf(to, `%[1]s acct unsubscribe - cancel an existing subscription

usage:
   unsubscribe [options]

options:
   -h, --help   show this help message
`, os.Args[0])
}

func (c *acctUnsubscribeCmd) parse(args []string) {
	parseOpts(args, c, nil)
}

func (*acctCmd) printUsage(to *os.File) {
	fmt.Fprintf(to, `%[1]s acct - account related commands

usage:
   acct <command> [args...]

options:
   -h, --help   show this help message

subcommands:
   init          create a lockbook account
   restore       restore an existing account from its secret account string
   privkey       print out the private key for this lockbook
   status        overview of your account
   subscribe     create a new subscription with a credit card
   unsubscribe   cancel an existing subscription
`, os.Args[0])
}

func (c *acctCmd) parse(args []string) {
	i := parseOpts(args, c, nil)
	if i >= len(args) {
		c.printUsage(os.Stderr)
		os.Exit(1)
	}
	switch args[i] {
	case "init":
		c.init = new(acctInitCmd)
		c.init.parse(args[i+1:])
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
	fmt.Fprintf(to, `%[1]s cat - print a document's content

usage:
   cat [options] <target>

options:
   -h, --help   show this help message

arguments:
   <target>   lockbook file path or ID
`, os.Args[0])
}

func (c *catCmd) parse(args []string) {
	i := parseOpts(args, c, nil)
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
   -h, --help   show this help message

arguments:
   <target>   the target can be a file path, UUID, or UUID prefix
`, os.Args[0])
}

func (c *debugFinfoCmd) parse(args []string) {
	i := parseOpts(args, c, nil)
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
   -h, --help   show this help message
`, os.Args[0])
}

func (c *debugValidateCmd) parse(args []string) {
	parseOpts(args, c, nil)
}

func (*debugWhoamiCmd) printUsage(to *os.File) {
	fmt.Fprintf(to, `%[1]s debug whoami - print user information for this lockbook

usage:
   whoami [options]

options:
   -h, --help   show this help message
`, os.Args[0])
}

func (c *debugWhoamiCmd) parse(args []string) {
	parseOpts(args, c, nil)
}

func (*debugCmd) printUsage(to *os.File) {
	fmt.Fprintf(to, `%[1]s debug - investigative commands mainly intended for devs

usage:
   debug [options] <command>

options:
   -h, --help   show this help message

subcommands:
   finfo      view info about a target file
   validate   find invalid states within your lockbook
   whoami     print user information for this lockbook
`, os.Args[0])
}

func (c *debugCmd) parse(args []string) {
	i := parseOpts(args, c, nil)
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
	case "whoami":
		c.whoami = new(debugWhoamiCmd)
		c.whoami.parse(args[i+1:])
	default:
		exitUnknownCmd(c, args[i])
	}
}

func (*exportCmd) printUsage(to *os.File) {
	fmt.Fprintf(to, `%[1]s export - copy a lockbook file to your file system

usage:
   export [--quiet] <target> [dest-dir]
   export [--img-fmt <fmt>] <drawing> [dest-dir]

options:
   -i, --img-fmt  <arg>   format for exporting a lockbook drawing (png|jpeg|pnm|tga|farbfeld|bmp)
   -q, --quiet            don't output progress on each file
   -h, --help             show this help message

arguments:
   <target>   lockbook file path or ID
   [dest]     disk file path (default ".")
`, os.Args[0])
}

func (c *exportCmd) parse(args []string) {
	i := parseOpts(args, c, []clapOpt{
		{"img-fmt", "i", &c.imgFmt},
		{"quiet", "q", &c.quiet},
	})
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

func (*importCmd) printUsage(to *os.File) {
	fmt.Fprintf(to, `%[1]s import - import files into lockbook from your system

usage:
   import [options] <diskpath> [dest]

options:
   -q, --quiet   don't output progress on each file
   -h, --help    show this help message

arguments:
   <diskpath>   the file to import into lockbook
   [dest]       where to put the imported files in lockbook
`, os.Args[0])
}

func (c *importCmd) parse(args []string) {
	i := parseOpts(args, c, []clapOpt{
		{"quiet", "q", &c.quiet},
	})
	args = args[i:]
	if len(args) < 1 {
		exitMissingArg(c, "<diskpath>")
	}
	c.diskPath = args[0]
	if len(args) < 2 {
		return
	}
	c.dest = args[1]
}

func (*jotCmd) printUsage(to *os.File) {
	fmt.Fprintf(to, `%[1]s jot - quickly record brief thoughts

usage:
   jot [options] <message>

options:
   -d, --dateit          prepend the date and time to the message
   -D, --dateit-after    append the date and time to the message
   -t, --target  <arg>   the target file (defaults to "/scratch.md")
   -h, --help            show this help message

arguments:
   <message>   the text you would like to jot down
`, os.Args[0])
}

func (c *jotCmd) parse(args []string) {
	i := parseOpts(args, c, []clapOpt{
		{"dateit", "d", &c.dateIt},
		{"dateit-after", "D", &c.dateItAfter},
		{"target", "t", &c.target},
	})
	args = args[i:]
	if len(args) < 1 {
		exitMissingArg(c, "<message>")
	}
	c.message = args[0]
}

func (*lsCmd) printUsage(to *os.File) {
	fmt.Fprintf(to, `%[1]s ls - list files in a directory

usage:
   ls [options] [target]

options:
   -s, --short       just display the name (or file path)
   -r, --recursive   recursively list children of the target directory
   -t, --tree        recursively list children of the target directory in a tree format
       --paths       show absolute file paths instead of file names
       --dirs        only show folders
       --docs        only show documents
       --ids         show full UUIDs instead of prefixes
   -h, --help        show this help message

arguments:
   [target]   path or ID of the target directory (defaults to root)
`, os.Args[0])
}

func (c *lsCmd) parse(args []string) {
	i := parseOpts(args, c, []clapOpt{
		{"short", "s", &c.short},
		{"recursive", "r", &c.recursive},
		{"tree", "t", &c.tree},
		{"paths", "", &c.paths},
		{"dirs", "", &c.onlyDirs},
		{"docs", "", &c.onlyDocs},
		{"ids", "", &c.fullIDs},
	})
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
   -h, --help   show this help message

arguments:
   <path>   the path at which to create the directory
`, os.Args[0])
}

func (c *mkdirCmd) parse(args []string) {
	i := parseOpts(args, c, nil)
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
   -h, --help   show this help message

arguments:
   <path>   the path at which to create the document
`, os.Args[0])
}

func (c *mkdocCmd) parse(args []string) {
	i := parseOpts(args, c, nil)
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
   -h, --help   show this help message

arguments:
   <src>    the file to move
   <dest>   the destination directory
`, os.Args[0])
}

func (c *mvCmd) parse(args []string) {
	i := parseOpts(args, c, nil)
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
   -f, --force   non-interactive (fail instead of prompting for corrections)
   -h, --help    show this help message

arguments:
   <target>    the file to rename
   <newname>   the desired new name
`, os.Args[0])
}

func (c *renameCmd) parse(args []string) {
	i := parseOpts(args, c, []clapOpt{
		{"force", "f", &c.force},
	})
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
   -f, --force   don't prompt for confirmation
   -h, --help    show this help message

arguments:
   <target>   lockbook path or ID to delete
`, os.Args[0])
}

func (c *rmCmd) parse(args []string) {
	i := parseOpts(args, c, []clapOpt{
		{"force", "f", &c.force},
	})
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
       --ro     the other user will not be able to edit the file
   -h, --help   show this help message

arguments:
   <target>     the path or ID of the lockbook file you'd like to share
   <username>   the username of the other lockbook user
`, os.Args[0])
}

func (c *shareCreateCmd) parse(args []string) {
	i := parseOpts(args, c, []clapOpt{
		{"ro", "", &c.readOnly},
	})
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
       --ids    print full UUIDs instead of prefixes
   -h, --help   show this help message
`, os.Args[0])
}

func (c *sharePendingCmd) parse(args []string) {
	parseOpts(args, c, []clapOpt{
		{"ids", "", &c.fullIDs},
	})
}

func (*shareAcceptCmd) printUsage(to *os.File) {
	fmt.Fprintf(to, `%[1]s share accept - accept a pending share

usage:
   accept [options] <target> <dest> [newname]

options:
   -h, --help   show this help message

arguments:
   <target>    the ID or ID prefix of a pending share
   <dest>      where to place this in your file tree
   [newname]   name this file something else
`, os.Args[0])
}

func (c *shareAcceptCmd) parse(args []string) {
	i := parseOpts(args, c, nil)
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
   -h, --help   show this help message

arguments:
   <target>   the ID or ID prefix of a pending share
`, os.Args[0])
}

func (c *shareRejectCmd) parse(args []string) {
	i := parseOpts(args, c, nil)
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
   -h, --help   show this help message

subcommands:
   create    share a file with another lockbook user
   pending   list pending shares
   accept    accept a pending share
   reject    reject a pending share
`, os.Args[0])
}

func (c *shareCmd) parse(args []string) {
	i := parseOpts(args, c, nil)
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

func (*syncCmd) printUsage(to *os.File) {
	fmt.Fprintf(to, `%[1]s sync - get updates from the server and push changes

usage:
   sync [options]

options:
   -s, --status    show last synced and which operations a sync would perform
   -v, --verbose   output every sync step and progress
   -h, --help      show this help message
`, os.Args[0])
}

func (c *syncCmd) parse(args []string) {
	parseOpts(args, c, []clapOpt{
		{"status", "s", &c.status},
		{"verbose", "v", &c.verbose},
	})
}

func (*usageCmd) printUsage(to *os.File) {
	fmt.Fprintf(to, `%[1]s usage - local and server disk utilization (uncompressed and compressed)

usage:
   usage [-e]

options:
   -e, --exact   show amounts in bytes instead of as human readable values
   -h, --help    show this help message
`, os.Args[0])
}

func (c *usageCmd) parse(args []string) {
	parseOpts(args, c, []clapOpt{
		{"exact", "e", &c.exact},
	})
}

func (*writeCmd) printUsage(to *os.File) {
	fmt.Fprintf(to, `%[1]s write - write data from stdin to a lockbook document

usage:
   write [--trunc] <target>

options:
       --trunc   truncate the file instead of appending to it
   -h, --help    show this help message

arguments:
   <target>   lockbook path or ID to write
`, os.Args[0])
}

func (c *writeCmd) parse(args []string) {
	i := parseOpts(args, c, []clapOpt{
		{"trunc", "", &c.trunc},
	})
	args = args[i:]
	if len(args) < 1 {
		exitMissingArg(c, "<target>")
	}
	c.target = args[0]
}

func (*lbcli) printUsage(to *os.File) {
	fmt.Fprintf(to, `%[1]s - an unofficial lockbook cli

usage:
   %[1]s [options] <command>

options:
   -h, --help   show this help message

subcommands:
   acct     account related commands
   cat      print a document's content
   debug    investigative commands mainly intended for devs
   export   copy a lockbook file to your file system
   import   import files into lockbook from your system
   jot      quickly record brief thoughts
   ls       list files in a directory
   mkdir    create a directory or do nothing if it exists
   mkdoc    create a document or do nothing if it exists
   mv       move a file to another parent
   rename   rename a file
   rm       delete a file
   share    sharing related commands
   sync     get updates from the server and push changes
   usage    local and server disk utilization (uncompressed and compressed)
   write    write data from stdin to a lockbook document

run '%[1]s <subcommand> -h' for more information on specific commands.
`, os.Args[0])
}

func (c *lbcli) parse(args []string) {
	if len(args) > 0 && len(args) == len(os.Args) {
		args = args[1:]
	}
	i := parseOpts(args, c, nil)
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
	case "export":
		c.export = new(exportCmd)
		c.export.parse(args[i+1:])
	case "import":
		c.imprt = new(importCmd)
		c.imprt.parse(args[i+1:])
	case "jot":
		c.jot = new(jotCmd)
		c.jot.parse(args[i+1:])
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
	case "sync":
		c.sync = new(syncCmd)
		c.sync.parse(args[i+1:])
	case "usage":
		c.usage = new(usageCmd)
		c.usage.parse(args[i+1:])
	case "write":
		c.write = new(writeCmd)
		c.write.parse(args[i+1:])
	default:
		exitUnknownCmd(c, args[i])
	}
}
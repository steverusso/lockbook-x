// generated by goclap; DO NOT EDIT

package main

import "github.com/steverusso/goclap/clap"

func (*acctInitCmd) UsageHelp() string {
	return `lbcli acct init - Create a lockbook account

usage:
   init [options]

options:
   -welcome   Include the welcome document
   -no-sync   Don't perform the initial sync
   -h         Show this help message`
}

func (c *acctInitCmd) Parse(args []string) {
	p := clap.NewCommandParser("lbcli acct init")
	p.CustomUsage = c.UsageHelp
	p.Flag("welcome", clap.NewBool(&c.welcome))
	p.Flag("no-sync", clap.NewBool(&c.noSync))
	p.Parse(args)
}

func (*acctRestoreCmd) UsageHelp() string {
	return `lbcli acct restore - Restore an existing account from its secret account string

overview:
   The restore command reads the secret account string from standard input (stdin).
   In other words, pipe your account string to this command like:
   'cat lbkey.txt | lbcli acct restore'.

usage:
   restore [options]

options:
   -no-sync   Don't perform the initial sync
   -h         Show this help message`
}

func (c *acctRestoreCmd) Parse(args []string) {
	p := clap.NewCommandParser("lbcli acct restore")
	p.CustomUsage = c.UsageHelp
	p.Flag("no-sync", clap.NewBool(&c.noSync))
	p.Parse(args)
}

func (*acctPrivKeyCmd) UsageHelp() string {
	return `lbcli acct privkey - Print out the private key for this lockbook

overview:
   Your private key should always be kept secret and should only be displayed when you are
   in a secure location. For that reason, this command will prompt you before printing
   your private key just to double check. However, this prompt can be satisfied by passing
   the '--no-prompt' option.

usage:
   privkey [options]

options:
   -no-prompt   Don't require confirmation before displaying the private key
   -h           Show this help message`
}

func (c *acctPrivKeyCmd) Parse(args []string) {
	p := clap.NewCommandParser("lbcli acct privkey")
	p.CustomUsage = c.UsageHelp
	p.Flag("no-prompt", clap.NewBool(&c.noPrompt))
	p.Parse(args)
}

func (*acctStatusCmd) UsageHelp() string {
	return `lbcli acct status - Overview of your account

usage:
   status [options]

options:
   -h   Show this help message`
}

func (c *acctStatusCmd) Parse(args []string) {
	p := clap.NewCommandParser("lbcli acct status")
	p.CustomUsage = c.UsageHelp
	p.Parse(args)
}

func (*acctSubscribeCmd) UsageHelp() string {
	return `lbcli acct subscribe - Create a new subscription with a credit card

usage:
   subscribe [options]

options:
   -h   Show this help message`
}

func (c *acctSubscribeCmd) Parse(args []string) {
	p := clap.NewCommandParser("lbcli acct subscribe")
	p.CustomUsage = c.UsageHelp
	p.Parse(args)
}

func (*acctUnsubscribeCmd) UsageHelp() string {
	return `lbcli acct unsubscribe - Cancel an existing subscription

usage:
   unsubscribe [options]

options:
   -h   Show this help message`
}

func (c *acctUnsubscribeCmd) Parse(args []string) {
	p := clap.NewCommandParser("lbcli acct unsubscribe")
	p.CustomUsage = c.UsageHelp
	p.Parse(args)
}

func (*acctCmd) UsageHelp() string {
	return `lbcli acct - Account related commands

usage:
   acct <command> [args...]

options:
   -h   Show this help message

subcommands:
   init          Create a lockbook account
   restore       Restore an existing account from its secret account string
   privkey       Print out the private key for this lockbook
   status        Overview of your account
   subscribe     Create a new subscription with a credit card
   unsubscribe   Cancel an existing subscription`
}

func (c *acctCmd) Parse(args []string) {
	p := clap.NewCommandParser("lbcli acct")
	p.CustomUsage = c.UsageHelp
	rest := p.Parse(args)

	if len(rest) == 0 {
		p.Fatalf("no subcommand provided")
	}
	switch rest[0] {
	case "init":
		c.init = &acctInitCmd{}
		c.init.Parse(rest[1:])
	case "restore":
		c.restore = &acctRestoreCmd{}
		c.restore.Parse(rest[1:])
	case "privkey":
		c.privkey = &acctPrivKeyCmd{}
		c.privkey.Parse(rest[1:])
	case "status":
		c.status = &acctStatusCmd{}
		c.status.Parse(rest[1:])
	case "subscribe":
		c.subscribe = &acctSubscribeCmd{}
		c.subscribe.Parse(rest[1:])
	case "unsubscribe":
		c.unsubscribe = &acctUnsubscribeCmd{}
		c.unsubscribe.Parse(rest[1:])
	default:
		p.Fatalf("unknown subcommand '%s'", rest[0])
	}
}

func (*catCmd) UsageHelp() string {
	return `lbcli cat - Print a document's content

usage:
   cat [options] <target>

options:
   -h   Show this help message

arguments:
   <target>   Lockbook file path or ID`
}

func (c *catCmd) Parse(args []string) {
	p := clap.NewCommandParser("lbcli cat")
	p.CustomUsage = c.UsageHelp
	p.Arg("<target>", clap.NewString(&c.target)).Require()
	p.Parse(args)
}

func (*debugFinfoCmd) UsageHelp() string {
	return `lbcli debug finfo - View info about a target file

usage:
   finfo [options] <target>

options:
   -h   Show this help message

arguments:
   <target>   The target can be a file path, UUID, or UUID prefix`
}

func (c *debugFinfoCmd) Parse(args []string) {
	p := clap.NewCommandParser("lbcli debug finfo")
	p.CustomUsage = c.UsageHelp
	p.Arg("<target>", clap.NewString(&c.target)).Require()
	p.Parse(args)
}

func (*debugValidateCmd) UsageHelp() string {
	return `lbcli debug validate - Find invalid states within your lockbook

usage:
   validate [options]

options:
   -h   Show this help message`
}

func (c *debugValidateCmd) Parse(args []string) {
	p := clap.NewCommandParser("lbcli debug validate")
	p.CustomUsage = c.UsageHelp
	p.Parse(args)
}

func (*debugWhoamiCmd) UsageHelp() string {
	return `lbcli debug whoami - Print user information for this lockbook

usage:
   whoami [options]

options:
   -h   Show this help message`
}

func (c *debugWhoamiCmd) Parse(args []string) {
	p := clap.NewCommandParser("lbcli debug whoami")
	p.CustomUsage = c.UsageHelp
	p.Parse(args)
}

func (*debugCmd) UsageHelp() string {
	return `lbcli debug - Investigative commands mainly intended for devs

usage:
   debug [options] <command>

options:
   -h   Show this help message

subcommands:
   finfo      View info about a target file
   validate   Find invalid states within your lockbook
   whoami     Print user information for this lockbook`
}

func (c *debugCmd) Parse(args []string) {
	p := clap.NewCommandParser("lbcli debug")
	p.CustomUsage = c.UsageHelp
	rest := p.Parse(args)

	if len(rest) == 0 {
		p.Fatalf("no subcommand provided")
	}
	switch rest[0] {
	case "finfo":
		c.finfo = &debugFinfoCmd{}
		c.finfo.Parse(rest[1:])
	case "validate":
		c.validate = &debugValidateCmd{}
		c.validate.Parse(rest[1:])
	case "whoami":
		c.whoami = &debugWhoamiCmd{}
		c.whoami.Parse(rest[1:])
	default:
		p.Fatalf("unknown subcommand '%s'", rest[0])
	}
}

func (*exportCmd) UsageHelp() string {
	return `lbcli export - Copy a lockbook file to your file system

usage:
   export [--quiet] <target> [dest-dir]
   export [--img-fmt <fmt>] <drawing> [dest-dir]

options:
   -img-fmt,i  <arg>   Format for exporting a lockbook drawing
                       (png|jpeg|pnm|tga|farbfeld|bmp)
   -quiet,q            Don't output progress on each file
   -h                  Show this help message

arguments:
   <target>   Lockbook file path or ID
   [dest]     Disk file path (default ".")`
}

func (c *exportCmd) Parse(args []string) {
	p := clap.NewCommandParser("lbcli export")
	p.CustomUsage = c.UsageHelp
	p.Flag("img-fmt,i", clap.NewString(&c.imgFmt))
	p.Flag("quiet,q", clap.NewBool(&c.quiet))
	p.Arg("<target>", clap.NewString(&c.target)).Require()
	p.Arg("[dest]", clap.NewString(&c.dest))
	p.Parse(args)
}

func (*importCmd) UsageHelp() string {
	return `lbcli import - Import files into lockbook from your system

usage:
   import [options] <diskpath> [dest]

options:
   -quiet,q   Don't output progress on each file
   -h         Show this help message

arguments:
   <diskpath>   The file to import into lockbook
   [dest]       Where to put the imported files in lockbook`
}

func (c *importCmd) Parse(args []string) {
	p := clap.NewCommandParser("lbcli import")
	p.CustomUsage = c.UsageHelp
	p.Flag("quiet,q", clap.NewBool(&c.quiet))
	p.Arg("<diskpath>", clap.NewString(&c.diskPath)).Require()
	p.Arg("[dest]", clap.NewString(&c.dest))
	p.Parse(args)
}

func (*jotCmd) UsageHelp() string {
	return `lbcli jot - Quickly record brief thoughts

usage:
   jot [options] <message>

options:
   -dateit,d          Prepend the date and time to the message
   -dateit-after,D    Append the date and time to the message
   -target,t  <arg>   The target file (defaults to "/scratch.md")
   -h                 Show this help message

arguments:
   <message>   The text you would like to jot down`
}

func (c *jotCmd) Parse(args []string) {
	p := clap.NewCommandParser("lbcli jot")
	p.CustomUsage = c.UsageHelp
	p.Flag("dateit,d", clap.NewBool(&c.dateIt))
	p.Flag("dateit-after,D", clap.NewBool(&c.dateItAfter))
	p.Flag("target,t", clap.NewString(&c.target))
	p.Arg("<message>", clap.NewString(&c.message)).Require()
	p.Parse(args)
}

func (*lsCmd) UsageHelp() string {
	return `lbcli ls - List files in a directory

usage:
   ls [options] [target]

options:
   -short,s       Just display the name (or file path)
   -recursive,r   Recursively list children of the target directory
   -tree,t        Recursively list children of the target directory in a tree format
   -paths         Show absolute file paths instead of file names
   -dirs          Only show folders
   -docs          Only show documents
   -ids           Show full UUIDs instead of prefixes
   -h             Show this help message

arguments:
   [target]   Path or ID of the target directory (defaults to root)`
}

func (c *lsCmd) Parse(args []string) {
	p := clap.NewCommandParser("lbcli ls")
	p.CustomUsage = c.UsageHelp
	p.Flag("short,s", clap.NewBool(&c.short))
	p.Flag("recursive,r", clap.NewBool(&c.recursive))
	p.Flag("tree,t", clap.NewBool(&c.tree))
	p.Flag("paths", clap.NewBool(&c.paths))
	p.Flag("dirs", clap.NewBool(&c.onlyDirs))
	p.Flag("docs", clap.NewBool(&c.onlyDocs))
	p.Flag("ids", clap.NewBool(&c.fullIDs))
	p.Arg("[target]", clap.NewString(&c.target))
	p.Parse(args)
}

func (*mkdirCmd) UsageHelp() string {
	return `lbcli mkdir - Create a directory or do nothing if it exists

usage:
   mkdir [options] <path>

options:
   -h   Show this help message

arguments:
   <path>   The path at which to create the directory`
}

func (c *mkdirCmd) Parse(args []string) {
	p := clap.NewCommandParser("lbcli mkdir")
	p.CustomUsage = c.UsageHelp
	p.Arg("<path>", clap.NewString(&c.path)).Require()
	p.Parse(args)
}

func (*mkdocCmd) UsageHelp() string {
	return `lbcli mkdoc - Create a document or do nothing if it exists

usage:
   mkdoc [options] <path>

options:
   -h   Show this help message

arguments:
   <path>   The path at which to create the document`
}

func (c *mkdocCmd) Parse(args []string) {
	p := clap.NewCommandParser("lbcli mkdoc")
	p.CustomUsage = c.UsageHelp
	p.Arg("<path>", clap.NewString(&c.path)).Require()
	p.Parse(args)
}

func (*mvCmd) UsageHelp() string {
	return `lbcli mv - Move a file to another parent

usage:
   mv [options] <src> <dest>

options:
   -h   Show this help message

arguments:
   <src>    The file to move
   <dest>   The destination directory`
}

func (c *mvCmd) Parse(args []string) {
	p := clap.NewCommandParser("lbcli mv")
	p.CustomUsage = c.UsageHelp
	p.Arg("<src>", clap.NewString(&c.src)).Require()
	p.Arg("<dest>", clap.NewString(&c.dest)).Require()
	p.Parse(args)
}

func (*renameCmd) UsageHelp() string {
	return `lbcli rename - Rename a file

usage:
   rename [-f] <target> [new-name]

options:
   -force,f   Non-interactive (fail instead of prompting for corrections)
   -h         Show this help message

arguments:
   <target>    The file to rename
   <newname>   The desired new name`
}

func (c *renameCmd) Parse(args []string) {
	p := clap.NewCommandParser("lbcli rename")
	p.CustomUsage = c.UsageHelp
	p.Flag("force,f", clap.NewBool(&c.force))
	p.Arg("<target>", clap.NewString(&c.target)).Require()
	p.Arg("<newname>", clap.NewString(&c.newName)).Require()
	p.Parse(args)
}

func (*rmCmd) UsageHelp() string {
	return `lbcli rm - Delete a file

usage:
   rm [options] <target>

options:
   -force,f   Don't prompt for confirmation
   -h         Show this help message

arguments:
   <target>   Lockbook path or ID to delete`
}

func (c *rmCmd) Parse(args []string) {
	p := clap.NewCommandParser("lbcli rm")
	p.CustomUsage = c.UsageHelp
	p.Flag("force,f", clap.NewBool(&c.force))
	p.Arg("<target>", clap.NewString(&c.target)).Require()
	p.Parse(args)
}

func (*shareCreateCmd) UsageHelp() string {
	return `lbcli share create - Share a file with another lockbook user

usage:
   create [options] <target> <username>

options:
   -ro   The other user will not be able to edit the file
   -h    Show this help message

arguments:
   <target>     The path or ID of the lockbook file you'd like to share
   <username>   The username of the other lockbook user`
}

func (c *shareCreateCmd) Parse(args []string) {
	p := clap.NewCommandParser("lbcli share create")
	p.CustomUsage = c.UsageHelp
	p.Flag("ro", clap.NewBool(&c.readOnly))
	p.Arg("<target>", clap.NewString(&c.target)).Require()
	p.Arg("<username>", clap.NewString(&c.username)).Require()
	p.Parse(args)
}

func (*sharePendingCmd) UsageHelp() string {
	return `lbcli share pending - List pending shares

usage:
   pending [options]

options:
   -ids   Print full UUIDs instead of prefixes
   -h     Show this help message`
}

func (c *sharePendingCmd) Parse(args []string) {
	p := clap.NewCommandParser("lbcli share pending")
	p.CustomUsage = c.UsageHelp
	p.Flag("ids", clap.NewBool(&c.fullIDs))
	p.Parse(args)
}

func (*shareAcceptCmd) UsageHelp() string {
	return `lbcli share accept - Accept a pending share

usage:
   accept [options] <target> <dest> [newname]

options:
   -h   Show this help message

arguments:
   <target>    The ID or ID prefix of a pending share
   <dest>      Where to place this in your file tree
   [newname]   Name this file something else`
}

func (c *shareAcceptCmd) Parse(args []string) {
	p := clap.NewCommandParser("lbcli share accept")
	p.CustomUsage = c.UsageHelp
	p.Arg("<target>", clap.NewString(&c.target)).Require()
	p.Arg("<dest>", clap.NewString(&c.dest)).Require()
	p.Arg("[newname]", clap.NewString(&c.newName))
	p.Parse(args)
}

func (*shareRejectCmd) UsageHelp() string {
	return `lbcli share reject - Reject a pending share

usage:
   reject [options] <target>

options:
   -h   Show this help message

arguments:
   <target>   The ID or ID prefix of a pending share`
}

func (c *shareRejectCmd) Parse(args []string) {
	p := clap.NewCommandParser("lbcli share reject")
	p.CustomUsage = c.UsageHelp
	p.Arg("<target>", clap.NewString(&c.target)).Require()
	p.Parse(args)
}

func (*shareCmd) UsageHelp() string {
	return `lbcli share - Sharing related commands

usage:
   share [options] <command>

options:
   -h   Show this help message

subcommands:
   create    Share a file with another lockbook user
   pending   List pending shares
   accept    Accept a pending share
   reject    Reject a pending share`
}

func (c *shareCmd) Parse(args []string) {
	p := clap.NewCommandParser("lbcli share")
	p.CustomUsage = c.UsageHelp
	rest := p.Parse(args)

	if len(rest) == 0 {
		p.Fatalf("no subcommand provided")
	}
	switch rest[0] {
	case "create":
		c.create = &shareCreateCmd{}
		c.create.Parse(rest[1:])
	case "pending":
		c.pending = &sharePendingCmd{}
		c.pending.Parse(rest[1:])
	case "accept":
		c.accept = &shareAcceptCmd{}
		c.accept.Parse(rest[1:])
	case "reject":
		c.reject = &shareRejectCmd{}
		c.reject.Parse(rest[1:])
	default:
		p.Fatalf("unknown subcommand '%s'", rest[0])
	}
}

func (*syncCmd) UsageHelp() string {
	return `lbcli sync - Get updates from the server and push changes

usage:
   sync [options]

options:
   -status,s    Show last synced and which operations a sync would perform
   -verbose,v   Output every sync step and progress
   -h           Show this help message`
}

func (c *syncCmd) Parse(args []string) {
	p := clap.NewCommandParser("lbcli sync")
	p.CustomUsage = c.UsageHelp
	p.Flag("status,s", clap.NewBool(&c.status))
	p.Flag("verbose,v", clap.NewBool(&c.verbose))
	p.Parse(args)
}

func (*usageCmd) UsageHelp() string {
	return `lbcli usage - Local and server disk utilization (uncompressed and compressed)

usage:
   usage [-e]

options:
   -exact,e   Show amounts in bytes instead of as human readable values
   -h         Show this help message`
}

func (c *usageCmd) Parse(args []string) {
	p := clap.NewCommandParser("lbcli usage")
	p.CustomUsage = c.UsageHelp
	p.Flag("exact,e", clap.NewBool(&c.exact))
	p.Parse(args)
}

func (*writeCmd) UsageHelp() string {
	return `lbcli write - Write data from stdin to a lockbook document

usage:
   write [--trunc] <target>

options:
   -trunc   Truncate the file instead of appending to it
   -h       Show this help message

arguments:
   <target>   Lockbook path or ID to write`
}

func (c *writeCmd) Parse(args []string) {
	p := clap.NewCommandParser("lbcli write")
	p.CustomUsage = c.UsageHelp
	p.Flag("trunc", clap.NewBool(&c.trunc))
	p.Arg("<target>", clap.NewString(&c.target)).Require()
	p.Parse(args)
}

func (*lbcli) UsageHelp() string {
	return `lbcli - An unofficial lockbook cli

usage:
   lbcli [options] <command>

options:
   -h   Show this help message

subcommands:
   acct     Account related commands
   cat      Print a document's content
   debug    Investigative commands mainly intended for devs
   export   Copy a lockbook file to your file system
   import   Import files into lockbook from your system
   jot      Quickly record brief thoughts
   ls       List files in a directory
   mkdir    Create a directory or do nothing if it exists
   mkdoc    Create a document or do nothing if it exists
   mv       Move a file to another parent
   rename   Rename a file
   rm       Delete a file
   share    Sharing related commands
   sync     Get updates from the server and push changes
   usage    Local and server disk utilization (uncompressed and compressed)
   write    Write data from stdin to a lockbook document

Run 'lbcli <subcommand> -h' for more information on specific commands.`
}

func (c *lbcli) Parse(args []string) {
	p := clap.NewCommandParser("lbcli")
	p.CustomUsage = c.UsageHelp
	rest := p.Parse(args)

	if len(rest) == 0 {
		p.Fatalf("no subcommand provided")
	}
	switch rest[0] {
	case "acct":
		c.acct = &acctCmd{}
		c.acct.Parse(rest[1:])
	case "cat":
		c.cat = &catCmd{}
		c.cat.Parse(rest[1:])
	case "debug":
		c.debug = &debugCmd{}
		c.debug.Parse(rest[1:])
	case "export":
		c.export = &exportCmd{}
		c.export.Parse(rest[1:])
	case "import":
		c.imprt = &importCmd{}
		c.imprt.Parse(rest[1:])
	case "jot":
		c.jot = &jotCmd{}
		c.jot.Parse(rest[1:])
	case "ls", "list":
		c.ls = &lsCmd{}
		c.ls.Parse(rest[1:])
	case "mkdir":
		c.mkdir = &mkdirCmd{}
		c.mkdir.Parse(rest[1:])
	case "mkdoc":
		c.mkdoc = &mkdocCmd{}
		c.mkdoc.Parse(rest[1:])
	case "mv", "move":
		c.mv = &mvCmd{}
		c.mv.Parse(rest[1:])
	case "rename":
		c.rename = &renameCmd{}
		c.rename.Parse(rest[1:])
	case "rm":
		c.rm = &rmCmd{}
		c.rm.Parse(rest[1:])
	case "share":
		c.share = &shareCmd{}
		c.share.Parse(rest[1:])
	case "sync":
		c.sync = &syncCmd{}
		c.sync.Parse(rest[1:])
	case "usage":
		c.usage = &usageCmd{}
		c.usage.Parse(rest[1:])
	case "write":
		c.write = &writeCmd{}
		c.write.Parse(rest[1:])
	default:
		p.Fatalf("unknown subcommand '%s'", rest[0])
	}
}

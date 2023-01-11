package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path"
	"path/filepath"

	lb "github.com/steverusso/lockbook-x/go-lockbook"
	"github.com/urfave/cli/v2"
)

const idPrefixLen = 8

func init() {
	cli.CommandHelpTemplate = lbCommandHelpTmpl
	cli.SubcommandHelpTemplate = lbSubcommandHelpTmpl
}

func main() {
	// Figure out data directory.
	dataDir := os.Getenv("LOCKBOOK_PATH")
	if dataDir == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			log.Fatalf("getting user home dir: %v", err)
		}
		dataDir = filepath.Join(home, ".lockbook/cli")
	}
	// Initialize a new lockbook Core instance.
	core, err := lb.NewCore(dataDir)
	if err != nil {
		log.Fatalf("error: initializing core: %v", err)
	}
	app := &cli.App{
		Usage:           "An unofficial Lockbook CLI.",
		UsageText:       path.Base(os.Args[0]) + " [global options] command [options] [args...]",
		HideHelpCommand: true,
		// Check for an account before every command besides `init`.
		Before: func(c *cli.Context) error {
			first := c.Args().First()
			if first != "" && first != "init" {
				_, err = core.GetAccount()
				if err, ok := err.(*lb.Error); ok && err.Code == lb.CodeNoAccount {
					fmt.Fprintf(os.Stderr, "no account! run 'init' or 'init --restore' to get started.\n")
					os.Exit(1)
				}
				return err
			}
			return nil
		},
		// Show help and exit with error status if no command is provided.
		Action: func(c *cli.Context) error {
			cli.ShowAppHelpAndExit(c, 1)
			return nil
		},
		Commands: []*cli.Command{
			{
				Name:            "acct",
				Usage:           "account key and subscription management",
				UsageText:       "acct <sub-command> [options] [args...]",
				HideHelpCommand: true,
				Subcommands: []*cli.Command{
					{
						Name:  "privkey",
						Usage: "print out the private key for this lockbook",
						Action: func(c *cli.Context) error {
							return acctPrivKey(core)
						},
					},
					{
						Name:  "status",
						Usage: "overview of your account",
						Action: func(c *cli.Context) error {
							return acctStatus(core)
						},
					},
					{
						Name:  "subscribe",
						Usage: "create a new subscription with a credit card",
						Action: func(c *cli.Context) error {
							return acctSubscribe(core)
						},
					},
					{
						Name:  "unsubscribe",
						Usage: "cancel an existing subscription",
						Action: func(c *cli.Context) error {
							return acctUnsubscribe(core)
						},
					},
				},
			},
			{
				Name:      "cat",
				Usage:     "print one or more documents to stdout",
				UsageText: "cat <targets>...",
				Action: func(c *cli.Context) error {
					return cat(core, c.Args().Slice())
				},
			},
			{
				Name:      "drawing",
				Usage:     "export a lockbook drawing as an image written to stdout",
				UsageText: "drawing <target> [png|jpeg|pnm|tga|farbfeld|bmp]",
				Action: func(c *cli.Context) error {
					if c.NArg() == 0 {
						cli.ShowSubcommandHelpAndExit(c, 1)
					}
					return exportDrawing(core, c.Args().First(), c.Args().Get(1))
				},
			},
			{
				Name:  "debug",
				Usage: "investigative commands mainly intended for devs",
				Subcommands: []*cli.Command{
					{
						Name:        "finfo",
						Usage:       "view info about a target file",
						Description: "the target can be a file path, uuid, or uuid prefix.",
						ArgsUsage:   "<target>",
						Action: func(c *cli.Context) error {
							if c.NArg() <= 0 {
								return errors.New("must provide a file path, uuid, or uuid prefix for finfo")
							}
							return debugFinfo(core, c.Args().First())
						},
					},
					{
						Name:  "validate",
						Usage: "find invalid states within your lockbook",
						Action: func(c *cli.Context) error {
							return debugValidate(core)
						},
					},
				},
			},
			{
				Name:      "export",
				Usage:     "copy a lockbook file to your file system",
				UsageText: "export <target> [dest-dir]",
				Flags: []cli.Flag{
					&cli.BoolFlag{
						Name:    "verbose",
						Aliases: []string{"v"},
						Usage:   "print out each file as it's being exported",
					},
				},
				Action: func(c *cli.Context) error {
					if c.NArg() == 0 {
						cli.ShowSubcommandHelpAndExit(c, 1)
					}
					return exportFile(core, c.Bool("verbose"), c.Args().First(), c.Args().Get(1))
				},
			},
			{
				Name:  "init",
				Usage: "create a new account or restore an existing one",
				Flags: []cli.Flag{
					&cli.BoolFlag{
						Name:  "restore",
						Usage: "read a private key from stdin to restore the account",
					},
					&cli.BoolFlag{
						Name:  "no-sync",
						Usage: "don't perform an initial sync after the account is initialized",
					},
					&cli.BoolFlag{Name: "welcome", Usage: "include the welcome document"},
				},
				Action: func(c *cli.Context) error {
					return acctInit(core, acctInitParams{
						isRestore:    c.Bool("restore"),
						isNoSync:     c.Bool("no-sync"),
						isWelcomeDoc: c.Bool("welcome"),
					})
				},
			},
			{
				Name:      "ls",
				Usage:     "list files in a directory",
				UsageText: "ls [options] <directory>",
				Flags: []cli.Flag{
					&cli.BoolFlag{
						Name:    "short",
						Aliases: []string{"s"},
						Usage:   "just display the name (or file path)",
					},
					&cli.BoolFlag{
						Name:    "recursive",
						Aliases: []string{"r"},
						Usage:   "recursively include all children of the target directory",
					},
					&cli.BoolFlag{
						Name:  "paths",
						Usage: "show absolute file paths instead of file names",
					},
					&cli.BoolFlag{Name: "dirs", Usage: "only show folders"},
					&cli.BoolFlag{Name: "docs", Usage: "only show documents"},
					&cli.BoolFlag{
						Name:  "full-ids",
						Usage: "show full file uuids instead of prefixes",
					},
				},
				Action: func(c *cli.Context) error {
					target := c.Args().First()
					if target == "" {
						target = "/"
					}
					return listFiles(core, lsParams{
						short:     c.Bool("short"),
						recursive: c.Bool("recursive"),
						paths:     c.Bool("paths"),
						onlyDirs:  c.Bool("dirs"),
						onlyDocs:  c.Bool("docs"),
						fullIDs:   c.Bool("full-ids"),
						target:    target,
					})
				},
			},
			{
				Name:      "mkdir",
				Usage:     "create a folder or do nothing if it exists",
				UsageText: "mkdir <path>",
				Action: func(c *cli.Context) error {
					if c.NArg() == 0 {
						cli.ShowSubcommandHelpAndExit(c, 1)
					}
					v := c.Args().First()
					if v != "/" && v[len(v)-1] != '/' {
						v += "/"
					}
					return mk(core, string(v))
				},
			},
			{
				Name:      "mkdoc",
				Usage:     "create a document or do nothing if it exists",
				UsageText: "mkdoc <path>",
				Action: func(c *cli.Context) error {
					if c.NArg() == 0 {
						cli.ShowSubcommandHelpAndExit(c, 1)
					}
					v := c.Args().First()
					if v != "/" && v[len(v)-1] == '/' {
						v = v[:len(v)-1]
					}
					return mk(core, string(v))
				},
			},
			{
				Name:      "rename",
				Usage:     "rename a file",
				UsageText: "rename [-f] <target> [new-name]",
				Flags: []cli.Flag{
					&cli.BoolFlag{
						Name:    "force",
						Aliases: []string{"f"},
						Usage:   "non-interactive (fail instead of prompting for corrections)",
					},
				},
				Action: func(c *cli.Context) error {
					if c.NArg() == 0 {
						cli.ShowSubcommandHelpAndExit(c, 1)
					}
					isForce := c.Bool("force")
					target := c.Args().First()
					newName := c.Args().Get(1)
					if newName == "" && isForce {
						fmt.Fprintf(os.Stderr, "error: must provide new name if --force\n")
						cli.ShowSubcommandHelpAndExit(c, 1)
					}
					return rename(core, target, newName, isForce)
				},
			},
			{
				Name:      "mv",
				Usage:     "move a file to another parent",
				UsageText: "mv <src> <dest>",
				Action: func(c *cli.Context) error {
					if c.NArg() < 2 {
						cli.ShowSubcommandHelpAndExit(c, 1)
					}
					srcTarget := c.Args().Get(0)
					destTarget := c.Args().Get(1)
					return moveFile(core, srcTarget, destTarget)
				},
			},
			{
				Name:      "rm",
				Usage:     "delete files",
				UsageText: "rm [-f] <targets>...",
				Flags: []cli.Flag{
					&cli.BoolFlag{
						Name:    "force",
						Aliases: []string{"f"},
						Usage:   "non-interactive (fail instead of prompting for corrections)",
					},
				},
				Action: func(c *cli.Context) error {
					if c.NArg() == 0 {
						fmt.Fprintf(os.Stderr, "error: no targets provided\n")
						cli.ShowSubcommandHelpAndExit(c, 1)
					}
					targets := c.Args().Slice()
					return deleteFiles(core, targets, c.Bool("force"))
				},
			},
			{
				Name:            "share",
				Usage:           "manage shared files",
				UsageText:       "share <sub-command> [options] [args...]",
				HideHelpCommand: true,
				Subcommands: []*cli.Command{
					{
						Name:  "new",
						Usage: "create a new shared file",
						Flags: []cli.Flag{
							&cli.BoolFlag{
								Name:  "ro",
								Usage: "read only: the other user will not be able to change this file",
							},
						},
						Action: func(c *cli.Context) error {
							if c.NArg() == 0 {
								fmt.Fprintf(os.Stderr, "error: no target or username provided\n")
								cli.ShowSubcommandHelpAndExit(c, 1)
							}
							if c.NArg() == 1 {
								fmt.Fprintf(os.Stderr, "error: no username provided\n")
								cli.ShowSubcommandHelpAndExit(c, 1)
							}
							target := c.Args().First()
							toWho := c.Args().Get(1)
							return createShare(core, target, toWho, c.Bool("ro"))
						},
					},
					{
						Name:  "list",
						Usage: "list any pending shares",
						Flags: []cli.Flag{
							&cli.BoolFlag{
								Name:  "full-ids",
								Usage: "print full uuids instead of prefixs",
							},
						},
						Action: func(c *cli.Context) error {
							return listPendingShares(core, c.Bool("full-ids"))
						},
					},
					{
						Name:      "accept",
						Usage:     "move a pending share in to your file tree",
						UsageText: "accept [--rename <name>] <target> [dest-dir]",
						Flags: []cli.Flag{
							&cli.StringFlag{
								Name:  "rename",
								Usage: "use a different name for the file in your file tree",
							},
						},
						Action: func(c *cli.Context) error {
							if c.NArg() == 0 {
								fmt.Fprintf(os.Stderr, "error: no pending share id provided\n")
								cli.ShowSubcommandHelpAndExit(c, 1)
							}
							dest := c.Args().Get(1)
							if dest == "" {
								dest = "/"
							}
							return acceptShare(core, c.Args().First(), dest, c.String("rename"))
						},
					},
					{
						Name:      "delete",
						Usage:     "delete a pending share",
						UsageText: "delete <id>",
						Action: func(c *cli.Context) error {
							return deletePendingShare(core, c.Args().First())
						},
					},
				},
			},
			{
				Name:  "status",
				Usage: "which operations a sync would perform",
				Action: func(c *cli.Context) error {
					return acctSyncStatus(core)
				},
			},
			{
				Name:  "sync",
				Usage: "get updates from the server and push changes",
				Flags: []cli.Flag{
					&cli.BoolFlag{
						Name:    "verbose",
						Aliases: []string{"v"},
						Usage:   "print each file as it's being synced",
					},
				},
				Action: func(c *cli.Context) error {
					return acctSyncAll(core, c.Bool("verbose"))
				},
			},
			{
				Name:  "usage",
				Usage: "local and server disk utilization (uncompressed and compressed)",
				Flags: []cli.Flag{
					&cli.BoolFlag{Name: "exact", Usage: "show amounts in bytes"},
				},
				Action: func(c *cli.Context) error {
					return acctUsage(core, c.Bool("exact"))
				},
			},
			{
				Name:  "whoami",
				Usage: "print information about this lockbook instance",
				Flags: []cli.Flag{
					&cli.BoolFlag{
						Name:    "long",
						Aliases: []string{"l"},
						Usage:   "print your data directory and server url as well",
					},
				},
				Action: func(c *cli.Context) error {
					return acctWhoAmI(core, c.Bool("long"), dataDir)
				},
			},
			{
				Name:      "write",
				Usage:     "write data from stdin to a lockbook document",
				UsageText: "write [options]",
				Flags: []cli.Flag{
					&cli.BoolFlag{
						Name:  "trunc",
						Usage: "truncate the file instead of appending to it",
					},
				},
				Action: func(c *cli.Context) error {
					if c.NArg() == 0 {
						fmt.Fprintf(os.Stderr, "error: no targets provided\n")
						cli.ShowSubcommandHelpAndExit(c, 1)
					}
					return writeDoc(core, c.Args().First(), c.Bool("trunc"))
				},
			},
		},
	}
	if err := app.Run(os.Args); err != nil {
		fmt.Printf("\033[1;31merror:\033[0m %v\n", err)
	}
}

const lbCommandHelpTmpl = `NAME:
   {{.HelpName}} - {{.Usage}}

USAGE:
   {{if .UsageText}}{{.UsageText}}{{else}}{{.HelpName}}{{if .VisibleFlags}} [options]{{end}} {{if .ArgsUsage}}{{.ArgsUsage}}{{else}}[arguments...]{{end}}{{end}}{{if .Category}}

CATEGORY:
   {{.Category}}{{end}}{{if .Description}}

DESCRIPTION:
   {{.Description}}{{end}}{{if .VisibleFlags}}

OPTIONS:{{range .VisibleFlags}}
   {{.}}{{end}}{{end}}
`

const lbSubcommandHelpTmpl = `NAME:
   {{.HelpName}} - {{if .Description}}{{.Description}}{{else}}{{.Usage}}{{end}}

USAGE:
   {{if .UsageText}}{{.UsageText}}{{else}}{{.HelpName}} command{{if .VisibleFlags}} [options]{{end}} {{if .ArgsUsage}}{{.ArgsUsage}}{{else}}[arguments...]{{end}}{{end}}

COMMANDS:{{range .VisibleCategories}}{{if .Name}}
   {{.Name}}:{{end}}{{range .VisibleCommands}}
   {{join .Names ", "}}{{"\t"}}{{.Usage}}{{end}}
{{end}}{{if .VisibleFlags}}
OPTIONS:{{range .VisibleFlags}}
   {{.}}{{end}}{{end}}
`

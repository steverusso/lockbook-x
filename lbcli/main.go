package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/gofrs/uuid"
	"github.com/steverusso/lockbook-x/go-lockbook"
)

const idPrefixLen = 8

//go:generate goclap -type lbcli

// an unofficial lockbook cli implemented in go
type lbcli struct {
	acct   *acctCmd
	cat    *catCmd
	debug  *debugCmd
	export *exportCmd
	imprt  *importCmd
	init   *initCmd
	jot    *jotCmd
	ls     *lsCmd
	mkdir  *mkdirCmd
	mkdoc  *mkdocCmd
	mv     *mvCmd
	rename *renameCmd
	rm     *rmCmd
	share  *shareCmd
	sync   *syncCmd
	usage  *usageCmd
	write  *writeCmd
}

// get updates from the server and push changes
type syncCmd struct {
	// show last synced and which operations a sync would perform
	//
	// clap:opt status,s
	status bool
	// output every sync step and progress
	//
	// clap:opt verbose,v
	verbose bool
}

func (c *syncCmd) run(core lockbook.Core) error {
	if c.status {
		if err := printSyncStatus(core); err != nil {
			return fmt.Errorf("getting sync status: %w", err)
		}
		return nil
	}
	var syncProgress func(lockbook.SyncProgress)
	if c.verbose {
		syncProgress = func(sp lockbook.SyncProgress) {
			fmt.Printf("(%d/%d) ", sp.Progress, sp.Total)
			cwu := sp.CurrentWorkUnit
			switch {
			case cwu.PullMetadata:
				fmt.Println("pulling metadata updates...")
			case cwu.PushMetadata:
				fmt.Println("pushing metadata updates...")
			case cwu.PullDocument != "":
				fmt.Printf("pulling %s...\n", cwu.PullDocument)
			case cwu.PushDocument != "":
				fmt.Printf("pushing %s...\n", cwu.PushDocument)
			}
		}
	}
	err := core.SyncAll(syncProgress)
	if err != nil {
		fmt.Println()
		return err
	}
	if c.verbose {
		fmt.Println("done")
	}
	return nil
}

func printSyncStatus(core lockbook.Core) error {
	wc, err := core.CalculateWork()
	if err != nil {
		return fmt.Errorf("calculating work: %w", err)
	}
	for _, wu := range wc.WorkUnits {
		pushOrPull := "pushed"
		if wu.Type == lockbook.WorkUnitTypeServer {
			pushOrPull = "pulled"
		}
		fmt.Printf("%s needs to be %s\n", wu.File.Name, pushOrPull)
	}
	lastSyncedAt, err := core.GetLastSyncedHumanString()
	if err != nil {
		return fmt.Errorf("getting last synced human string: %w", err)
	}
	fmt.Printf("last synced: %s\n", lastSyncedAt)
	return nil
}

// local and server disk utilization (uncompressed and compressed)
//
// clap:cmd_usage [-e]
type usageCmd struct {
	// show amounts in bytes
	//
	// clap:opt exact,e
	exact bool
}

func (c *usageCmd) run(core lockbook.Core) error {
	u, err := core.GetUsage()
	if err != nil {
		return fmt.Errorf("getting usage: %w", err)
	}
	uu, err := core.GetUncompressedUsage()
	if err != nil {
		return fmt.Errorf("getting uncompressed usage: %w", err)
	}

	uncompressed := uu.Readable
	serverUsage := u.ServerUsage.Readable
	dataCap := u.DataCap.Readable
	if c.exact {
		uncompressed = fmt.Sprintf("%d B", uu.Exact)
		serverUsage = fmt.Sprintf("%d B", u.ServerUsage.Exact)
		dataCap = fmt.Sprintf("%d B", u.DataCap.Exact)
	}

	fmt.Printf("uncompressed file size: %s\n", uncompressed)
	fmt.Printf("server utilization: %s\n", serverUsage)
	fmt.Printf("server data cap: %s\n", dataCap)
	return nil
}

func idFromSomething(core lockbook.Core, v string) (uuid.UUID, error) {
	if id := uuid.FromStringOrNil(v); !id.IsNil() {
		return id, nil
	}
	f, err := core.FileByPath(v)
	if err == nil {
		return f.ID, nil
	}
	if err, ok := err.(*lockbook.Error); ok && err.Code != lockbook.CodeFileNonexistent {
		return uuid.Nil, fmt.Errorf("trying to get a file by path: %w", err)
	}
	// Not a full UUID and not a path, so that leaves UUID prefix.
	files, err := core.ListMetadatas()
	if err != nil {
		return uuid.Nil, fmt.Errorf("listing metadatas to check ids: %w", err)
	}
	possibs := make([]lockbook.File, 0, 5)
	for i := range files {
		if strings.HasPrefix(files[i].ID.String(), v) {
			possibs = append(possibs, files[i])
		}
	}
	n := len(possibs)
	if n == 0 {
		return uuid.Nil, fmt.Errorf("value %q is not a path, uuid, or uuid prefix", v)
	}
	if n == 1 {
		return possibs[0].ID, nil
	}
	// Multiple ID prefix matches.
	errMsg := fmt.Sprintf("value %q is not a path and matches %d file ID prefixes:\n", v, n)
	for _, f := range possibs {
		pathOrErr, err := core.PathByID(f.ID)
		if err != nil {
			pathOrErr = fmt.Sprintf("error getting path: %v", err)
		}
		errMsg += fmt.Sprintf("  %s  %s\n", f.ID, pathOrErr)
	}
	return uuid.Nil, errors.New(errMsg)
}

func run() error {
	// Figure out data directory.
	dataDir := os.Getenv("LOCKBOOK_PATH")
	if dataDir == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("getting user home dir: %v", err)
		}
		dataDir = filepath.Join(home, ".lockbook/cli")
	}

	// Initialize a new lockbook Core instance.
	core, err := lockbook.NewCore(dataDir)
	if err != nil {
		return fmt.Errorf("initializing core: %v", err)
	}

	a := lbcli{}
	a.parse(os.Args)

	// Check for an account before every command besides `init`.
	if a.init == nil && (a.acct == nil || a.acct.restore == nil) {
		_, err = core.GetAccount()
		if err, ok := err.(*lockbook.Error); ok && err.Code == lockbook.CodeAccountNonexistent {
			return errors.New("no account! run 'init' or 'init --restore' to get started.")
		}
		if err != nil {
			return fmt.Errorf("getting account: %v", err)
		}
	}

	switch {
	case a.acct != nil:
		switch {
		case a.acct.restore != nil:
			return a.acct.restore.run(core)
		case a.acct.privkey != nil:
			return a.acct.privkey.run(core)
		case a.acct.status != nil:
			return a.acct.status.run(core)
		case a.acct.subscribe != nil:
			return a.acct.subscribe.run(core)
		case a.acct.unsubscribe != nil:
			return a.acct.unsubscribe.run(core)
		}
	case a.cat != nil:
		return a.cat.run(core)
	case a.debug != nil:
		switch {
		case a.debug.finfo != nil:
			return a.debug.finfo.run(core)
		case a.debug.validate != nil:
			return a.debug.validate.run(core)
		case a.debug.whoami != nil:
			return a.debug.whoami.run(core)
		}
	case a.export != nil:
		return a.export.run(core)
	case a.imprt != nil:
		return a.imprt.run(core)
	case a.init != nil:
		return a.init.run(core)
	case a.jot != nil:
		return a.jot.run(core)
	case a.ls != nil:
		return a.ls.run(core)
	case a.mkdir != nil:
		return a.mkdir.run(core)
	case a.mkdoc != nil:
		return a.mkdoc.run(core)
	case a.mv != nil:
		return a.mv.run(core)
	case a.rename != nil:
		return a.rename.run(core)
	case a.rm != nil:
		return a.rm.run(core)
	case a.share != nil:
		switch {
		case a.share.create != nil:
			return a.share.create.run(core)
		case a.share.pending != nil:
			return a.share.pending.run(core)
		case a.share.accept != nil:
			return a.share.accept.run(core)
		case a.share.reject != nil:
			return a.share.reject.run(core)
		}
	case a.sync != nil:
		return a.sync.run(core)
	case a.usage != nil:
		return a.usage.run(core)
	case a.write != nil:
		return a.write.run(core)
	}
	return nil
}

func main() {
	if err := run(); err != nil {
		fmt.Printf("\033[1;31merror:\033[0m %v\n", err)
		os.Exit(1)
	}
}

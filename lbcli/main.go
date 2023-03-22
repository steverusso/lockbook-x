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

// An unofficial lockbook cli.
type lbcli struct {
	acct   *acctCmd
	cat    *catCmd
	debug  *debugCmd
	export *exportCmd
	imprt  *importCmd
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

// Get updates from the server and push changes.
type syncCmd struct {
	// Show last synced and which operations a sync would perform.
	//
	// clap:opt status,s
	status bool
	// Output every sync step and progress.
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
				fmt.Println("-> file tree updates...")
			case cwu.PushMetadata:
				fmt.Println("<- file tree updates...")
			case cwu.PullDocument != nil:
				idPrefix := cwu.PullDocument.ID.String()[:idPrefixLen]
				fmt.Printf("-> %s (%s)...\n", cwu.PullDocument.Name, idPrefix)
			case cwu.PushDocument != nil:
				idPrefix := cwu.PushDocument.ID.String()[:idPrefixLen]
				fmt.Printf("<- %s (%s)...\n", cwu.PushDocument.Name, idPrefix)
			}
		}
	}
	err := core.SyncAll(syncProgress)
	if err != nil {
		return fmt.Errorf("syncing: %w", err)
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

// Local and server disk utilization (uncompressed and compressed).
//
// clap:cmd_usage [-e]
type usageCmd struct {
	// Show amounts in bytes instead of as human readable values.
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
	if err, ok := asLbErr(err); ok && err.Code != lockbook.CodeFileNonexistent {
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

func asLbErr(err error) (*lockbook.Error, bool) {
	var lberr *lockbook.Error
	if errors.As(err, &lberr) {
		return lberr, true
	}
	return nil, false
}

func isStdinPipe() bool {
	fi, err := os.Stdin.Stat()
	if err != nil {
		panic(err)
	}
	return fi.Mode()&os.ModeNamedPipe != 0
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

	lb := lbcli{}
	lb.parse(os.Args)

	// Check for an account before every command besides `init` or `restore`.
	if lb.acct == nil || lb.acct.init == nil || lb.acct.restore == nil {
		_, err = core.GetAccount()
		if err, ok := err.(*lockbook.Error); ok && err.Code == lockbook.CodeAccountNonexistent {
			return errors.New("no account! run 'acct init' or 'acct restore' to get started.")
		}
		if err != nil {
			return fmt.Errorf("getting account: %v", err)
		}
	}

	switch {
	case lb.acct != nil:
		return lb.acct.run(core)
	case lb.cat != nil:
		return lb.cat.run(core)
	case lb.debug != nil:
		return lb.debug.run(core)
	case lb.export != nil:
		return lb.export.run(core)
	case lb.imprt != nil:
		return lb.imprt.run(core)
	case lb.jot != nil:
		return lb.jot.run(core)
	case lb.ls != nil:
		return lb.ls.run(core)
	case lb.mkdir != nil:
		return lb.mkdir.run(core)
	case lb.mkdoc != nil:
		return lb.mkdoc.run(core)
	case lb.mv != nil:
		return lb.mv.run(core)
	case lb.rename != nil:
		return lb.rename.run(core)
	case lb.rm != nil:
		return lb.rm.run(core)
	case lb.share != nil:
		return lb.share.run(core)
	case lb.sync != nil:
		return lb.sync.run(core)
	case lb.usage != nil:
		return lb.usage.run(core)
	case lb.write != nil:
		return lb.write.run(core)
	}
	return nil
}

func main() {
	if err := run(); err != nil {
		fmt.Printf("\033[1;31merror:\033[0m %v\n", err)
		if err, ok := asLbErr(err); ok && err.Trace != "" {
			fmt.Println(err.Trace)
			os.Exit(int(err.Code))
		}
		os.Exit(1)
	}
}

package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"

	lb "github.com/steverusso/lockbook-x/go-lockbook"
)

type acctInitParams struct {
	isRestore    bool
	isNoSync     bool
	isWelcomeDoc bool
}

func acctInit(core lb.Core, ip acctInitParams) error {
	fi, err := os.Stdin.Stat()
	if err != nil {
		panic(err)
	}
	if ip.isRestore {
		if fi.Mode()&os.ModeNamedPipe == 0 {
			return errors.New("to restore an existing lockbook account, pipe your account string into this command, e.g.:\npbpaste | lockbook init --restore")
		}
		data, err := io.ReadAll(os.Stdin)
		if err != nil {
			return fmt.Errorf("trying to read from stdin: %w", err)
		}
		_, err = core.ImportAccount(string(data))
		if err != nil {
			return fmt.Errorf("importing account: %w", err)
		}
	} else {
		if fi.Mode()&os.ModeNamedPipe != 0 {
			return errors.New("cannot create a new account without ability to prompt for terminal input")
		}
		scnr := bufio.NewScanner(os.Stdin)
		fmt.Printf("please choose a username: ")
		if !scnr.Scan() {
			return scnr.Err()
		}
		uname := scnr.Text()
		apiURL := os.Getenv("API_URL")
		if apiURL == "" {
			apiURL = lb.DefaultAPILocation()
		}
		fmt.Println("generating keys and checking for username availability...")
		_, err = core.CreateAccount(uname, ip.isWelcomeDoc)
		if err != nil {
			return fmt.Errorf("creating account: %w", err)
		}
		fmt.Println("account created!")
	}
	if !ip.isNoSync {
		fmt.Print("doing initial sync... ")
		if err := core.SyncAll(nil); err != nil {
			return fmt.Errorf("syncing after init: %w", err)
		}
		fmt.Println("done")
	}
	return nil
}

func acctWhoAmI(core lb.Core, long bool, dir string) error {
	acct, err := core.GetAccount()
	if err != nil {
		return fmt.Errorf("getting account: %w", err)
	}
	if !long {
		fmt.Println(acct.Username)
		return nil
	}
	fmt.Printf("data-dir: %s\n", dir)
	fmt.Printf("username: %s\n", acct.Username)
	fmt.Printf("server:   %s\n", acct.APIURL)
	return nil
}

func acctPrivKey(core lb.Core) error {
	acctStr, err := core.ExportAccount()
	if err != nil {
		return fmt.Errorf("exporting account: %w", err)
	}
	answer := ""
	fmt.Print("your private key is about to be visible. do you want to proceed? [y/N]: ")
	fmt.Scanln(&answer)
	if answer != "y" && answer != "Y" {
		fmt.Println("aborted")
		return nil
	}
	fmt.Println(acctStr)
	return nil
}

func acctStatus(core lb.Core) error {
	u, err := core.GetUsage()
	if err != nil {
		return fmt.Errorf("getting usage: %w", err)
	}
	info, err := core.GetSubscriptionInfo()
	if err != nil {
		return fmt.Errorf("getting subscription info: %w", err)
	}
	switch {
	case info.StripeLast4 != "":
		fmt.Printf("type: Stripe, *%s\n", info.StripeLast4)
	case info.GooglePlay != lb.GooglePlayNone:
		fmt.Println("type: Google Play")
		fmt.Printf("state: %s\n", info.GooglePlay)
	case info.AppStore != lb.AppStoreNone:
		fmt.Println("type: App Store")
		fmt.Printf("state: %s\n", info.AppStore)
	default:
		fmt.Println("trial tier")
	}
	if !info.PeriodEnd.IsZero() {
		fmt.Printf("renews on: %s\n", info.PeriodEnd)
	}
	pct := (u.ServerUsage.Exact * 100) / u.DataCap.Exact
	fmt.Printf("data cap: %s, %d%% utilized\n", u.DataCap.Readable, pct)
	return nil
}

func acctSubscribe(core lb.Core) error {
	fmt.Print("checking for existing card... ")
	subInfo, err := core.GetSubscriptionInfo()
	if err != nil {
		return fmt.Errorf("getting subscription info: %w", err)
	}
	useNewCard := true
	if subInfo.StripeLast4 != "" {
		answer := ""
		fmt.Printf("do you want to use card ending in *%s? [y/N]: ", subInfo.StripeLast4)
		fmt.Scanln(&answer)
		if answer == "y" || answer == "Y" {
			useNewCard = false
		}
	} else {
		fmt.Println("no existing cards found.")
	}
	var card *lb.CreditCard
	if useNewCard {
		card := &lb.CreditCard{}
		fmt.Print("enter your card number: ")
		fmt.Scanln(&card.Number)
		for {
			fmt.Print("expiration month: ")
			_, err := fmt.Scanln("%d", &card.ExpiryMonth)
			if err == nil {
				break
			}
			fmt.Println(err)
		}
		for {
			fmt.Print("expiration year: ")
			_, err := fmt.Scanf("%d", &card.ExpiryYear)
			if err == nil {
				break
			}
			fmt.Println(err)
		}
		fmt.Print("cvc: ")
		fmt.Scanln(&card.CVC)
	}
	if err = core.UpgradeViaStripe(card); err != nil {
		return err
	}
	fmt.Println("subscribed!")
	return nil
}

func acctUnsubscribe(core lb.Core) error {
	fmt.Print("cancelling subscription... ")
	err := core.CancelSubscription()
	if err != nil {
		return err
	}
	fmt.Println("done")
	return nil
}

func acctSyncAll(core lb.Core, isVerbose bool) error {
	var syncProgress func(lb.SyncProgress)
	if isVerbose {
		syncProgress = func(sp lb.SyncProgress) {
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
	if isVerbose {
		fmt.Println("done")
	}
	return nil
}

func acctSyncStatus(core lb.Core) error {
	wc, err := core.CalculateWork()
	if err != nil {
		return fmt.Errorf("calculating work: %w", err)
	}
	for _, wu := range wc.WorkUnits {
		pushOrPull := "pushed"
		if wu.Type == lb.WorkUnitTypeServer {
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

func acctUsage(core lb.Core, isExact bool) error {
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
	if isExact {
		uncompressed = fmt.Sprintf("%d B", uu.Exact)
		serverUsage = fmt.Sprintf("%d B", u.ServerUsage.Exact)
		dataCap = fmt.Sprintf("%d B", u.DataCap.Exact)
	}

	fmt.Printf("uncompressed file size: %s\n", uncompressed)
	fmt.Printf("server utilization: %s\n", serverUsage)
	fmt.Printf("server data cap: %s\n", dataCap)
	return nil
}

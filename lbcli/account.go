package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/steverusso/lockbook-x/go-lockbook"
)

// account related commands
type acctCmd struct {
	restore     *acctRestoreCmd
	privkey     *acctPrivKeyCmd
	status      *acctStatusCmd
	subscribe   *acctSubscribeCmd
	unsubscribe *acctUnsubscribeCmd
}

// restore an existing account from its secret account string
type acctRestoreCmd struct {
	noSync bool `opt:"no-sync" desc:"don't perform the initial sync"`
}

// print out the private key for this lockbook
type acctPrivKeyCmd struct{}

// overview of your account
type acctStatusCmd struct{}

// create a new subscription with a credit card
type acctSubscribeCmd struct{}

// cancel an existing subscription
type acctUnsubscribeCmd struct{}

// create a lockbook account
type initCmd struct {
	welcome bool `opt:"welcome" desc:"include the welcome document"`
	noSync  bool `opt:"no-sync" desc:"don't perform the initial sync"`
}

func (c *initCmd) run(core lockbook.Core) error {
	fi, err := os.Stdin.Stat()
	if err != nil {
		panic(err)
	}
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
		apiURL = lockbook.DefaultAPILocation
	}
	fmt.Println("generating keys and checking for username availability...")
	_, err = core.CreateAccount(uname, c.welcome)
	if err != nil {
		return fmt.Errorf("creating account: %w", err)
	}
	fmt.Println("account created!")
	if !c.noSync {
		fmt.Print("doing initial sync... ")
		if err := core.SyncAll(nil); err != nil {
			return fmt.Errorf("syncing after init: %w", err)
		}
		fmt.Println("done")
	}
	return nil
}

func (c *acctRestoreCmd) run(core lockbook.Core) error {
	fi, err := os.Stdin.Stat()
	if err != nil {
		panic(err)
	}
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
	if !c.noSync {
		fmt.Print("doing initial sync... ")
		if err := core.SyncAll(nil); err != nil {
			return fmt.Errorf("syncing after init: %w", err)
		}
		fmt.Println("done")
	}
	return nil
}

func (acctPrivKeyCmd) run(core lockbook.Core) error {
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

func (acctStatusCmd) run(core lockbook.Core) error {
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
	case info.GooglePlay != lockbook.GooglePlayNone:
		fmt.Println("type: Google Play")
		fmt.Printf("state: %s\n", info.GooglePlay)
	case info.AppStore != lockbook.AppStoreNone:
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

func (acctSubscribeCmd) run(core lockbook.Core) error {
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
	var card *lockbook.CreditCard
	if useNewCard {
		card := &lockbook.CreditCard{}
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

func (acctUnsubscribeCmd) run(core lockbook.Core) error {
	fmt.Print("cancelling subscription... ")
	err := core.CancelSubscription()
	if err != nil {
		return err
	}
	fmt.Println("done")
	return nil
}

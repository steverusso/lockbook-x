use std::io::Write;

use structopt::StructOpt;

use lb::Core;

use crate::sync;
use crate::util::input;
use crate::CliError;

#[derive(StructOpt)]
pub enum AcctCmd {
    /// Export a private key to stdout.
    Privkey,
    /// Prints out information about your account
    Status,
    /// Create a new subscription using a credit card
    Subscribe,
    /// Terminate a lockbook subscription
    Unsubscribe,
}

#[derive(StructOpt)]
pub struct InitArgs {
    /// Restore a lockbook instance by providing the private key
    #[structopt(long)]
    restore: bool,
    /// Include the welcome doc
    #[structopt(long)]
    welcome: bool,
    /// Don't perform an initial sync
    #[structopt(long)]
    no_sync: bool,
}

pub fn init(core: &Core, args: InitArgs) -> Result<(), CliError> {
    if args.restore {
        if atty::is(atty::Stream::Stdin) {
            return Err(CliError::new("to restore an existing lockbook account, pipe your account string into this command, e.g.:\npbpaste | lockbook init --restore"));
        }

        let mut account_string = String::new();
        std::io::stdin()
            .read_line(&mut account_string)
            .expect("failed to read from stdin");
        account_string.retain(|c| !c.is_whitespace());

        println!("restoring...");
        core.import_account(&account_string)?;

        println!("account restored! next, try syncing.");
    } else {
        print!("please enter a username: ");
        std::io::stdout().flush()?;

        let mut username = String::new();
        std::io::stdin()
            .read_line(&mut username)
            .expect("failed to read from stdin");
        username.retain(|c| c != '\n' && c != '\r');

        let server =
            std::env::var("API_URL").unwrap_or_else(|_| lb::DEFAULT_API_LOCATION.to_string());

        println!("generating keys and checking for username availability...");
        core.create_account(&username, &server, args.welcome)?;

        println!("account created!");
    }
    if !args.no_sync {
        sync(core, false)?;
    }
    Ok(())
}

pub fn acct(core: &Core, cmd: AcctCmd) -> Result<(), CliError> {
    match cmd {
        AcctCmd::Privkey => privkey(core),
        AcctCmd::Status => status(core),
        AcctCmd::Subscribe => subscribe(core),
        AcctCmd::Unsubscribe => cancel_subscription(core),
    }
}

fn privkey(core: &Core) -> Result<(), CliError> {
    let account_string = core.export_account()?;

    let answer: String =
        input("your private key is about to be visible. do you want to proceed? [y/n]: ")?;
    if answer != "y" && answer != "Y" {
        return Ok(());
    }

    println!("{}", account_string);
    Ok(())
}

fn status(core: &Core) -> Result<(), CliError> {
    let cap = core.get_usage()?;
    let pct = (cap.server_usage.exact * 100) / cap.data_cap.exact;

    if let Some(info) = core.get_subscription_info()? {
        match info.payment_platform {
            lb::PaymentPlatform::Stripe { card_last_4_digits } => {
                println!("type: Stripe, *{}", card_last_4_digits)
            }
            lb::PaymentPlatform::GooglePlay { account_state } => {
                println!("type: Google Play");
                println!("state: {:?}", account_state);
            }
        }
        println!("renews on: {}", info.period_end);
    } else {
        println!("trial tier");
    }
    println!("data cap: {}, {}% utilized", cap.data_cap.readable, pct);
    Ok(())
}

fn subscribe(core: &Core) -> Result<(), CliError> {
    println!("checking for existing payment methods...");
    let existing_card =
        core.get_subscription_info()?
            .and_then(|info| match info.payment_platform {
                lb::PaymentPlatform::Stripe { card_last_4_digits } => Some(card_last_4_digits),
                lb::PaymentPlatform::GooglePlay { .. } => None,
            });

    let mut use_old_card = false;
    if let Some(card) = existing_card {
        let answer: String = input(format!("do you want use *{}? [y/n]: ", card))?;
        if answer == "y" || answer == "Y" {
            use_old_card = true;
        }
    } else {
        println!("no existing cards found...");
    }

    let payment_method = if use_old_card {
        lb::PaymentMethod::OldCard
    } else {
        lb::PaymentMethod::NewCard {
            number: input("enter your card number: ")?,
            exp_year: input("expiration year: ")?,
            exp_month: input("expiration month: ")?,
            cvc: input("cvc: ")?,
        }
    };

    core.upgrade_account_stripe(lb::StripeAccountTier::Premium(payment_method))?;
    println!("subscribed!");
    Ok(())
}

fn cancel_subscription(core: &Core) -> Result<(), CliError> {
    println!("cancelling subscription... ");
    core.cancel_subscription()?;
    Ok(())
}

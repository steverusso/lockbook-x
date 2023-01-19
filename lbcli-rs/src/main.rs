mod account;
mod debug;
mod error;
mod imex;
mod ls;
mod share;
mod util;

use std::io::{self, Write};

use structopt::StructOpt;

use lb::Core;

use crate::error::CliError;

const ID_PREFIX_LEN: usize = 8;

#[derive(StructOpt)]
#[structopt(about = "Unofficial cli for Lockbook.")]
enum LbCli {
    /// Account related commands such as key management and billing
    Acct(account::AcctCmd),
    /// Print the contents of lockbook docs to stdout
    Cat {
        /// The paths or IDs of lockbook documents
        targets: Vec<String>,
    },
    /// Mostly investigative lockbook commands
    Debug(debug::DebugCmd),
    /// Export a drawing as an image
    Drawing(imex::DrawingArgs),
    /// Export files from lockbook to your file system
    Export(imex::ExportArgs),
    /// Import files from your file system into lockbook
    Import(imex::ImportArgs),
    /// Initialize a lockbook instance by creating a new account or restoring an existing one
    Init(account::InitArgs),
    /// List info about lockbook files
    Ls(ls::LsArgs),
    /// Create a new folder if it doesn't exist
    Mkdir { path: String },
    /// Create a new doc if it doesn't exist
    Mkdoc { path: String },
    /// Move a file to another parent
    Mv {
        /// Lockbook file path or ID to move
        src: String,
        /// Lockbook folder path or ID of the desired destination
        dest: String,
    },
    /// Rename a lockbook file
    Rename {
        /// Lockbook file path or ID
        target: String,
        /// The new name
        name: String,
    },
    /// Delete a file
    Rm {
        /// Lockbook file path or ID
        target: String,
        /// Don't prompt confirmation before deleting the files
        #[structopt(short, long)]
        force: bool,
    },
    /// Manage shared files
    Share(share::ShareCmd),
    /// What operations a sync would perform
    Status,
    /// Get updates and push changes
    Sync {
        /// Print information about each step
        #[structopt(long, short)]
        verbose: bool,
    },
    /// Prints local & server disk utilization (uncompressed & compressed)
    Usage {
        /// Show amounts in bytes
        #[structopt(long)]
        exact: bool,
    },
    /// Prints your lockbook username
    Whoami {
        /// Show data directory and server url as well
        #[structopt(long, short)]
        long: bool,
    },
    /// Write content from stdin to a lockbook doc
    Write {
        /// Overwrite the doc content instead of appending
        #[structopt(long)]
        trunc: bool,
        path: String,
    },
}

fn mk(core: &Core, p: &str) -> Result<(), CliError> {
    if util::maybe_get_by_path(core, p)?.is_none() {
        let _ = core.create_at_path(p)?;
    }
    Ok(())
}

fn rm(core: &Core, target: &str, force: bool) -> Result<(), CliError> {
    let f = util::file_from_path_or_id(core, target)?;

    if !force {
        let mut phrase = format!("delete '{}'", target);
        if f.is_folder() {
            let count = core
                .get_and_get_children_recursively(f.id)
                .map_err(|err| (err, f.id))?
                .len();
            phrase = format!("{phrase} and its {count} children")
        }

        let answer: String = util::input(format!("are you sure you want to {phrase}? [y/n]: "))?;
        if answer != "y" && answer != "Y" {
            println!("aborted.");
            return Ok(());
        }
    }

    core.delete_file(f.id).map_err(|err| (err, f.id).into())
}

fn status(core: &Core) -> Result<(), CliError> {
    for wu in core.calculate_work()?.work_units {
        let action = match wu {
            lb::WorkUnit::LocalChange { .. } => "pushed",
            lb::WorkUnit::ServerChange { .. } => "pulled",
        };
        println!("{} needs to be {}", wu.get_metadata().name, action)
    }
    println!("last synced: {}", core.get_last_synced_human_string()?);
    Ok(())
}

fn sync(core: &Core, verbose: bool) -> Result<(), CliError> {
    println!("syncing...");
    core.sync(if verbose {
        Some(Box::new(|sp: lb::SyncProgress| {
            use lb::ClientWorkUnit::*;
            match sp.current_work_unit {
                PullMetadata => println!("pulling file tree updates"),
                PushMetadata => println!("pushing file tree updates"),
                PullDocument(name) => println!("pulling: {}", name),
                PushDocument(name) => println!("pushing: {}", name),
            };
        }))
    } else {
        None
    })?;
    Ok(())
}

fn usage(core: &Core, exact: bool) -> Result<(), CliError> {
    let u = core.get_usage()?;
    let uncompr_u = core.get_uncompressed_usage()?;

    let (uncompressed, server_usage, data_cap) = if exact {
        (
            format!("{} B", uncompr_u.exact),
            format!("{} B", u.server_usage.exact),
            format!("{} B", u.data_cap.exact),
        )
    } else {
        (
            uncompr_u.readable,
            u.server_usage.readable,
            u.data_cap.readable,
        )
    };

    println!("uncompressed file size: {}", uncompressed);
    println!("server utilization: {}", server_usage);
    println!("server data cap: {}", data_cap);
    Ok(())
}

fn write(core: &Core, trunc: bool, path: &str) -> Result<(), CliError> {
    if path.ends_with('/') {
        return Err(CliError::new("target file path cannot be a directory"));
    }
    if atty::is(atty::Stream::Stdin) {
        return Err(CliError::new("to write some file content, pipe the content into this command, e.g.:\ncat hello.txt | lockbook write /path/to/doc"));
    }
    // Read the new content from stdin. If this fails, nothing in lockbook has changed at this
    // point.
    let mut new_content = Vec::new();
    for ln in io::stdin().lines() {
        writeln!(&mut new_content, "{}", ln?)?;
    }
    // Create the target file if it doesn't exist.
    let doc_id = if let Some(f) = util::maybe_get_by_path(core, path)? {
        f.id
    } else {
        core.create_at_path(path)?.id
    };
    // The overall content is everything if we are appending (default), or just the new content if
    // truncating.
    let content = if trunc {
        new_content
    } else {
        let mut tmp = core.read_document(doc_id).map_err(|err| (err, doc_id))?;
        tmp.append(&mut new_content);
        tmp
    };
    core.write_document(doc_id, &content)
        .map_err(|err| (err, doc_id).into())
}

fn run() -> Result<(), CliError> {
    let writeable_path = match (std::env::var("LOCKBOOK_PATH"), std::env::var("HOME")) {
        (Ok(s), _) => s,
        (Err(_), Ok(s)) => format!("{}/.lockbook/cli", s),
        _ => return Err(CliError::new("no cli location")),
    };

    let core = Core::init(&lb::Config {
        writeable_path,
        logs: true,
        colored_logs: true,
    })?;

    let cli = LbCli::from_args();
    if !matches!(cli, LbCli::Init(_)) {
        let _ = core.get_account().map_err(|err| match err {
            lb::Error::UiError(lb::GetAccountError::NoAccount) => {
                CliError::new("no account! run 'init' or 'init --restore' to get started.")
            }
            err => err.into(),
        })?;
    }

    match cli {
        LbCli::Acct(cmd) => account::acct(&core, cmd),
        LbCli::Cat { targets } => {
            for v in &targets {
                let id = util::id_from_path_or_id(&core, v)?;
                let content = core.read_document(id).map_err(|err| (err, id))?;
                print!("{}", String::from_utf8_lossy(&content));
                io::stdout().flush()?;
            }
            Ok(())
        }
        LbCli::Debug(cmd) => debug::debug(&core, cmd),
        LbCli::Drawing(args) => imex::drawing(&core, args),
        LbCli::Export(args) => imex::export(&core, args),
        LbCli::Import(args) => imex::import(&core, args),
        LbCli::Init(args) => account::init(&core, args),
        LbCli::Ls(args) => ls::ls(&core, args),
        LbCli::Mkdir { mut path } => {
            if !path.ends_with('/') {
                path += "/";
            }
            mk(&core, &path)
        }
        LbCli::Mkdoc { mut path } => {
            if path.ends_with('/') {
                path.pop();
            }
            mk(&core, &path)
        }
        LbCli::Mv { src, dest } => {
            let src_id = util::id_from_path_or_id(&core, &src)?;
            let dest_id = util::id_from_path_or_id(&core, &dest)?;
            core.move_file(src_id, dest_id)
                .map_err(|err| (err, src_id, dest_id).into())
        }
        LbCli::Rename { target, name } => {
            let id = util::id_from_path_or_id(&core, &target)?;
            core.rename_file(id, &name).map_err(|err| (err, id).into())
        }
        LbCli::Rm { target, force } => rm(&core, &target, force),
        LbCli::Share(cmd) => share::share(&core, cmd),
        LbCli::Status => status(&core),
        LbCli::Sync { verbose } => sync(&core, verbose),
        LbCli::Usage { exact } => usage(&core, exact),
        LbCli::Whoami { long } => account::whoami(&core, long),
        LbCli::Write { trunc, path } => write(&core, trunc, &path),
    }
}

fn main() {
    if let Err(err) = run() {
        eprintln!("{}", err);
        std::process::exit(1)
    }
}

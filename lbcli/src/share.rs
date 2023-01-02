use structopt::StructOpt;

use lb::Core;
use lb::Uuid;

use crate::util::id_from_something;
use crate::util::maybe_get_by_path;
use crate::CliError;
use crate::ID_PREFIX_LEN;

#[derive(StructOpt)]
pub enum ShareCmd {
    /// Share a specified document with a user
    New(NewArgs),
    /// List share requests for you, from other people
    Pending {
        /// Display full file IDs instead of prefixes
        #[structopt(long)]
        full_ids: bool,
    },
    /// Accept a share and place it within your lockbook
    Accept(AcceptArgs),
}

#[derive(StructOpt)]
pub struct NewArgs {
    /// Lockbook file path or ID
    target: String,
    username: String,
    /// Read-only (the other user will not be able to edit the shared file)
    #[structopt(long = "ro")]
    read_only: bool,
}

#[derive(StructOpt)]
pub struct AcceptArgs {
    /// ID (full or prefix) of a pending share
    id: String,
    /// Lockbook file path or ID
    #[structopt(default_value = "/")]
    dest: String,
    #[structopt(long)]
    name: Option<String>,
}

pub fn share(core: &Core, share: ShareCmd) -> Result<(), CliError> {
    match share {
        ShareCmd::New(args) => new(core, args),
        ShareCmd::Pending { full_ids } => pending(core, full_ids),
        ShareCmd::Accept(args) => accept(core, args),
    }
}

struct ShareInfo {
    id: Uuid,
    mode: String,
    name: String,
    from: String,
}

fn to_share_infos(files: Vec<lb::File>) -> Vec<ShareInfo> {
    let mut infos: Vec<ShareInfo> = files
        .into_iter()
        .map(|f| {
            let from = f
                .shares
                .get(0)
                .map(|sh| sh.shared_by.clone())
                .unwrap_or_default();
            let mode = f
                .shares
                .get(0)
                .map(|sh| sh.mode)
                .unwrap_or(lb::ShareMode::Write);
            ShareInfo {
                id: f.id,
                mode: mode.to_string().to_lowercase(),
                name: f.name,
                from,
            }
        })
        .collect();
    infos.sort_by(|a, b| a.from.cmp(&b.from));
    infos
}

fn print_share_infos(infos: &Vec<ShareInfo>, full_ids: bool) {
    // Determine each column's max width.
    let w_id = if full_ids {
        Uuid::nil().to_string().len()
    } else {
        ID_PREFIX_LEN
    };
    let mut w_from = 0;
    let mut w_name = 0;
    let mut w_mode = 0;
    for info in infos {
        let n = info.mode.len();
        if n > w_mode {
            w_mode = n;
        }
        let n = info.from.len();
        if n > w_from {
            w_from = n;
        }
        let n = info.name.len();
        if n > w_name {
            w_name = n;
        }
    }
    // Print the table column headers.
    println!(
        " {:<w_id$} | {:<w_mode$} | {:<w_from$} | file",
        "id",
        "mode",
        "from",
        w_id = w_id,
        w_mode = w_mode,
        w_from = w_from
    );
    println!(
        "-{:-<w_id$}-+-{:-<w_mode$}-+-{:-<w_from$}-+-{:-<w_name$}-",
        "",
        "",
        "",
        "",
        w_id = w_id,
        w_mode = w_mode,
        w_from = w_from,
        w_name = w_name
    );
    // Print the table rows of pending share infos.
    for info in infos {
        println!(
            " {:<w_id$} | {:<w_mode$} | {:<w_from$} | {}",
            &info.id.to_string()[..w_id],
            info.mode,
            info.from,
            info.name,
            w_id = w_id,
            w_mode = w_mode,
            w_from = w_from
        );
    }
}

fn new(core: &Core, args: NewArgs) -> Result<(), CliError> {
    let id = id_from_something(core, &args.target)?;
    let mode = if args.read_only {
        lb::ShareMode::Read
    } else {
        lb::ShareMode::Write
    };
    core.share_file(id, &args.username, mode)?;
    println!("done!\nfile '{}' will be shared next time you sync.", id);
    Ok(())
}

fn pending(core: &Core, full_ids: bool) -> Result<(), CliError> {
    let pending_shares = to_share_infos(core.get_pending_shares()?);
    if pending_shares.is_empty() {
        println!("no pending shares.");
        return Ok(());
    }
    print_share_infos(&pending_shares, full_ids);
    Ok(())
}

fn accept(core: &Core, args: AcceptArgs) -> Result<(), CliError> {
    let pendings = core.get_pending_shares()?;

    let maybe_share = if let Ok(id) = Uuid::parse_str(&args.id) {
        pendings.iter().find(|f| f.id == id).cloned()
    } else {
        let possibs: Vec<lb::File> = pendings
            .into_iter()
            .filter(|f| f.id.to_string().starts_with(&args.id))
            .collect();
        match possibs.len() {
            0 => None,
            1 => Some(possibs[0].clone()),
            n => {
                println!(
                    "id prefix '{}' matched the following {} pending shares:\n",
                    args.id, n
                );
                let possib_infos = to_share_infos(possibs);
                print_share_infos(&possib_infos, true);
                return Ok(());
            }
        }
    };

    let share = match maybe_share {
        Some(s) => s,
        None => {
            return Err(CliError(format!(
                "unable to find share with id '{}'",
                args.id
            )))
        }
    };

    // If a destination ID is provided, it must be of an existing directory.
    let parent_id = if let Ok(id) = Uuid::parse_str(&args.dest) {
        let f = core.get_file_by_id(id).map_err(|err| (err, id))?;
        if !f.is_folder() {
            return Err(CliError::new(
                "destination ID must be of an existing folder",
            ));
        }
        id
    } else {
        // If the destination path exists, it must be a directory. The link will be dropped in it.
        let mut path = args.dest.clone();
        if let Some(f) = maybe_get_by_path(core, &path)? {
            if !f.is_folder() {
                return Err(CliError::new(
                    "existing destination path is a doc, must be a folder",
                ));
            }
            f.id
        } else {
            // If the destination path doesn't exist, then it's just treated as a non-existent
            // directory path. The user can set the name with the `--name` input option.
            if !path.ends_with('/') {
                path += "/";
            }
            let f = core.create_at_path(&path)?;
            f.id
        }
    };

    let mut name = args.name.unwrap_or_else(|| share.name.clone());
    if name.ends_with('/') {
        name.pop(); // Prevent "name contains slash" error.
    }

    core.create_file(&name, parent_id, lb::FileType::Link { target: share.id })
        .map_err(|err| CliError(format!("{:?}", err)))?;
    Ok(())
}

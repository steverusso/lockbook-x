use structopt::StructOpt;

use crate::util::file_from_path_or_id;
use crate::CliError;

#[derive(StructOpt)]
pub enum DebugCmd {
    /// Prints metadata of a lockbook file via path or ID
    Finfo(FinfoArgs),
    /// Helps find invalid states within lockbook
    Validate,
}

#[derive(StructOpt)]
pub struct FinfoArgs {
    /// Lockbook file path or ID
    target: String,
    /// Print out the raw Rust `Debug` text of the lockbook file
    #[structopt(long)]
    raw: bool,
}

pub fn debug(core: &lb::Core, cmd: DebugCmd) -> Result<(), CliError> {
    match cmd {
        DebugCmd::Finfo(args) => finfo(core, args),
        DebugCmd::Validate => validate(core),
    }
}

fn finfo(core: &lb::Core, args: FinfoArgs) -> Result<(), CliError> {
    let my_name = core.get_account()?.username;
    match file_from_path_or_id(core, &args.target) {
        Ok(f) => print_file(&f, args.raw, &my_name),
        Err(err) => {
            let ids = core
                .list_metadatas()
                .map_err(|err| CliError(format!("listing metadatas: {:?}", err)))?
                .into_iter()
                .filter_map(|f| {
                    if f.id.to_string().starts_with(&args.target) {
                        Some(f.id)
                    } else {
                        None
                    }
                })
                .collect::<Vec<lb::Uuid>>();
            match ids.len() {
                0 => return Err(err),
                1 => {
                    let id = ids[0];
                    let f = core.get_file_by_id(id).map_err(|err| (err, id))?;
                    print_file(&f, args.raw, &my_name);
                }
                _ => {
                    println!("id prefix matched the following files:\n");
                    for id in ids {
                        let msg = match core.get_path_by_id(id) {
                            Ok(p) => p,
                            Err(err) => format!("error getting path: {:?}", err),
                        };
                        println!(" {}  {}", id, msg);
                    }
                }
            }
        }
    }
    Ok(())
}

fn print_file(f: &lb::File, raw: bool, my_name: &str) {
    if raw {
        println!("{:#?}", f);
        return;
    }
    // Build the text that will contain share info.
    let mut shares = String::new();
    for sh in &f.shares {
        let shared_by = if sh.shared_by == my_name {
            String::from("me")
        } else {
            format!("@{}", sh.shared_by)
        };
        let shared_with = if sh.shared_with == my_name {
            String::from("me")
        } else {
            format!("@{}", sh.shared_with)
        };
        shares += &format!("\n    {} -> {} ({:?})", shared_by, shared_with, sh.mode);
    }
    let data = &[
        ("name", f.name.to_string()),
        ("id", f.id.to_string()),
        ("parent", f.parent.to_string()),
        ("type", f.file_type.to_string().to_lowercase()),
        ("lastmod", f.last_modified.to_string()),
        ("lastmod_by", f.last_modified_by.to_string()),
        (&format!("shares ({})", f.shares.len()), shares),
    ];
    // Determine widest key name.
    let mut w_name = 0;
    for (k, _) in data {
        let n = k.len();
        if n > w_name {
            w_name = n;
        }
    }
    for (k, v) in data {
        println!("  {:<w_name$} : {}", k, v, w_name = w_name);
    }
}

fn validate(core: &lb::Core) -> Result<(), CliError> {
    let warnings = core
        .validate()
        .map_err(|err| CliError(format!("validating: {:?}", err)))?;
    if warnings.is_empty() {
        return Ok(());
    }
    for w in &warnings {
        eprintln!("{:#?}", w);
    }
    Err(CliError(format!("{} warnings found", warnings.len())))
}

use std::cell::Cell;
use std::env;
use std::fs;
use std::io::Write;
use std::path::PathBuf;

use structopt::StructOpt;

use lb::ImportStatus;

use crate::util::file_from_path_or_id;
use crate::util::id_from_path_or_id;
use crate::CliError;

#[derive(StructOpt)]
pub struct ExportArgs {
    /// At least one filesystem location
    target: String,
    /// The path or id of a lockbook folder
    dest: Option<PathBuf>,
}

#[derive(StructOpt)]
pub struct ImportArgs {
    /// At least one filesystem location
    disk_paths: Vec<PathBuf>,
    /// Path or ID of a lockbook folder
    dest: String,
}

#[derive(StructOpt)]
pub struct DrawingArgs {
    /// Path or ID of a lockbook drawing
    target: String,
    /// Exported image format (png, jpeg, bmp, tga, pnm, farbfeld) (default is png)
    format: String,
}

pub fn export(core: &lb::Core, args: ExportArgs) -> Result<(), CliError> {
    let target_file = file_from_path_or_id(core, &args.target)?;

    let dest = if let Some(path) = args.dest {
        path
    } else {
        // If no destination path is provided, it'll be a file with the target name in the current
        // directory. If it's root, it'll be the account's username.
        let name = if target_file.id == target_file.parent {
            core.get_account()?.username
        } else {
            target_file.name.clone()
        };
        let mut dir = env::current_dir()?;
        dir.push(name);
        dir
    };

    println!("exporting '{}'...", args.target);
    fs::create_dir(&dest)?;

    core.export_file(target_file.id, dest.clone(), false, None)
        .map_err(|err| (err, dest).into())
}

pub fn import(core: &lb::Core, flags: ImportArgs) -> Result<(), CliError> {
    let dest_id = id_from_path_or_id(core, &flags.dest)?;

    let total = Cell::new(0);
    let nth_file = Cell::new(0);
    let update_status = move |status: ImportStatus| match status {
        ImportStatus::CalculatedTotal(n_files) => total.set(n_files),
        ImportStatus::Error(disk_path, err) => match err {
            lb::CoreError::DiskPathInvalid => {
                eprintln!("invalid disk path '{}'", disk_path.display())
            }
            _ => eprintln!("unexpected error: {:#?}", err),
        },
        ImportStatus::StartingItem(disk_path) => {
            nth_file.set(nth_file.get() + 1);
            print!(
                "({}/{}) importing: {}... ",
                nth_file.get(),
                total.get(),
                disk_path
            );
            std::io::stdout().flush().unwrap();
        }
        ImportStatus::FinishedItem(_meta) => println!("done."),
    };

    core.import_files(&flags.disk_paths, dest_id, &update_status)
        .map_err(|err| (err, dest_id).into())
}

pub fn drawing(core: &lb::Core, args: DrawingArgs) -> Result<(), CliError> {
    let id = id_from_path_or_id(core, &args.target)?;
    let img_fmt = args.format.parse().unwrap_or_else(|_| {
        if !args.format.is_empty() {
            eprintln!("'{}' is an unsupported format, but feel free to make a github issue! falling back to png for now.", args.format);
        }
        lb::SupportedImageFormats::Png
    });
    let drawing_bytes = core
        .export_drawing(id, img_fmt, None)
        .map_err(|err| (err, id))?;

    std::io::stdout().write_all(drawing_bytes.as_slice())?;
    Ok(())
}

use std::fmt;
use std::io::{self, Write};
use std::str::FromStr;

use crate::CliError;

pub fn input<T>(prompt: impl fmt::Display) -> Result<T, CliError>
where
    T: FromStr,
    <T as FromStr>::Err: fmt::Debug,
{
    print!("{}", prompt);
    io::stdout().flush()?;

    let mut answer = String::new();
    io::stdin()
        .read_line(&mut answer)
        .expect("failed to read from stdin");
    answer.retain(|c| c != '\n' && c != '\r');

    Ok(answer.parse::<T>().unwrap())
}

pub fn file_from_path_or_id(core: &lb::Core, v: &str) -> Result<lb::File, CliError> {
    Ok(match lb::Uuid::parse_str(v) {
        Ok(id) => core.get_file_by_id(id).map_err(|err| (err, id))?,
        Err(_) => core.get_by_path(v).map_err(|err| (err, v))?,
    })
}

pub fn id_from_something(core: &lb::Core, v: &str) -> Result<lb::Uuid, CliError> {
    if let Ok(id) = lb::Uuid::parse_str(v) {
        return Ok(id);
    }
    if let Some(f) = maybe_get_by_path(core, v)? {
        return Ok(f.id);
    }
    let ids: Vec<lb::Uuid> = core
        .list_metadatas()
        .map_err(|err| CliError(format!("list metadatas: {:?}", err)))?
        .into_iter()
        .filter_map(|f| {
            if f.id.to_string().starts_with(v) {
                Some(f.id)
            } else {
                None
            }
        })
        .collect();
    match ids.len() {
        0 => Err(CliError(format!("cannot find any file via '{}'", v))),
        1 => Ok(ids[0]),
        n => {
            let mut err_msg = format!(
                "value '{}' was not a path and matched the following {} file IDs:\n",
                v, n
            );
            for id in ids {
                let msg = core
                    .get_path_by_id(id)
                    .unwrap_or_else(|err| format!("error getting path: {:?}", err));
                err_msg += &format!("\n {}  {}", id, msg);
            }
            Err(CliError(err_msg))
        }
    }
}

pub fn id_from_path_or_id(core: &lb::Core, v: &str) -> Result<lb::Uuid, CliError> {
    match lb::Uuid::parse_str(v) {
        Ok(id) => Ok(id),
        Err(_) => Ok(core.get_by_path(v).map_err(|err| (err, v))?.id),
    }
}

pub fn maybe_get_by_path(core: &lb::Core, p: &str) -> Result<Option<lb::File>, CliError> {
    match core.get_by_path(p) {
        Ok(f) => Ok(Some(f)),
        Err(lb::Error::UiError(lb::GetFileByPathError::NoFileAtThatPath)) => Ok(None),
        Err(err) => Err((err, p).into()),
    }
}

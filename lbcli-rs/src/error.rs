use std::fmt;
use std::io;
use std::path::PathBuf;

use lb::Error as LbError;
use lb::Uuid;

pub struct CliError(pub String);

impl CliError {
    pub fn new(msg: impl ToString) -> Self {
        Self(msg.to_string())
    }
}

impl fmt::Display for CliError {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        write!(f, "error: {}", self.0)
    }
}

impl From<lb::UnexpectedError> for CliError {
    fn from(err: lb::UnexpectedError) -> Self {
        Self(format!("unexpected: {:?}", err))
    }
}

impl From<LbError<lb::AccountExportError>> for CliError {
    fn from(err: LbError<lb::AccountExportError>) -> Self {
        Self(format!("exporting account: {:?}", err))
    }
}

impl From<LbError<lb::CalculateWorkError>> for CliError {
    fn from(err: LbError<lb::CalculateWorkError>) -> Self {
        Self(format!("calculating work: {:?}", err))
    }
}

impl From<LbError<lb::CancelSubscriptionError>> for CliError {
    fn from(err: LbError<lb::CancelSubscriptionError>) -> Self {
        Self(format!("canceling subscription: {:?}", err))
    }
}

impl From<LbError<lb::CreateAccountError>> for CliError {
    fn from(err: LbError<lb::CreateAccountError>) -> Self {
        Self(format!("creating account: {:?}", err))
    }
}

impl From<LbError<lb::CreateFileAtPathError>> for CliError {
    fn from(err: LbError<lb::CreateFileAtPathError>) -> Self {
        Self(format!("creating file at path: {:?}", err))
    }
}

impl From<(LbError<lb::ExportDrawingError>, Uuid)> for CliError {
    fn from(v: (LbError<lb::ExportDrawingError>, Uuid)) -> Self {
        let (err, id) = v;
        Self(format!("exporting drawing with id '{}': {:?}", id, err))
    }
}

impl From<(LbError<lb::ExportFileError>, PathBuf)> for CliError {
    fn from(v: (LbError<lb::ExportFileError>, PathBuf)) -> Self {
        let (err, disk_dir) = v;
        Self(format!("exporting file to {:?}: {:?}", disk_dir, err))
    }
}

impl From<LbError<lb::FeatureFlagError>> for CliError {
    fn from(err: LbError<lb::FeatureFlagError>) -> Self {
        Self(format!("feature flag err: {:?}", err))
    }
}

impl From<(LbError<lb::FileDeleteError>, Uuid)> for CliError {
    fn from(v: (LbError<lb::FileDeleteError>, Uuid)) -> Self {
        let (err, id) = v;
        Self(format!("deleting file with id '{}': {:?}", id, err))
    }
}

impl From<LbError<lb::GetAccountError>> for CliError {
    fn from(err: LbError<lb::GetAccountError>) -> Self {
        Self(format!("getting account: {:?}", err))
    }
}

impl From<(LbError<lb::GetAndGetChildrenError>, Uuid)> for CliError {
    fn from(v: (LbError<lb::GetAndGetChildrenError>, Uuid)) -> Self {
        let (err, id) = v;
        Self(format!("get and get children of '{}': {:?}", id, err))
    }
}

impl From<(LbError<lb::GetFileByIdError>, Uuid)> for CliError {
    fn from(v: (LbError<lb::GetFileByIdError>, Uuid)) -> Self {
        let (err, id) = v;
        Self(format!("get file by id '{}': {:?}", id, err))
    }
}

impl From<(LbError<lb::GetFileByPathError>, &str)> for CliError {
    fn from(v: (LbError<lb::GetFileByPathError>, &str)) -> Self {
        let (err, path) = v;
        Self(format!("get file by path '{}': {:?}", path, err))
    }
}

impl From<LbError<lb::GetRootError>> for CliError {
    fn from(err: LbError<lb::GetRootError>) -> Self {
        Self(format!("getting root: {:?}", err))
    }
}

impl From<LbError<lb::GetSubscriptionInfoError>> for CliError {
    fn from(err: LbError<lb::GetSubscriptionInfoError>) -> Self {
        Self(format!("getting subscription info: {:?}", err))
    }
}

impl From<LbError<lb::GetUsageError>> for CliError {
    fn from(err: LbError<lb::GetUsageError>) -> Self {
        Self(format!("getting usage: {:?}", err))
    }
}

impl From<LbError<lb::ImportError>> for CliError {
    fn from(err: LbError<lb::ImportError>) -> Self {
        Self(format!("importing account: {:?}", err))
    }
}

impl From<(LbError<lb::ImportFileError>, Uuid)> for CliError {
    fn from(v: (LbError<lb::ImportFileError>, Uuid)) -> Self {
        let (err, id) = v;
        Self(format!("importing file to '{}': {:?}", id, err))
    }
}

impl From<(LbError<lb::MoveFileError>, Uuid, Uuid)> for CliError {
    fn from(v: (LbError<lb::MoveFileError>, Uuid, Uuid)) -> Self {
        let (err, src_id, dest_id) = v;
        Self(format!("moving '{}' -> '{}': {:?}", src_id, dest_id, err))
    }
}

impl From<(LbError<lb::ReadDocumentError>, Uuid)> for CliError {
    fn from(v: (LbError<lb::ReadDocumentError>, Uuid)) -> Self {
        let (err, id) = v;
        Self(format!("reading doc '{}': {:?}", id, err))
    }
}

impl From<(LbError<lb::RenameFileError>, Uuid)> for CliError {
    fn from(v: (LbError<lb::RenameFileError>, Uuid)) -> Self {
        let (err, id) = v;
        Self(format!("renaming file '{}': {:?}", id, err))
    }
}

impl From<LbError<lb::ShareFileError>> for CliError {
    fn from(err: LbError<lb::ShareFileError>) -> Self {
        Self(format!("sharing file: {:?}", err))
    }
}

impl From<LbError<lb::SyncAllError>> for CliError {
    fn from(err: LbError<lb::SyncAllError>) -> Self {
        Self(format!("syncing: {:?}", err))
    }
}

impl From<LbError<lb::UpgradeAccountStripeError>> for CliError {
    fn from(err: LbError<lb::UpgradeAccountStripeError>) -> Self {
        Self(format!("upgrading account via stripe: {:?}", err))
    }
}

impl From<(LbError<lb::WriteToDocumentError>, Uuid)> for CliError {
    fn from(v: (LbError<lb::WriteToDocumentError>, Uuid)) -> Self {
        let (err, id) = v;
        Self(format!("writing doc '{}': {:?}", id, err))
    }
}

impl From<io::Error> for CliError {
    fn from(err: io::Error) -> Self {
        Self(format!("{:?}", err))
    }
}

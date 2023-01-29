use lockbook_core::{AccountExportError, CreateAccountError, GetAccountError, ImportError};

use crate::*;

#[repr(C)]
pub struct LbAccount {
    username: *mut c_char,
    api_url: *mut c_char,
}

fn lb_account_new() -> LbAccount {
    LbAccount {
        username: null_mut(),
        api_url: null_mut(),
    }
}

/// # Safety
#[no_mangle]
pub unsafe extern "C" fn lb_account_free(a: LbAccount) {
    libc::free(a.username as *mut c_void);
    libc::free(a.api_url as *mut c_void);
}

#[repr(C)]
pub struct LbAccountResult {
    ok: LbAccount,
    err: LbError,
}

fn lb_account_result_new() -> LbAccountResult {
    LbAccountResult {
        ok: lb_account_new(),
        err: lb_error_none(),
    }
}

/// # Safety
#[no_mangle]
pub unsafe extern "C" fn lb_account_result_free(r: LbAccountResult) {
    if r.err.code == LbErrorCode::Success {
        lb_account_free(r.ok);
    } else {
        lb_error_free(r.err);
    }
}

/// # Safety
///
/// The returned value must be passed to `lb_account_result_free` to avoid a memory leak.
/// Alternatively, the `ok` value or `err` value can be passed to `lb_account_free` or
/// `lb_error_free` respectively depending on whether there's an error or not.
#[no_mangle]
pub unsafe extern "C" fn lb_create_account(
    core: *mut c_void,
    username: *const c_char,
    api_url: *const c_char,
    welcome_doc: bool,
) -> LbAccountResult {
    let mut r = lb_account_result_new();
    match core!(core).create_account(rstr(username), rstr(api_url), welcome_doc) {
        Ok(acct) => {
            r.ok.username = cstr(acct.username);
            r.ok.api_url = cstr(acct.api_url);
        }
        Err(err) => {
            use CreateAccountError::*;
            r.err.msg = cstr(format!("{:?}", err));
            r.err.code = match err {
                Error::UiError(err) => match err {
                    UsernameTaken => LbErrorCode::UsernameTaken,
                    InvalidUsername => LbErrorCode::UsernameInvalid,
                    AccountExistsAlready => LbErrorCode::AccountExistsAlready,
                    CouldNotReachServer => LbErrorCode::CouldNotReachServer,
                    ClientUpdateRequired => LbErrorCode::ClientUpdateRequired,
                    ServerDisabled => LbErrorCode::ServerDisabled,
                },
                Error::Unexpected(_) => LbErrorCode::Unexpected,
            };
        }
    }
    r
}

/// # Safety
///
/// The returned value must be passed to `lb_account_result_free` to avoid a memory leak.
/// Alternatively, the `ok` value or `err` value can be passed to `lb_account_free` or
/// `lb_error_free` respectively depending on whether there's an error or not.
#[no_mangle]
pub unsafe extern "C" fn lb_import_account(
    core: *mut c_void,
    account_string: *const c_char,
) -> LbAccountResult {
    let mut r = lb_account_result_new();
    match core!(core).import_account(rstr(account_string)) {
        Ok(acct) => {
            r.ok.username = cstr(acct.username);
            r.ok.api_url = cstr(acct.api_url);
        }
        Err(err) => {
            use ImportError::*;
            r.err.msg = cstr(format!("{:?}", err));
            r.err.code = match err {
                Error::UiError(err) => match err {
                    AccountStringCorrupted => LbErrorCode::AccountStringCorrupted,
                    AccountExistsAlready => LbErrorCode::AccountExistsAlready,
                    AccountDoesNotExist => LbErrorCode::AccountDoesNotExist,
                    UsernamePKMismatch => LbErrorCode::UsernamePubKeyMismatch,
                    CouldNotReachServer => LbErrorCode::CouldNotReachServer,
                    ClientUpdateRequired => LbErrorCode::ClientUpdateRequired,
                },
                Error::Unexpected(_) => LbErrorCode::Unexpected,
            };
        }
    }
    r
}

/// # Safety
///
/// The returned value must be passed to `lb_string_result_free` to avoid a memory leak.
/// Alternatively, the `ok` value or `err` value can be passed to `free` or `lb_error_free`
/// respectively depending on whether there's an error or not.
#[no_mangle]
pub unsafe extern "C" fn lb_export_account(core: *mut c_void) -> LbStringResult {
    let mut r = lb_string_result_new();
    match core!(core).export_account() {
        Ok(acct_str) => r.ok = cstr(acct_str),
        Err(err) => {
            use AccountExportError::*;
            r.err.msg = cstr(format!("{:?}", err));
            r.err.code = match err {
                Error::UiError(NoAccount) => LbErrorCode::NoAccount,
                Error::Unexpected(_) => LbErrorCode::Unexpected,
            };
        }
    }
    r
}

/// # Safety
///
/// The returned value must be passed to `lb_account_result_free` to avoid a memory leak.
/// Alternatively, the `ok` value or `err` value can be passed to `lb_account_free` or
/// `lb_error_free` respectively depending on whether there's an error or not.
#[no_mangle]
pub unsafe extern "C" fn lb_get_account(core: *mut c_void) -> LbAccountResult {
    let mut r = lb_account_result_new();
    match core!(core).get_account() {
        Ok(acct) => {
            r.ok.username = cstr(acct.username);
            r.ok.api_url = cstr(acct.api_url);
        }
        Err(err) => {
            r.err.msg = cstr(format!("{:?}", err));
            r.err.code = match err {
                Error::UiError(err) => match err {
                    GetAccountError::NoAccount => LbErrorCode::NoAccount,
                },
                Error::Unexpected(_) => LbErrorCode::Unexpected,
            };
        }
    }
    r
}

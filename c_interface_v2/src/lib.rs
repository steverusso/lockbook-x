mod account;
mod files;
mod subscription;
mod sync_and_usage;

use std::ffi::{c_char, c_void};
use std::ptr::null_mut;

use lockbook_core::{Config, Core, Error};

use crate::files::*;

unsafe fn cstr(value: String) -> *mut c_char {
    let len = value.len();
    let s = libc::malloc(len + 1) as *mut c_char;
    libc::memcpy(
        s as *mut c_void,
        value.as_bytes().as_ptr() as *mut c_void,
        len,
    );
    std::ptr::write(s.add(len) as *mut u8, 0);
    s
}

unsafe fn rstr<'a>(s: *const c_char) -> &'a str {
    std::ffi::CStr::from_ptr(s)
        .to_str()
        .expect("*const char -> &str")
}

macro_rules! core {
    ($ptr:ident) => {
        &*($ptr as *mut Core)
    };
}

pub(crate) use core;

#[no_mangle]
pub static C_DEFAULT_API_LOCATION: &[u8; 30] = b"https://api.prod.lockbook.net\0";

#[repr(C)]
pub struct LbError {
    code: LbErrorCode,
    msg: *mut c_char,
}

fn lb_error_none() -> LbError {
    LbError {
        code: LbErrorCode::Zero_,
        msg: null_mut(),
    }
}

/// # Safety
#[no_mangle]
pub unsafe extern "C" fn lb_error_free(err: LbError) {
    libc::free(err.msg as *mut c_void);
}

#[repr(C)]
pub enum LbErrorCode {
    Zero_ = 0,
    Unexpected,
    AccountExistsAlready,
    AccountDoesNotExist,
    AccountStringCorrupted,
    ClientUpdateRequired,
    CouldNotReachServer,
    FileExists,
    FileIsNotDocument,
    FileIsNotFolder,
    FileNameContainsSlash,
    FileNameEmpty,
    FileNameUnavailable,
    FileNotFound,
    FolderMovedIntoItself,
    InsufficientPermission,
    InvalidDrawing,
    LinkInSharedFolder,
    NoAccount,
    NoRoot,
    NoRootOps,
    PathContainsEmptyFile,
    TargetParentNotFound,
    UsernameInvalid,
    UsernamePubKeyMismatch,
    UsernameTaken,
    ServerDisabled,

    NotPremium,
    SubscriptionAlreadyCanceled,
    UsageIsOverFreeTierDataCap,
    ExistingRequestPending,
    CannotCancelForAppStore,
}

#[repr(C)]
pub struct LbStringResult {
    ok: *mut c_char,
    err: LbError,
}

fn lb_string_result_new() -> LbStringResult {
    LbStringResult {
        ok: null_mut(),
        err: lb_error_none(),
    }
}

/// # Safety
#[no_mangle]
pub unsafe extern "C" fn lb_string_result_free(r: LbStringResult) {
    libc::free(r.ok as *mut c_void);
    lb_error_free(r.err);
}

#[repr(C)]
pub struct LbBytesResult {
    bytes: *mut u8,
    count: usize,
    err: LbError,
}

fn lb_bytes_result_new() -> LbBytesResult {
    LbBytesResult {
        bytes: null_mut(),
        count: 0,
        err: lb_error_none(),
    }
}

/// # Safety
#[no_mangle]
pub unsafe extern "C" fn lb_bytes_result_free(r: LbBytesResult) {
    let _ = Vec::from_raw_parts(r.bytes, r.count, r.count);
    lb_error_free(r.err);
}

#[repr(C)]
pub struct LbInitResult {
    core: *mut c_void,
    err: LbError,
}

/// # Safety
///
/// The returned value must be passed to `lb_error_free` to avoid a memory leak.
#[no_mangle]
pub unsafe extern "C" fn lb_init(writeable_path: *const c_char, logs: bool) -> LbInitResult {
    let mut r = LbInitResult {
        core: null_mut(),
        err: lb_error_none(),
    };
    match Core::init(&Config {
        writeable_path: rstr(writeable_path).to_string(),
        logs,
        colored_logs: true,
    }) {
        Ok(core) => r.core = Box::into_raw(Box::new(core)) as *mut c_void,
        Err(err) => {
            r.err.code = LbErrorCode::Unexpected;
            r.err.msg = cstr(format!("{:?}", err));
        }
    }
    r
}

/// # Safety
///
/// The returned value must be passed to `lb_string_free` to avoid a memory leak.
#[no_mangle]
pub unsafe extern "C" fn lb_writeable_path(core: *mut c_void) -> *mut c_char {
    cstr(core!(core).config.writeable_path.clone())
}

#[repr(C)]
pub struct LbValidateResult {
    warnings: *mut *mut c_char,
    n_warnings: usize,
    err: LbError,
}

/// # Safety
#[no_mangle]
pub unsafe extern "C" fn lb_validate_result_index(r: LbValidateResult, i: usize) -> *mut c_char {
    *r.warnings.add(i)
}

/// # Safety
#[no_mangle]
pub unsafe extern "C" fn lb_validate_result_free(r: LbValidateResult) {
    let warnings = Vec::from_raw_parts(r.warnings, r.n_warnings, r.n_warnings);
    for w in warnings {
        libc::free(w as *mut c_void);
    }
    lb_error_free(r.err);
}

/// # Safety
///
/// The returned value must be passed to `lb_validate_result_free` to avoid a memory leak.
#[no_mangle]
pub unsafe extern "C" fn lb_validate(core: *mut c_void) -> LbValidateResult {
    let mut r = LbValidateResult {
        warnings: null_mut(),
        n_warnings: 0,
        err: lb_error_none(),
    };
    match core!(core).validate() {
        Ok(warnings) => {
            let mut c_warnings = Vec::with_capacity(warnings.len());
            for w in warnings {
                c_warnings.push(cstr(w.to_string()));
            }
            let mut c_warnings = std::mem::ManuallyDrop::new(c_warnings);
            r.warnings = c_warnings.as_mut_ptr();
            r.n_warnings = c_warnings.len();
        }
        Err(err) => {
            r.err.msg = cstr(err.to_string());
            r.err.code = LbErrorCode::Unexpected;
        }
    }
    r
}

#[cfg(test)]
mod tests;

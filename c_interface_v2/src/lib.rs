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
    let s = libc::malloc(len + 1);
    libc::memcpy(s, value.as_bytes().as_ptr() as *mut c_void, len);
    std::ptr::write(s.add(len) as *mut c_char, 0);
    s as *mut c_char
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
pub static LB_DEFAULT_API_LOCATION: &[u8; 30] = b"https://api.prod.lockbook.net\0";

#[repr(C)]
pub struct LbError {
    code: LbErrorCode,
    msg: *mut c_char,
}

/// # Safety
#[no_mangle]
pub extern "C" fn lb_error_none() -> LbError {
    LbError {
        code: LbErrorCode::Success,
        msg: null_mut(),
    }
}

/// # Safety
#[no_mangle]
pub unsafe extern "C" fn lb_error_copy(err: LbError) -> LbError {
    LbError {
        code: err.code,
        msg: libc::strdup(err.msg),
    }
}

/// # Safety
#[no_mangle]
pub unsafe extern "C" fn lb_error_free(err: LbError) {
    libc::free(err.msg as *mut c_void);
}

#[derive(PartialEq)]
#[repr(C)]
pub enum LbErrorCode {
    Success = 0,
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
    if r.err.code == LbErrorCode::Success {
        libc::free(r.ok as *mut c_void);
    } else {
        lb_error_free(r.err);
    }
}

#[repr(C)]
pub struct LbBytes {
    data: *mut u8,
    size: usize,
}

/// # Safety
#[no_mangle]
pub unsafe extern "C" fn lb_bytes_free(b: LbBytes) {
    let _ = Vec::from_raw_parts(b.data, b.size, b.size);
}

#[repr(C)]
pub struct LbBytesResult {
    ok: LbBytes,
    err: LbError,
}

fn lb_bytes_result_new() -> LbBytesResult {
    LbBytesResult {
        ok: LbBytes {
            data: null_mut(),
            size: 0,
        },
        err: lb_error_none(),
    }
}

/// # Safety
#[no_mangle]
pub unsafe extern "C" fn lb_bytes_result_free(r: LbBytesResult) {
    if r.err.code == LbErrorCode::Success {
        lb_bytes_free(r.ok);
    } else {
        lb_error_free(r.err);
    }
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
/// The returned value must be passed to `free` to avoid a memory leak.
#[no_mangle]
pub unsafe extern "C" fn lb_writeable_path(core: *mut c_void) -> *mut c_char {
    cstr(core!(core).get_config().unwrap().writeable_path.clone()) // todo(steve)
}

#[repr(C)]
pub struct LbStringList {
    data: *mut *mut c_char,
    size: usize,
}

fn lb_string_list_new() -> LbStringList {
    LbStringList {
        data: null_mut(),
        size: 0,
    }
}

/// # Safety
#[no_mangle]
pub unsafe extern "C" fn lb_string_list_index(sl: LbStringList, i: usize) -> *mut c_char {
    *sl.data.add(i)
}

/// # Safety
#[no_mangle]
pub unsafe extern "C" fn lb_string_list_free(sl: LbStringList) {
    let data = Vec::from_raw_parts(sl.data, sl.size, sl.size);
    for s in data {
        libc::free(s as *mut c_void);
    }
}

#[repr(C)]
pub struct LbValidateResult {
    ok: LbStringList,
    err: LbError,
}

/// # Safety
#[no_mangle]
pub unsafe extern "C" fn lb_validate_result_free(r: LbValidateResult) {
    if r.err.code == LbErrorCode::Success {
        lb_string_list_free(r.ok);
    } else {
        lb_error_free(r.err);
    }
}

/// # Safety
///
/// The returned value must be passed to `lb_validate_result_free` to avoid a memory leak.
/// Alternatively, the `ok` value or `err` value can be passed to `lb_string_list_free` or
/// `lb_error_free` respectively depending on whether there's an error or not.
#[no_mangle]
pub unsafe extern "C" fn lb_validate(core: *mut c_void) -> LbValidateResult {
    let mut r = LbValidateResult {
        ok: lb_string_list_new(),
        err: lb_error_none(),
    };
    match core!(core).validate() {
        Ok(warnings) => {
            let mut c_warnings = Vec::with_capacity(warnings.len());
            for w in warnings {
                c_warnings.push(cstr(w.to_string()));
            }
            let mut c_warnings = std::mem::ManuallyDrop::new(c_warnings);
            r.ok.data = c_warnings.as_mut_ptr();
            r.ok.size = c_warnings.len();
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

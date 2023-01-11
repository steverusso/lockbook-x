use std::ffi::{c_char, c_void, CStr, CString};
use std::ptr::null_mut;

use lockbook_core::{
    AccountExportError, CalculateWorkError, CancelSubscriptionError, CreateAccountError,
    CreateFileAtPathError, Error, ExportDrawingError, ExportFileError, FileDeleteError,
    GetAccountError, GetAndGetChildrenError, GetFileByIdError, GetFileByPathError, GetUsageError,
    ImportError, MoveFileError, ReadDocumentError, RenameFileError, SyncAllError,
    WriteToDocumentError,
};
use lockbook_core::{
    ClientWorkUnit, Config, Core, File, FileType, PaymentMethod, PaymentPlatform, Share, ShareMode,
    StripeAccountTier, SupportedImageFormats, Uuid, WorkUnit,
};

fn cstr(value: String) -> *mut c_char {
    CString::new(value).expect("rust -> c string").into_raw()
}

unsafe fn rstr<'a>(s: *const c_char) -> &'a str {
    CStr::from_ptr(s).to_str().expect("*const char -> &str")
}

unsafe fn parse_c_uuid(c_id: *const c_char) -> Result<Uuid, LbError> {
    let s = rstr(c_id);
    s.parse().map_err(|_| LbError {
        code: LbErrorCode::Unexpected,
        msg: cstr(format!("unable to parse uuid '{}'", s)),
    })
}

macro_rules! core {
    ($ptr:ident) => {
        &*($ptr as *mut Core)
    };
}

macro_rules! uuid_or_return {
    ($id:expr) => {
        match parse_c_uuid($id) {
            Ok(id) => id,
            Err(err) => return err,
        }
    };
    ($id:expr, $result:ident) => {
        match parse_c_uuid($id) {
            Ok(id) => id,
            Err(err) => {
                $result.err = err;
                return $result;
            }
        }
    };
}

#[no_mangle]
pub unsafe extern "C" fn is_uuid(s: *const c_char) -> bool {
    Uuid::parse_str(rstr(s)).is_ok()
}

#[no_mangle]
pub extern "C" fn lb_default_api_location() -> *const c_char {
    static C_DEFAULT_API_LOCATION: &str = "https://api.prod.lockbook.net\0";

    C_DEFAULT_API_LOCATION.as_ptr() as *const c_char
}

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
    if !err.msg.is_null() {
        let _ = CString::from_raw(err.msg);
    }
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

#[no_mangle]
pub unsafe extern "C" fn lb_string_result_free(r: LbStringResult) {
    if !r.ok.is_null() {
        let _ = CString::from_raw(r.ok);
    }
    lb_error_free(r.err);
}

#[repr(C)]
pub struct LbFile {
    id: *mut c_char,
    parent: *mut c_char,
    name: *mut c_char,
    typ: LbFileType,
    lastmod_by: *mut c_char,
    lastmod: u64,
    shares: LbShareList,
}

fn lb_file_new(f: File) -> LbFile {
    let mut typ = lb_file_type_doc();
    if let FileType::Folder = f.file_type {
        typ.tag = LbFileTypeTag::LbFileTypeTagFolder;
    }
    if let FileType::Link { target } = f.file_type {
        typ.tag = LbFileTypeTag::LbFileTypeTagLink;
        typ.link_target = cstr(target.to_string());
    }
    LbFile {
        id: cstr(f.id.to_string()),
        parent: cstr(f.parent.to_string()),
        name: cstr(f.name),
        typ,
        lastmod_by: cstr(f.last_modified_by),
        lastmod: f.last_modified,
        shares: lb_share_list_new(f.shares),
    }
}

unsafe fn lb_file_free(f: LbFile) {
    if !f.id.is_null() {
        let _ = CString::from_raw(f.id);
    }
    if !f.parent.is_null() {
        let _ = CString::from_raw(f.parent);
    }
    if !f.name.is_null() {
        let _ = CString::from_raw(f.name);
    }
    if !f.lastmod_by.is_null() {
        let _ = CString::from_raw(f.lastmod_by);
    }
    lb_file_type_free(f.typ);
    lb_share_list_free(f.shares);
}

/// The zero value represents a document.
#[repr(C)]
pub struct LbFileType {
    tag: LbFileTypeTag,
    link_target: *mut c_char,
}

#[repr(C)]
pub enum LbFileTypeTag {
    LbFileTypeTagDocument,
    LbFileTypeTagFolder,
    LbFileTypeTagLink,
}

#[no_mangle]
pub extern "C" fn lb_file_type_doc() -> LbFileType {
    LbFileType {
        tag: LbFileTypeTag::LbFileTypeTagDocument,
        link_target: null_mut(),
    }
}

#[no_mangle]
pub unsafe extern "C" fn lb_file_type_free(t: LbFileType) {
    if !t.link_target.is_null() {
        let _ = CString::from_raw(t.link_target);
    }
}

#[repr(C)]
pub struct LbShareList {
    list: *mut LbShare,
    count: usize,
}

fn lb_share_list_new(shares: Vec<Share>) -> LbShareList {
    let mut list = Vec::with_capacity(shares.len());
    for sh in shares {
        list.push(LbShare {
            by: cstr(sh.shared_by),
            with: cstr(sh.shared_with),
            mode: match sh.mode {
                ShareMode::Read => LbShareMode::LbShareModeRead,
                ShareMode::Write => LbShareMode::LbShareModeWrite,
            },
        });
    }
    let mut list = std::mem::ManuallyDrop::new(list);
    LbShareList {
        list: list.as_mut_ptr(),
        count: list.len(),
    }
}

#[no_mangle]
pub unsafe extern "C" fn lb_share_list_index(sl: LbShareList, i: usize) -> *mut LbShare {
    sl.list.add(i)
}

unsafe fn lb_share_list_free(sl: LbShareList) {
    let list = Vec::from_raw_parts(sl.list, sl.count, sl.count);
    for sh in list {
        if !sh.by.is_null() {
            let _ = CString::from_raw(sh.by);
        }
        if !sh.with.is_null() {
            let _ = CString::from_raw(sh.with);
        }
    }
}

#[repr(C)]
pub struct LbShare {
    by: *mut c_char,
    with: *mut c_char,
    mode: LbShareMode,
}

#[repr(C)]
pub enum LbShareMode {
    LbShareModeRead,
    LbShareModeWrite,
}

#[repr(C)]
pub struct LbFileResult {
    ok: LbFile,
    err: LbError,
}

fn lb_file_result_new() -> LbFileResult {
    LbFileResult {
        ok: LbFile {
            id: null_mut(),
            parent: null_mut(),
            name: null_mut(),
            typ: lb_file_type_doc(),
            lastmod_by: null_mut(),
            lastmod: 0,
            shares: LbShareList {
                list: std::ptr::null_mut(),
                count: 0,
            },
        },
        err: lb_error_none(),
    }
}

#[no_mangle]
pub unsafe extern "C" fn lb_file_result_free(r: LbFileResult) {
    lb_file_free(r.ok);
    lb_error_free(r.err);
}

#[repr(C)]
pub struct LbFileListResult {
    ok: LbFileList,
    err: LbError,
}

fn lb_file_list_result_new() -> LbFileListResult {
    LbFileListResult {
        ok: LbFileList {
            list: null_mut(),
            count: 0,
        },
        err: lb_error_none(),
    }
}

#[no_mangle]
pub unsafe extern "C" fn lb_file_list_result_free(r: LbFileListResult) {
    let list = Vec::from_raw_parts(r.ok.list, r.ok.count, r.ok.count);
    for f in list {
        lb_file_free(f);
    }
    lb_error_free(r.err);
}

#[repr(C)]
pub struct LbFileList {
    list: *mut LbFile,
    count: usize,
}

#[no_mangle]
pub unsafe extern "C" fn lb_file_list_index(fl: LbFileList, i: usize) -> *mut LbFile {
    fl.list.add(i)
}

unsafe fn lb_file_list_init(fl: &mut LbFileList, from: Vec<File>) {
    let mut into = Vec::with_capacity(from.len());
    for f in from {
        into.push(lb_file_new(f));
    }
    let mut into = std::mem::ManuallyDrop::new(into);
    fl.list = into.as_mut_ptr();
    fl.count = into.len();
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

#[no_mangle]
pub unsafe extern "C" fn lb_bytes_result_free(r: LbBytesResult) {
    if !r.bytes.is_null() {
        let _ = Vec::from_raw_parts(r.bytes, r.count, r.count);
    }
    lb_error_free(r.err);
}

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

#[no_mangle]
pub unsafe extern "C" fn lb_account_free(a: LbAccount) {
    if !a.username.is_null() {
        let _ = CString::from_raw(a.username);
    }
    if !a.api_url.is_null() {
        let _ = CString::from_raw(a.api_url);
    }
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

#[no_mangle]
pub unsafe extern "C" fn lb_account_result_free(r: LbAccountResult) {
    lb_account_free(r.ok);
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
/// The returned value must be passed to `lb_account_result_free` to avoid a memory leak.
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

/// # Safety
///
/// The returned value must be passed to `lb_account_result_free` to avoid a memory leak.
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
/// The returned value must be passed to `lb_file_result_free` to avoid a memory leak.
#[no_mangle]
pub unsafe extern "C" fn lb_create_file(
    core: *mut c_void,
    name: *const c_char,
    parent: *const c_char,
    ft: LbFileType,
) -> LbFileResult {
    let mut r = lb_file_result_new();
    let parent = uuid_or_return!(parent, r);
    let ftype = match ft.tag {
        LbFileTypeTag::LbFileTypeTagDocument => FileType::Document,
        LbFileTypeTag::LbFileTypeTagFolder => FileType::Folder,
        LbFileTypeTag::LbFileTypeTagLink => {
            let target = uuid_or_return!(ft.link_target, r);
            FileType::Link { target }
        }
    };
    match core!(core).create_file(rstr(name), parent, ftype) {
        Ok(f) => r.ok = lb_file_new(f),
        Err(err) => {
            r.err.msg = cstr(format!("{:?}", err));
            r.err.code = LbErrorCode::Unexpected;
        }
    }
    r
}

/// # Safety
///
/// The returned value must be passed to `lb_file_result_free` to avoid a memory leak.
#[no_mangle]
pub unsafe extern "C" fn lb_create_file_at_path(
    core: *mut c_void,
    path_and_name: *const c_char,
) -> LbFileResult {
    let mut r = lb_file_result_new();
    match core!(core).create_at_path(rstr(path_and_name)) {
        Ok(f) => r.ok = lb_file_new(f),
        Err(err) => {
            use CreateFileAtPathError::*;
            r.err.msg = cstr(format!("{:?}", err));
            r.err.code = match err {
                Error::UiError(err) => match err {
                    FileAlreadyExists => LbErrorCode::FileExists,
                    NoRoot => LbErrorCode::NoRoot,
                    PathContainsEmptyFile => LbErrorCode::PathContainsEmptyFile,
                    DocumentTreatedAsFolder => LbErrorCode::FileIsNotFolder,
                    InsufficientPermission => LbErrorCode::InsufficientPermission,
                },
                Error::Unexpected(_) => LbErrorCode::Unexpected,
            };
        }
    }
    r
}

/// # Safety
///
/// The returned value must be passed to `lb_error_free` to avoid a memory leak.
#[no_mangle]
pub unsafe extern "C" fn lb_write_document(
    core: *mut c_void,
    id: *const c_char,
    data: *const u8,
    len: i32,
) -> LbError {
    let mut e = lb_error_none();
    let id = uuid_or_return!(id);
    let data = std::slice::from_raw_parts(data, len as usize);
    if let Err(err) = core!(core).write_document(id, data) {
        use WriteToDocumentError::*;
        e.msg = cstr(format!("{:?}", err));
        e.code = match err {
            Error::UiError(err) => match err {
                FileDoesNotExist => LbErrorCode::FileNotFound,
                FolderTreatedAsDocument => LbErrorCode::FileIsNotDocument,
                InsufficientPermission => LbErrorCode::InsufficientPermission,
            },
            Error::Unexpected(_) => LbErrorCode::Unexpected,
        };
    }
    e
}

/// # Safety
///
/// The returned value must be passed to `lb_file_result_free` to avoid a memory leak.
#[no_mangle]
pub unsafe extern "C" fn lb_get_file_by_id(core: *mut c_void, id: *const c_char) -> LbFileResult {
    let mut r = lb_file_result_new();
    let id = uuid_or_return!(id, r);
    match core!(core).get_file_by_id(id) {
        Ok(f) => r.ok = lb_file_new(f),
        Err(err) => {
            r.err.msg = cstr(format!("{:?}", err));
            r.err.code = match err {
                Error::UiError(err) => match err {
                    GetFileByIdError::NoFileWithThatId => LbErrorCode::FileNotFound,
                },
                Error::Unexpected(_) => LbErrorCode::Unexpected,
            };
        }
    }
    r
}

/// # Safety
///
/// The returned value must be passed to `lb_file_result_free` to avoid a memory leak.
#[no_mangle]
pub unsafe extern "C" fn lb_get_file_by_path(
    core: *mut c_void,
    path: *const c_char,
) -> LbFileResult {
    let mut r = lb_file_result_new();
    match core!(core).get_by_path(rstr(path)) {
        Ok(f) => r.ok = lb_file_new(f),
        Err(err) => {
            r.err.msg = cstr(format!("{:?}", err));
            r.err.code = match err {
                Error::UiError(err) => match err {
                    GetFileByPathError::NoFileAtThatPath => LbErrorCode::FileNotFound,
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
#[no_mangle]
pub unsafe extern "C" fn lb_get_path_by_id(core: *mut c_void, id: *const c_char) -> LbStringResult {
    let mut r = lb_string_result_new();
    let id = uuid_or_return!(id, r);
    match core!(core).get_path_by_id(id) {
        Ok(path) => r.ok = cstr(path),
        Err(err) => {
            r.err.msg = cstr(format!("{:?}", err));
            r.err.code = LbErrorCode::Unexpected;
        }
    }
    r
}

/// # Safety
///
/// The returned value must be passed to `lb_file_list_result_free` to avoid a memory leak.
#[no_mangle]
pub unsafe extern "C" fn lb_get_children(core: *mut c_void, id: *const c_char) -> LbFileListResult {
    let mut r = lb_file_list_result_new();
    let id = uuid_or_return!(id, r);
    match core!(core).get_children(id) {
        Ok(files) => lb_file_list_init(&mut r.ok, files),
        Err(err) => {
            r.err.msg = cstr(format!("{:?}", err));
            r.err.code = LbErrorCode::Unexpected;
        }
    }
    r
}

/// # Safety
///
/// The returned value must be passed to `lb_file_list_result_free` to avoid a memory leak.
#[no_mangle]
pub unsafe extern "C" fn lb_get_and_get_children_recursively(
    core: *mut c_void,
    id: *const c_char,
) -> LbFileListResult {
    let mut r = lb_file_list_result_new();
    let id = uuid_or_return!(id, r);
    match core!(core).get_and_get_children_recursively(id) {
        Ok(files) => lb_file_list_init(&mut r.ok, files),
        Err(err) => {
            use GetAndGetChildrenError::*;
            r.err.msg = cstr(format!("{:?}", err));
            r.err.code = match err {
                Error::UiError(err) => match err {
                    FileDoesNotExist => LbErrorCode::FileNotFound,
                    DocumentTreatedAsFolder => LbErrorCode::FileIsNotFolder,
                },
                Error::Unexpected(_) => LbErrorCode::Unexpected,
            };
        }
    }
    r
}

/// # Safety
///
/// The returned value must be passed to `lb_file_list_result_free` to avoid a memory leak.
#[no_mangle]
pub unsafe extern "C" fn lb_list_metadatas(core: *mut c_void) -> LbFileListResult {
    let mut r = lb_file_list_result_new();
    match core!(core).list_metadatas() {
        Ok(files) => lb_file_list_init(&mut r.ok, files),
        Err(err) => {
            r.err.msg = cstr(format!("{:?}", err));
            r.err.code = LbErrorCode::Unexpected;
        }
    }
    r
}

/// # Safety
///
/// The returned value must be passed to `lb_bytes_result_free` to avoid a memory leak.
#[no_mangle]
pub unsafe extern "C" fn lb_read_document(core: *mut c_void, id: *const c_char) -> LbBytesResult {
    let mut r = lb_bytes_result_new();
    let id = uuid_or_return!(id, r);
    match core!(core).read_document(id) {
        Ok(mut data) => {
            data.shrink_to_fit();
            let mut data = std::mem::ManuallyDrop::new(data);
            r.bytes = data.as_mut_ptr();
            r.count = data.len();
        }
        Err(err) => {
            use ReadDocumentError::*;
            r.err.msg = cstr(format!("{:?}", err));
            r.err.code = match err {
                Error::UiError(err) => match err {
                    TreatedFolderAsDocument => LbErrorCode::FileIsNotDocument,
                    FileDoesNotExist => LbErrorCode::FileNotFound,
                },
                Error::Unexpected(_) => LbErrorCode::Unexpected,
            };
        }
    }
    r
}

#[repr(C)]
pub struct LbImexFileInfo {
    pub disk_path: *mut c_char,
    pub lb_path: *mut c_char,
}

#[no_mangle]
pub unsafe extern "C" fn lb_imex_file_info_free(fi: LbImexFileInfo) {
    if !fi.disk_path.is_null() {
        let _ = CString::from_raw(fi.disk_path);
    }
    if !fi.lb_path.is_null() {
        let _ = CString::from_raw(fi.lb_path);
    }
}

pub type LbImexCallback = unsafe extern "C" fn(LbImexFileInfo, *mut c_void);

/// # Safety
///
/// The returned value must be passed to `lb_error_free` to avoid a memory leak.
#[no_mangle]
pub unsafe extern "C" fn lb_export_file(
    core: *mut c_void,
    id: *const c_char,
    dest: *const c_char,
    progress: LbImexCallback,
    user_data: *mut c_void,
) -> LbError {
    let mut e = lb_error_none();
    let id = uuid_or_return!(id);
    if let Err(err) = core!(core).export_file(
        id,
        rstr(dest).into(),
        false,
        Some(Box::new(move |info| {
            let c_info = LbImexFileInfo {
                disk_path: cstr(info.disk_path.to_string_lossy().to_string()),
                lb_path: cstr(info.lockbook_path),
            };
            progress(c_info, user_data)
        })),
    ) {
        use ExportFileError::*;
        e.msg = cstr(format!("{:?}", err));
        e.code = match err {
            Error::UiError(err) => match err {
                ParentDoesNotExist => LbErrorCode::Unexpected,
                DiskPathTaken => LbErrorCode::Unexpected,
                DiskPathInvalid => LbErrorCode::Unexpected,
            },
            Error::Unexpected(_) => LbErrorCode::Unexpected,
        };
    }
    e
}

/// # Safety
///
/// The returned value must be passed to `lb_bytes_result_free` to avoid a memory leak.
#[no_mangle]
pub unsafe extern "C" fn lb_export_drawing(
    core: *mut c_void,
    id: *const c_char,
    fmt_code: u8,
) -> LbBytesResult {
    let mut r = lb_bytes_result_new();
    let id = uuid_or_return!(id, r);
    // These values are bound together in a unit test at the end of this file.
    let img_fmt = match fmt_code {
        0 => SupportedImageFormats::Png,
        1 => SupportedImageFormats::Jpeg,
        2 => SupportedImageFormats::Pnm,
        3 => SupportedImageFormats::Tga,
        4 => SupportedImageFormats::Farbfeld,
        5 => SupportedImageFormats::Bmp,
        n => {
            r.err.msg = cstr(format!("unknown image format code {}", n));
            r.err.code = LbErrorCode::Unexpected;
            return r;
        }
    };
    match core!(core).export_drawing(id, img_fmt, None) {
        Ok(mut data) => {
            data.shrink_to_fit();
            let mut data = std::mem::ManuallyDrop::new(data);
            r.bytes = data.as_mut_ptr();
            r.count = data.len();
        }
        Err(err) => {
            use ExportDrawingError::*;
            r.err.msg = cstr(format!("{:?}", err));
            r.err.code = match err {
                Error::UiError(err) => match err {
                    FolderTreatedAsDrawing => LbErrorCode::FileIsNotDocument,
                    FileDoesNotExist => LbErrorCode::FileNotFound,
                    InvalidDrawing => LbErrorCode::InvalidDrawing,
                },
                Error::Unexpected(_) => LbErrorCode::Unexpected,
            };
        }
    }
    r
}

/// # Safety
///
/// The returned value must be passed to `lb_error_free` to avoid a memory leak.
#[no_mangle]
pub unsafe extern "C" fn lb_delete_file(core: *mut c_void, id: *const c_char) -> LbError {
    let mut e = lb_error_none();
    let id = uuid_or_return!(id);
    if let Err(err) = core!(core).delete_file(id) {
        use FileDeleteError::*;
        e.msg = cstr(format!("{:?}", err));
        e.code = match err {
            Error::UiError(err) => match err {
                CannotDeleteRoot => LbErrorCode::NoRootOps,
                FileDoesNotExist => LbErrorCode::FileNotFound,
                InsufficientPermission => LbErrorCode::InsufficientPermission,
            },
            Error::Unexpected(_) => LbErrorCode::Unexpected,
        };
    }
    e
}

/// # Safety
///
/// The returned value must be passed to `lb_error_free` to avoid a memory leak.
#[no_mangle]
pub unsafe extern "C" fn lb_move_file(
    core: *mut c_void,
    id: *const c_char,
    new_parent: *const c_char,
) -> LbError {
    let mut e = lb_error_none();
    let id = uuid_or_return!(id);
    let new_parent = uuid_or_return!(new_parent);
    if let Err(err) = core!(core).move_file(id, new_parent) {
        use MoveFileError::*;
        e.msg = cstr(format!("{:?}", err));
        e.code = match err {
            Error::UiError(err) => match err {
                CannotMoveRoot => LbErrorCode::NoRootOps,
                DocumentTreatedAsFolder => LbErrorCode::FileIsNotFolder,
                FileDoesNotExist => LbErrorCode::FileNotFound,
                FolderMovedIntoItself => LbErrorCode::FolderMovedIntoItself,
                TargetParentDoesNotExist => LbErrorCode::TargetParentNotFound,
                TargetParentHasChildNamedThat => LbErrorCode::FileNameUnavailable,
                LinkInSharedFolder => LbErrorCode::LinkInSharedFolder,
                InsufficientPermission => LbErrorCode::InsufficientPermission,
            },
            Error::Unexpected(_) => LbErrorCode::Unexpected,
        };
    }
    e
}

/// # Safety
///
/// The returned value must be passed to `lb_error_free` to avoid a memory leak.
#[no_mangle]
pub unsafe extern "C" fn lb_rename_file(
    core: *mut c_void,
    id: *const c_char,
    new_name: *const c_char,
) -> LbError {
    let mut e = lb_error_none();
    let id = uuid_or_return!(id);
    if let Err(err) = core!(core).rename_file(id, rstr(new_name)) {
        use RenameFileError::*;
        e.msg = cstr(format!("{:?}", err));
        e.code = match err {
            Error::UiError(err) => match err {
                FileDoesNotExist => LbErrorCode::FileNotFound,
                NewNameEmpty => LbErrorCode::FileNameEmpty,
                NewNameContainsSlash => LbErrorCode::FileNameContainsSlash,
                FileNameNotAvailable => LbErrorCode::FileNameUnavailable,
                CannotRenameRoot => LbErrorCode::NoRootOps,
                InsufficientPermission => LbErrorCode::InsufficientPermission,
            },
            Error::Unexpected(_) => LbErrorCode::Unexpected,
        };
    }
    e
}

#[repr(C)]
pub struct LbCalcWorkResult {
    units: *mut LbWorkUnit,
    num_units: usize,
    last_server_update_at: u64,
    err: LbError,
}

#[no_mangle]
pub unsafe extern "C" fn lb_calc_work_result_index(
    r: LbCalcWorkResult,
    i: usize,
) -> *mut LbWorkUnit {
    r.units.add(i)
}

#[no_mangle]
pub unsafe extern "C" fn lb_calc_work_result_free(r: LbCalcWorkResult) {
    if !r.units.is_null() {
        let units = Vec::from_raw_parts(r.units, r.num_units, r.num_units);
        for wu in units {
            lb_file_free(wu.file);
        }
    }
    lb_error_free(r.err);
}

#[repr(C)]
pub struct LbWorkUnit {
    typ: LbWorkUnitType,
    file: LbFile,
}

#[repr(C)]
pub enum LbWorkUnitType {
    LbWorkUnitLocal,
    LbWorkUnitServer,
}

/// # Safety
///
/// The returned value must be passed to `lb_calc_work_result_free` to avoid a memory leak.
#[no_mangle]
pub unsafe extern "C" fn lb_calculate_work(core: *mut c_void) -> LbCalcWorkResult {
    let mut r = LbCalcWorkResult {
        units: null_mut(),
        num_units: 0,
        last_server_update_at: 0,
        err: lb_error_none(),
    };
    match core!(core).calculate_work() {
        Ok(work) => {
            let mut list = Vec::with_capacity(work.work_units.len());
            for wu in work.work_units {
                let typ = match wu {
                    WorkUnit::LocalChange { .. } => LbWorkUnitType::LbWorkUnitLocal,
                    WorkUnit::ServerChange { .. } => LbWorkUnitType::LbWorkUnitServer,
                };
                let file = lb_file_new(match wu {
                    WorkUnit::LocalChange { metadata } => metadata,
                    WorkUnit::ServerChange { metadata } => metadata,
                });
                list.push(LbWorkUnit { typ, file });
            }
            let mut list = std::mem::ManuallyDrop::new(list);
            r.units = list.as_mut_ptr();
            r.num_units = list.len();
            r.last_server_update_at = work.most_recent_update_from_server;
        }
        Err(err) => {
            use CalculateWorkError::*;
            r.err.msg = cstr(format!("{:?}", err));
            r.err.code = match err {
                Error::UiError(err) => match err {
                    CouldNotReachServer => LbErrorCode::CouldNotReachServer,
                    ClientUpdateRequired => LbErrorCode::ClientUpdateRequired,
                },
                Error::Unexpected(_) => LbErrorCode::Unexpected,
            };
        }
    }
    r
}

#[repr(C)]
pub struct LbSyncProgress {
    total: u64,
    progress: u64,
    current_wu: LbClientWorkUnit,
}

#[repr(C)]
pub struct LbClientWorkUnit {
    pull_meta: bool,
    push_meta: bool,
    pull_doc: *mut c_char,
    push_doc: *mut c_char,
}

#[no_mangle]
pub unsafe extern "C" fn lb_sync_progress_free(sp: LbSyncProgress) {
    if !sp.current_wu.pull_doc.is_null() {
        let _ = CString::from_raw(sp.current_wu.pull_doc);
    }
    if !sp.current_wu.push_doc.is_null() {
        let _ = CString::from_raw(sp.current_wu.push_doc);
    }
}

pub type LbSyncProgressCallback = unsafe extern "C" fn(LbSyncProgress, *mut c_void);

/// # Safety
///
/// The returned value must be passed to `lb_error_free` to avoid a memory leak.
#[no_mangle]
pub unsafe extern "C" fn lb_sync_all(
    core: *mut c_void,
    progress: LbSyncProgressCallback,
    user_data: *mut c_void,
) -> LbError {
    let mut e = lb_error_none();
    if let Err(err) = core!(core).sync(Some(Box::new(move |sp| {
        let mut cwu = LbClientWorkUnit {
            pull_meta: false,
            push_meta: false,
            pull_doc: null_mut(),
            push_doc: null_mut(),
        };
        match sp.current_work_unit {
            ClientWorkUnit::PullMetadata => cwu.pull_meta = true,
            ClientWorkUnit::PushMetadata => cwu.push_meta = true,
            ClientWorkUnit::PullDocument(v) => cwu.pull_doc = cstr(v),
            ClientWorkUnit::PushDocument(v) => cwu.push_doc = cstr(v),
        };
        let c_sp = LbSyncProgress {
            total: sp.total as u64,
            progress: sp.progress as u64,
            current_wu: cwu,
        };
        progress(c_sp, user_data);
    }))) {
        use SyncAllError::*;
        e.msg = cstr(format!("{:?}", err));
        e.code = match err {
            Error::UiError(err) => match err {
                Retry => LbErrorCode::Unexpected,
                ClientUpdateRequired => LbErrorCode::ClientUpdateRequired,
                CouldNotReachServer => LbErrorCode::CouldNotReachServer,
            },
            Error::Unexpected(_) => LbErrorCode::Unexpected,
        };
    }
    e
}

/// # Safety
///
/// The returned value must be passed to `lb_string_result_free` to avoid a memory leak.
#[no_mangle]
pub unsafe extern "C" fn lb_get_last_synced_human_string(core: *mut c_void) -> LbStringResult {
    let mut r = lb_string_result_new();
    match core!(core).get_last_synced_human_string() {
        Ok(acct_str) => r.ok = cstr(acct_str),
        Err(err) => {
            r.err.msg = cstr(format!("{:?}", err));
            r.err.code = LbErrorCode::Unexpected;
        }
    }
    r
}

#[repr(C)]
pub struct LbUsageResult {
    usages: *mut LbFileUsage,
    num_usages: usize,
    server_usage: LbUsageItemMetric,
    data_cap: LbUsageItemMetric,
    err: LbError,
}

#[no_mangle]
pub unsafe extern "C" fn lb_usage_result_index(r: LbUsageResult, i: usize) -> *mut LbFileUsage {
    r.usages.add(i)
}

#[no_mangle]
pub unsafe extern "C" fn lb_usage_result_free(r: LbUsageResult) {
    let usages = Vec::from_raw_parts(r.usages, r.num_usages, r.num_usages);
    for fu in usages {
        if !fu.id.is_null() {
            let _ = CString::from_raw(fu.id);
        }
    }
    lb_usage_item_metric_free(r.server_usage);
    lb_usage_item_metric_free(r.data_cap);
    lb_error_free(r.err);
}

#[repr(C)]
pub struct LbUsageItemMetric {
    exact: u64,
    readable: *mut c_char,
}

fn lb_usage_item_metric_none() -> LbUsageItemMetric {
    LbUsageItemMetric {
        exact: 0,
        readable: null_mut(),
    }
}

#[no_mangle]
pub unsafe extern "C" fn lb_usage_item_metric_free(m: LbUsageItemMetric) {
    if !m.readable.is_null() {
        let _ = CString::from_raw(m.readable);
    }
}

#[repr(C)]
pub struct LbFileUsage {
    id: *mut c_char,
    size_bytes: u64,
}

/// # Safety
///
/// The returned value must be passed to `lb_usage_result_free` to avoid a memory leak.
#[no_mangle]
pub unsafe extern "C" fn lb_get_usage(core: *mut c_void) -> LbUsageResult {
    let mut r = LbUsageResult {
        usages: null_mut(),
        num_usages: 0,
        server_usage: lb_usage_item_metric_none(),
        data_cap: lb_usage_item_metric_none(),
        err: lb_error_none(),
    };
    match core!(core).get_usage() {
        Ok(m) => {
            let mut usages = Vec::with_capacity(m.usages.len());
            for fu in m.usages {
                usages.push(LbFileUsage {
                    id: cstr(fu.file_id.to_string()),
                    size_bytes: fu.size_bytes,
                });
            }
            let mut usages = std::mem::ManuallyDrop::new(usages);
            r.usages = usages.as_mut_ptr();
            r.num_usages = usages.len();
            r.server_usage.exact = m.server_usage.exact;
            r.server_usage.readable = cstr(m.server_usage.readable);
            r.data_cap.exact = m.data_cap.exact;
            r.data_cap.readable = cstr(m.data_cap.readable);
        }
        Err(err) => {
            use GetUsageError::*;
            r.err.msg = cstr(format!("{:?}", err));
            r.err.code = match err {
                Error::UiError(err) => match err {
                    ClientUpdateRequired => LbErrorCode::ClientUpdateRequired,
                    CouldNotReachServer => LbErrorCode::CouldNotReachServer,
                },
                Error::Unexpected(_) => LbErrorCode::Unexpected,
            };
        }
    }
    r
}

#[repr(C)]
pub struct LbUncUsageResult {
    ok: LbUsageItemMetric,
    err: LbError,
}

#[no_mangle]
pub unsafe extern "C" fn lb_unc_usage_result_free(r: LbUncUsageResult) {
    lb_usage_item_metric_free(r.ok);
    lb_error_free(r.err);
}

/// # Safety
///
/// The returned value must be passed to `lb_unc_usage_result_free` to avoid a memory leak.
#[no_mangle]
pub unsafe extern "C" fn lb_get_uncompressed_usage(core: *mut c_void) -> LbUncUsageResult {
    let mut r = LbUncUsageResult {
        ok: lb_usage_item_metric_none(),
        err: lb_error_none(),
    };
    match core!(core).get_uncompressed_usage() {
        Ok(im) => {
            r.ok.exact = im.exact;
            r.ok.readable = cstr(im.readable);
        }
        Err(err) => {
            use GetUsageError::*;
            r.err.msg = cstr(format!("{:?}", err));
            r.err.code = match err {
                Error::UiError(err) => match err {
                    ClientUpdateRequired => LbErrorCode::ClientUpdateRequired,
                    CouldNotReachServer => LbErrorCode::CouldNotReachServer,
                },
                Error::Unexpected(_) => LbErrorCode::Unexpected,
            };
        }
    }
    r
}

/// # Safety
///
/// The returned value must be passed to `lb_error_free` to avoid a memory leak.
#[no_mangle]
pub unsafe extern "C" fn lb_share_file(
    core: *mut c_void,
    id: *const c_char,
    uname: *const c_char,
    mode: LbShareMode,
) -> LbError {
    let mut e = lb_error_none();
    let id = uuid_or_return!(id);
    let mode = match mode {
        LbShareMode::LbShareModeRead => ShareMode::Read,
        LbShareMode::LbShareModeWrite => ShareMode::Write,
    };
    if let Err(err) = core!(core).share_file(id, rstr(uname), mode) {
        e.msg = cstr(format!("{:?}", err));
        e.code = LbErrorCode::Unexpected;
    }
    e
}

/// # Safety
///
/// The returned value must be passed to `lb_file_list_result_free` to avoid a memory leak.
#[no_mangle]
pub unsafe extern "C" fn lb_get_pending_shares(core: *mut c_void) -> LbFileListResult {
    let mut r = lb_file_list_result_new();
    match core!(core).get_pending_shares() {
        Ok(files) => lb_file_list_init(&mut r.ok, files),
        Err(err) => {
            r.err.msg = cstr(format!("{:?}", err));
            r.err.code = LbErrorCode::Unexpected;
        }
    }
    r
}

/// # Safety
///
/// The returned value must be passed to `lb_error_free` to avoid a memory leak.
#[no_mangle]
pub unsafe extern "C" fn lb_delete_pending_share(core: *mut c_void, id: *const c_char) -> LbError {
    let mut e = lb_error_none();
    let id = uuid_or_return!(id);
    if let Err(err) = core!(core).delete_pending_share(id) {
        e.msg = cstr(format!("{:?}", err));
        e.code = match err {
            Error::UiError(_) => LbErrorCode::FileNotFound,
            Error::Unexpected(_) => LbErrorCode::Unexpected,
        };
    }
    e
}

#[repr(C)]
pub struct LbSubInfoResult {
    stripe_last4: *mut c_char,
    google_play_state: u8,
    app_store_state: u8,
    period_end: u64,
    err: LbError,
}

#[no_mangle]
pub unsafe extern "C" fn lb_sub_info_result_free(r: LbSubInfoResult) {
    if !r.stripe_last4.is_null() {
        let _ = CString::from_raw(r.stripe_last4);
    }
    lb_error_free(r.err);
}

/// # Safety
///
/// The returned value must be passed to `lb_sub_info_result_free` to avoid a memory leak.
#[no_mangle]
pub unsafe extern "C" fn lb_get_subscription_info(core: *mut c_void) -> LbSubInfoResult {
    let mut r = LbSubInfoResult {
        stripe_last4: null_mut(),
        google_play_state: 0,
        app_store_state: 0,
        period_end: 0,
        err: lb_error_none(),
    };
    match core!(core).get_subscription_info() {
        Ok(None) => {} // Leave zero values for no subscription info.
        Ok(Some(info)) => {
            use PaymentPlatform::*;
            match info.payment_platform {
                Stripe { card_last_4_digits } => r.stripe_last4 = cstr(card_last_4_digits),
                // The integer representations of both the google play and app store account
                // state enums are bound together by a unit test at the end of this file.
                GooglePlay { account_state } => r.google_play_state = account_state as u8 + 1,
                AppStore { account_state } => r.app_store_state = account_state as u8 + 1,
            }
            r.period_end = info.period_end;
        }
        Err(err) => {
            r.err.msg = cstr(format!("{:?}", err));
            r.err.code = LbErrorCode::Unexpected;
        }
    }
    r
}

/// # Safety
///
/// The returned value must be passed to `lb_error_free` to avoid a memory leak.
#[no_mangle]
pub unsafe extern "C" fn lb_upgrade_account_stripe_old_card(core: *mut c_void) -> LbError {
    let mut e = lb_error_none();
    if let Err(err) =
        core!(core).upgrade_account_stripe(StripeAccountTier::Premium(PaymentMethod::OldCard))
    {
        e.msg = cstr(format!("{:?}", err));
        e.code = LbErrorCode::Unexpected;
    }
    e
}

/// # Safety
///
/// The returned value must be passed to `lb_error_free` to avoid a memory leak.
#[no_mangle]
pub unsafe extern "C" fn lb_upgrade_account_stripe_new_card(
    core: *mut c_void,
    num: *const c_char,
    exp_year: i32,
    exp_month: i32,
    cvc: *const c_char,
) -> LbError {
    let mut e = lb_error_none();
    let number = rstr(num).to_string();
    let cvc = rstr(cvc).to_string();

    if let Err(err) =
        core!(core).upgrade_account_stripe(StripeAccountTier::Premium(PaymentMethod::NewCard {
            number,
            exp_year,
            exp_month,
            cvc,
        }))
    {
        e.msg = cstr(format!("{:?}", err));
        e.code = LbErrorCode::Unexpected;
    }
    e
}

/// # Safety
///
/// The returned value must be passed to `lb_error_free` to avoid a memory leak.
#[no_mangle]
pub unsafe extern "C" fn lb_cancel_subscription(core: *mut c_void) -> LbError {
    let mut e = lb_error_none();
    if let Err(err) = core!(core).cancel_subscription() {
        use CancelSubscriptionError::*;
        e.msg = cstr(format!("{:?}", err));
        e.code = match err {
            Error::UiError(err) => match err {
                NotPremium => LbErrorCode::NotPremium,
                AlreadyCanceled => LbErrorCode::SubscriptionAlreadyCanceled,
                UsageIsOverFreeTierDataCap => LbErrorCode::UsageIsOverFreeTierDataCap,
                ExistingRequestPending => LbErrorCode::ExistingRequestPending,
                CouldNotReachServer => LbErrorCode::CouldNotReachServer,
                ClientUpdateRequired => LbErrorCode::ClientUpdateRequired,
                CannotCancelForAppStore => LbErrorCode::CannotCancelForAppStore,
            },
            Error::Unexpected(_) => LbErrorCode::Unexpected,
        };
    }
    e
}

#[repr(C)]
pub struct LbValidateResult {
    warnings: *mut *mut c_char,
    n_warnings: usize,
    err: LbError,
}

#[no_mangle]
pub unsafe extern "C" fn lb_validate_result_index(r: LbValidateResult, i: usize) -> *mut c_char {
    *r.warnings.add(i)
}

#[no_mangle]
pub unsafe extern "C" fn lb_validate_result_free(r: LbValidateResult) {
    let warnings = Vec::from_raw_parts(r.warnings, r.n_warnings, r.n_warnings);
    for w in warnings {
        if !w.is_null() {
            let _ = CString::from_raw(w);
        }
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

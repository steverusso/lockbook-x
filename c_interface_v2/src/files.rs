use lockbook_core::{
    CreateFileAtPathError, ExportDrawingError, ExportFileError, FileDeleteError,
    GetAndGetChildrenError, GetFileByIdError, GetFileByPathError, MoveFileError, ReadDocumentError,
    RenameFileError, WriteToDocumentError,
};

use crate::*;

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

pub fn lb_file_new(f: File) -> LbFile {
    let mut typ = lb_file_type_doc();
    if let FileType::Folder = f.file_type {
        typ.tag = LbFileTypeTag::Folder;
    }
    if let FileType::Link { target } = f.file_type {
        typ.tag = LbFileTypeTag::Link;
        typ.link_target = target.into_bytes();
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

pub unsafe fn lb_file_free(f: LbFile) {
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
    lb_share_list_free(f.shares);
}

/// The zero value represents a document.
#[repr(C)]
pub struct LbFileType {
    tag: LbFileTypeTag,
    link_target: [u8; 16],
}

#[repr(C)]
pub enum LbFileTypeTag {
    Document,
    Folder,
    Link,
}

#[no_mangle]
pub extern "C" fn lb_file_type_doc() -> LbFileType {
    LbFileType {
        tag: LbFileTypeTag::Document,
        link_target: [0; 16],
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
                ShareMode::Read => LbShareMode::Read,
                ShareMode::Write => LbShareMode::Write,
            },
        });
    }
    let mut list = std::mem::ManuallyDrop::new(list);
    LbShareList {
        list: list.as_mut_ptr(),
        count: list.len(),
    }
}

/// # Safety
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
    Read,
    Write,
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

/// # Safety
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

/// # Safety
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

/// # Safety
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
        LbFileTypeTag::Document => FileType::Document,
        LbFileTypeTag::Folder => FileType::Folder,
        LbFileTypeTag::Link => FileType::Link {
            target: Uuid::from_bytes(ft.link_target),
        },
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
/// The returned value must be passed to `lb_file_result_free` to avoid a memory leak.
#[no_mangle]
pub unsafe extern "C" fn lb_get_root(core: *mut c_void) -> LbFileResult {
    let mut r = lb_file_result_new();
    match core!(core).get_root() {
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

#[repr(C)]
pub struct LbImexFileInfo {
    pub disk_path: *mut c_char,
    pub lb_path: *mut c_char,
}

/// # Safety
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
    // These values are bound together in a unit test in this crate.
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
        LbShareMode::Read => ShareMode::Read,
        LbShareMode::Write => ShareMode::Write,
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

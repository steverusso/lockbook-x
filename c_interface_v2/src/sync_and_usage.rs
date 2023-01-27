use lockbook_core::{CalculateWorkError, ClientWorkUnit, GetUsageError, SyncAllError, WorkUnit};

use crate::*;

#[repr(C)]
pub struct LbCalcWorkResult {
    units: *mut LbWorkUnit,
    num_units: usize,
    last_server_update_at: u64,
    err: LbError,
}

#[repr(C)]
pub struct LbWorkUnit {
    pub typ: LbWorkUnitType,
    pub file: LbFile,
}

#[repr(C)]
pub enum LbWorkUnitType {
    Local,
    Server,
}

/// # Safety
#[no_mangle]
pub unsafe extern "C" fn lb_calc_work_result_index(
    r: LbCalcWorkResult,
    i: usize,
) -> *mut LbWorkUnit {
    r.units.add(i)
}

/// # Safety
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
                    WorkUnit::LocalChange { .. } => LbWorkUnitType::Local,
                    WorkUnit::ServerChange { .. } => LbWorkUnitType::Server,
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

/// # Safety
#[no_mangle]
pub unsafe extern "C" fn lb_sync_progress_free(sp: LbSyncProgress) {
    if !sp.current_wu.pull_doc.is_null() {
        libc::free(sp.current_wu.pull_doc as *mut c_void);
    }
    if !sp.current_wu.push_doc.is_null() {
        libc::free(sp.current_wu.push_doc as *mut c_void);
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

/// # Safety
#[no_mangle]
pub unsafe extern "C" fn lb_usage_result_index(r: LbUsageResult, i: usize) -> *mut LbFileUsage {
    r.usages.add(i)
}

/// # Safety
#[no_mangle]
pub unsafe extern "C" fn lb_usage_result_free(r: LbUsageResult) {
    let usages = Vec::from_raw_parts(r.usages, r.num_usages, r.num_usages);
    for fu in usages {
        if !fu.id.is_null() {
            libc::free(fu.id as *mut c_void);
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

/// # Safety
#[no_mangle]
pub unsafe extern "C" fn lb_usage_item_metric_free(m: LbUsageItemMetric) {
    if !m.readable.is_null() {
        libc::free(m.readable as *mut c_void);
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

/// # Safety
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

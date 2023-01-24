package lockbook

/*
#cgo LDFLAGS: ${SRCDIR}/../target/release/libc_interface_v2.a -lm
#include "../lockbook_core.h"

extern void go_imex_callback(struct LbImexFileInfo info, void *h);

extern void go_sync_callback(struct LbSyncProgress sp, void *h);
*/
import "C"

import (
	"os"
	"runtime/cgo"
	"time"
	"unsafe"
)

type lbCoreFFI struct {
	ref unsafe.Pointer
}

func initLbCoreFFI(fpath string) (*lbCoreFFI, error) {
	cPath := C.CString(fpath)
	defer C.free(unsafe.Pointer(cPath))
	r := C.lb_init(cPath, true)
	defer C.lb_error_free(r.err)

	if r.err.code != 0 {
		return nil, newErrorFromC(r.err)
	}
	return &lbCoreFFI{ref: r.core}, nil
}

func (l *lbCoreFFI) GetAccount() (Account, error) {
	r := C.lb_get_account(l.ref)
	defer C.lb_account_result_free(r)
	return goAccountResult(r)
}

func (l *lbCoreFFI) CreateAccount(uname string, welcome bool) (Account, error) {
	cUsername := C.CString(uname)
	defer C.free(unsafe.Pointer(cUsername))
	cAPIURL := C.CString(os.Getenv("API_URL"))
	defer C.free(unsafe.Pointer(cAPIURL))
	r := C.lb_create_account(l.ref, cUsername, cAPIURL, C.bool(welcome))
	defer C.lb_account_result_free(r)
	return goAccountResult(r)
}

func (l *lbCoreFFI) ImportAccount(acctStr string) (Account, error) {
	cAcctStr := C.CString(acctStr)
	defer C.free(unsafe.Pointer(cAcctStr))
	r := C.lb_import_account(l.ref, cAcctStr)
	defer C.lb_account_result_free(r)
	return goAccountResult(r)
}

func (l *lbCoreFFI) ExportAccount() (string, error) {
	r := C.lb_export_account(l.ref)
	defer C.lb_string_result_free(r)
	return goStringResult(r)
}

func (l *lbCoreFFI) FileByID(id FileID) (File, error) {
	r := C.lb_get_file_by_id(l.ref, cFileID(id))
	defer C.lb_file_result_free(r)
	return goFileResult(r)
}

func (l *lbCoreFFI) FileByPath(lbPath string) (File, error) {
	cPath := C.CString(lbPath)
	defer C.free(unsafe.Pointer(cPath))
	r := C.lb_get_file_by_path(l.ref, cPath)
	defer C.lb_file_result_free(r)
	return goFileResult(r)
}

func (l *lbCoreFFI) PathByID(id FileID) (string, error) {
	r := C.lb_get_path_by_id(l.ref, cFileID(id))
	defer C.lb_string_result_free(r)
	return goStringResult(r)
}

func (l *lbCoreFFI) GetRoot() (File, error) {
	r := C.lb_get_root(l.ref)
	defer C.lb_file_result_free(r)
	return goFileResult(r)
}

func (l *lbCoreFFI) GetChildren(id FileID) ([]File, error) {
	r := C.lb_get_children(l.ref, cFileID(id))
	defer C.lb_file_list_result_free(r)
	return goFileListResult(r)
}

func (l *lbCoreFFI) GetAndGetChildrenRecursively(id FileID) ([]File, error) {
	r := C.lb_get_and_get_children_recursively(l.ref, cFileID(id))
	defer C.lb_file_list_result_free(r)
	return goFileListResult(r)
}

func (l *lbCoreFFI) ListMetadatas() ([]File, error) {
	r := C.lb_list_metadatas(l.ref)
	defer C.lb_file_list_result_free(r)
	return goFileListResult(r)
}

func (l *lbCoreFFI) ReadDocument(id FileID) ([]byte, error) {
	r := C.lb_read_document(l.ref, cFileID(id))
	defer C.lb_bytes_result_free(r)
	return goBytesResult(r)
}

func (l *lbCoreFFI) WriteDocument(id FileID, data []byte) error {
	cData := C.CBytes(data)
	defer C.free(unsafe.Pointer(cData))
	e := C.lb_write_document(l.ref, cFileID(id), (*C.uchar)(cData), C.int(len(data)))
	defer C.lb_error_free(e)
	return newErrorFromC(e)
}

func (l *lbCoreFFI) CreateFile(name string, parentID FileID, typ FileType) (File, error) {
	cName := C.CString(name)
	defer C.free(unsafe.Pointer(cName))

	cTyp := C.lb_file_type_doc()
	switch typ := typ.(type) {
	case FileTypeFolder:
		cTyp.tag = C.LB_FILE_TYPE_TAG_FOLDER
	case FileTypeLink:
		cTyp.tag = C.LB_FILE_TYPE_TAG_LINK
		var i C.size_t
		for i = 0; i < 16; i++ {
			cTyp.link_target[i] = C.uint8_t(typ.Target[i])
		}
	}

	r := C.lb_create_file(l.ref, cName, cFileID(parentID), cTyp)
	defer C.lb_file_result_free(r)
	return goFileResult(r)
}

func (l *lbCoreFFI) CreateFileAtPath(lbPath string) (File, error) {
	cPath := C.CString(lbPath)
	defer C.free(unsafe.Pointer(cPath))
	r := C.lb_create_file_at_path(l.ref, cPath)
	defer C.lb_file_result_free(r)
	return goFileResult(r)
}

func (l *lbCoreFFI) DeleteFile(id FileID) error {
	e := C.lb_delete_file(l.ref, cFileID(id))
	defer C.lb_error_free(e)
	return newErrorFromC(e)
}

func (l *lbCoreFFI) RenameFile(id FileID, newName string) error {
	cNewName := C.CString(newName)
	defer C.free(unsafe.Pointer(cNewName))
	e := C.lb_rename_file(l.ref, cFileID(id), cNewName)
	defer C.lb_error_free(e)
	return newErrorFromC(e)
}

func (l *lbCoreFFI) MoveFile(srcID, destID FileID) error {
	e := C.lb_move_file(l.ref, cFileID(srcID), cFileID(destID))
	defer C.lb_error_free(e)
	return newErrorFromC(e)
}

//export go_imex_callback
func go_imex_callback(info C.LbImexFileInfo, handlePtr unsafe.Pointer) {
	h := (*C.uintptr_t)(handlePtr)
	fn := cgo.Handle(*h).Value().(func(C.LbImexFileInfo))
	fn(info)
}

func (l *lbCoreFFI) ExportFile(id FileID, dest string, fn func(ImportExportFileInfo)) error {
	handle := cgo.NewHandle(func(cInfo C.LbImexFileInfo) {
		defer C.lb_imex_file_info_free(cInfo)
		if fn == nil {
			return
		}
		fn(ImportExportFileInfo{
			DiskPath: C.GoString(cInfo.disk_path),
			LbPath:   C.GoString(cInfo.lb_path),
		})
	})
	defer handle.Delete()
	handlePtr := C.malloc(C.sizeof_uintptr_t)
	defer C.free(handlePtr)
	*(*C.uintptr_t)(handlePtr) = C.uintptr_t(handle)
	cDest := C.CString(dest)
	defer C.free(unsafe.Pointer(cDest))
	e := C.lb_export_file(l.ref, cFileID(id), cDest, C.LbImexCallback(C.go_imex_callback), handlePtr)
	defer C.lb_error_free(e)
	return newErrorFromC(e)
}

func (l *lbCoreFFI) ExportDrawing(id FileID, imgFmt ImageFormat) ([]byte, error) {
	r := C.lb_export_drawing(l.ref, cFileID(id), C.uchar(imgFmt))
	defer C.lb_bytes_result_free(r)
	return goBytesResult(r)
}

func (l *lbCoreFFI) GetLastSyncedHumanString() (string, error) {
	r := C.lb_get_last_synced_human_string(l.ref)
	defer C.lb_string_result_free(r)
	return goStringResult(r)
}

func (l *lbCoreFFI) GetUsage() (UsageMetrics, error) {
	r := C.lb_get_usage(l.ref)
	defer C.lb_usage_result_free(r)

	if r.err.code != 0 {
		return UsageMetrics{}, newErrorFromC(r.err)
	}
	fileUsages := make([]FileUsage, int(r.num_usages))
	var i C.size_t
	for i = 0; i < r.num_usages; i++ {
		fu := C.lb_usage_result_index(r, i)
		fileUsages[i] = FileUsage{
			FileID:    C.GoString(fu.id),
			SizeBytes: uint64(fu.size_bytes),
		}
	}
	return UsageMetrics{
		Usages: fileUsages,
		ServerUsage: UsageItemMetric{
			Exact:    uint64(r.server_usage.exact),
			Readable: C.GoString(r.server_usage.readable),
		},
		DataCap: UsageItemMetric{
			Exact:    uint64(r.data_cap.exact),
			Readable: C.GoString(r.data_cap.readable),
		},
	}, nil
}

func (l *lbCoreFFI) GetUncompressedUsage() (UsageItemMetric, error) {
	r := C.lb_get_uncompressed_usage(l.ref)
	defer C.lb_unc_usage_result_free(r)

	if r.err.code != 0 {
		return UsageItemMetric{}, newErrorFromC(r.err)
	}
	return UsageItemMetric{
		Exact:    uint64(r.ok.exact),
		Readable: C.GoString(r.ok.readable),
	}, nil
}

func (l *lbCoreFFI) CalculateWork() (WorkCalculated, error) {
	r := C.lb_calculate_work(l.ref)
	defer C.lb_calc_work_result_free(r)

	if r.err.code != 0 {
		return WorkCalculated{}, newErrorFromC(r.err)
	}
	workUnits := make([]WorkUnit, int(r.num_units))
	var i C.size_t
	for i = 0; i < r.num_units; i++ {
		cUnit := C.lb_calc_work_result_index(r, i)
		workUnits[i] = WorkUnit{
			Type: WorkUnitType(cUnit.typ),
			File: newFileFromC(&cUnit.file),
		}
	}
	return WorkCalculated{
		LastServerUpdateAt: uint64(r.last_server_update_at),
		WorkUnits:          workUnits,
	}, nil
}

//export go_sync_callback
func go_sync_callback(info C.LbSyncProgress, handlePtr unsafe.Pointer) {
	h := (*C.uintptr_t)(handlePtr)
	fn := cgo.Handle(*h).Value().(func(C.LbSyncProgress))
	fn(info)
}

func (l *lbCoreFFI) SyncAll(fn func(SyncProgress)) error {
	handle := cgo.NewHandle(func(cSP C.LbSyncProgress) {
		defer C.lb_sync_progress_free(cSP)
		if fn == nil {
			return
		}
		fn(SyncProgress{
			Total:    uint64(cSP.total),
			Progress: uint64(cSP.progress),
			CurrentWorkUnit: ClientWorkUnit{
				PullMetadata: bool(cSP.current_wu.pull_meta),
				PushMetadata: bool(cSP.current_wu.push_meta),
				PullDocument: C.GoString(cSP.current_wu.pull_doc),
				PushDocument: C.GoString(cSP.current_wu.push_doc),
			},
		})
	})
	defer handle.Delete()
	h := C.uintptr_t(handle)
	e := C.lb_sync_all(l.ref, C.LbSyncProgressCallback(C.go_sync_callback), unsafe.Pointer(&h))
	defer C.lb_error_free(e)
	return newErrorFromC(e)
}

func (l *lbCoreFFI) ShareFile(id FileID, uname string, mode ShareMode) error {
	cUname := C.CString(uname)
	defer C.free(unsafe.Pointer(cUname))
	cMode := C.LB_SHARE_MODE_READ
	if mode == ShareModeWrite {
		cMode = C.LB_SHARE_MODE_WRITE
	}
	e := C.lb_share_file(l.ref, cFileID(id), cUname, uint32(cMode))
	defer C.lb_error_free(e)
	return newErrorFromC(e)
}

func (l *lbCoreFFI) GetPendingShares() ([]File, error) {
	r := C.lb_get_pending_shares(l.ref)
	defer C.lb_file_list_result_free(r)
	return goFileListResult(r)
}

func (l *lbCoreFFI) DeletePendingShare(id FileID) error {
	e := C.lb_delete_pending_share(l.ref, cFileID(id))
	defer C.lb_error_free(e)
	return newErrorFromC(e)
}

func (l *lbCoreFFI) GetSubscriptionInfo() (SubscriptionInfo, error) {
	r := C.lb_get_subscription_info(l.ref)
	defer C.lb_sub_info_result_free(r)

	if r.err.code != 0 {
		return SubscriptionInfo{}, newErrorFromC(r.err)
	}
	info := SubscriptionInfo{}
	if r.period_end != 0 {
		info.PeriodEnd = time.UnixMilli(int64(r.period_end))
	}
	if stripeLast4 := C.GoString(r.stripe_last4); stripeLast4 != "" {
		info.StripeLast4 = stripeLast4
	}
	if r.google_play_state != 0 {
		info.GooglePlay = GooglePlayAccountState(r.google_play_state)
	}
	if r.app_store_state != 0 {
		info.AppStore = AppStoreAccountState(r.app_store_state)
	}
	return info, nil
}

func (l *lbCoreFFI) UpgradeViaStripe(card *CreditCard) error {
	if card == nil {
		e := C.lb_upgrade_account_stripe_old_card(l.ref)
		defer C.lb_error_free(e)
		return newErrorFromC(e)
	}
	cNumber := C.CString(card.Number)
	defer C.free(unsafe.Pointer(cNumber))
	cCVC := C.CString(card.CVC)
	defer C.free(unsafe.Pointer(cCVC))
	e := C.lb_upgrade_account_stripe_new_card(l.ref, cNumber, C.int(card.ExpiryYear), C.int(card.ExpiryMonth), cCVC)
	defer C.lb_error_free(e)
	return newErrorFromC(e)
}

func (l *lbCoreFFI) CancelSubscription() error {
	e := C.lb_cancel_subscription(l.ref)
	defer C.lb_error_free(e)
	return newErrorFromC(e)
}

func (l *lbCoreFFI) Validate() ([]string, error) {
	r := C.lb_validate(l.ref)
	defer C.lb_validate_result_free(r)
	if r.err.code != 0 {
		return nil, newErrorFromC(r.err)
	}
	warnings := make([]string, int(r.n_warnings))
	var i C.size_t
	for i = 0; i < r.n_warnings; i++ {
		msg := C.lb_validate_result_index(r, i)
		warnings[i] = C.GoString(msg)
	}
	return warnings, nil
}

func DefaultAPILocation() string {
	return C.GoString((*C.char)(unsafe.Pointer(&C.C_DEFAULT_API_LOCATION[0])))
}

func newErrorFromC(e C.LbError) error {
	if e.code == 0 {
		return nil
	}
	return &Error{
		Code: ErrorCode(e.code),
		Msg:  C.GoString(e.msg),
	}
}

func cFileID(v FileID) (r C.LbFileId) {
	var i C.size_t
	for i = 0; i < 16; i++ {
		r.data[i] = C.uint8_t(v[i])
	}
	return
}

func goFileID(v [16]C.uint8_t) (r FileID) {
	var i C.size_t
	for i = 0; i < 16; i++ {
		r[i] = byte(v[i])
	}
	return
}

func newFileFromC(f *C.LbFile) File {
	shares := make([]Share, int(f.shares.count))
	var i C.size_t
	for i = 0; i < f.shares.count; i++ {
		sh := C.lb_share_list_index(f.shares, i)
		shares[i] = Share{
			SharedBy:   C.GoString(sh.by),
			SharedWith: C.GoString(sh.with),
			Mode:       ShareMode(sh.mode),
		}
	}
	var ft FileType
	switch {
	case f.typ.tag == C.LB_FILE_TYPE_TAG_DOCUMENT:
		ft = FileTypeDocument{}
	case f.typ.tag == C.LB_FILE_TYPE_TAG_FOLDER:
		ft = FileTypeFolder{}
	case f.typ.tag == C.LB_FILE_TYPE_TAG_LINK:
		target := FileID{}
		var i C.size_t
		for i = 0; i < 16; i++ {
			target[i] = byte(f.typ.link_target[i])
		}
		ft = FileTypeLink{Target: target}
	}
	return File{
		ID:        goFileID(f.id),
		Parent:    goFileID(f.parent),
		Name:      C.GoString(f.name),
		Type:      ft,
		Lastmod:   time.UnixMilli(int64(f.lastmod)),
		LastmodBy: C.GoString(f.lastmod_by),
		Shares:    shares,
	}
}

func goAccountResult(r C.LbAccountResult) (Account, error) {
	if r.err.code != 0 {
		return Account{}, newErrorFromC(r.err)
	}
	return Account{
		Username: C.GoString(r.ok.username),
		APIURL:   C.GoString(r.ok.api_url),
	}, nil
}

func goBytesResult(r C.LbBytesResult) ([]byte, error) {
	if r.err.code != 0 {
		return nil, newErrorFromC(r.err)
	}
	return C.GoBytes(unsafe.Pointer(r.bytes), C.int(r.count)), nil
}

func goFileResult(r C.LbFileResult) (File, error) {
	if r.err.code != 0 {
		return File{}, newErrorFromC(r.err)
	}
	return newFileFromC(&r.ok), nil
}

func goFileListResult(r C.LbFileListResult) ([]File, error) {
	if err := newErrorFromC(r.err); err != nil {
		return nil, err
	}
	files := make([]File, int(r.ok.count))
	var i C.size_t
	for i = 0; i < r.ok.count; i++ {
		f := C.lb_file_list_index(r.ok, i)
		files[i] = newFileFromC(f)
	}
	return files, nil
}

func goStringResult(r C.LbStringResult) (string, error) {
	if r.err.code != 0 {
		return "", newErrorFromC(r.err)
	}
	return C.GoString(r.ok), nil
}

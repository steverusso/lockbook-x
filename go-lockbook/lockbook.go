package lockbook

import (
	"fmt"
	"strconv"
	"time"
)

type Core interface {
	GetAccount() (Account, error)
	CreateAccount(uname string, welcome bool) (Account, error)
	ImportAccount(acctStr string) (Account, error)
	ExportAccount() (string, error)

	FileByID(id string) (File, error)
	FileByPath(lbPath string) (File, error)
	GetRoot() (File, error)
	GetChildren(id string) ([]File, error)
	GetAndGetChildrenRecursively(id string) ([]File, error)
	ListMetadatas() ([]File, error)
	PathByID(id string) (string, error)

	ReadDocument(id string) ([]byte, error)
	WriteDocument(id string, data []byte) error

	CreateFile(name, parentID string, typ FileType) (File, error)
	CreateFileAtPath(lbPath string) (File, error)
	DeleteFile(id string) error
	RenameFile(id string, newName string) error
	MoveFile(srcID, destID string) error

	ExportFile(id, dest string, fn func(ImportExportFileInfo)) error
	ExportDrawing(id string, imgFmt ImageFormat) ([]byte, error)

	GetLastSyncedHumanString() (string, error)
	GetUsage() (UsageMetrics, error)
	GetUncompressedUsage() (UsageItemMetric, error)
	CalculateWork() (WorkCalculated, error)
	SyncAll(fn func(SyncProgress)) error

	ShareFile(id, uname string, mode ShareMode) error
	GetPendingShares() ([]File, error)
	DeletePendingShare(id string) error

	GetSubscriptionInfo() (SubscriptionInfo, error)
	UpgradeViaStripe(card *CreditCard) error
	CancelSubscription() error

	Validate() ([]string, error)
}

func NewCore(fpath string) (Core, error) {
	return initLbCoreFFI(fpath)
}

type ErrorCode uint32

const (
	CodeEmpty ErrorCode = iota
	CodeUnexpected
	CodeAccountExistsAlready
	CodeAccountDoesNotExist
	CodeAccountStringCorrupted
	CodeClientUpdateRequired
	CodeCouldNotReachServer
	CodeFileExists
	CodeFileIsNotDocument
	CodeFileIsNotFolder
	CodeFileNameContainsSlash
	CodeFileNameEmpty
	CodeFileNameUnavailable
	CodeFileNotFound
	CodeFolderMovedIntoItself
	CodeInsufficientPermission
	CodeInvalidDrawing
	CodeLinkInSharedFolder
	CodeNoAccount
	CodeNoRoot
	CodeNoRootOps
	CodePathContainsEmptyFile
	CodeTargetParentNotFound
	CodeUsernameInvalid
	CodeUsernamePubKeyMismatch
	CodeUsernameTaken
	CodeServerDisabled

	CodeNotPremium
	CodeSubscriptionAlreadyCanceled
	CodeUsageIsOverFreeTierDataCap
	CodeExistingRequestPending
	CodeCannotCancelForAppStore
)

type Error struct {
	Code ErrorCode
	Msg  string
}

func (e *Error) Error() string {
	return e.Msg
}

type Account struct {
	Username string
	APIURL   string
}

type File struct {
	ID        string
	Parent    string
	Name      string
	Type      FileType
	Lastmod   int64
	LastmodBy string
	Shares    []Share
}

func (f *File) IsDir() bool {
	_, ok := f.Type.(FileTypeFolder)
	return ok
}

func (f *File) IsRoot() bool {
	return f.ID == f.Parent
}

type (
	FileType interface{ implsFileType() }

	FileTypeDocument struct{}
	FileTypeFolder   struct{}
	FileTypeLink     struct{ Target string }
)

func (_ FileTypeDocument) implsFileType() {}
func (_ FileTypeFolder) implsFileType()   {}
func (_ FileTypeLink) implsFileType()     {}

func FileTypeString(t FileType) string {
	switch t := t.(type) {
	case FileTypeDocument:
		return "Document"
	case FileTypeFolder:
		return "Folder"
	case FileTypeLink:
		return "Link('" + t.Target + "')"
	default:
		return fmt.Sprintf("FileType(%v)", t)
	}
}

type Share struct {
	Mode       ShareMode
	SharedBy   string
	SharedWith string
}

type ShareMode int

const (
	ShareModeRead ShareMode = iota
	ShareModeWrite
)

func (s ShareMode) String() string {
	switch s {
	case ShareModeRead:
		return "Read"
	case ShareModeWrite:
		return "Write"
	default:
		return "ShareMode(" + strconv.FormatInt(int64(s), 10) + ")"
	}
}

type WorkCalculated struct {
	LastServerUpdateAt uint64
	WorkUnits          []WorkUnit
}

type WorkUnit struct {
	Type WorkUnitType
	File File
}

type WorkUnitType int

const (
	WorkUnitTypeLocal WorkUnitType = iota
	WorkUnitTypeServer
)

type SyncProgress struct {
	Total           uint64
	Progress        uint64
	CurrentWorkUnit ClientWorkUnit
}

type ClientWorkUnit struct {
	PullMetadata bool
	PushMetadata bool
	PullDocument string
	PushDocument string
}

type UsageMetrics struct {
	Usages      []FileUsage
	ServerUsage UsageItemMetric
	DataCap     UsageItemMetric
}

type UsageItemMetric struct {
	Exact    uint64
	Readable string
}

type FileUsage struct {
	FileID    string
	SizeBytes uint64
}

type ImportExportFileInfo struct {
	DiskPath string
	LbPath   string
}

type ImageFormat int

const (
	ImgFmtPNG ImageFormat = iota
	ImgFmtJPEG
	ImgFmtPNM
	ImgFmtTGA
	ImgFmtFarbfeld
	ImgFmtBMP
)

type CreditCard struct {
	Number      string
	ExpiryYear  int
	ExpiryMonth int
	CVC         string
}

type SubscriptionInfo struct {
	StripeLast4 string
	GooglePlay  GooglePlayAccountState
	AppStore    AppStoreAccountState
	PeriodEnd   time.Time
}

type StripeInfo struct {
	Last4 string
}

type GooglePlayAccountState int

const (
	GooglePlayNone GooglePlayAccountState = iota
	GooglePlayOk
	GooglePlayCanceled
	GooglePlayGracePeriod
	GooglePlayOnHold
)

func (s GooglePlayAccountState) String() string {
	switch s {
	case GooglePlayNone:
		return "None"
	case GooglePlayOk:
		return "Ok"
	case GooglePlayCanceled:
		return "Canceled"
	case GooglePlayGracePeriod:
		return "Grace Period"
	case GooglePlayOnHold:
		return "On Hold"
	default:
		return "GooglePlayAccountState(" + strconv.FormatInt(int64(s), 10) + ")"
	}
}

type AppStoreAccountState int

const (
	AppStoreNone AppStoreAccountState = iota
	AppStoreOk
	AppStoreGracePeriod
	AppStoreFailedToRenew
	AppStoreExpired
)

func (s AppStoreAccountState) String() string {
	switch s {
	case AppStoreNone:
		return "None"
	case AppStoreOk:
		return "Ok"
	case AppStoreGracePeriod:
		return "Grace Period"
	case AppStoreFailedToRenew:
		return "Failed To Renew"
	case AppStoreExpired:
		return "Expired"
	default:
		return "AppStoreAccountState(" + strconv.FormatInt(int64(s), 10) + ")"
	}
}

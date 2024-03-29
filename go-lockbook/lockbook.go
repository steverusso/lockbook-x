package lockbook

import (
	"fmt"
	"sort"
	"strconv"
	"time"

	"github.com/gofrs/uuid"
)

type Core interface {
	WriteablePath() string

	GetAccount() (Account, error)
	CreateAccount(uname, apiURL string, welcome bool) (Account, error)
	ImportAccount(acctStr string) (Account, error)
	ExportAccount() (string, error)

	FileByID(id FileID) (File, error)
	FileByPath(lbPath string) (File, error)
	GetRoot() (File, error)
	GetChildren(id FileID) ([]File, error)
	GetAndGetChildrenRecursively(id FileID) ([]File, error)
	ListMetadatas() ([]File, error)
	PathByID(id FileID) (string, error)

	ReadDocument(id FileID) ([]byte, error)
	WriteDocument(id FileID, data []byte) error

	CreateFile(name string, parentID FileID, typ FileType) (File, error)
	CreateFileAtPath(lbPath string) (File, error)
	DeleteFile(id FileID) error
	RenameFile(id FileID, newName string) error
	MoveFile(srcID, destID FileID) error

	ImportFile(src string, dest FileID, fn func(ImportFileInfo)) error
	ExportFile(id FileID, dest string, fn func(ExportFileInfo)) error
	ExportDrawing(id FileID, imgFmt ImageFormat) ([]byte, error)
	ExportDrawingToDisk(id FileID, imgFmt ImageFormat, dest string) error

	GetLastSynced() (time.Time, error)
	GetLastSyncedHumanString() (string, error)
	GetUsage() (UsageMetrics, error)
	GetUncompressedUsage() (UsageItemMetric, error)
	CalculateWork() (WorkCalculated, error)
	SyncAll(fn func(SyncProgress)) error

	ShareFile(id FileID, uname string, mode ShareMode) error
	GetPendingShares() ([]File, error)
	DeletePendingShare(id FileID) error

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
	CodeSuccess ErrorCode = iota
	CodeUnexpected
	CodeAccountExists
	CodeAccountNonexistent
	CodeAccountStringCorrupted
	CodeAlreadyCanceled
	CodeAlreadyPremium
	CodeAppStoreAccountAlreadyLinked
	CodeCannotCancelSubscriptionForAppStore
	CodeCardDecline
	CodeCardExpired
	CodeCardInsufficientFunds
	CodeCardInvalidCvc
	CodeCardInvalidExpMonth
	CodeCardInvalidExpYear
	CodeCardInvalidNumber
	CodeCardNotSupported
	CodeClientUpdateRequired
	CodeCurrentUsageIsMoreThanNewTier
	CodeDiskPathInvalid
	CodeDiskPathTaken
	CodeDrawingInvalid
	CodeExistingRequestPending
	CodeFileNameContainsSlash
	CodeFileNameEmpty
	CodeFileNonexistent
	CodeFileNotDocument
	CodeFileNotFolder
	CodeFileParentNonexistent
	CodeFolderMovedIntoSelf
	CodeInsufficientPermission
	CodeInvalidPurchaseToken
	CodeInvalidAuthDetails
	CodeLinkInSharedFolder
	CodeLinkTargetIsOwned
	CodeLinkTargetNonexistent
	CodeMultipleLinksToSameFile
	CodeNotPremium
	CodeOldCardDoesNotExist
	CodePathContainsEmptyFileName
	CodePathTaken
	CodeRootModificationInvalid
	CodeRootNonexistent
	CodeServerDisabled
	CodeServerUnreachable
	CodeShareAlreadyExists
	CodeShareNonexistent
	CodeTryAgain
	CodeUsageIsOverFreeTierDataCap
	CodeUsernameInvalid
	CodeUsernameNotFound
	CodeUsernamePublicKeyMismatch
	CodeUsernameTaken
)

type Error struct {
	Code  ErrorCode
	Msg   string
	Trace string
}

func (e *Error) Error() string {
	return e.Msg
}

type Account struct {
	Username string
	APIURL   string
}

type FileID = uuid.UUID

type File struct {
	ID        FileID
	Parent    FileID
	Name      string
	Type      FileType
	Lastmod   time.Time
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
	FileTypeLink     struct{ Target FileID }
)

func (FileTypeDocument) implsFileType() {}
func (FileTypeFolder) implsFileType()   {}
func (FileTypeLink) implsFileType()     {}

func FileTypeString(t FileType) string {
	switch t := t.(type) {
	case FileTypeDocument:
		return "Document"
	case FileTypeFolder:
		return "Folder"
	case FileTypeLink:
		return "Link('" + t.Target.String() + "')"
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

func SortFiles(files []File) {
	sort.SliceStable(files, func(i, j int) bool {
		a, b := files[i], files[j]
		if a.IsDir() == b.IsDir() {
			return a.Name < b.Name
		}
		return a.IsDir()
	})
}

type WorkCalculated struct {
	LastServerUpdateAt uint64
	WorkUnits          []WorkUnit
}

type WorkUnit struct {
	Type WorkUnitType
	ID   FileID
}

type WorkUnitType int

const (
	WorkUnitTypeLocal WorkUnitType = iota
	WorkUnitTypeServer
)

// SyncProgress is the data sent (via closure) at certain stages of sync.
type SyncProgress struct {
	Total    uint64
	Progress uint64
	Msg      string
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
	FileID    FileID
	SizeBytes uint64
}

// ImportFileInfo is the data sent (via closure) at certain stages of file import. The
// stage and type of information is determined by the zero value of each field. A non-zero
// `Total` means a "total calculated" update. A non-empty `DiskPath` means a "file
// started" update. A non-nil `FileDone` means a "file finished" update.
type ImportFileInfo struct {
	Total    int
	DiskPath string
	FileDone *File
}

type ExportFileInfo struct {
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

func MaybeFileByPath(core Core, p string) (File, bool, error) {
	f, err := core.FileByPath(p)
	if err != nil {
		if err, ok := err.(*Error); ok && err.Code == CodeFileNonexistent {
			return File{}, false, nil
		}
		return File{}, false, err
	}
	return f, true, nil
}

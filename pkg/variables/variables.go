package variables

// Server Errors
const (
	JsonPackFailedError     = "Failed to marshal JSON object"
	ResponseSendFailedError = "Failed to send response to client"
	ListenAndServeError     = "Failed to listen and serve"
)

// Authorization Errors
const (
	UserNotAuthorized = "User not authorized"
)

// API Messages
const (
	StatusMethodNotAllowedError = "Method not allowed"
	StatusBadRequestError       = "Bad request"
	StatusInternalServerError   = "Internal server error"
	StatusUnauthorizedError     = "Unauthorized"
	SessionCreateError          = "Session create failed"
	StatusOkMessage             = "Succesful response"
	SessionKilledError          = "Session killed failed"
	SessionNotFoundError        = "Session not found"
	UserAlreadyExistsError      = "User already exists"
	StatusForbiddenError        = "Forbidden"
	GrpcListenAndServeError     = "Failed grpc to listen and serve"
	GrpcConnectError            = "Failed grpc to connect"
	InvalidImageError           = "Invalid image"
	ValidateStringError         = "Validate string error"
	FeatureIdError              = "invalid or missing 'feature_id' parameter"
	TagIdError                  = "invalid or missing 'tag_id' parameter"
	LastRevisionError           = "invalid value for 'use_last_revision' parameter"
	BannerNotFoundError         = "Banner not found"
	InvalidLimit                = "Limit must be a positive number"
	InvalidOffset               = "Offset must be non-negative"
)

// Middleware types
type (
	contextKey string
	sessionKey string
)

// Middleware keys constants
const (
	UserIDKey    contextKey = "userId"
	SessionIDKey sessionKey = "sessionId"
)

// Configs types
type (
	AppConfig struct {
		Address string `yaml:"address"`
	}

	CacheDataBaseConfig struct {
		Host     string `yaml:"host"`
		Password string `yaml:"password"`
		DbNumber int    `yaml:"db"`
		Timer    int    `yaml:"timer"`
	}

	RelationalDataBaseConfig struct {
		User         string `yaml:"user"`
		DbName       string `yaml:"dbname"`
		Password     string `yaml:"password"`
		Host         string `yaml:"host"`
		Port         int    `yaml:"port"`
		Sslmode      string `yaml:"sslmode"`
		MaxOpenConns int    `yaml:"max_open_conns"`
		Timer        uint32 `yaml:"timer"`
	}

	GrpcConfig struct {
		Address        string `yaml:"address"`
		Port           string `yaml:"port"`
		ConnectionType string `yaml:"connection_type"`
	}
)

// Cookies data
const (
	SessionCookieName = "session_id"
	HttpOnly          = true
)

// Repository messages
const (
	AuthorizationCachePingRetryError      = "Authorization cache: ping failed"
	AuthorizationCachePingMaxRetriesError = "Authorization cache: ping error. Maximum number of retries reached"
	SessionRemoveError                    = "Delete session request could not be completed:"
	SqlOpenError                          = "Open SQL connection failed:"
	SqlPingError                          = "Ping SQL connection failed:"
	SqlMaxPingRetriesError                = "Maximum number of retries reached:"
	SqlProfileCreateError                 = "Profile create failed:"
	FindProfileIdByLoginError             = "Find profile id by login failed:"
	ProfileIdNotFoundByLoginError         = "Profile id not found:"
	ProfileRoleNotFoundByLoginError       = "Profile role not found:"
)

// Repository constants
const (
	MaxRetries  = 5
	UserRoleId  = 1
	AdminRoleId = 2
	PageSize    = 10
)

// Core Messages
const (
	InvalidLoginOrPasswordError     = "Invalid email or password"
	SessionRepositoryNotActiveError = "Session repository not active"
	ProfileRepositoryNotActiveError = "Profile repository not active"
	CreateProfileError              = "Create profile failed"
	ProfileNotFoundError            = "Profile not found"
	GetProfileError                 = "Get profile failed"
	GetProfileRoleError             = "Get profile role failed"
	GrpcRecievError                 = "gRPC recieve error"
)

// Core variables
var (
	LetterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
)

// Logger constants
const (
	ModuleLogger     = "Module"
	CoreModuleLogger = "CoreModuleLogger"
)

// Main messages
const (
	ReadAuthConfigError      = "Read auth config failed"
	ReadAuthSqlConfigError   = "Read auth sql config failed"
	ReadAuthCacheConfigError = "Read auth cache config failed"
	ReadGrpcConfigError      = "Grpc config file error"
	CoreInitializeError      = "Core initialize failed"
)

// Regexp
const (
	LoginRegexp = `^[a-zA-Z0-9]+$`
)

// Roles
const (
	AdminRole = "admin"
)

// Query params
const (
	PaginationPageNumber = "page"
	PaginationPageSize   = "page_size"
)

// Validate params
var (
	ValidImageTypes    = []string{"image/jpeg", "image/png", "image/gif"}
	MaxTitleSize       = 150
	MinTitleSize       = 5
	MaxDescriptionSize = 900
	MinDescriptionSize = 5
)

package auth

type AccessMode uint8

func (a AccessMode) Value() uint8 {
	return uint8(a)
}

const (
	Delete AccessMode = 1 << iota
	Edit
	Read
	Add

	NONE AccessMode = 0
	FULL            = Add | Read | Edit | Delete
)

type SuperPermission string

func (a SuperPermission) Value() string {
	return string(a)
}

const (
	SuperSelectPermission SuperPermission = "read"
	SuperEditPermission   SuperPermission = "write"
	SuperCreatePermission SuperPermission = "add"
	SuperDeletePermission SuperPermission = "delete"
)

type Session interface {
	GetRoles() []int64
	GetUserId() int64
	GetUserIp() string
	GetDomainId() int64
	GetPermissions() []string
	GetObjectScope(string) ObjectScoper
	GetAllObjectScopes() []ObjectScoper
	CheckLicenseAccess(string) bool
	CheckObacAccess(string, AccessMode) bool
	IsRbacCheckRequired(string, AccessMode) bool
	HasSuperPermission(permission SuperPermission) bool

	GetMainAccessMode() AccessMode
	GetMainObjClassName() string
}

type ObjectScoper interface {
	IsRbacUsed() bool
	IsObacUsed() bool
	GetAccess() string
	GetObjectName() string
}

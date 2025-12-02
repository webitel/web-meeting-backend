package user

import (
	"github.com/webitel/meetings/infra/auth"
	"strings"
	"time"
)

type UserAuthSession struct {
	User             *User
	Permissions      []string
	Scopes           map[string]*Scope
	License          map[string]bool
	Roles            []*Role
	DomainId         int64
	ExpiresAt        int64
	SuperCreate      bool
	SuperEdit        bool
	SuperDelete      bool
	SuperSelect      bool
	MainAccess       auth.AccessMode
	MainObjClassName string
	UserIp           string
}

// region Auther interface implementation

func (s *UserAuthSession) GetUserId() int64 {
	if s.User == nil || s.User.Id <= 0 {
		return 0
	}
	return s.User.Id
}

func (s *UserAuthSession) GetUserIp() string {
	if s.UserIp == "" {
		return "unknown"
	}
	return s.UserIp
}

func (s *UserAuthSession) GetDomainId() int64 {
	return s.DomainId
}

func (s *UserAuthSession) GetRoles() []int64 {
	roles := []int64{s.GetUserId()}
	for _, role := range s.Roles {
		roles = append(
			roles,
			role.Id,
		)
	}
	return roles
}

func (s *UserAuthSession) GetObjectScope(sc string) auth.ObjectScoper {
	if sc == "" {
		return nil
	}
	scope, found := s.Scopes[sc]
	if !found {
		return nil
	}
	return scope
}

func (s *UserAuthSession) GetAllObjectScopes() []auth.ObjectScoper {
	var res []auth.ObjectScoper
	for _, scope := range s.Scopes {
		res = append(res, scope)
	}
	return res
}

func (s *UserAuthSession) GetPermissions() []string {
	return s.Permissions
}

func (s *UserAuthSession) CheckLicenseAccess(name string) bool {
	if legit, found := s.License[name]; found {
		return legit
	}
	return false
}

func (s *UserAuthSession) GetMainAccessMode() auth.AccessMode {
	return s.MainAccess
}

func (s *UserAuthSession) GetMainObjClassName() string {
	return s.MainObjClassName
}

func (s *UserAuthSession) CheckObacAccess(scopeName string, accessType auth.AccessMode) bool {
	scope := s.GetObjectScope(scopeName)
	if scope == nil {
		return false
	}

	if scope.IsObacUsed() {
		var (
			bypass  bool
			require string
		)

		switch accessType {
		case auth.Delete, auth.Read | auth.Delete:
			require, bypass = "d", s.SuperDelete
		case auth.Edit, auth.Read | auth.Edit:
			require, bypass = "w", s.SuperEdit
		case auth.Read, auth.NONE:
			require, bypass = "r", s.SuperSelect
		case auth.Add, auth.Read | auth.Add:
			require, bypass = "x", s.SuperCreate
		}
		if bypass {
			return true
		}
		for i := len(require) - 1; i >= 0; i-- {
			mode := require[i]
			if strings.IndexByte(scope.GetAccess(), mode) < 0 {
				return false
			}
		}
	}

	return true
}

func (s *UserAuthSession) IsRbacCheckRequired(scopeName string, accessType auth.AccessMode) bool {
	scope := s.GetObjectScope(scopeName)
	if scope == nil {
		return false
	}
	rbacEnabled := scope.IsRbacUsed()
	if rbacEnabled {
		var bypass bool

		switch accessType {
		case auth.Delete, auth.Read | auth.Delete:
			bypass = s.SuperDelete
		case auth.Edit, auth.Read | auth.Edit:
			bypass = s.SuperEdit
		case auth.Read, auth.NONE:
			bypass = s.SuperSelect
		case auth.Add, auth.Read | auth.Add:
			bypass = s.SuperCreate
		}
		if bypass {
			return false
		}
	}
	return rbacEnabled

}

// endregion

func (s *UserAuthSession) IsExpired() bool {
	return time.Now().Unix() > s.ExpiresAt
}

func (s *UserAuthSession) HasSuperPermission(permission auth.SuperPermission) bool {
	switch permission {
	case auth.SuperCreatePermission:
		return s.SuperCreate
	case auth.SuperDeletePermission:
		return s.SuperDelete
	case auth.SuperEditPermission:

		return s.SuperEdit
	case auth.SuperSelectPermission:
		return s.SuperSelect
	}
	return false
}

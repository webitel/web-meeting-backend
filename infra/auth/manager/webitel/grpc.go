package webitel

import (
	"context"
	"net"
	"strings"
	"time"

	"golang.org/x/sync/singleflight"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/peer"

	authclient "buf.build/gen/go/webitel/webitel-go/grpc/go/_gogrpc"
	authmodel "buf.build/gen/go/webitel/webitel-go/protocolbuffers/go"

	"github.com/webitel/web-meeting-backend/infra/auth"
	session "github.com/webitel/web-meeting-backend/infra/auth/session/user"
	"github.com/webitel/web-meeting-backend/internal/errors"
)

var _ auth.Manager = &Manager{}

type Manager struct {
	Client     authclient.AuthClient
	Group      singleflight.Group
	Connection *grpc.ClientConn
}

func New(conn *grpc.ClientConn) (*Manager, error) {
	return &Manager{Client: authclient.NewAuthClient(conn), Group: singleflight.Group{}, Connection: conn}, nil
}

func (i *Manager) AuthorizeFromContext(ctx context.Context, mainObjClassName string, mainAccessMode auth.AccessMode) (auth.Session, error) {
	var token []string

	var info metadata.MD

	var ok bool

	v := ctx.Value(session.RequestContextName)
	info, ok = v.(metadata.MD)

	if !ok {
		info, ok = metadata.FromIncomingContext(ctx)
	}

	if !ok {
		return nil, errors.Forbidden("internal.grpc.get_context: Not found")
	} else {
		token = info.Get(session.AuthTokenName)
	}

	newContext := metadata.NewOutgoingContext(ctx, info)

	if len(token) < 1 {
		return nil, errors.Internal("webitel_manager.authorize_from_from_context.search_token.not_found: token not found")
	}

	userToken := token[0]

	sess, err, _ := i.Group.Do(userToken, func() (any, error) {
		return i.Client.UserInfo(newContext, nil)
	})
	if err != nil {
		return nil, errors.Internal("webitel_manager.authorize_from_from_context.user_info.err", errors.WithCause(err))
	}
	return ConstructSessionFromUserInfo(sess.(*authmodel.Userinfo), mainObjClassName, mainAccessMode, getClientIP(ctx)), nil
}

func ConstructSessionFromUserInfo(userinfo *authmodel.Userinfo, mainObjClass string, mainAccess auth.AccessMode, ip string) *session.UserAuthSession {
	sess := &session.UserAuthSession{
		User: &session.User{
			Id:        userinfo.UserId,
			Name:      userinfo.Name,
			Username:  userinfo.Username,
			Extension: userinfo.Extension,
		},
		ExpiresAt:        userinfo.ExpiresAt,
		DomainId:         userinfo.Dc,
		Permissions:      make([]string, 0),
		License:          map[string]bool{},
		Scopes:           map[string]*session.Scope{},
		MainAccess:       mainAccess,
		MainObjClassName: mainObjClass,
		UserIp:           ip,
	}
	for _, lic := range userinfo.License {
		sess.License[lic.Id] = lic.ExpiresAt > time.Now().UnixMilli()
	}
	for _, permission := range userinfo.Permissions {
		switch auth.SuperPermission(permission.GetId()) {
		case auth.SuperCreatePermission:
			sess.SuperCreate = true
		case auth.SuperDeletePermission:
			sess.SuperDelete = true
		case auth.SuperEditPermission:
			sess.SuperEdit = true
		case auth.SuperSelectPermission:
			sess.SuperSelect = true
		}
		sess.Permissions = append(sess.Permissions, permission.GetId())
	}
	for _, scope := range userinfo.Scope {
		sess.Scopes[scope.Class] = &session.Scope{
			Id:     scope.GetId(),
			Name:   scope.GetName(),
			Abac:   scope.Abac,
			Obac:   scope.Obac,
			Rbac:   scope.Rbac,
			Class:  scope.Class,
			Access: scope.Access,
		}
	}

	for i, role := range userinfo.Roles {
		if i == 0 {
			sess.Roles = make([]*session.Role, 0)
		}

		sess.Roles = append(sess.Roles, &session.Role{
			Id:   role.GetId(),
			Name: role.GetName(),
		})
	}

	return sess
}

func getClientIP(ctx context.Context) string {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return ""
	}

	// First try to get IP from headers
	ip := strings.Join(md.Get("x-real-ip"), ",")
	if ip == "" {
		ip = strings.Join(md.Get("x-forwarded-for"), ",")
	}

	// If no IP from headers, try to get from peer
	if ip == "" {
		if p, ok := peer.FromContext(ctx); ok {
			if addr, ok := p.Addr.(*net.TCPAddr); ok {
				ip = addr.IP.String()
			} else if addr, ok := p.Addr.(*net.UDPAddr); ok {
				ip = addr.IP.String()
			}
		}
	}

	return ip
}

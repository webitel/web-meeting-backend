package interceptor

import (
	"context"
	"regexp"
	"strings"

	api "github.com/webitel/web-meeting-backend/gen/go/api/meetings"
	"github.com/webitel/web-meeting-backend/infra/auth"
	errors "github.com/webitel/web-meeting-backend/internal/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
)

// Define a header constant for the token
const (
	hdrTokenAccess = "X-Webitel-Access"
	SessionHeader  = "session"
)

// Regular expression to parse gRPC method information
var reg = regexp.MustCompile(`^(.*\.)`)

// AuthUnaryServerInterceptor authenticates and authorizes unary RPCs.
func AuthUnaryServerInterceptor(authManager auth.Manager) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		// Retrieve authorization details
		objClass, licenses, action := objClassWithAction(info)

		// Authorize session with the token
		session, err := authManager.AuthorizeFromContext(ctx, objClass, action)
		if err != nil {
			return nil, errors.New(
				"unauthorized",
				errors.WithCause(err),
				errors.WithCode(codes.Unauthenticated),
				errors.WithID("auth.interceptor.unauthorized"),
			)
		}
		//  License validation
		if missingLicenses := checkLicenses(session, licenses); len(missingLicenses) > 0 {
			return nil, errors.New(
				"permission denied",
				errors.WithCode(codes.PermissionDenied),
				errors.WithCause(errors.New("missing required licenses "+strings.Join(missingLicenses, ", "))),
				errors.WithID("auth.interceptor.license"),
			)
		}

		// Permission validation
		if ok := validateSessionPermission(session, objClass, action); !ok {
			return nil, errors.New(
				"permission denied",
				errors.WithCode(codes.PermissionDenied),
				errors.WithCause(errors.New("missing required permissions "+objClass)),
				errors.WithID("auth.interceptor.permission"),
			)
		}

		ctx = context.WithValue(ctx, SessionHeader, session)

		// Proceed with api_handler after successful validation
		resp, err := handler(ctx, req)
		if err != nil {
			return nil, err
		}

		return resp, nil
	}
}

// tokenFromContext extracts the authorization token from metadata.
func tokenFromContext(ctx context.Context) (string, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return "", errors.New(
			"Metadata is empty; authorization token required",
			errors.WithID("auth.metadata.missing"),
			errors.WithCode(codes.Unauthenticated),
		)
	}
	token := md.Get(hdrTokenAccess)
	if len(token) < 1 || token[0] == "" {
		return "", errors.New(
			"Authorization token is missing",
			errors.WithID("auth.token.missing"),
			errors.WithCode(codes.Unauthenticated),
		)
	}
	return token[0], nil
}

func objClassWithAction(info *grpc.UnaryServerInfo) (string, []string, auth.AccessMode) {
	serviceName, methodName := splitFullMethodName(info.FullMethod)
	service := api.WebitelAPI[serviceName]
	objClass := service.ObjClass
	licenses := service.AdditionalLicenses
	action := service.WebitelMethods[methodName].Access
	var accessMode auth.AccessMode
	switch action {
	case 0:
		accessMode = auth.Add
	case 1:
		accessMode = auth.Read
	case 2:
		accessMode = auth.Edit
	case 3:
		accessMode = auth.Delete
	}

	return objClass, licenses, accessMode
}

// checkLicenses verifies that the session has all required licenses.
func checkLicenses(session auth.Session, licenses []string) []string {
	var missing []string
	for _, license := range licenses {
		if !session.CheckLicenseAccess(license) {
			missing = append(missing, license)
		}
	}
	return missing
}

// validateSessionPermission checks if the session has the required permissions.
func validateSessionPermission(session auth.Session, objClass string, accessMode auth.AccessMode) bool {
	return session.CheckObacAccess(objClass, accessMode)
}

// splitFullMethodName extracts service and method names from the full gRPC method name.
func splitFullMethodName(fullMethod string) (string, string) {
	fullMethod = strings.TrimPrefix(fullMethod, "/")
	if i := strings.Index(fullMethod, "/"); i >= 0 {
		return reg.ReplaceAllString(fullMethod[:i], ""), fullMethod[i+1:]
	}
	return "unknown", "unknown"
}

package grpc_srv

import (
	"github.com/webitel/wlog"
	"google.golang.org/grpc"
	"net"
	"os"
	"strconv"
)

type Server struct {
	Addr string
	host string
	port int
	log  *wlog.Logger
	*grpc.Server
	listener net.Listener
}

// New provides a new gRPC server.
func New(addr string, log *wlog.Logger) (*Server, error) {

	s := grpc.NewServer(grpc.ChainUnaryInterceptor())

	l, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, err
	}

	h, p, err := net.SplitHostPort(l.Addr().String())
	if err != nil {
		return nil, err
	}
	port, _ := strconv.Atoi(p)

	if h == "::" {
		h = publicAddr()
	}

	return &Server{
		Addr:     addr,
		Server:   s,
		log:      log,
		host:     h,
		port:     port,
		listener: l,
	}, nil
}

func (s *Server) Listen() error {
	return s.Serve(s.listener)
}

func (s *Server) Shutdown() error {
	s.log.Debug("receive shutdown grpc")
	err := s.listener.Close()
	s.Server.GracefulStop()
	return err
}

func (s *Server) Host() string {
	if e, ok := os.LookupEnv("PROXY_GRPC_HOST"); ok {
		return e
	}
	return s.host
}

func (s *Server) Port() int {
	return s.port
}

func publicAddr() string {
	ifaces, err := net.Interfaces()

	if err != nil {
		return ""
	}
	for _, i := range ifaces {
		addrs, err := i.Addrs()
		if err != nil {
			continue
		}

		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			default:
				continue
			}

			if isPublicIP(ip) {
				return ip.String()
			}
			// process IP address
		}
	}
	return ""
}

func isPublicIP(IP net.IP) bool {
	if IP.IsLoopback() || IP.IsLinkLocalMulticast() || IP.IsLinkLocalUnicast() {
		return false
	}
	return true
}

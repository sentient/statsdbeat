package beater

import (
	"net"

	"github.com/elastic/beats/v7/libbeat/logp"
)

type HealthServer struct {
	addr string
	log  *logp.Logger
}

//NewHealthCheck returns a new server that responds to health checks with a status of Http.OK
func NewHealthCheck(address string, log *logp.Logger) *HealthServer {
	s := &HealthServer{
		addr: address,
		log:  log,
	}
	return s
}

//Serve sends a small responds to the client that tries to connect.
func (s *HealthServer) Serve() {
	s.log.Infof("health check listining at %s", s.addr)
	l, err := net.Listen("tcp", s.addr)
	if err != nil {
		s.log.Error("failed to setup health check")
		return
	}
	defer l.Close()
	for {
		// Listen for an incoming connection.
		conn, err := l.Accept()
		if err != nil {
			s.log.Errorf("failed to accept tcp connection. Error %v", err)
			return
		}
		s.log.Debug("Response ok on healthcheck")
		conn.Write([]byte("ok"))
		conn.Close()
	}

}

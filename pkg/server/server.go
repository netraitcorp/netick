package server

type Server interface {
	Options() *Options
}

/*
func (srv *Server) ListenAndServe() error {
	ln, err := net.Listen("tcp", srv.Addr)
	if err != nil {
		return err
	}
	return srv.serve(ln)
}

func (srv *Server) serve(ln net.Listener) error {
	var tempDelay time.Duration
	for {
		rw, err := ln.Accept()
		if err != nil {
			if ne, ok := err.(net.Error); ok && ne.Temporary() {
				if tempDelay == 0 {
					tempDelay = 5 * time.Millisecond
				} else {
					tempDelay *= 2
				}
				if max := 1 * time.Second; tempDelay > max {
					tempDelay = max
				}
				//log.Printf("[ERROR] tcp: Accept error: %v; retrying in %v", err, tempDelay)
				time.Sleep(tempDelay)
				continue
			}
			return err
		}
		tempDelay = 0

		c := srv.createConn(rw)
		c.Accept()
	}
}

func (srv *Server) createConn(rw net.Conn) *Conn {
	return NewConnection(rw)
}
*/

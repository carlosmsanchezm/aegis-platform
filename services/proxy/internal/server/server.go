package server

import (
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/websocket"
	"go.uber.org/zap"

	"github.com/yourorg/aegis/services/proxy/internal/jti"
	"github.com/yourorg/aegis/services/proxy/internal/jwtutil"
)

type ProxyServer struct {
	log      *zap.Logger
	cfg      Config
	verifier *jwtutil.Verifier
	jtiStore *jti.Store
	upgrader websocket.Upgrader
}

func NewProxyServer(log *zap.Logger, cfg Config) *ProxyServer {
	store := jti.New(cfg.TokenReuseTTL)
	verifier := jwtutil.NewVerifier(cfg.JWTSecret).WithAudience(cfg.ExpectedAudience)
	return &ProxyServer{
		log:      log,
		cfg:      cfg,
		verifier: verifier,
		jtiStore: store,
		upgrader: websocket.Upgrader{
			CheckOrigin: func(*http.Request) bool { return true },
		},
	}
}

func (s *ProxyServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	wid := extractWorkloadID(r.URL.Path)
	if wid == "" {
		http.NotFound(w, r)
		return
	}

	tokenStr, err := bearerToken(r.Header.Get("Authorization"))
	if err != nil {
		s.deny(w, r, wid, "missing_token", err)
		return
	}

	claims, err := s.verifier.Verify(tokenStr)
	if err != nil {
		s.deny(w, r, wid, "invalid_token", err)
		return
	}

	if claims.Wid != wid {
		s.deny(w, r, wid, "wid_mismatch", errors.New("token wid mismatch"))
		return
	}

	if s.cfg.Cluster != "" && claims.Cluster != "" && claims.Cluster != s.cfg.Cluster {
		s.deny(w, r, wid, "cluster_mismatch", errors.New("token cluster mismatch"))
		return
	}

	checkHost := claims.Dest
	if host, _, err := net.SplitHostPort(claims.Dest); err == nil {
		checkHost = host
	}
	if !strings.HasSuffix(checkHost, s.cfg.DestSuffix) {
		s.deny(w, r, wid, "dest_invalid", fmt.Errorf("dest must end with %s", s.cfg.DestSuffix))
		return
	}

	if !s.jtiStore.Use(claims.ID) {
		s.deny(w, r, wid, "token_reused", errors.New("token already used"))
		return
	}

	s.audit(claims, r, "auth_ok", "")

	switch {
	case r.Method == http.MethodConnect:
		s.handleConnect(w, r, claims)
	case isWebSocketRequest(r):
		s.handleWebSocket(w, r, claims)
	default:
		http.Error(w, "unsupported method", http.StatusMethodNotAllowed)
	}
}

func (s *ProxyServer) handleConnect(w http.ResponseWriter, r *http.Request, claims *jwtutil.Claims) {
	hijacker, ok := w.(http.Hijacker)
	if !ok {
		http.Error(w, "hijacking not supported", http.StatusInternalServerError)
		s.audit(claims, r, "deny", "no_hijack")
		return
	}
	clientConn, _, err := hijacker.Hijack()
	if err != nil {
		s.audit(claims, r, "deny", fmt.Sprintf("hijack_error:%v", err))
		return
	}

	_, err = io.WriteString(clientConn, "HTTP/1.1 200 Connection Established\r\n\r\n")
	if err != nil {
		clientConn.Close()
		s.audit(claims, r, "deny", fmt.Sprintf("write_resp_error:%v", err))
		return
	}

	s.tunnelConnections(clientConn, claims, r)
}

func (s *ProxyServer) handleWebSocket(w http.ResponseWriter, r *http.Request, claims *jwtutil.Claims) {
	conn, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		s.audit(claims, r, "deny", fmt.Sprintf("ws_upgrade:%v", err))
		return
	}
	s.tunnelWebSocket(conn, claims, r)
}

func (s *ProxyServer) tunnelConnections(clientConn net.Conn, claims *jwtutil.Claims, r *http.Request) {
	upstream, err := net.DialTimeout("tcp", claims.Dest, 10*time.Second)
	if err != nil {
		s.audit(claims, r, "deny", fmt.Sprintf("dial_error:%v", err))
		clientConn.Close()
		return
	}

	s.audit(claims, r, "allow", "tcp")

	go func() {
		_, _ = io.Copy(upstream, clientConn)
		upstream.Close()
	}()
	_, _ = io.Copy(clientConn, upstream)
	clientConn.Close()
}

func (s *ProxyServer) tunnelWebSocket(ws *websocket.Conn, claims *jwtutil.Claims, r *http.Request) {
	upstream, err := net.DialTimeout("tcp", claims.Dest, 10*time.Second)
	if err != nil {
		s.audit(claims, r, "deny", fmt.Sprintf("ws_dial_error:%v", err))
		ws.Close()
		return
	}

	s.audit(claims, r, "allow", "websocket")

	done := make(chan struct{})

	go func() {
		defer close(done)
		for {
			type_, message, err := ws.ReadMessage()
			if err != nil {
				return
			}
			if type_ != websocket.BinaryMessage {
				continue
			}
			if _, err := upstream.Write(message); err != nil {
				return
			}
		}
	}()

	buf := make([]byte, 32*1024)
	for {
		n, err := upstream.Read(buf)
		if n > 0 {
			if err := ws.WriteMessage(websocket.BinaryMessage, buf[:n]); err != nil {
				break
			}
		}
		if err != nil {
			break
		}
	}

	upstream.Close()
	ws.Close()
	<-done
}

func (s *ProxyServer) deny(w http.ResponseWriter, r *http.Request, wid, reason string, err error) {
	status := http.StatusForbidden
	if errors.Is(err, errBadTokenFormat) {
		status = http.StatusUnauthorized
	}
	http.Error(w, "access denied", status)

	s.log.Warn("proxy deny",
		zap.String("wid", wid),
		zap.String("reason", reason),
		zap.String("remote", r.RemoteAddr),
		zap.Error(err),
	)
}

func (s *ProxyServer) audit(claims *jwtutil.Claims, r *http.Request, decision, note string) {
	if claims == nil {
		return
	}
	s.log.Info("proxy decision",
		zap.String("decision", decision),
		zap.String("jti", claims.ID),
		zap.String("wid", claims.Wid),
		zap.String("sub", claims.Sub),
		zap.String("dest", claims.Dest),
		zap.String("remote", r.RemoteAddr),
		zap.String("note", note),
	)
}

var errBadTokenFormat = errors.New("authorization header must be Bearer token")

func bearerToken(header string) (string, error) {
	parts := strings.SplitN(header, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		return "", errBadTokenFormat
	}
	return strings.TrimSpace(parts[1]), nil
}

func extractWorkloadID(path string) string {
	if !strings.HasPrefix(path, "/proxy/") {
		return ""
	}
	wid := strings.TrimPrefix(path, "/proxy/")
	wid = strings.Trim(wid, "/")
	if wid == "" {
		return ""
	}
	if idx := strings.IndexByte(wid, '/'); idx >= 0 {
		wid = wid[:idx]
	}
	return wid
}

func isWebSocketRequest(r *http.Request) bool {
	return strings.EqualFold(r.Header.Get("Connection"), "Upgrade") &&
		strings.EqualFold(r.Header.Get("Upgrade"), "websocket")
}

func (s *ProxyServer) Start() error {
	server := &http.Server{
		Addr:         s.cfg.ListenAddr,
		Handler:      s,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
	}
	s.log.Info("proxy listening", zap.String("addr", s.cfg.ListenAddr))
	return server.ListenAndServe()
}

package server

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
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

type countingReader struct {
	reader io.Reader
	total  uint64
}

func (cr *countingReader) Read(p []byte) (int, error) {
	n, err := cr.reader.Read(p)
	cr.total += uint64(n)
	return n, err
}

type sessionError struct {
	mu  sync.Mutex
	err error
}

func (s *sessionError) set(err error) {
	if err == nil {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.err == nil {
		s.err = err
	}
}

func (s *sessionError) get() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.err
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
	// Enforce client certificate SAN suffix allow-list policy (AC-17a/b)
	if err := s.enforceClientCertSanSuffixPolicy(r); err != nil {
		s.log.Warn("proxy.audit", zap.String("event", "session.deny"), zap.String("remote", r.RemoteAddr), zap.String("reason", "client_cert_san_suffix_denied"), zap.Error(err))
		http.Error(w, "client certificate SAN suffix not allowed", http.StatusForbidden)
		return
	}

	wid := extractWorkloadID(r.URL.Path)
	if wid == "" {
		http.NotFound(w, r)
		return
	}

	tokenStr, err := bearerToken(r.Header.Get("Authorization"))
	if err != nil {
		s.deny(w, r, nil, wid, "missing_token", err)
		return
	}

	claims, err := s.verifier.Verify(tokenStr)
	if err != nil {
		s.deny(w, r, nil, wid, "invalid_token", err)
		return
	}

	if claims.Wid != wid {
		s.deny(w, r, claims, wid, "wid_mismatch", errors.New("token wid mismatch"))
		return
	}

	if s.cfg.Cluster != "" && claims.Cluster != "" && claims.Cluster != s.cfg.Cluster {
		s.deny(w, r, claims, wid, "cluster_mismatch", errors.New("token cluster mismatch"))
		return
	}
	destAddr, err := s.resolveDestination(claims)
	if err != nil {
		s.deny(w, r, claims, wid, "dest_invalid", err)
		return
	}

	if !s.jtiStore.Use(claims.ID) {
		s.deny(w, r, claims, wid, "token_reused", errors.New("token already used"))
		return
	}

	sessCert := s.extractClientCert(r)

	switch {
	case r.Method == http.MethodConnect:
		s.handleConnect(w, r, claims, destAddr, sessCert)
	case isWebSocketRequest(r):
		s.handleWebSocket(w, r, claims, destAddr, sessCert)
	default:
		s.auditEvent("session.deny", claims, true,
			zap.String("wid", wid),
			zap.String("remote", r.RemoteAddr),
			zap.String("reason", "method_not_allowed"),
			zap.String("method", r.Method),
		)
		http.Error(w, "unsupported method", http.StatusMethodNotAllowed)
	}
}

// enforceClientCertSanSuffixPolicy checks the client certificate SANs against the configured allow-list.
func (s *ProxyServer) enforceClientCertSanSuffixPolicy(r *http.Request) error {
	if len(s.cfg.ClientCertSanSuffixAllowList) == 0 {
		return nil // No policy enforced
	}
	connState := getTLSConnectionState(r)
	if connState == nil || len(connState.PeerCertificates) == 0 {
		return errors.New("no client certificate presented")
	}
	cert := connState.PeerCertificates[0]
	allowed := false
	for _, san := range cert.DNSNames {
		for _, suffix := range s.cfg.ClientCertSanSuffixAllowList {
			if strings.HasSuffix(san, suffix) {
				allowed = true
				break
			}
		}
		if allowed {
			break
		}
	}
	if !allowed {
		return fmt.Errorf("client certificate SANs %v do not match allowed suffixes %v", cert.DNSNames, s.cfg.ClientCertSanSuffixAllowList)
	}
	return nil
}

// extractClientCert returns the verified client certificate (if present) from the request.
func (s *ProxyServer) extractClientCert(r *http.Request) *x509.Certificate {
	connState := getTLSConnectionState(r)
	if connState == nil || len(connState.PeerCertificates) == 0 {
		return nil
	}
	return connState.PeerCertificates[0]
}

// getTLSConnectionState extracts the TLS connection state from the request.
func getTLSConnectionState(r *http.Request) *tls.ConnectionState {
	if r == nil || r.TLS == nil {
		return nil
	}
	return r.TLS
}

func (s *ProxyServer) handleConnect(w http.ResponseWriter, r *http.Request, claims *jwtutil.Claims, dest string, clientCert *x509.Certificate) {
	hijacker, ok := w.(http.Hijacker)
	if !ok {
		http.Error(w, "hijacking not supported", http.StatusInternalServerError)
		s.auditEvent("session.deny", claims, true,
			zap.String("remote", r.RemoteAddr),
			zap.String("dest", dest),
			zap.String("reason", "hijack_unsupported"),
		)
		return
	}
	clientConn, _, err := hijacker.Hijack()
	if err != nil {
		s.auditEvent("session.deny", claims, true,
			zap.String("remote", r.RemoteAddr),
			zap.String("dest", dest),
			zap.String("reason", "hijack_failed"),
			zap.Error(err),
		)
		http.Error(w, "failed to hijack connection", http.StatusInternalServerError)
		return
	}
	defer clientConn.Close()

	upstream, err := net.DialTimeout("tcp", dest, 10*time.Second)
	if err != nil {
		s.auditEvent("session.deny", claims, true,
			zap.String("remote", r.RemoteAddr),
			zap.String("dest", dest),
			zap.String("reason", "dial_failed"),
			zap.Error(err),
		)
		http.Error(w, "upstream unavailable", http.StatusBadGateway)
		return
	}
	defer upstream.Close()

	if _, err := io.WriteString(clientConn, "HTTP/1.1 200 Connection Established\r\n\r\n"); err != nil {
		s.auditEvent("session.stop", claims, true,
			zap.String("remote", r.RemoteAddr),
			zap.String("dest", dest),
			zap.String("mode", "connect"),
			zap.Error(fmt.Errorf("handshake_failed: %w", err)),
		)
		return
	}

	start := time.Now()
	s.auditSessionStart("session.start", claims, false, r.RemoteAddr, dest, "connect", clientCert)

	var txBytes, rxBytes uint64
	sessErr := &sessionError{}
	defer func() {
		fields := []zap.Field{
			zap.String("remote", r.RemoteAddr),
			zap.String("dest", dest),
			zap.String("mode", "connect"),
			zap.Duration("duration", time.Since(start)),
			zap.Uint64("bytes_tx", txBytes),
			zap.Uint64("bytes_rx", rxBytes),
		}
		if err := sessErr.get(); err != nil {
			fields = append(fields, zap.Error(err))
		}
		s.auditEvent("session.stop", claims, sessErr.get() != nil, fields...)
	}()

	clientReader := &countingReader{reader: clientConn}
	upstreamReader := &countingReader{reader: upstream}

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		if _, err := io.Copy(upstream, clientReader); err != nil && !errors.Is(err, io.EOF) {
			sessErr.set(fmt.Errorf("copy_client_to_upstream: %w", err))
		}
	}()

	if _, err := io.Copy(clientConn, upstreamReader); err != nil && !errors.Is(err, io.EOF) {
		sessErr.set(fmt.Errorf("copy_upstream_to_client: %w", err))
	}

	wg.Wait()
	txBytes = clientReader.total
	rxBytes = upstreamReader.total
}

func (s *ProxyServer) handleWebSocket(w http.ResponseWriter, r *http.Request, claims *jwtutil.Claims, dest string, clientCert *x509.Certificate) {
	conn, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		s.auditEvent("session.deny", claims, true,
			zap.String("remote", r.RemoteAddr),
			zap.String("dest", dest),
			zap.String("reason", "ws_upgrade_failed"),
			zap.Error(err),
		)
		return
	}

	upstream, err := net.DialTimeout("tcp", dest, 10*time.Second)
	if err != nil {
		s.auditEvent("session.deny", claims, true,
			zap.String("remote", r.RemoteAddr),
			zap.String("dest", dest),
			zap.String("reason", "ws_dial_failed"),
			zap.Error(err),
		)
		conn.Close()
		return
	}

	start := time.Now()
	s.auditSessionStart("session.start", claims, false, r.RemoteAddr, dest, "websocket", clientCert)

	var txBytes, rxBytes atomic.Uint64
	sessErr := &sessionError{}
	defer func() {
		fields := []zap.Field{
			zap.String("remote", r.RemoteAddr),
			zap.String("dest", dest),
			zap.String("mode", "websocket"),
			zap.Duration("duration", time.Since(start)),
			zap.Uint64("bytes_tx", txBytes.Load()),
			zap.Uint64("bytes_rx", rxBytes.Load()),
		}
		if err := sessErr.get(); err != nil {
			fields = append(fields, zap.Error(err))
		}
		s.auditEvent("session.stop", claims, sessErr.get() != nil, fields...)
		upstream.Close()
		conn.Close()
	}()
	done := make(chan struct{})
	go func() {
		defer close(done)
		for {
			typ, message, err := conn.ReadMessage()
			if err != nil {
				if !websocket.IsCloseError(err, websocket.CloseNormalClosure, websocket.CloseGoingAway) {
					sessErr.set(fmt.Errorf("ws_read: %w", err))
				}
				return
			}
			if typ != websocket.BinaryMessage {
				continue
			}
			if len(message) == 0 {
				continue
			}
			if _, err := upstream.Write(message); err != nil {
				sessErr.set(fmt.Errorf("ws_forward: %w", err))
				return
			}
			txBytes.Add(uint64(len(message)))
		}
	}()

	buf := make([]byte, 32*1024)
	for {
		n, err := upstream.Read(buf)
		if n > 0 {
			if err := conn.WriteMessage(websocket.BinaryMessage, buf[:n]); err != nil {
				sessErr.set(fmt.Errorf("ws_write: %w", err))
				break
			}
			rxBytes.Add(uint64(n))
		}
		if err != nil {
			if !errors.Is(err, io.EOF) {
				sessErr.set(fmt.Errorf("upstream_read: %w", err))
			}
			break
		}
	}

	<-done
}

func (s *ProxyServer) resolveDestination(claims *jwtutil.Claims) (string, error) {
	dest := strings.TrimSpace(claims.Dest)
	if dest == "" {
		return "", errors.New("dest claim missing")
	}
	host, portStr, err := net.SplitHostPort(dest)
	if err != nil {
		return "", fmt.Errorf("invalid dest: %w", err)
	}
	if !strings.HasSuffix(host, s.cfg.DestSuffix) {
		return "", fmt.Errorf("dest must end with %s", s.cfg.DestSuffix)
	}
	if claims.DNS != "" && !strings.EqualFold(claims.DNS, host) {
		return "", errors.New("dns claim mismatch")
	}
	port, err := strconv.Atoi(portStr)
	if err != nil || port <= 0 || port > 65535 {
		return "", errors.New("invalid dest port")
	}
	return net.JoinHostPort(host, strconv.Itoa(port)), nil
}

func (s *ProxyServer) deny(w http.ResponseWriter, r *http.Request, claims *jwtutil.Claims, wid, reason string, err error) {
	status := http.StatusForbidden
	if errors.Is(err, errBadTokenFormat) {
		status = http.StatusUnauthorized
	}
	http.Error(w, "access denied", status)

	fields := []zap.Field{
		zap.String("remote", r.RemoteAddr),
		zap.String("reason", reason),
	}
	if wid != "" && claims == nil {
		fields = append(fields, zap.String("wid", wid))
	}
	if err != nil {
		fields = append(fields, zap.Error(err))
	}
	s.auditEvent("session.deny", claims, true, fields...)
}

func (s *ProxyServer) auditEvent(event string, claims *jwtutil.Claims, warn bool, extra ...zap.Field) {
	fields := []zap.Field{zap.String("event", event)}
	if claims != nil {
		fields = append(fields,
			zap.String("wid", claims.Wid),
			zap.String("sub", claims.Sub),
			zap.String("jti", claims.ID),
			zap.String("dest", claims.Dest),
		)
		if claims.DNS != "" {
			fields = append(fields, zap.String("dns", claims.DNS))
		}
		if claims.Cluster != "" {
			fields = append(fields, zap.String("cluster", claims.Cluster))
		}
	}
	fields = append(fields, extra...)
	if warn {
		s.log.Warn("proxy.audit", fields...)
	} else {
		s.log.Info("proxy.audit", fields...)
	}
}

// auditSessionStart logs session.start with client certificate details for audit (AC-17a.03, AC-17b)
func (s *ProxyServer) auditSessionStart(event string, claims *jwtutil.Claims, warn bool, remote, dest, mode string, clientCert *x509.Certificate) {
	extra := []zap.Field{
		zap.String("remote", remote),
		zap.String("dest", dest),
		zap.String("mode", mode),
	}
	if clientCert != nil {
		extra = append(extra,
			zap.String("client_cert_subject", clientCert.Subject.String()),
			zap.String("client_cert_serial", clientCert.SerialNumber.Text(16)),
			zap.String("mtls_subject", clientCert.Subject.String()),
			zap.String("mtls_serial", clientCert.SerialNumber.Text(16)),
		)
		if len(clientCert.DNSNames) > 0 {
			extra = append(extra, zap.Strings("client_cert_sans", clientCert.DNSNames))
		}
		if len(clientCert.Subject.CommonName) > 0 {
			extra = append(extra, zap.String("client_cert_cn", clientCert.Subject.CommonName))
		}
		if len(clientCert.AuthorityKeyId) > 0 {
			extra = append(extra, zap.String("client_cert_akid", hex.EncodeToString(clientCert.AuthorityKeyId)))
		}
	}
	s.auditEvent(event, claims, warn, extra...)
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
	cert, err := tls.LoadX509KeyPair(s.cfg.TLSCertFile, s.cfg.TLSKeyFile)
	if err != nil {
		return fmt.Errorf("load tls certificate: %w", err)
	}

	tlsConfig := &tls.Config{
		Certificates:             []tls.Certificate{cert},
		MinVersion:               tls.VersionTLS12,
		PreferServerCipherSuites: true,
		CipherSuites: []uint16{
			tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
			tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
		},
		CurvePreferences: []tls.CurveID{
			tls.CurveP256,
			tls.CurveP384,
		},
		// Enforce client certificate authentication
		ClientAuth: tls.RequireAndVerifyClientCert,
	}

	// Add VerifyConnection to enforce SAN suffix allow-list at handshake time (AC-17b)
	if len(s.cfg.ClientCertSanSuffixAllowList) > 0 {
		tlsConfig.VerifyConnection = func(cs tls.ConnectionState) error {
			if len(cs.PeerCertificates) == 0 {
				return errors.New("no client certificate presented")
			}
			cert := cs.PeerCertificates[0]
			allowed := false
			for _, san := range cert.DNSNames {
				for _, suffix := range s.cfg.ClientCertSanSuffixAllowList {
					if strings.HasSuffix(san, suffix) {
						allowed = true
						break
					}
				}
				if allowed {
					break
				}
			}
			if !allowed {
				return fmt.Errorf("client certificate SANs %v do not match allowed suffixes %v", cert.DNSNames, s.cfg.ClientCertSanSuffixAllowList)
			}
			return nil
		}
	}

	server := &http.Server{
		Addr:              s.cfg.ListenAddr,
		Handler:           s,
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      30 * time.Second,
		ReadHeaderTimeout: 15 * time.Second,
		TLSConfig:         tlsConfig,
	}

	listener, err := net.Listen("tcp", s.cfg.ListenAddr)
	if err != nil {
		return fmt.Errorf("listen: %w", err)
	}
	defer listener.Close()

	s.log.Info("proxy listening", zap.String("addr", s.cfg.ListenAddr), zap.String("tls_cert", s.cfg.TLSCertFile))
	if err := server.Serve(tls.NewListener(listener, tlsConfig)); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	return nil
}

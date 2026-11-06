package pinger

import (
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"crypto/tls"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/coder/websocket"
	"github.com/quic-go/quic-go"
)

const probeTimeout = 3 * time.Second

type probeStream interface {
	io.ReadWriter
	SetDeadline(time.Time) error
}

func dial(ctx context.Context, peer *url.URL) (probeStream, time.Duration, func(), error) {
	addr := peer.Host
	tlsConfig := &tls.Config{
		ServerName:         peer.Hostname(),
		InsecureSkipVerify: true,
	}
	if sni := peer.Query().Get("sni"); sni != "" && net.ParseIP(sni) == nil {
		tlsConfig.ServerName = sni
	}
	start := time.Now()
	switch peer.Scheme {
	case "tcp", "tls":
		d := &net.Dialer{}
		var conn net.Conn
		var err error
		if peer.Scheme == "tls" {
			conn, err = (&tls.Dialer{NetDialer: d, Config: tlsConfig}).DialContext(ctx, "tcp", addr)
		} else {
			conn, err = d.DialContext(ctx, "tcp", addr)
		}
		elapsed := time.Since(start)
		if err != nil {
			return nil, 0, nil, err
		}
		return conn, elapsed, func() { _ = conn.Close() }, nil
	case "quic":
		conn, err := quic.DialAddr(ctx, addr, tlsConfig, nil)
		elapsed := time.Since(start)
		if err != nil {
			return nil, 0, nil, err
		}
		stream, err := conn.OpenStreamSync(ctx)
		if err != nil {
			_ = conn.CloseWithError(0, "")
			return nil, 0, nil, err
		}
		return stream, elapsed, func() { _ = conn.CloseWithError(0, "") }, nil
	case "ws", "wss":
		transport := &http.Transport{TLSClientConfig: tlsConfig, DisableKeepAlives: true}
		ws, _, err := websocket.Dial(ctx, peer.String(), &websocket.DialOptions{
			HTTPClient:   &http.Client{Transport: transport},
			Subprotocols: []string{"ygg-ws"},
		})
		elapsed := time.Since(start)
		if err != nil {
			transport.CloseIdleConnections()
			return nil, 0, nil, err
		}
		if ws.Subprotocol() != "ygg-ws" {
			_ = ws.CloseNow()
			transport.CloseIdleConnections()
			return nil, 0, nil, fmt.Errorf("server did not negotiate ygg-ws")
		}
		return websocket.NetConn(context.Background(), ws, websocket.MessageBinary), elapsed, func() {
			_ = ws.CloseNow()
			transport.CloseIdleConnections()
		}, nil
	default:
		return nil, 0, nil, fmt.Errorf("unsupported protocol %q", peer.Scheme)
	}
}

func Ping(uri string, count int) (time.Duration, int, error) {
	if count <= 0 {
		return 0, 0, fmt.Errorf("ping count must be positive")

	}
	peer, err := url.Parse(uri)
	if err != nil {
		return 0, 0, fmt.Errorf("parse peer URI: %w", err)
	}
	port, err := strconv.Atoi(peer.Port())
	if peer.Hostname() == "" || err != nil || port < 1 || port > 65535 {
		return 0, 0, fmt.Errorf("invalid peer host or port")
	}
	switch peer.Scheme {
	case "tcp", "tls", "quic", "ws", "wss":
	default:
		return 0, 0, fmt.Errorf("unsupported protocol %q", peer.Scheme)
	}
	if peer.User != nil {
		return 0, 0, fmt.Errorf("userinfo in peer URI is not supported")
	}
	password := []byte(peer.Query().Get("password"))
	if len(password) > 64 {
		return 0, 0, fmt.Errorf("peer password is too long")
	}
	priority := uint64(0)
	if value := peer.Query().Get("priority"); value != "" {
		priority, err = strconv.ParseUint(value, 10, 8)
		if err != nil {
			return 0, 0, fmt.Errorf("invalid peer priority: %w", err)
		}
	}
	publicKey, privateKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return 0, 0, fmt.Errorf("generate probe key: %w", err)
	}
	metadata, err := encodeMetadata(publicKey, privateKey, password, byte(priority))
	if err != nil {
		return 0, 0, err
	}

	var total time.Duration
	var lost int
	for i := 0; i < count; i++ {
		elapsed, err := probe(peer, metadata, password)
		if err != nil {
			slog.Debug("Peer probe failed", "host", peer.Host, "error", err)
			//lost++
			//continue
			lost = 999
			break // Do not try unavailable peers again
		}
		total += elapsed
	}
	if lost == count {
		return 0, lost, nil
	}
	return total / time.Duration(count-lost), lost, nil
}

func probe(peer *url.URL, metadata, password []byte) (time.Duration, error) {
	ctx, cancel := context.WithTimeout(context.Background(), probeTimeout)
	defer cancel()
	conn, elapsed, closeConn, err := dial(ctx, peer)
	if err != nil {
		return 0, fmt.Errorf("dial: %w", err)
	}
	defer closeConn()
	if err := conn.SetDeadline(time.Now().Add(probeTimeout)); err != nil {
		return 0, fmt.Errorf("set handshake deadline: %w", err)
	}
	if err := exchangeMetadata(conn, metadata, password); err != nil {
		return 0, fmt.Errorf("Yggdrasil handshake: %w", err)
	}
	return elapsed, nil
}

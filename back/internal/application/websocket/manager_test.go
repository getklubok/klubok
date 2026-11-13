package websocket

import (
	"bufio"
	"context"
	"errors"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"

	ws "github.com/lesismal/nbio/nbhttp/websocket"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"

	"klubok/internal/application/websocket/mocks"
	"klubok/internal/domain"
)

var errDB = errors.New("db error")

func TestManager_getSession(t *testing.T) {
	tests := []struct {
		name    string
		cid     domain.CookieID
		mgr     func(t *testing.T) *Manager
		want    *domain.Session
		wantErr error
	}{
		{
			name: "Cache hit",
			cid:  domain.CookieID("cid-123"),
			mgr: func(t *testing.T) *Manager {
				t.Helper()
				cache := mocks.NewSessionCacheMock(t)

				expected := &domain.Session{
					ID:       domain.SessionID("session-1"),
					CookieID: domain.CookieID("cid-123"),
					UserID:   domain.UserID("user-1"),
				}
				cache.EXPECT().Get(domain.CookieID("cid-123")).Return(expected, true)

				return &Manager{
					sessionCache: cache,
					log:          zap.NewNop(),
				}
			},
			want: &domain.Session{
				ID:       domain.SessionID("session-1"),
				CookieID: domain.CookieID("cid-123"),
				UserID:   domain.UserID("user-1"),
			},
		},
		{
			name: "Cache miss - found in service",
			cid:  domain.CookieID("cid-123"),
			mgr: func(t *testing.T) *Manager {
				t.Helper()
				cache := mocks.NewSessionCacheMock(t)
				sess := mocks.NewSessionServiceMock(t)

				cache.EXPECT().Get(domain.CookieID("cid-123")).Return(nil, false)
				sess.EXPECT().
					FindByCID(mock.Anything, domain.CookieID("cid-123")).
					Return(domain.Session{
						ID:       domain.SessionID("session-1"),
						CookieID: domain.CookieID("cid-123"),
						UserID:   domain.UserID("user-1"),
					}, nil)
				cache.EXPECT().Add(domain.CookieID("cid-123"), mock.Anything).Return(false)

				return &Manager{
					sessionCache:   cache,
					sessionService: sess,
					log:            zap.NewNop(),
				}
			},
			want: &domain.Session{
				ID:       domain.SessionID("session-1"),
				CookieID: domain.CookieID("cid-123"),
				UserID:   domain.UserID("user-1"),
			},
		},
		{
			name: "Session not found",
			cid:  domain.CookieID("cid-123"),
			mgr: func(t *testing.T) *Manager {
				t.Helper()
				cache := mocks.NewSessionCacheMock(t)
				sess := mocks.NewSessionServiceMock(t)

				cache.EXPECT().Get(domain.CookieID("cid-123")).Return(nil, false)
				sess.EXPECT().
					FindByCID(mock.Anything, domain.CookieID("cid-123")).
					Return(domain.Session{}, domain.ErrSessionNotFound)

				return &Manager{
					sessionCache:   cache,
					sessionService: sess,
					log:            zap.NewNop(),
				}
			},
			wantErr: errors.New("find session by cid: session not found"),
		},
		{
			name: "Service error",
			cid:  domain.CookieID("cid-123"),
			mgr: func(t *testing.T) *Manager {
				t.Helper()
				cache := mocks.NewSessionCacheMock(t)
				sess := mocks.NewSessionServiceMock(t)

				cache.EXPECT().Get(domain.CookieID("cid-123")).Return(nil, false)
				sess.EXPECT().
					FindByCID(mock.Anything, domain.CookieID("cid-123")).
					Return(domain.Session{}, errDB)

				return &Manager{
					sessionCache:   cache,
					sessionService: sess,
					log:            zap.NewNop(),
				}
			},
			wantErr: errors.New("find session by cid: db error"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wm := tt.mgr(t)
			got, err := wm.getSession(context.Background(), tt.cid)
			if tt.wantErr != nil {
				assert.EqualError(t, err, tt.wantErr.Error())
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestManager_getUser(t *testing.T) {
	tests := []struct {
		name    string
		id      domain.UserID
		mgr     func(t *testing.T) *Manager
		want    *domain.User
		wantErr error
	}{
		{
			name: "Cache hit",
			id:   domain.UserID("user-1"),
			mgr: func(t *testing.T) *Manager {
				t.Helper()
				cache := mocks.NewUserCacheMock(t)

				expected := &domain.User{
					ID:       domain.UserID("user-1"),
					Email:    "test@example.com",
					Username: "testuser",
				}
				cache.EXPECT().Get(domain.UserID("user-1")).Return(expected, true)

				return &Manager{
					userCache: cache,
					log:       zap.NewNop(),
				}
			},
			want: &domain.User{
				ID:       domain.UserID("user-1"),
				Email:    "test@example.com",
				Username: "testuser",
			},
		},
		{
			name: "Cache miss - found in service",
			id:   domain.UserID("user-1"),
			mgr: func(t *testing.T) *Manager {
				t.Helper()
				cache := mocks.NewUserCacheMock(t)
				user := mocks.NewUserServiceMock(t)

				cache.EXPECT().Get(domain.UserID("user-1")).Return(nil, false)
				user.EXPECT().
					FindByID(mock.Anything, domain.UserID("user-1")).
					Return(&domain.User{
						ID:       domain.UserID("user-1"),
						Email:    "test@example.com",
						Username: "testuser",
					}, nil)
				cache.EXPECT().Add(domain.UserID("user-1"), mock.Anything).Return(false)

				return &Manager{
					userCache:   cache,
					userService: user,
					log:         zap.NewNop(),
				}
			},
			want: &domain.User{
				ID:       domain.UserID("user-1"),
				Email:    "test@example.com",
				Username: "testuser",
			},
		},
		{
			name: "User not found",
			id:   domain.UserID("user-1"),
			mgr: func(t *testing.T) *Manager {
				t.Helper()
				cache := mocks.NewUserCacheMock(t)
				user := mocks.NewUserServiceMock(t)

				cache.EXPECT().Get(domain.UserID("user-1")).Return(nil, false)
				user.EXPECT().
					FindByID(mock.Anything, domain.UserID("user-1")).
					Return(nil, domain.ErrUserNotFound)

				return &Manager{
					userCache:   cache,
					userService: user,
					log:         zap.NewNop(),
				}
			},
			wantErr: errors.New("find user by id: user not found"),
		},
		{
			name: "Service error",
			id:   domain.UserID("user-1"),
			mgr: func(t *testing.T) *Manager {
				t.Helper()
				cache := mocks.NewUserCacheMock(t)
				user := mocks.NewUserServiceMock(t)

				cache.EXPECT().Get(domain.UserID("user-1")).Return(nil, false)
				user.EXPECT().
					FindByID(mock.Anything, domain.UserID("user-1")).
					Return(nil, errDB)

				return &Manager{
					userCache:   cache,
					userService: user,
					log:         zap.NewNop(),
				}
			},
			wantErr: errors.New("find user by id: db error"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wm := tt.mgr(t)
			got, err := wm.getUser(context.Background(), tt.id)
			if tt.wantErr != nil {
				assert.EqualError(t, err, tt.wantErr.Error())
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestManager_GetDiagramUsers(t *testing.T) {
	tests := []struct {
		name      string
		diagramID domain.DiagramID
		mgr       func(t *testing.T) *Manager
		want      []domain.WebSocketUser
	}{
		{
			name:      "Success",
			diagramID: domain.DiagramID("diagram-1"),
			mgr: func(t *testing.T) *Manager {
				t.Helper()

				conn := &ws.Conn{}

				return &Manager{
					connections: map[domain.DiagramID][]*ws.Conn{
						domain.DiagramID("diagram-1"): {conn},
					},
					connInfo: map[*ws.Conn]ConnInfo{
						conn: {
							UserID:    domain.UserID("user-1"),
							Username:  "username",
							DiagramID: domain.DiagramID("diagram-1"),
						},
					},
				}
			},
			want: []domain.WebSocketUser{
				{
					ID:       domain.UserID("user-1"),
					Username: "username",
				},
			},
		},
		{
			name:      "No connections",
			diagramID: domain.DiagramID("diagram-1"),
			mgr: func(t *testing.T) *Manager {
				t.Helper()
				return &Manager{
					connections: make(map[domain.DiagramID][]*ws.Conn),
					connInfo:    make(map[*ws.Conn]ConnInfo),
				}
			},
			want: nil,
		},
		{
			name:      "Empty diagram",
			diagramID: domain.DiagramID("diagram-1"),
			mgr: func(t *testing.T) *Manager {
				t.Helper()
				return &Manager{
					connections: map[domain.DiagramID][]*ws.Conn{
						domain.DiagramID("diagram-2"): {},
					},
					connInfo: make(map[*ws.Conn]ConnInfo),
				}
			},
			want: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wm := tt.mgr(t)
			got := wm.GetDiagramUsers(tt.diagramID)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestManager_GetDiagramStats(t *testing.T) {
	tests := []struct {
		name      string
		diagramID domain.DiagramID
		mgr       func(t *testing.T) *Manager
		want      domain.WebSocketStatDiagram
	}{
		{
			name:      "Empty diagram",
			diagramID: domain.DiagramID("diagram-1"),
			mgr: func(t *testing.T) *Manager {
				t.Helper()
				return &Manager{
					connections:     make(map[domain.DiagramID][]*ws.Conn),
					cursorPositions: make(map[domain.DiagramID]map[domain.UserID]domain.WebSocketCursorPosition),
				}
			},
			want: domain.WebSocketStatDiagram{
				DiagramID:   domain.DiagramID("diagram-1"),
				Connections: 0,
			},
		},
		{
			name:      "With connections",
			diagramID: domain.DiagramID("diagram-1"),
			mgr: func(t *testing.T) *Manager {
				t.Helper()
				wm := &Manager{
					connections:     make(map[domain.DiagramID][]*ws.Conn),
					cursorPositions: make(map[domain.DiagramID]map[domain.UserID]domain.WebSocketCursorPosition),
				}
				wm.connections[domain.DiagramID("diagram-1")] = []*ws.Conn{nil, nil}
				wm.cursorPositions[domain.DiagramID("diagram-1")] = map[domain.UserID]domain.WebSocketCursorPosition{
					domain.UserID("user-1"): {
						UserID:   domain.UserID("user-1"),
						Username: "testuser",
						X:        10.5,
						Y:        20.3,
					},
				}
				return wm
			},
			want: domain.WebSocketStatDiagram{
				DiagramID:   domain.DiagramID("diagram-1"),
				Connections: 2,
				CursorPositions: []domain.WebSocketCursorPosition{
					{
						UserID:   domain.UserID("user-1"),
						Username: "testuser",
						X:        10.5,
						Y:        20.3,
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wm := tt.mgr(t)
			got := wm.GetDiagramStats(tt.diagramID)
			assert.Equal(t, tt.want.DiagramID, got.DiagramID)
			assert.Equal(t, tt.want.Connections, got.Connections)
			if tt.want.CursorPositions == nil {
				assert.Nil(t, got.CursorPositions)
			} else {
				assert.ElementsMatch(t, tt.want.CursorPositions, got.CursorPositions)
			}
		})
	}
}

func TestClientIP(t *testing.T) {
	tests := []struct {
		name   string
		header string
		remote string
		want   string
	}{
		{
			name:   "X-Real-IP has priority",
			header: "203.0.113.7",
			remote: "127.0.0.1:5678",
			want:   "203.0.113.7",
		},
		{
			name:   "fallback to RemoteAddr",
			remote: "127.0.0.1:5678",
			want:   "127.0.0.1:5678",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/ws", nil)
			r.RemoteAddr = tt.remote
			if tt.header != "" {
				r.Header.Set("X-Real-IP", tt.header)
			}

			assert.Equal(t, tt.want, clientIP(r))
		})
	}
}

func TestManager_MaskRemoteAddrInLogs(t *testing.T) {
	tests := []struct {
		name       string
		masker     domain.IPMasker
		remote     string
		clientIP   string
		registered bool
		wantAddr   string
	}{
		{
			name:       "Masker disabled - full client ip in logs",
			masker:     domain.IPMasker{Enabled: false, V4Octets: 3, V6Octets: 7},
			remote:     "127.0.0.1:5678",
			clientIP:   "1.2.3.4",
			registered: true,
			wantAddr:   "1.2.3.4",
		},
		{
			name:       "Masker enabled - truncated ipv4 client ip in logs",
			masker:     domain.IPMasker{Enabled: true, V4Octets: 3, V6Octets: 7},
			remote:     "127.0.0.1:5678",
			clientIP:   "1.2.3.4",
			registered: true,
			wantAddr:   "1.2.3.x",
		},
		{
			name:       "Masker enabled - truncated ipv6 client ip in logs",
			masker:     domain.IPMasker{Enabled: true, V4Octets: 3, V6Octets: 7},
			remote:     "[::1]:5678",
			clientIP:   "2001:db8:85a3:0:0:8a2e:370:7334",
			registered: true,
			wantAddr:   "2001:db8:85a3:0:0:8a2e:370:x",
		},
		{
			name:       "Unregistered connection - falls back to remote addr",
			masker:     domain.IPMasker{Enabled: true, V4Octets: 3, V6Octets: 7},
			remote:     "127.0.0.1:5678",
			registered: false,
			wantAddr:   "127.0.0.x",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// net.Pipe синхронный: handshake-запись сервера блокируется,
			// пока клиент не читает. Сливаем клиентский конец.
			server, client := net.Pipe()
			go func() {
				_, _ = io.Copy(io.Discard, client)
				_ = client.Close()
			}()

			core, logs := observer.New(zapcore.InfoLevel)
			wm := &Manager{
				ipMasker: tt.masker,
				log:      zap.New(core),
				connInfo: make(map[*ws.Conn]ConnInfo),
			}
			wm.setupUpgrader()
			wm.setupHandlers()

			wsc := upgradeConn(t, wm.upgrader, tt.remote, server)
			if tt.registered {
				wm.connInfo[wsc] = ConnInfo{ClientIP: tt.clientIP}
			}

			// OnClose вызывается из CloseAndClean и логирует remote_addr.
			wsc.CloseAndClean(nil)

			_ = client.Close()

			entries := logs.FilterMessage("OnClose:").All()
			assert.Len(t, entries, 1)
			if len(entries) == 1 {
				assert.Equal(t, tt.wantAddr, entries[0].ContextMap()["remote_addr"])
			}
		})
	}
}

func TestExtractCoordinates(t *testing.T) {
	tests := []struct {
		name      string
		data      any
		fieldName string
		want      float64
	}{
		{
			name:      "Valid x coordinate",
			data:      map[string]any{"x": 10.5, "y": 20.3},
			fieldName: "x",
			want:      10.5,
		},
		{
			name:      "Valid y coordinate",
			data:      map[string]any{"x": 10.5, "y": 20.3},
			fieldName: "y",
			want:      20.3,
		},
		{
			name:      "Field not found",
			data:      map[string]any{"x": 10.5},
			fieldName: "y",
			want:      0,
		},
		{
			name:      "Not a map",
			data:      "not a map",
			fieldName: "x",
			want:      0,
		},
		{
			name:      "Wrong type for field",
			data:      map[string]any{"x": "not a number"},
			fieldName: "x",
			want:      0,
		},
		{
			name:      "Nil data",
			data:      nil,
			fieldName: "x",
			want:      0,
		},
		{
			name:      "Zero coordinate",
			data:      map[string]any{"x": 0.0, "y": 0.0},
			fieldName: "x",
			want:      0.0,
		},
		{
			name:      "Negative coordinate",
			data:      map[string]any{"x": -15.7},
			fieldName: "x",
			want:      -15.7,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractCoordinates(tt.data, tt.fieldName)
			assert.InDelta(t, tt.want, got, 0.001)
		})
	}
}

// fakeAddr подменяет RemoteAddr у соединения: в проде r.RemoteAddr и
// conn.RemoteAddr() берутся из реального TCP-соединения, в тесте задаём
// адрес явно.
type fakeAddr struct {
	addr string
}

func (a fakeAddr) Network() string { return "tcp" }
func (a fakeAddr) String() string  { return a.addr }

// fakeAddrConn — обёртка над net.Conn с управляемым RemoteAddr.
type fakeAddrConn struct {
	net.Conn

	addr fakeAddr
}

func (c *fakeAddrConn) RemoteAddr() net.Addr { return c.addr }

// hijackRecorder позволяет Upgrader.Upgrade перехватить net.Conn
// с подменённым RemoteAddr и выполняет запись handshake в буфер.
type hijackRecorder struct {
	*httptest.ResponseRecorder

	conn net.Conn
}

func (h *hijackRecorder) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	rw := bufio.NewReadWriter(bufio.NewReader(h.conn), bufio.NewWriter(h.conn))

	return h.conn, rw, nil
}

// upgradeConn выполняет websocket-handshake с заданным RemoteAddr
// и возвращает полученный *ws.Conn. Аргументы (false, false) отключают
// передачу conn в poller и чтение из соединения в горутине.
func upgradeConn(t *testing.T, upgrader *ws.Upgrader, remote string, conn net.Conn) *ws.Conn {
	t.Helper()

	ctx := context.Background()
	req := httptest.NewRequestWithContext(ctx, http.MethodGet, "/ws", nil)
	req.Header.Set("Upgrade", "websocket")
	req.Header.Set("Connection", "Upgrade")
	req.Header.Set("Sec-WebSocket-Version", "13")
	req.Header.Set("Sec-WebSocket-Key", "dGhlIHNhbXBsZSBub25jZQ==")
	req.RemoteAddr = remote

	w := &hijackRecorder{
		ResponseRecorder: httptest.NewRecorder(),
		conn:             &fakeAddrConn{Conn: conn, addr: fakeAddr{addr: remote}},
	}

	wsc, err := upgrader.Upgrade(w, req, nil, false, false)
	assert.NoError(t, err)

	return wsc
}

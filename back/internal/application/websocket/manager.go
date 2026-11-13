package websocket

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"maps"
	"net/http"
	"sync"
	"time"

	"github.com/hashicorp/golang-lru/v2/expirable"
	"github.com/lesismal/nbio/nbhttp"
	ws "github.com/lesismal/nbio/nbhttp/websocket"
	"go.uber.org/zap"

	"klubok/internal/domain"
)

const (
	defaultCacheSize = 1000
	defaultCacheTTL  = 10 * time.Minute

	defaultMaxConnPerUser = 10
	defaultMaxOnline      = 1_000_000

	defaultStopTimeout = 3 * time.Second

	defaultKeepaliveTime     = 120 * time.Second
	defaultHeartbeatInterval = 60 * time.Second
)

// Message структура для сообщений.
type Message struct {
	Type      domain.WebSocketMessageType `json:"type"`
	DiagramID domain.DiagramID            `json:"diagramId"`
	Timestamp int64                       `json:"timestamp"`
	Data      any                         `json:"data"`
}

type ConnInfo struct {
	DiagramID domain.DiagramID
	UserID    domain.UserID
	Username  string
	ClientIP  string
}

// Manager менеджер WebSocket соединений.
type Manager struct {
	port              int
	cookieManager     CookieManager
	sessionService    SessionService
	userService       UserService
	permissionService PermissionService
	ipMasker          domain.IPMasker
	log               *zap.Logger

	sessionCache SessionCache
	userCache    UserCache

	engine   *nbhttp.Engine
	upgrader *ws.Upgrader

	// Мапа для хранения соединений по diagramID
	connections map[domain.DiagramID][]*ws.Conn
	connInfo    map[*ws.Conn]ConnInfo
	muConn      sync.RWMutex

	// Мапа для хранения позиций курсоров: diagramID -> userID -> cursorPosition
	cursorPositions map[domain.DiagramID]map[domain.UserID]domain.WebSocketCursorPosition
	muCursor        sync.RWMutex

	heartbeatTicker *time.Ticker
	heartbeatStop   chan struct{}
}

// NewManager создает новый менеджер WebSocket.
func NewManager(
	port int,
	cookieManager CookieManager,
	sessionService SessionService,
	userService UserService,
	permissionService PermissionService,
	ipMasker domain.IPMasker,
	log *zap.Logger,
) *Manager {
	wm := &Manager{
		port:              port,
		cookieManager:     cookieManager,
		sessionService:    sessionService,
		userService:       userService,
		permissionService: permissionService,
		ipMasker:          ipMasker,
		log:               log.Named("websocket_server"),

		sessionCache: expirable.NewLRU[domain.CookieID, *domain.Session](defaultCacheSize, nil, defaultCacheTTL),
		userCache:    expirable.NewLRU[domain.UserID, *domain.User](defaultCacheSize, nil, defaultCacheTTL),

		connections:     make(map[domain.DiagramID][]*ws.Conn),
		connInfo:        make(map[*ws.Conn]ConnInfo),
		cursorPositions: make(map[domain.DiagramID]map[domain.UserID]domain.WebSocketCursorPosition),
	}

	wm.setupUpgrader()
	wm.setupEngine()
	wm.setupHandlers()

	return wm
}

// Start запускает WebSocket сервер.
func (wm *Manager) Start() error {
	// Запускаем heartbeat
	wm.startHeartbeat()

	// Запускаем сервер
	if err := wm.engine.Start(); err != nil {
		return fmt.Errorf("nbio.Start failed: %w", err)
	}

	wm.log.Info("WebSocket started",
		zap.Int("port", wm.port),
	)
	return nil
}

// Stop останавливает WebSocket сервер.
func (wm *Manager) Stop() {
	ctx, cancel := context.WithTimeout(context.Background(), defaultStopTimeout)
	defer cancel()

	if wm.heartbeatStop != nil {
		close(wm.heartbeatStop)
	}
	if wm.heartbeatTicker != nil {
		wm.heartbeatTicker.Stop()
	}

	if err := wm.engine.Shutdown(ctx); err != nil {
		wm.log.Error("Shutdown failed",
			zap.Error(err),
		)
	}

	wm.log.Info("WebSocket stopped")
}

// BroadcastToDiagram отправляет сообщение всем клиентам конкретного diagramID, кроме notForUserID (если не nil).
func (wm *Manager) BroadcastToDiagram(diagramID domain.DiagramID, msg Message, notForUserID *domain.UserID) error {
	wm.muConn.RLock()

	conns := make([]*ws.Conn, len(wm.connections[diagramID]))
	copy(conns, wm.connections[diagramID])

	connsInfo := make(map[*ws.Conn]ConnInfo, len(wm.connInfo))
	maps.Copy(connsInfo, wm.connInfo)

	wm.muConn.RUnlock()

	if len(conns) == 0 {
		return nil
	}

	// Сериализуем сообщение один раз
	data, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("marshal message: %w", err)
	}

	var errs []error

	// Отправляем всем клиентам
	for _, conn := range conns {
		connInfo, ok := connsInfo[conn]
		if !ok {
			continue
		}

		if notForUserID != nil && connInfo.UserID == *notForUserID {
			continue
		}

		if err := conn.WriteMessage(ws.TextMessage, data); err != nil {
			errs = append(errs, err)
			wm.log.Error("Write message failed",
				zap.String("remote_addr", wm.ipMasker.Mask(connInfo.ClientIP)),
				zap.Error(err),
			)

			// Удаляем проблемное соединение
			wm.unregisterConnection(conn)
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("send message: %w", errors.Join(errs...))
	}

	return nil
}

// GetDiagramUsers возвращает список пользователей для diagramID.
func (wm *Manager) GetDiagramUsers(diagramID domain.DiagramID) []domain.WebSocketUser {
	wm.muConn.RLock()

	conns := make([]*ws.Conn, len(wm.connections[diagramID]))
	copy(conns, wm.connections[diagramID])

	connsInfo := make(map[*ws.Conn]ConnInfo, len(wm.connInfo))
	maps.Copy(connsInfo, wm.connInfo)

	wm.muConn.RUnlock()

	if len(conns) == 0 {
		return nil
	}

	userIDs := make(map[domain.UserID]struct{}, len(conns))
	users := make([]domain.WebSocketUser, 0, len(conns))
	for _, conn := range conns {
		cInfo, cOK := connsInfo[conn]
		if !cOK {
			continue
		}

		if _, ok := userIDs[cInfo.UserID]; ok {
			continue
		}

		userIDs[cInfo.UserID] = struct{}{}
		users = append(users, domain.WebSocketUser{
			ID:       cInfo.UserID,
			Username: cInfo.Username,
		})
	}

	return users
}

// GetDiagramStats возвращает статистику по diagramID.
func (wm *Manager) GetDiagramStats(diagramID domain.DiagramID) domain.WebSocketStatDiagram {
	wm.muConn.RLock()
	connections := len(wm.connections[diagramID])
	wm.muConn.RUnlock()

	wm.muCursor.RLock()
	var cursorPositions []domain.WebSocketCursorPosition
	if positions, ok := wm.cursorPositions[diagramID]; ok {
		cursorPositions = make([]domain.WebSocketCursorPosition, 0, len(positions))
		for _, position := range positions {
			cursorPositions = append(cursorPositions, position)
		}
	}
	wm.muCursor.RUnlock()

	return domain.WebSocketStatDiagram{
		DiagramID:       diagramID,
		Connections:     connections,
		CursorPositions: cursorPositions,
	}
}

// Устанавливаем WebSocket upgrader.
func (wm *Manager) setupUpgrader() {
	wm.upgrader = ws.NewUpgrader()
	wm.upgrader.CheckOrigin = func(r *http.Request) bool {
		// Разрешаем все origin для упрощения
		// В продакшене нужно настроить правильно
		return true
	}
}

func (wm *Manager) setupEngine() {
	mux := &http.ServeMux{}
	mux.HandleFunc("/ws", wm.onWebsocket)

	// Создаем nbio engine
	wm.engine = nbhttp.NewEngine(nbhttp.Config{
		Name:                    "WebSocketServer",
		Network:                 "tcp",
		Addrs:                   []string{fmt.Sprintf(":%d", wm.port)},
		MaxLoad:                 defaultMaxOnline,
		KeepaliveTime:           defaultKeepaliveTime,
		ReleaseWebsocketPayload: true,
		Handler:                 mux,
	})
}

func (wm *Manager) onWebsocket(w http.ResponseWriter, r *http.Request) { //nolint:gocognit // TODO: refactor
	ip := clientIP(r)

	wm.log.Info("onWebsocket:",
		zap.String("remote_addr", wm.ipMasker.Mask(ip)),
	)

	ctx := r.Context()

	cookieID, err := wm.cookieManager.Get(r)
	if err != nil {
		if !errors.Is(err, http.ErrNoCookie) {
			wm.log.Error("Get session id from cookie",
				zap.Error(err),
			)
		}

		wm.log.Warn("No cookieID in request")
		return
	}

	// Получаем diagramID из query параметров
	diagramID := domain.DiagramID(r.URL.Query().Get("diagramID"))
	if diagramID == "" {
		wm.log.Warn("No diagramID in query params")
		return
	}

	session, err := wm.getSession(ctx, cookieID)
	if err != nil {
		if !errors.Is(err, domain.ErrSessionNotFound) {
			wm.log.Error("Get session",
				zap.Error(err),
			)
		}

		wm.log.Warn("Session not found")
		return
	}

	ok, err := wm.permissionService.HasDiagramPermission(ctx, diagramID, session.UserID, []domain.DiagramCollaboratorPermission{
		domain.DiagramCollaboratorPermissionOwner,
		domain.DiagramCollaboratorPermissionView,
		domain.DiagramCollaboratorPermissionEdit,
	})
	if err != nil {
		wm.log.Error("Check diagram collaborator permission",
			zap.Error(err),
		)
		return
	}
	if !ok {
		wm.log.Warn("User is not diagram owner or has not access edit to diagram")
		return
	}

	conn, err := wm.upgrader.Upgrade(w, r, nil)
	if err != nil {
		panic(err)
	}

	wm.log.Info("OnOpen:",
		zap.String("remote_addr", wm.ipMasker.Mask(ip)),
	)

	user, err := wm.getUser(ctx, session.UserID)
	if err != nil {
		if !errors.Is(err, domain.ErrUserNotFound) {
			wm.log.Error("Get user",
				zap.Error(err),
			)
		}

		wm.log.Warn("User not found")
		return
	}

	// Регистрируем соединение
	if err := wm.registerConnection(conn, diagramID, session.UserID, user.Username, ip); err != nil {
		wm.log.Warn("Failed to register connection",
			zap.String("diagramID", diagramID.String()),
			zap.String("userID", session.UserID.String()),
			zap.Error(err),
		)

		if err := conn.Close(); err != nil {
			wm.log.Error("Close connection failed",
				zap.Error(err),
			)
		}

		return
	}
}

// setupHandlers настраивает обработчики WebSocket.
func (wm *Manager) setupHandlers() {
	// Обработчик закрытия соединения
	wm.upgrader.OnClose(func(c *ws.Conn, err error) {
		wm.log.Info("OnClose:",
			zap.String("remote_addr", wm.ipMasker.Mask(wm.clientIPForConn(c))),
		)
		wm.unregisterConnection(c)
	})

	wm.upgrader.OnMessage(func(c *ws.Conn, messageType ws.MessageType, data []byte) {
		var msg Message
		if err := json.Unmarshal(data, &msg); err != nil {
			wm.log.Error("Unmarshal message failed",
				zap.Error(err),
			)
			return
		}

		switch msg.Type {
		case domain.WebSocketMessageTypePing:
			pongMsg := Message{
				Type:      domain.WebSocketMessageTypePong,
				DiagramID: msg.DiagramID,
				Timestamp: time.Now().Unix(),
			}
			pongData, err := json.Marshal(pongMsg)
			if err != nil {
				wm.log.Error("Marshal pong message failed",
					zap.Error(err),
				)
				return
			}
			if err := c.WriteMessage(ws.TextMessage, pongData); err != nil {
				wm.log.Error("Send pong failed",
					zap.Error(err),
				)
				wm.unregisterConnection(c)
			}

		case domain.WebSocketMessageTypeCursorMove:
			wm.handleCursorMove(c, msg)

		default:
			wm.log.Info("OnMessage:",
				zap.Any("message_type", messageType),
				zap.String("data", string(data)),
			)
		}
	})
}

// clientIP берёт IP клиента из X-Real-IP (его ставит nginx), иначе RemoteAddr.
func clientIP(r *http.Request) string {
	if ip := r.Header.Get("X-Real-IP"); ip != "" {
		return ip
	}

	return r.RemoteAddr
}

// clientIPForConn возвращает сохранённый IP клиента, иначе адрес TCP-соединения.
func (wm *Manager) clientIPForConn(conn *ws.Conn) string {
	wm.muConn.RLock()
	info, ok := wm.connInfo[conn]
	wm.muConn.RUnlock()

	if ok {
		return info.ClientIP
	}

	return conn.RemoteAddr().String()
}

// registerConnection регистрирует новое соединение.
func (wm *Manager) registerConnection(conn *ws.Conn, diagramID domain.DiagramID, userID domain.UserID, username, ip string) error {
	wm.muConn.Lock()

	// Проверяем лимит соединений на userID
	count := 0
	for _, info := range wm.connInfo {
		if info.UserID == userID {
			count++
		}
	}
	if count >= defaultMaxConnPerUser {
		wm.muConn.Unlock()
		return fmt.Errorf("max connections per user reached: %d", defaultMaxConnPerUser)
	}

	// Добавляем соединение
	wm.connections[diagramID] = append(wm.connections[diagramID], conn)
	wm.connInfo[conn] = ConnInfo{
		DiagramID: diagramID,
		UserID:    userID,
		Username:  username,
		ClientIP:  ip,
	}
	connCount := len(wm.connections[diagramID])

	wm.muConn.Unlock()

	wm.log.Info("Registered connection",
		zap.String("diagramID", diagramID.String()),
		zap.String("userID", userID.String()),
		zap.Int("connections", connCount),
	)

	// Отправляем приветственное сообщение
	welcomeMsg := Message{
		Type:      domain.WebSocketMessageTypeConnected,
		DiagramID: diagramID,
		Timestamp: time.Now().Unix(),
	}
	data, err := json.Marshal(welcomeMsg)
	if err != nil {
		return fmt.Errorf("marshal message: %w", err)
	}
	if err := conn.WriteMessage(ws.TextMessage, data); err != nil {
		wm.log.Error("Send ping failed",
			zap.Error(err),
		)
		// Удаляем проблемное соединение
		wm.unregisterConnection(conn)
	}

	// Отправляем сообщение о подключении
	connectedMsg := Message{
		Type:      domain.WebSocketMessageTypeUserConnected,
		DiagramID: diagramID,
		Timestamp: time.Now().Unix(),
		Data: domain.WebSocketUser{
			ID:       userID,
			Username: username,
		},
	}
	if err := wm.BroadcastToDiagram(diagramID, connectedMsg, nil); err != nil {
		wm.log.Warn("Broadcast to diagram",
			zap.String("diagramID", diagramID.String()),
			zap.String("userID", userID.String()),
			zap.Error(err),
		)
	}

	return nil
}

// unregisterConnection удаляет соединение.
func (wm *Manager) unregisterConnection(conn *ws.Conn) {
	wm.muConn.Lock()

	cInfo, exists := wm.connInfo[conn]
	if !exists {
		wm.muConn.Unlock()
		return
	}

	// Удаляем соединение из списка для diagramID
	if connections, ok := wm.connections[cInfo.DiagramID]; ok {
		for i, c := range connections {
			if c == conn {
				wm.connections[cInfo.DiagramID] = append(connections[:i], connections[i+1:]...)
				break
			}
		}

		// Если для diagramID больше нет соединений, удаляем ключ
		if len(wm.connections[cInfo.DiagramID]) == 0 {
			delete(wm.connections, cInfo.DiagramID)
		}
	}

	delete(wm.connInfo, conn)

	wm.muConn.Unlock()

	// Удаляем позицию курсора пользователя
	wm.muCursor.Lock()
	if cursorUsers, ok := wm.cursorPositions[cInfo.DiagramID]; ok {
		delete(cursorUsers, cInfo.UserID)
		if len(cursorUsers) == 0 {
			delete(wm.cursorPositions, cInfo.DiagramID)
		}
	}
	wm.muCursor.Unlock()

	wm.log.Info("Unregistered connection",
		zap.String("diagramID", cInfo.DiagramID.String()),
	)

	// Отправляем сообщение об отключении
	disconnectedMsg := Message{
		Type:      domain.WebSocketMessageTypeUserDisconnected,
		DiagramID: cInfo.DiagramID,
		Timestamp: time.Now().Unix(),
		Data: domain.WebSocketUser{
			ID:       cInfo.UserID,
			Username: cInfo.Username,
		},
	}
	if err := wm.BroadcastToDiagram(cInfo.DiagramID, disconnectedMsg, nil); err != nil {
		wm.log.Warn("Broadcast to diagram",
			zap.String("diagramID", cInfo.DiagramID.String()),
			zap.String("userID", cInfo.UserID.String()),
			zap.Error(err),
		)
	}
}

func (wm *Manager) getSession(ctx context.Context, cid domain.CookieID) (*domain.Session, error) {
	if s, ok := wm.sessionCache.Get(cid); ok {
		return s, nil
	}

	s, err := wm.sessionService.FindByCID(ctx, cid)
	if err != nil {
		return nil, fmt.Errorf("find session by cid: %w", err)
	}

	wm.sessionCache.Add(cid, &s)

	return &s, nil
}

func (wm *Manager) getUser(ctx context.Context, id domain.UserID) (*domain.User, error) {
	if u, ok := wm.userCache.Get(id); ok {
		return u, nil
	}

	u, err := wm.userService.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("find user by id: %w", err)
	}

	wm.userCache.Add(id, u)

	return u, nil
}

// startHeartbeat запускает heartbeat для всех соединений.
func (wm *Manager) startHeartbeat() {
	wm.heartbeatTicker = time.NewTicker(defaultHeartbeatInterval)
	wm.heartbeatStop = make(chan struct{})

	go func() {
		for {
			select {
			case <-wm.heartbeatTicker.C:
				wm.sendHeartbeat()
			case <-wm.heartbeatStop:
				return
			}
		}
	}()
}

// sendHeartbeat отправляет ping всем активным соединениям.
func (wm *Manager) sendHeartbeat() {
	wm.muConn.RLock()
	allConns := make(map[*ws.Conn]ConnInfo, len(wm.connInfo))
	maps.Copy(allConns, wm.connInfo)
	wm.muConn.RUnlock()

	for conn, info := range allConns {
		pingMsg := Message{
			Type:      domain.WebSocketMessageTypePing,
			DiagramID: info.DiagramID,
			Timestamp: time.Now().Unix(),
		}
		data, err := json.Marshal(pingMsg)
		if err != nil {
			wm.log.Error("Marshal ping message failed",
				zap.Error(err),
			)
			continue
		}

		if err := conn.WriteMessage(ws.TextMessage, data); err != nil {
			wm.log.Error("Send ping failed",
				zap.Error(err),
			)
			// Удаляем проблемное соединение
			wm.unregisterConnection(conn)
		}
	}
}

func (wm *Manager) handleCursorMove(conn *ws.Conn, msg Message) {
	wm.muConn.Lock()
	connInfo, exists := wm.connInfo[conn]
	wm.muConn.Unlock()
	if !exists {
		return
	}

	cursorMove := domain.WebSocketCursorPosition{
		UserID:   connInfo.UserID,
		Username: connInfo.Username,
		X:        extractCoordinates(msg.Data, "x"),
		Y:        extractCoordinates(msg.Data, "y"),
	}

	// Сохраняем позицию курсора
	wm.muCursor.Lock()
	if wm.cursorPositions[msg.DiagramID] == nil {
		wm.cursorPositions[msg.DiagramID] = make(map[domain.UserID]domain.WebSocketCursorPosition)
	}
	wm.cursorPositions[msg.DiagramID][connInfo.UserID] = cursorMove
	wm.muCursor.Unlock()

	// Отправляем сообщение о движении курсора всем остальным пользователям диаграммы
	cursorMsg := Message{
		Type:      domain.WebSocketMessageTypeCursorMove,
		DiagramID: msg.DiagramID,
		Timestamp: time.Now().Unix(),
		Data:      cursorMove,
	}
	if err := wm.BroadcastToDiagram(msg.DiagramID, cursorMsg, &connInfo.UserID); err != nil {
		wm.log.Warn("Broadcast cursor move failed",
			zap.String("diagramID", msg.DiagramID.String()),
			zap.String("userID", connInfo.UserID.String()),
			zap.Error(err),
		)
	}
}

func extractCoordinates(data any, fieldName string) float64 {
	if dataMap, okMap := data.(map[string]any); okMap {
		if val, exists := dataMap[fieldName]; exists {
			if typedVal, okVal := val.(float64); okVal {
				return typedVal
			}
		}
	}

	return 0
}

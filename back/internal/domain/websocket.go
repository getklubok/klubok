package domain

type WebSocketMessageType string

func (v WebSocketMessageType) String() string {
	return string(v)
}

const (
	WebSocketMessageTypeConnected          WebSocketMessageType = "ws:connected"
	WebSocketMessageTypePing               WebSocketMessageType = "ws:ping"
	WebSocketMessageTypePong               WebSocketMessageType = "ws:pong"
	WebSocketMessageTypeShapeUpserted      WebSocketMessageType = "shape:upserted"
	WebSocketMessageTypeShapeDeleted       WebSocketMessageType = "shape:deleted"
	WebSocketMessageTypeConnectionUpserted WebSocketMessageType = "connection:upserted"
	WebSocketMessageTypeConnectionDeleted  WebSocketMessageType = "connection:deleted"
	WebSocketMessageTypeCommentAdded       WebSocketMessageType = "comment:added"
	WebSocketMessageTypeCursorMove         WebSocketMessageType = "cursor:move"
	WebSocketMessageTypeUserConnected      WebSocketMessageType = "user:connected"
	WebSocketMessageTypeUserDisconnected   WebSocketMessageType = "user:disconnected"
)

type WebSocketUser struct {
	ID       UserID `json:"id"`
	Username string `json:"username"`
}

type WebSocketStatDiagram struct {
	DiagramID       DiagramID                 `json:"diagramId"`
	Connections     int                       `json:"connections"`
	CursorPositions []WebSocketCursorPosition `json:"cursorPositions"`
}

type WebSocketCursorPosition struct {
	UserID   UserID  `json:"userId"`
	Username string  `json:"username"`
	X        float64 `json:"x"`
	Y        float64 `json:"y"`
}

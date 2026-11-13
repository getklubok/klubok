package domain

// UserID — ULID пользователя.
type UserID string

func (id UserID) String() string { return string(id) }

// SessionID — ULID сессии.
type SessionID string

func (id SessionID) String() string { return string(id) }

// CookieID — 30-символьный случайный идентификатор cookie сессии (не ULID,
// см. sessions.cid в migrations/00001_init.sql).
type CookieID string

func (id CookieID) String() string { return string(id) }

// DiagramID — ULID диаграммы.
type DiagramID string

func (id DiagramID) String() string { return string(id) }

// ShapeID — ULID элемента (shape) диаграммы.
type ShapeID string

func (id ShapeID) String() string { return string(id) }

// CommentID — ULID комментария.
type CommentID string

func (id CommentID) String() string { return string(id) }

// CollaboratorID — ULID записи «диаграмма-коллаборатор».
type CollaboratorID string

func (id CollaboratorID) String() string { return string(id) }

// EmbedID — ULID записи embed.
type EmbedID string

func (id EmbedID) String() string { return string(id) }

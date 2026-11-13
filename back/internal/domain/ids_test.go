package domain

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIDTypesString(t *testing.T) {
	// все 8 типов: String() возвращает исходное значение
	ids := []struct {
		name string
		got  string
	}{
		{"UserID", UserID("01J6GP155TZP3FMQ6F86KHE34Y").String()},
		{"SessionID", SessionID("01J6GP155TZP3FMQ6F86KHE34Y").String()},
		{"CookieID", CookieID("cookie").String()},
		{"DiagramID", DiagramID("01J6GP155TZP3FMQ6F86KHE34Y").String()},
		{"ShapeID", ShapeID("01J6GP155TZP3FMQ6F86KHE34Y").String()},
		{"CommentID", CommentID("01J6GP155TZP3FMQ6F86KHE34Y").String()},
		{"CollaboratorID", CollaboratorID("01J6GP155TZP3FMQ6F86KHE34Y").String()},
		{"EmbedID", EmbedID("01J6GP155TZP3FMQ6F86KHE34Y").String()},
	}
	for _, tt := range ids {
		want := "01J6GP155TZP3FMQ6F86KHE34Y"
		if tt.name == "CookieID" {
			want = "cookie"
		}
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, want, tt.got)
		})
	}
}

func TestIDTypesJSON(t *testing.T) {
	in := struct {
		UserID         UserID         `json:"userId"`
		SessionID      SessionID      `json:"sessionId"`
		CookieID       CookieID       `json:"cookieId"`
		DiagramID      DiagramID      `json:"diagramId"`
		ShapeID        ShapeID        `json:"shapeId"`
		CommentID      CommentID      `json:"commentId"`
		CollaboratorID CollaboratorID `json:"collaboratorId"`
		EmbedID        EmbedID        `json:"embedId"`
	}{
		UserID:         "u1",
		SessionID:      "s1",
		CookieID:       "c1",
		DiagramID:      "d1",
		ShapeID:        "sh1",
		CommentID:      "cm1",
		CollaboratorID: "co1",
		EmbedID:        "e1",
	}

	data, err := json.Marshal(in)
	require.NoError(t, err)

	var out map[string]string
	require.NoError(t, json.Unmarshal(data, &out))
	assert.Equal(t, "u1", out["userId"])
	assert.Equal(t, "s1", out["sessionId"])
	assert.Equal(t, "c1", out["cookieId"])
	assert.Equal(t, "d1", out["diagramId"])
	assert.Equal(t, "sh1", out["shapeId"])
	assert.Equal(t, "cm1", out["commentId"])
	assert.Equal(t, "co1", out["collaboratorId"])
	assert.Equal(t, "e1", out["embedId"])
}

package image

import (
	"strings"
	"unicode"
)

// sanitizeOriginalName убирает компоненты пути и управляющие символы из имени файла.
func sanitizeOriginalName(name, ext string) string {
	name = strings.ReplaceAll(name, "\\", "/")

	if idx := strings.LastIndex(name, "/"); idx >= 0 {
		name = name[idx+1:]
	}

	name = strings.Map(func(r rune) rune {
		if unicode.IsControl(r) {
			return -1
		}

		return r
	}, name)
	name = strings.TrimSpace(name)

	if name == "" || name == "." || name == ".." {
		return "image." + ext
	}

	return name
}

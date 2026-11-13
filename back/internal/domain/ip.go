package domain

import (
	"context"
	"net/netip"
	"strings"

	"github.com/go-chi/chi/v5/middleware"
)

type IP string

func (ip IP) String() string {
	return string(ip)
}

func IPFromContext(ctx context.Context) IP {
	return IP(middleware.GetClientIP(ctx))
}

// IPMasker маскирует IP-адреса перед записью в сохраняемые места:
// оставляет первые V4Octets октетов (IPv4) или V6Octets групп (IPv6),
// остальное заменяет на "x"/":x".
type IPMasker struct {
	Enabled  bool
	V4Octets int
	V6Octets int
}

const (
	defaultV4Octets = 3
	defaultV6Octets = 7

	ipv4Octets = 4
	ipv6Groups = 8
)

// Mask усекает адрес до первых октетов/групп: IPv4 — "1.2.3.x",
// IPv6 — "2001:db8:85a3:0:0:8a2e:370:x" (ровно 8 групп). Порт отбрасывается.
// Принимаются формы "ip", "ip:port" и "[ipv6]:port". При Enabled=false или
// неразбираемом входе строка возвращается без изменений.
func (m IPMasker) Mask(addr string) string {
	if !m.Enabled {
		return addr
	}

	a := parseAddr(addr)
	if !a.IsValid() {
		return addr
	}

	// Зона (например, "%eth0") — имя локального интерфейса, в логи не нужна.
	a = a.WithZone("")

	if a.Is4() || a.Is4In6() {
		return maskV4(a, clamp(m.V4Octets, ipv4Octets, defaultV4Octets))
	}

	return maskV6(a, clamp(m.V6Octets, ipv6Groups, defaultV6Octets), addr)
}

// clamp ограничивает значение диапазоном 1..upper, при выходе за него
// возвращая fallback.
func clamp(v, upper, fallback int) int {
	if v < 1 || v > upper {
		return fallback
	}

	return v
}

// parseAddr разбирает формы "ip", "ip:port" и "[ipv6]:port".
func parseAddr(addr string) netip.Addr {
	ap, err := netip.ParseAddrPort(addr)
	if err == nil {
		return ap.Addr()
	}

	a, err := netip.ParseAddr(addr)
	if err != nil {
		return netip.Addr{}
	}

	return a
}

// maskV4 заменяет октеты с позиции octets на "x".
func maskV4(a netip.Addr, octets int) string {
	parts := strings.Split(a.Unmap().String(), ".")
	for i := octets; i < len(parts); i++ {
		parts[i] = "x"
	}

	return strings.Join(parts, ".")
}

// maskV6 нормализует адрес в развёрнутую форму из 8 групп и заменяет
// группы с позиции groups на "x". Неразворачиваемый адрес возвращается как есть.
func maskV6(a netip.Addr, groups int, orig string) string {
	expanded := expandV6(strings.Trim(a.String(), "[]"))
	if len(expanded) != ipv6Groups {
		return orig
	}

	for i := groups; i < ipv6Groups; i++ {
		expanded[i] = "x"
	}

	return strings.Join(expanded, ":")
}

// expandV6 разворачивает "::" в нулевые группы.
func expandV6(text string) []string {
	parts := strings.Split(text, "::")
	const singlePart = 1

	head := splitV6Groups(parts[0])
	if len(parts) == singlePart {
		return head
	}

	tail := splitV6Groups(parts[1])

	expanded := head
	for range ipv6Groups - len(head) - len(tail) {
		expanded = append(expanded, "0")
	}

	return append(expanded, tail...)
}

// splitV6Groups разбивает текст адреса на группы, пустые группы (от "::")
// заменяя нулями.
func splitV6Groups(text string) []string {
	if text == "" {
		return nil
	}

	groups := strings.Split(text, ":")
	for i := range groups {
		if groups[i] == "" {
			groups[i] = "0"
		}
	}

	return groups
}

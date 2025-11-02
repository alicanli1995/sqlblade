package dialect

import (
	"unsafe"
)

func fastIntToString(n int) string {
	if n == 0 {
		return "0"
	}
	var buf [20]byte
	i := len(buf) - 1
	neg := n < 0
	if neg {
		n = -n
	}
	for n > 0 {
		buf[i] = byte('0' + n%10)
		i--
		n /= 10
	}
	if neg {
		buf[i] = '-'
		i--
	}
	return bytesToString(buf[i+1:])
}

func bytesToString(b []byte) string {
	return *(*string)(unsafe.Pointer(&b))
}

// OptimizedPlaceholder returns placeholder using fast int to string conversion
func (p *PostgreSQL) OptimizedPlaceholder(index int) string {
	return "$" + fastIntToString(index)
}

// OptimizedPlaceholder for MySQL/SQLite (returns "?" but kept for consistency)
func (m *MySQL) OptimizedPlaceholder(index int) string {
	return "?"
}

func (s *SQLite) OptimizedPlaceholder(index int) string {
	return "?"
}

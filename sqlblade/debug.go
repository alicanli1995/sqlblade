package sqlblade

import (
	"fmt"
	"strings"
	"time"
)

// QueryDebugger provides SQL query debugging and logging capabilities
type QueryDebugger struct {
	enabled            bool
	logger             Logger
	showArgs           bool
	colorize           bool
	indentSQL          bool
	showTiming         bool
	slowQueryThreshold time.Duration
}

// Logger interface for custom logging
type Logger interface {
	Log(query *DebugQuery)
}

// DebugQuery contains information about a query for debugging
type DebugQuery struct {
	SQL          string
	Args         []interface{}
	Table        string
	Operation    string // SELECT, INSERT, UPDATE, DELETE
	Duration     time.Duration
	RowsAffected int64
	Error        error
	Timestamp    time.Time
}

// DefaultLogger is a simple logger that prints to stdout
type DefaultLogger struct{}

func (l *DefaultLogger) Log(query *DebugQuery) {
	fmt.Println(formatQuery(query))
}

// NewQueryDebugger creates a new query debugger
func NewQueryDebugger() *QueryDebugger {
	return &QueryDebugger{
		enabled:            false,
		logger:             &DefaultLogger{},
		showArgs:           true,
		colorize:           true,
		indentSQL:          true,
		showTiming:         true,
		slowQueryThreshold: 100 * time.Millisecond,
	}
}

// Enable enables query debugging
func (qd *QueryDebugger) Enable() *QueryDebugger {
	qd.enabled = true
	return qd
}

// Disable disables query debugging
func (qd *QueryDebugger) Disable() *QueryDebugger {
	qd.enabled = false
	return qd
}

// SetLogger sets a custom logger
func (qd *QueryDebugger) SetLogger(logger Logger) *QueryDebugger {
	qd.logger = logger
	return qd
}

// ShowArgs enables/disables showing query arguments
func (qd *QueryDebugger) ShowArgs(show bool) *QueryDebugger {
	qd.showArgs = show
	return qd
}

// Colorize enables/disables colorized output
func (qd *QueryDebugger) Colorize(enable bool) *QueryDebugger {
	qd.colorize = enable
	return qd
}

// IndentSQL enables/disables SQL indentation
func (qd *QueryDebugger) IndentSQL(enable bool) *QueryDebugger {
	qd.indentSQL = enable
	return qd
}

// ShowTiming enables/disables timing information
func (qd *QueryDebugger) ShowTiming(enable bool) *QueryDebugger {
	qd.showTiming = enable
	return qd
}

// SetSlowQueryThreshold sets the threshold for slow query warnings
func (qd *QueryDebugger) SetSlowQueryThreshold(threshold time.Duration) *QueryDebugger {
	qd.slowQueryThreshold = threshold
	return qd
}

// Log logs a query if debugging is enabled
func (qd *QueryDebugger) Log(query *DebugQuery) {
	if !qd.enabled {
		return
	}
	qd.logger.Log(query)
}

var globalDebugger = NewQueryDebugger()

// EnableDebug enables global query debugging
func EnableDebug() {
	globalDebugger.Enable()
}

// DisableDebug disables global query debugging
func DisableDebug() {
	globalDebugger.Disable()
}

// SetDebugLogger sets a custom logger for global debugging
func SetDebugLogger(logger Logger) {
	globalDebugger.SetLogger(logger)
}

// ConfigureDebug configures global debug settings
func ConfigureDebug(config func(*QueryDebugger)) {
	config(globalDebugger)
	globalDebugger.Enable()
}

// formatQuery formats a query for display
func formatQuery(query *DebugQuery) string {
	var sb strings.Builder

	if query.Timestamp.IsZero() {
		query.Timestamp = time.Now()
	}

	// Header
	sb.WriteString("\n")
	sb.WriteString("═══════════════════════════════════════════════════════════════\n")
	sb.WriteString(fmt.Sprintf("SQL Query Debug - %s\n", query.Timestamp.Format("2006-01-02 15:04:05.000")))
	sb.WriteString("═══════════════════════════════════════════════════════════════\n")

	// Operation
	sb.WriteString(fmt.Sprintf("Operation: %s\n", query.Operation))
	if query.Table != "" {
		sb.WriteString(fmt.Sprintf("Table:     %s\n", query.Table))
	}

	// Timing
	if globalDebugger.showTiming && query.Duration > 0 {
		sb.WriteString(fmt.Sprintf("Duration:  %s", query.Duration))
		if query.Duration > globalDebugger.slowQueryThreshold {
			sb.WriteString(" ⚠️  SLOW QUERY")
		}
		sb.WriteString("\n")
	}

	// Rows affected
	if query.RowsAffected > 0 {
		sb.WriteString(fmt.Sprintf("Rows:      %d\n", query.RowsAffected))
	}

	// Error
	if query.Error != nil {
		sb.WriteString(fmt.Sprintf("Error:     %v\n", query.Error))
	}

	sb.WriteString("───────────────────────────────────────────────────────────────\n")

	// SQL
	sqlStr := query.SQL
	if globalDebugger.indentSQL {
		sqlStr = indentSQL(sqlStr)
	}
	sb.WriteString("SQL:\n")
	sb.WriteString(sqlStr)
	sb.WriteString("\n")

	// Args
	if globalDebugger.showArgs && len(query.Args) > 0 {
		sb.WriteString("───────────────────────────────────────────────────────────────\n")
		sb.WriteString("Parameters:\n")
		for i, arg := range query.Args {
			sb.WriteString(fmt.Sprintf("  $%d = %v (%T)\n", i+1, arg, arg))
		}
	}

	// Footer
	sb.WriteString("═══════════════════════════════════════════════════════════════\n")
	sb.WriteString("\n")

	return sb.String()
}

// indentSQL attempts to format SQL with basic indentation
func indentSQL(sql string) string {
	sql = strings.TrimSpace(sql)

	// Simple indentation based on keywords
	lines := strings.Split(sql, "\n")
	if len(lines) == 1 {
		// Single line, try to format
		sql = strings.ReplaceAll(sql, "SELECT ", "\nSELECT ")
		sql = strings.ReplaceAll(sql, " FROM ", "\nFROM ")
		sql = strings.ReplaceAll(sql, " WHERE ", "\nWHERE ")
		sql = strings.ReplaceAll(sql, " JOIN ", "\nJOIN ")
		sql = strings.ReplaceAll(sql, " LEFT JOIN ", "\nLEFT JOIN ")
		sql = strings.ReplaceAll(sql, " RIGHT JOIN ", "\nRIGHT JOIN ")
		sql = strings.ReplaceAll(sql, " INNER JOIN ", "\nINNER JOIN ")
		sql = strings.ReplaceAll(sql, " GROUP BY ", "\nGROUP BY ")
		sql = strings.ReplaceAll(sql, " HAVING ", "\nHAVING ")
		sql = strings.ReplaceAll(sql, " ORDER BY ", "\nORDER BY ")
		sql = strings.ReplaceAll(sql, " LIMIT ", "\nLIMIT ")
		sql = strings.ReplaceAll(sql, " OFFSET ", "\nOFFSET ")
		sql = strings.ReplaceAll(sql, " INSERT INTO ", "\nINSERT INTO ")
		sql = strings.ReplaceAll(sql, " UPDATE ", "\nUPDATE ")
		sql = strings.ReplaceAll(sql, " DELETE FROM ", "\nDELETE FROM ")
		sql = strings.ReplaceAll(sql, " SET ", "\nSET ")
		sql = strings.ReplaceAll(sql, " VALUES ", "\nVALUES ")
		sql = strings.ReplaceAll(sql, " RETURNING ", "\nRETURNING ")

		lines = strings.Split(sql, "\n")
	}

	var result []string
	indent := 0
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		upper := strings.ToUpper(line)

		// Decrease indent before certain keywords
		if strings.HasPrefix(upper, "FROM ") ||
			strings.HasPrefix(upper, "WHERE ") ||
			strings.HasPrefix(upper, "GROUP BY ") ||
			strings.HasPrefix(upper, "HAVING ") ||
			strings.HasPrefix(upper, "ORDER BY ") ||
			strings.HasPrefix(upper, "LIMIT ") ||
			strings.HasPrefix(upper, "OFFSET ") ||
			strings.HasPrefix(upper, "RETURNING ") {
			indent = 0
		}

		// Apply indent
		indented := strings.Repeat("  ", indent) + line
		result = append(result, indented)

		// Increase indent after certain keywords
		if strings.HasPrefix(upper, "SELECT ") ||
			strings.HasPrefix(upper, "INSERT INTO ") ||
			strings.HasPrefix(upper, "UPDATE ") ||
			strings.HasPrefix(upper, "DELETE FROM ") {
			indent = 1
		} else if strings.Contains(upper, " JOIN ") ||
			strings.HasPrefix(upper, "SET ") ||
			strings.HasPrefix(upper, "VALUES ") {
			indent = 2
		}
	}

	return strings.Join(result, "\n")
}

// SubstituteArgs substitutes parameters in SQL for easier reading
func SubstituteArgs(sql string, args []interface{}) string {
	result := sql
	for i, arg := range args {
		placeholder := fmt.Sprintf("$%d", i+1)
		var valueStr string
		switch v := arg.(type) {
		case string:
			valueStr = fmt.Sprintf("'%s'", v)
		case nil:
			valueStr = "NULL"
		default:
			valueStr = fmt.Sprintf("%v", v)
		}
		result = strings.ReplaceAll(result, placeholder, valueStr)
	}
	return result
}

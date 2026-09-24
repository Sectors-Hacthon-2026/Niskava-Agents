// Package db provides SQLite persistence using pure-Go zero-CGO modernc.org/sqlite.
// Follows Local-First Data Sovereignty (Law 4) and schema specifications.
package db

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	_ "modernc.org/sqlite"
)

// SchemaDDL defines the core database schema.
const SchemaDDL = `
PRAGMA journal_mode = WAL;
PRAGMA foreign_keys = ON;
PRAGMA synchronous = NORMAL;

CREATE TABLE IF NOT EXISTS investigations (
    id TEXT PRIMARY KEY,
    ticker TEXT NOT NULL,
    market TEXT NOT NULL DEFAULT 'IDX',
    timeframe_days INTEGER NOT NULL DEFAULT 30,
    status TEXT NOT NULL,
    started_at TEXT NOT NULL,
    completed_at TEXT,
    summary_text TEXT,
    created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS anomalies (
    id TEXT PRIMARY KEY,
    investigation_id TEXT NOT NULL REFERENCES investigations(id) ON DELETE CASCADE,
    anomaly_date TEXT NOT NULL,
    metric_type TEXT NOT NULL,
    metric_value REAL NOT NULL,
    baseline_value REAL NOT NULL,
    z_score REAL NOT NULL,
    description TEXT,
    created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS findings (
    id TEXT PRIMARY KEY,
    investigation_id TEXT NOT NULL REFERENCES investigations(id) ON DELETE CASCADE,
    title TEXT NOT NULL,
    claim_text TEXT NOT NULL,
    verification_status TEXT NOT NULL,
    confidence_score REAL NOT NULL,
    causality_status TEXT NOT NULL,
    created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS evidence_items (
    id TEXT PRIMARY KEY,
    finding_id TEXT NOT NULL REFERENCES findings(id) ON DELETE CASCADE,
    source_type TEXT NOT NULL,
    source_name TEXT NOT NULL,
    source_url TEXT NOT NULL,
    publication_date TEXT NOT NULL,
    snippet_text TEXT NOT NULL,
    created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS timeline_events (
    id TEXT PRIMARY KEY,
    investigation_id TEXT NOT NULL REFERENCES investigations(id) ON DELETE CASCADE,
    event_timestamp TEXT NOT NULL,
    event_type TEXT NOT NULL,
    headline TEXT NOT NULL,
    details TEXT,
    created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS sectors_cache (
    cache_key TEXT PRIMARY KEY,
    endpoint TEXT NOT NULL,
    payload_json TEXT NOT NULL,
    created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
    expires_at TEXT
);

CREATE TABLE IF NOT EXISTS memory_nodes (
    id TEXT PRIMARY KEY,
    label TEXT NOT NULL,
    node_type TEXT NOT NULL,
    metadata_json TEXT,
    last_observed_at TEXT NOT NULL,
    created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS memory_edges (
    source_id TEXT NOT NULL REFERENCES memory_nodes(id) ON DELETE CASCADE,
    target_id TEXT NOT NULL REFERENCES memory_nodes(id) ON DELETE CASCADE,
    relation TEXT NOT NULL,
    context_snippet TEXT,
    session_id TEXT,
    weight REAL NOT NULL DEFAULT 1.0,
    confidence_score REAL NOT NULL DEFAULT 1.0,
    last_observed_at TEXT NOT NULL,
    created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (source_id, target_id, relation)
);

CREATE TABLE IF NOT EXISTS chat_sessions (
    id TEXT PRIMARY KEY,
    title TEXT NOT NULL,
    model TEXT NOT NULL DEFAULT 'hermes',
    status TEXT NOT NULL DEFAULT 'IDLE',
    message_count INTEGER NOT NULL DEFAULT 0,
    last_message_preview TEXT,
    is_pinned INTEGER NOT NULL DEFAULT 0,
    parent_session_id TEXT REFERENCES chat_sessions(id) ON DELETE SET NULL,
    created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS chat_messages (
    id TEXT PRIMARY KEY,
    session_id TEXT NOT NULL,
    role TEXT NOT NULL,
    content TEXT NOT NULL,
    thought TEXT,
    tool_calls_json TEXT,
    status TEXT NOT NULL DEFAULT 'COMPLETED',
    created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_investigations_ticker ON investigations(ticker);
CREATE INDEX IF NOT EXISTS idx_investigations_status ON investigations(status);
CREATE INDEX IF NOT EXISTS idx_anomalies_inv_id ON anomalies(investigation_id);
CREATE INDEX IF NOT EXISTS idx_findings_inv_id ON findings(investigation_id);
CREATE INDEX IF NOT EXISTS idx_evidence_finding_id ON evidence_items(finding_id);
CREATE INDEX IF NOT EXISTS idx_timeline_inv_id ON timeline_events(investigation_id);
CREATE INDEX IF NOT EXISTS idx_sectors_cache_endpoint ON sectors_cache(endpoint);
CREATE INDEX IF NOT EXISTS idx_memory_edges_source ON memory_edges(source_id);
CREATE INDEX IF NOT EXISTS idx_memory_edges_target ON memory_edges(target_id);
CREATE INDEX IF NOT EXISTS idx_chat_sessions_updated ON chat_sessions(updated_at DESC);
CREATE INDEX IF NOT EXISTS idx_chat_sessions_parent ON chat_sessions(parent_session_id);
CREATE INDEX IF NOT EXISTS idx_chat_messages_session ON chat_messages(session_id);

CREATE TABLE IF NOT EXISTS suspension_records (
    id TEXT PRIMARY KEY,
    symbol TEXT NOT NULL,
    suspension_date TEXT NOT NULL,
    unsuspension_date TEXT,
    session TEXT,
    market_type TEXT,
    reason TEXT NOT NULL,
    pdf_url TEXT,
    created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS insider_filings (
    id TEXT PRIMARY KEY,
    symbol TEXT NOT NULL,
    holder_name TEXT NOT NULL,
    holder_type TEXT NOT NULL,
    transaction_type TEXT NOT NULL,
    transaction_date TEXT NOT NULL,
    shares_transacted REAL NOT NULL,
    price_per_share REAL,
    percentage_after_transaction REAL,
    purpose TEXT,
    created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS news_cache (
    cache_key TEXT PRIMARY KEY,
    source_type TEXT NOT NULL,
    query_or_url TEXT NOT NULL,
    content_text TEXT NOT NULL,
    metadata_json TEXT,
    expires_at TEXT,
    created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS telegram_chats (
    chat_id INTEGER PRIMARY KEY,
    current_session_id TEXT NOT NULL REFERENCES chat_sessions(id) ON DELETE CASCADE,
    user_id INTEGER NOT NULL DEFAULT 0,
    username TEXT NOT NULL DEFAULT '',
    created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_suspensions_symbol ON suspension_records(symbol, suspension_date DESC);
CREATE INDEX IF NOT EXISTS idx_insider_filings_symbol ON insider_filings(symbol, transaction_date DESC);
CREATE INDEX IF NOT EXISTS idx_news_cache_type ON news_cache(source_type);
CREATE INDEX IF NOT EXISTS idx_telegram_chats_session ON telegram_chats(current_session_id);
`

// DB wraps the SQL database pool and provides high-level domain operations.
type DB struct {
	conn *sql.DB
	Path string
}

// Investigation represents a single investigation session record.
type Investigation struct {
	ID            string  `json:"id"`
	Ticker        string  `json:"ticker"`
	Market        string  `json:"market"`
	TimeframeDays int     `json:"timeframe_days"`
	Status        string  `json:"status"` // 'PENDING', 'RUNNING', 'COMPLETED', 'FAILED'
	StartedAt     string  `json:"started_at"`
	CompletedAt   *string `json:"completed_at,omitempty"`
	SummaryText   *string `json:"summary_text,omitempty"`
	CreatedAt     string  `json:"created_at"`
}

// Anomaly represents a detected quantitative market anomaly.
type Anomaly struct {
	ID              string  `json:"id"`
	InvestigationID string  `json:"investigation_id"`
	AnomalyDate     string  `json:"anomaly_date"`
	MetricType      string  `json:"metric_type"`
	MetricValue     float64 `json:"metric_value"`
	BaselineValue   float64 `json:"baseline_value"`
	ZScore          float64 `json:"z_score"`
	Description     string  `json:"description"`
}

// Finding represents a validated claim or intelligence point.
type Finding struct {
	ID                 string  `json:"id"`
	InvestigationID    string  `json:"investigation_id"`
	Title              string  `json:"title"`
	ClaimText          string  `json:"claim_text"`
	VerificationStatus string  `json:"verification_status"` // 'SUPPORTED', 'UNCERTAIN', 'CONTRADICTED'
	ConfidenceScore    float64 `json:"confidence_score"`    // Discrete rubric (1.00, 0.95, etc.)
	CausalityStatus    string  `json:"causality_status"`    // 'LIKELY_CATALYST', 'PRECEDED_ANNOUNCEMENT', 'UNEXPLAINED_BY_NEWS'
}

// Open initializes and migrates the SQLite database at dbPath.
func Open(dbPath string) (*DB, error) {
	absPath, err := filepath.Abs(dbPath)
	if err != nil {
		absPath = dbPath
	}

	// Ensure parent directory exists
	dir := filepath.Dir(absPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create db directory %s: %w", dir, err)
	}

	dsn := fmt.Sprintf("%s?_pragma=journal_mode(WAL)&_pragma=foreign_keys(ON)&_pragma=synchronous(NORMAL)", absPath)
	conn, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open sqlite database at %s: %w", absPath, err)
	}

	// Migration: rename legacy osint_cache -> news_cache if exists
	_, _ = conn.Exec("ALTER TABLE osint_cache RENAME TO news_cache;")

	// Run migration DDL
	if _, err := conn.Exec(SchemaDDL); err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to apply database migrations: %w", err)
	}

	// Self-healing migration for confidence_score column if existing database was created prior
	_, _ = conn.Exec("ALTER TABLE memory_edges ADD COLUMN confidence_score REAL NOT NULL DEFAULT 1.0;")

	// Ensure memory_edges is decoupled from investigations FK so CHAT sessions can persist edges
	var hasInvFK bool
	rows, err := conn.Query("PRAGMA foreign_key_list(memory_edges)")
	if err == nil {
		for rows.Next() {
			var id, seq int
			var table, from, to, onUpdate, onDelete, match string
			if scanErr := rows.Scan(&id, &seq, &table, &from, &to, &onUpdate, &onDelete, &match); scanErr == nil {
				if table == "investigations" {
					hasInvFK = true
					break
				}
			}
		}
		rows.Close()
	}
	if hasInvFK {
		migrationSQL := `
			PRAGMA foreign_keys = OFF;
			CREATE TABLE IF NOT EXISTS memory_edges_v2 (
				source_id TEXT NOT NULL REFERENCES memory_nodes(id) ON DELETE CASCADE,
				target_id TEXT NOT NULL REFERENCES memory_nodes(id) ON DELETE CASCADE,
				relation TEXT NOT NULL,
				context_snippet TEXT,
				session_id TEXT,
				weight REAL NOT NULL DEFAULT 1.0,
				confidence_score REAL NOT NULL DEFAULT 1.0,
				last_observed_at TEXT NOT NULL,
				created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
				PRIMARY KEY (source_id, target_id, relation)
			);
			INSERT OR IGNORE INTO memory_edges_v2 SELECT source_id, target_id, relation, context_snippet, session_id, weight, confidence_score, last_observed_at, created_at FROM memory_edges;
			DROP TABLE memory_edges;
			ALTER TABLE memory_edges_v2 RENAME TO memory_edges;
			CREATE INDEX IF NOT EXISTS idx_mem_edges_src ON memory_edges(source_id);
			CREATE INDEX IF NOT EXISTS idx_mem_edges_tgt ON memory_edges(target_id);
			CREATE INDEX IF NOT EXISTS idx_mem_edges_session ON memory_edges(session_id);
			PRAGMA foreign_keys = ON;
		`
		_, _ = conn.Exec(migrationSQL)
	}

	// Self-healing migration for chat_messages status column
	_, _ = conn.Exec("ALTER TABLE chat_messages ADD COLUMN status TEXT NOT NULL DEFAULT 'COMPLETED';")

	// Self-healing backfill from existing chat_messages into chat_sessions
	backfillSQL := `
		INSERT OR IGNORE INTO chat_sessions (id, title, model, status, message_count, last_message_preview, created_at, updated_at)
		SELECT 
			session_id,
			COALESCE(SUBSTR(MIN(CASE WHEN role = 'user' THEN content END), 1, 40), session_id) as title,
			'hermes',
			'IDLE',
			COUNT(id) as message_count,
			COALESCE(MAX(content), ''),
			MIN(created_at),
			MAX(created_at)
		FROM chat_messages
		GROUP BY session_id;
	`
	_, _ = conn.Exec(backfillSQL)

	// Self-healing reset for zombie BUSY chat sessions (e.g. from server crash or abrupt restart)
	_, _ = conn.Exec("UPDATE chat_sessions SET status = 'IDLE' WHERE status = 'BUSY';")

	return &DB{conn: conn, Path: absPath}, nil
}

// Close closes the underlying database connection.
func (d *DB) Close() error {
	if d.conn != nil {
		return d.conn.Close()
	}
	return nil
}

// CreateInvestigation creates a new investigation session record.
func (d *DB) CreateInvestigation(inv *Investigation) error {
	if inv.StartedAt == "" {
		inv.StartedAt = time.Now().UTC().Format(time.RFC3339)
	}
	if inv.Status == "" {
		inv.Status = "PENDING"
	}
	if inv.Market == "" {
		inv.Market = "IDX"
	}
	if inv.TimeframeDays == 0 {
		inv.TimeframeDays = 30
	}

	query := `
		INSERT INTO investigations (id, ticker, market, timeframe_days, status, started_at, summary_text)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`
	_, err := d.conn.Exec(query, inv.ID, inv.Ticker, inv.Market, inv.TimeframeDays, inv.Status, inv.StartedAt, inv.SummaryText)
	if err != nil {
		return fmt.Errorf("failed to insert investigation %s: %w", inv.ID, err)
	}
	return nil
}

// UpdateInvestigationStatus transitions an investigation status and optional summary.
func (d *DB) UpdateInvestigationStatus(id, status string, summaryText *string) error {
	now := time.Now().UTC().Format(time.RFC3339)
	query := `
		UPDATE investigations 
		SET status = ?, completed_at = CASE WHEN ? IN ('COMPLETED', 'FAILED') THEN ? ELSE completed_at END,
		    summary_text = COALESCE(?, summary_text)
		WHERE id = ?
	`
	_, err := d.conn.Exec(query, status, status, now, summaryText, id)
	if err != nil {
		return fmt.Errorf("failed to update status for investigation %s: %w", id, err)
	}
	return nil
}

// ListInvestigations returns past investigation sessions ordered by start time desc.
func (d *DB) ListInvestigations(limit int) ([]Investigation, error) {
	if limit <= 0 {
		limit = 50
	}
	query := `
		SELECT id, ticker, market, timeframe_days, status, started_at, completed_at, summary_text, created_at
		FROM investigations
		ORDER BY started_at DESC
		LIMIT ?
	`
	rows, err := d.conn.Query(query, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to list investigations: %w", err)
	}
	defer rows.Close()

	var results []Investigation
	for rows.Next() {
		var inv Investigation
		if err := rows.Scan(&inv.ID, &inv.Ticker, &inv.Market, &inv.TimeframeDays, &inv.Status, &inv.StartedAt, &inv.CompletedAt, &inv.SummaryText, &inv.CreatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan investigation row: %w", err)
		}
		results = append(results, inv)
	}
	return results, nil
}

// GetInvestigation retrieves a specific session by its ID.
func (d *DB) GetInvestigation(id string) (*Investigation, error) {
	query := `
		SELECT id, ticker, market, timeframe_days, status, started_at, completed_at, summary_text, created_at
		FROM investigations
		WHERE id = ?
	`
	row := d.conn.QueryRow(query, id)
	var inv Investigation
	if err := row.Scan(&inv.ID, &inv.Ticker, &inv.Market, &inv.TimeframeDays, &inv.Status, &inv.StartedAt, &inv.CompletedAt, &inv.SummaryText, &inv.CreatedAt); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get investigation %s: %w", id, err)
	}
	return &inv, nil
}

// ListFindingsByInvestigation retrieves all findings for a given session.
func (d *DB) ListFindingsByInvestigation(invID string) ([]Finding, error) {
	query := `
		SELECT id, investigation_id, title, claim_text, verification_status, confidence_score, causality_status
		FROM findings
		WHERE investigation_id = ?
		ORDER BY confidence_score DESC
	`
	rows, err := d.conn.Query(query, invID)
	if err != nil {
		return nil, fmt.Errorf("failed to list findings for %s: %w", invID, err)
	}
	defer rows.Close()

	var results []Finding
	for rows.Next() {
		var f Finding
		if err := rows.Scan(&f.ID, &f.InvestigationID, &f.Title, &f.ClaimText, &f.VerificationStatus, &f.ConfidenceScore, &f.CausalityStatus); err != nil {
			return nil, fmt.Errorf("failed to scan finding row: %w", err)
		}
		results = append(results, f)
	}
	if results == nil {
		results = []Finding{}
	}
	return results, nil
}

// GetAnomaliesByInvestigation returns all quantitative anomalies for a specific investigation.
func (d *DB) GetAnomaliesByInvestigation(invID string) ([]Anomaly, error) {
	query := `
		SELECT id, investigation_id, anomaly_date, metric_type, metric_value, baseline_value, z_score, description
		FROM anomalies
		WHERE investigation_id = ?
		ORDER BY anomaly_date ASC
	`
	rows, err := d.conn.Query(query, invID)
	if err != nil {
		return nil, fmt.Errorf("failed to get anomalies for %s: %w", invID, err)
	}
	defer rows.Close()

	var results []Anomaly
	for rows.Next() {
		var a Anomaly
		if err := rows.Scan(&a.ID, &a.InvestigationID, &a.AnomalyDate, &a.MetricType, &a.MetricValue, &a.BaselineValue, &a.ZScore, &a.Description); err != nil {
			return nil, fmt.Errorf("failed to scan anomaly: %w", err)
		}
		results = append(results, a)
	}
	if results == nil {
		results = []Anomaly{}
	}
	return results, nil
}

// CreateAnomaly records a quantitative anomaly in the database.
func (d *DB) CreateAnomaly(a *Anomaly) error {
	query := `
		INSERT INTO anomalies (id, investigation_id, anomaly_date, metric_type, metric_value, baseline_value, z_score, description)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`
	_, err := d.conn.Exec(query, a.ID, a.InvestigationID, a.AnomalyDate, a.MetricType, a.MetricValue, a.BaselineValue, a.ZScore, a.Description)
	if err != nil {
		return fmt.Errorf("failed to insert anomaly: %w", err)
	}
	return nil
}

// CreateFinding records an investigation finding in the database.
func (d *DB) CreateFinding(f *Finding) error {
	query := `
		INSERT INTO findings (id, investigation_id, title, claim_text, verification_status, confidence_score, causality_status)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`
	_, err := d.conn.Exec(query, f.ID, f.InvestigationID, f.Title, f.ClaimText, f.VerificationStatus, f.ConfidenceScore, f.CausalityStatus)
	if err != nil {
		return fmt.Errorf("failed to insert finding: %w", err)
	}
	return nil
}

// Conn returns the underlying database connection pool.
func (d *DB) Conn() *sql.DB {
	return d.conn
}

// ChatSession represents an explicit conversational research session (Hermes/OpenCode pattern).
type ChatSession struct {
	ID                 string  `json:"id"`
	Title              string  `json:"title"`
	Model              string  `json:"model"`
	Status             string  `json:"status"` // "IDLE", "BUSY", "ERROR"
	MessageCount       int     `json:"message_count"`
	LastMessagePreview string  `json:"last_message_preview"`
	IsPinned           bool    `json:"is_pinned"`
	ParentSessionID    *string `json:"parent_session_id,omitempty"`
	CreatedAt          string  `json:"created_at"`
	UpdatedAt          string  `json:"updated_at"`
}

// ChatMessage represents a single conversational turn in a research session.
type ChatMessage struct {
	ID            string  `json:"id"`
	SessionID     string  `json:"session_id"`
	Role          string  `json:"role"` // "user", "assistant", "system", "tool"
	Content       string  `json:"content"`
	Thought       *string `json:"thought,omitempty"`
	ToolCallsJSON *string `json:"tool_calls_json,omitempty"`
	Status        string  `json:"status"` // "COMPLETED", "ABORTED", "FAILED"
	CreatedAt     string  `json:"created_at"`
}

// SaveChatMessage records a user or assistant message to SQLite and touches parent session.
func (d *DB) SaveChatMessage(msg *ChatMessage) error {
	if msg.Status == "" {
		msg.Status = "COMPLETED"
	}
	query := `
		INSERT INTO chat_messages (id, session_id, role, content, thought, tool_calls_json, status, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, COALESCE(?, CURRENT_TIMESTAMP))
	`
	var createdAt interface{} = msg.CreatedAt
	if msg.CreatedAt == "" {
		createdAt = nil
	}
	_, err := d.conn.Exec(query, msg.ID, msg.SessionID, msg.Role, msg.Content, msg.Thought, msg.ToolCallsJSON, msg.Status, createdAt)
	if err != nil {
		return fmt.Errorf("failed to save chat message: %w", err)
	}

	// Touch and upsert parent chat session
	now := time.Now().UTC().Format(time.RFC3339)
	preview := msg.Content
	if len(preview) > 120 {
		preview = preview[:117] + "..."
	}

	defaultTitle := msg.Content
	if len(defaultTitle) > 40 {
		defaultTitle = defaultTitle[:37] + "..."
	}
	if defaultTitle == "" {
		defaultTitle = "Sesi Riset Pasar"
	}

	upsertQuery := `
		INSERT INTO chat_sessions (id, title, model, status, message_count, last_message_preview, is_pinned, created_at, updated_at)
		VALUES (?, ?, 'hermes', 'IDLE', 1, ?, 0, ?, ?)
		ON CONFLICT(id) DO UPDATE SET
			title = CASE WHEN title = 'Sesi Riset Pasar' OR title = '' OR title IS NULL THEN excluded.title ELSE title END,
			message_count = message_count + 1,
			last_message_preview = excluded.last_message_preview,
			updated_at = excluded.updated_at
	`
	_, _ = d.conn.Exec(upsertQuery, msg.SessionID, defaultTitle, preview, now, now)

	return nil
}

// GetChatHistory retrieves messages for a specific session ordered chronologically.
func (d *DB) GetChatHistory(sessionID string, limit int) ([]ChatMessage, error) {
	if limit <= 0 {
		limit = 50
	}
	query := `
		SELECT id, session_id, role, content, thought, tool_calls_json, COALESCE(status, 'COMPLETED'), created_at
		FROM chat_messages
		WHERE session_id = ?
		ORDER BY created_at ASC
		LIMIT ?
	`
	rows, err := d.conn.Query(query, sessionID, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get chat history for session %s: %w", sessionID, err)
	}
	defer rows.Close()

	var history []ChatMessage
	for rows.Next() {
		var m ChatMessage
		if err := rows.Scan(&m.ID, &m.SessionID, &m.Role, &m.Content, &m.Thought, &m.ToolCallsJSON, &m.Status, &m.CreatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan chat message: %w", err)
		}
		history = append(history, m)
	}
	return history, nil
}

// ChatSearchResult represents a matched chat message across conversation history.
type ChatSearchResult struct {
	MessageID    string `json:"message_id"`
	SessionID    string `json:"session_id"`
	SessionTitle string `json:"session_title"`
	Role         string `json:"role"`
	Content      string `json:"content"`
	CreatedAt    string `json:"created_at"`
}

// SearchChatMessages searches for messages matching a text query across all chat sessions.
func (d *DB) SearchChatMessages(query string, limit int) ([]ChatSearchResult, error) {
	if limit <= 0 {
		limit = 50
	}
	trimmed := strings.TrimSpace(query)
	if trimmed == "" {
		return []ChatSearchResult{}, nil
	}

	searchSQL := `
		SELECT 
			m.id, 
			m.session_id, 
			COALESCE(s.title, 'Sesi Riset Pasar') AS session_title, 
			m.role, 
			m.content, 
			m.created_at
		FROM chat_messages m
		LEFT JOIN chat_sessions s ON m.session_id = s.id
		WHERE m.content LIKE ?
		ORDER BY m.created_at DESC
		LIMIT ?
	`
	rows, err := d.conn.Query(searchSQL, "%"+trimmed+"%", limit)
	if err != nil {
		return nil, fmt.Errorf("failed to search chat messages: %w", err)
	}
	defer rows.Close()

	var results []ChatSearchResult
	for rows.Next() {
		var r ChatSearchResult
		if err := rows.Scan(&r.MessageID, &r.SessionID, &r.SessionTitle, &r.Role, &r.Content, &r.CreatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan chat search result: %w", err)
		}
		results = append(results, r)
	}
	if results == nil {
		results = []ChatSearchResult{}
	}
	return results, nil
}

// CreateChatSession inserts a new chat session record.
func (d *DB) CreateChatSession(s *ChatSession) error {
	if s.Status == "" {
		s.Status = "IDLE"
	}
	if s.Model == "" {
		s.Model = "hermes"
	}
	now := time.Now().UTC().Format(time.RFC3339)
	if s.CreatedAt == "" {
		s.CreatedAt = now
	}
	if s.UpdatedAt == "" {
		s.UpdatedAt = now
	}
	pinnedInt := 0
	if s.IsPinned {
		pinnedInt = 1
	}

	query := `
		INSERT INTO chat_sessions (id, title, model, status, message_count, last_message_preview, is_pinned, parent_session_id, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`
	_, err := d.conn.Exec(query, s.ID, s.Title, s.Model, s.Status, s.MessageCount, s.LastMessagePreview, pinnedInt, s.ParentSessionID, s.CreatedAt, s.UpdatedAt)
	if err != nil {
		return fmt.Errorf("failed to create chat session %s: %w", s.ID, err)
	}
	return nil
}

// ListChatSessions lists chat sessions ordered by is_pinned desc and updated_at desc, supporting pagination and search.
func (d *DB) ListChatSessions(limit, offset int, search string) ([]ChatSession, int, error) {
	if limit <= 0 {
		limit = 50
	}
	if offset < 0 {
		offset = 0
	}

	var (
		whereClause string
		args        []interface{}
	)
	trimmed := strings.TrimSpace(search)
	if trimmed != "" {
		whereClause = "WHERE title LIKE ? OR last_message_preview LIKE ? OR id LIKE ?"
		pattern := "%" + trimmed + "%"
		args = append(args, pattern, pattern, pattern)
	}

	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM chat_sessions %s", whereClause)
	var total int
	if err := d.conn.QueryRow(countQuery, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("failed to count chat sessions: %w", err)
	}

	query := fmt.Sprintf(`
		SELECT id, title, model, status, message_count, COALESCE(last_message_preview, ''), is_pinned, parent_session_id, created_at, updated_at
		FROM chat_sessions
		%s
		ORDER BY is_pinned DESC, updated_at DESC
		LIMIT ? OFFSET ?
	`, whereClause)

	queryArgs := append(args, limit, offset)
	rows, err := d.conn.Query(query, queryArgs...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list chat sessions: %w", err)
	}
	defer rows.Close()

	var sessions []ChatSession
	for rows.Next() {
		var (
			s        ChatSession
			isPinned int
			parentID sql.NullString
		)
		if err := rows.Scan(&s.ID, &s.Title, &s.Model, &s.Status, &s.MessageCount, &s.LastMessagePreview, &isPinned, &parentID, &s.CreatedAt, &s.UpdatedAt); err != nil {
			return nil, 0, fmt.Errorf("failed to scan chat session: %w", err)
		}
		s.IsPinned = isPinned == 1
		if parentID.Valid {
			s.ParentSessionID = &parentID.String
		}
		sessions = append(sessions, s)
	}

	return sessions, total, nil
}

// GetChatSession retrieves a single chat session by ID.
func (d *DB) GetChatSession(id string) (*ChatSession, error) {
	query := `
		SELECT id, title, model, status, message_count, COALESCE(last_message_preview, ''), is_pinned, parent_session_id, created_at, updated_at
		FROM chat_sessions
		WHERE id = ?
	`
	row := d.conn.QueryRow(query, id)
	var (
		s        ChatSession
		isPinned int
		parentID sql.NullString
	)
	if err := row.Scan(&s.ID, &s.Title, &s.Model, &s.Status, &s.MessageCount, &s.LastMessagePreview, &isPinned, &parentID, &s.CreatedAt, &s.UpdatedAt); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get chat session %s: %w", id, err)
	}
	s.IsPinned = isPinned == 1
	if parentID.Valid {
		s.ParentSessionID = &parentID.String
	}
	return &s, nil
}

// UpdateChatSession updates mutable properties of a chat session.
func (d *DB) UpdateChatSession(id string, title *string, isPinned *bool, status *string) error {
	now := time.Now().UTC().Format(time.RFC3339)
	var (
		clauses []string
		args    []interface{}
	)

	if title != nil {
		clauses = append(clauses, "title = ?")
		args = append(args, *title)
	}
	if isPinned != nil {
		pinnedInt := 0
		if *isPinned {
			pinnedInt = 1
		}
		clauses = append(clauses, "is_pinned = ?")
		args = append(args, pinnedInt)
	}
	if status != nil {
		clauses = append(clauses, "status = ?")
		args = append(args, *status)
	}

	if len(clauses) == 0 {
		return nil
	}

	clauses = append(clauses, "updated_at = ?")
	args = append(args, now)
	args = append(args, id)

	query := fmt.Sprintf("UPDATE chat_sessions SET %s WHERE id = ?", strings.Join(clauses, ", "))
	res, err := d.conn.Exec(query, args...)
	if err != nil {
		return fmt.Errorf("failed to update chat session %s: %w", id, err)
	}
	rowsAffected, _ := res.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("chat session %s not found", id)
	}
	return nil
}

// TouchChatSession updates a session's updated_at timestamp and preview snippet.
func (d *DB) TouchChatSession(id string, preview string) error {
	now := time.Now().UTC().Format(time.RFC3339)
	if len(preview) > 120 {
		preview = preview[:117] + "..."
	}
	query := `
		UPDATE chat_sessions
		SET updated_at = ?, last_message_preview = CASE WHEN ? != '' THEN ? ELSE last_message_preview END
		WHERE id = ?
	`
	_, err := d.conn.Exec(query, now, preview, preview, id)
	return err
}

// DeleteChatSession permanently removes a session and cascades deletion to chat_messages and memory_edges.
func (d *DB) DeleteChatSession(id string) error {
	tx, err := d.conn.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	if _, err := tx.Exec("DELETE FROM chat_messages WHERE session_id = ?", id); err != nil {
		return fmt.Errorf("failed to delete messages for session %s: %w", id, err)
	}
	if _, err := tx.Exec("DELETE FROM memory_edges WHERE session_id = ?", id); err != nil {
		return fmt.Errorf("failed to delete memory edges for session %s: %w", id, err)
	}
	res, err := tx.Exec("DELETE FROM chat_sessions WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("failed to delete chat session %s: %w", id, err)
	}
	rowsAffected, _ := res.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("chat session %s not found", id)
	}

	return tx.Commit()
}

// ClearAllChatSessions permanently deletes all chat sessions, messages, and associated session memory edges.
func (d *DB) ClearAllChatSessions() error {
	tx, err := d.conn.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	if _, err := tx.Exec("DELETE FROM chat_messages"); err != nil {
		return fmt.Errorf("failed to clear chat_messages: %w", err)
	}
	if _, err := tx.Exec("DELETE FROM memory_edges WHERE session_id IS NOT NULL"); err != nil {
		return fmt.Errorf("failed to clear session memory_edges: %w", err)
	}
	if _, err := tx.Exec("DELETE FROM chat_sessions"); err != nil {
		return fmt.Errorf("failed to clear chat_sessions: %w", err)
	}

	return tx.Commit()
}

// ForkChatSession clones conversation history up to upToMessageID into a new branched session (OpenCode pattern).
func (d *DB) ForkChatSession(sourceID, newID, newTitle, upToMessageID string) error {
	tx, err := d.conn.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	var sourceModel string
	err = tx.QueryRow("SELECT model FROM chat_sessions WHERE id = ?", sourceID).Scan(&sourceModel)
	if err != nil {
		return fmt.Errorf("source session %s not found: %w", sourceID, err)
	}

	now := time.Now().UTC().Format(time.RFC3339)
	if newTitle == "" {
		newTitle = fmt.Sprintf("Cabang dari %s", sourceID)
	}

	var (
		msgQuery string
		msgArgs  []interface{}
	)
	if upToMessageID != "" {
		msgQuery = `
			SELECT id, role, content, thought, tool_calls_json, COALESCE(status, 'COMPLETED'), created_at
			FROM chat_messages
			WHERE session_id = ? AND created_at <= (SELECT created_at FROM chat_messages WHERE id = ?)
			ORDER BY created_at ASC
		`
		msgArgs = []interface{}{sourceID, upToMessageID}
	} else {
		msgQuery = `
			SELECT id, role, content, thought, tool_calls_json, COALESCE(status, 'COMPLETED'), created_at
			FROM chat_messages
			WHERE session_id = ?
			ORDER BY created_at ASC
		`
		msgArgs = []interface{}{sourceID}
	}

	rows, err := tx.Query(msgQuery, msgArgs...)
	if err != nil {
		return fmt.Errorf("failed to query messages for fork: %w", err)
	}
	defer rows.Close()

	var (
		copiedMsgs  []ChatMessage
		lastPreview string
	)
	for rows.Next() {
		var m ChatMessage
		if err := rows.Scan(&m.ID, &m.Role, &m.Content, &m.Thought, &m.ToolCallsJSON, &m.Status, &m.CreatedAt); err != nil {
			return fmt.Errorf("failed to scan message for fork: %w", err)
		}
		copiedMsgs = append(copiedMsgs, m)
		lastPreview = m.Content
	}
	msgCount := len(copiedMsgs)
	if len(lastPreview) > 120 {
		lastPreview = lastPreview[:117] + "..."
	}

	insertSessionQuery := `
		INSERT INTO chat_sessions (id, title, model, status, message_count, last_message_preview, is_pinned, parent_session_id, created_at, updated_at)
		VALUES (?, ?, ?, 'IDLE', ?, ?, 0, ?, ?, ?)
	`
	if _, err := tx.Exec(insertSessionQuery, newID, newTitle, sourceModel, msgCount, lastPreview, sourceID, now, now); err != nil {
		return fmt.Errorf("failed to insert forked session: %w", err)
	}

	insertMsgQuery := `
		INSERT INTO chat_messages (id, session_id, role, content, thought, tool_calls_json, status, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`
	for idx, m := range copiedMsgs {
		newMsgID := fmt.Sprintf("%s-M%d", newID, idx+1)
		if _, err := tx.Exec(insertMsgQuery, newMsgID, newID, m.Role, m.Content, m.Thought, m.ToolCallsJSON, m.Status, m.CreatedAt); err != nil {
			return fmt.Errorf("failed to copy message during fork: %w", err)
		}
	}

	return tx.Commit()
}

// ClearSessionHistory deletes all messages for a specific session without removing the session record.
func (d *DB) ClearSessionHistory(sessionID string) error {
	tx, err := d.conn.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	if _, err := tx.Exec("DELETE FROM chat_messages WHERE session_id = ?", sessionID); err != nil {
		return fmt.Errorf("failed to delete messages: %w", err)
	}
	if _, err := tx.Exec("DELETE FROM memory_edges WHERE session_id = ?", sessionID); err != nil {
		return fmt.Errorf("failed to delete memory edges: %w", err)
	}

	now := time.Now().UTC().Format(time.RFC3339)
	_, err = tx.Exec("UPDATE chat_sessions SET message_count = 0, last_message_preview = '', updated_at = ? WHERE id = ?", now, sessionID)
	if err != nil {
		return fmt.Errorf("failed to reset session stats: %w", err)
	}

	return tx.Commit()
}

// MemoryNode represents an entity in the local conversational graph memory.
type MemoryNode struct {
	ID             string  `json:"id"`
	Label          string  `json:"label"`
	NodeType       string  `json:"node_type"`
	MetadataJSON   *string `json:"metadata_json,omitempty"`
	LastObservedAt string  `json:"last_observed_at"`
	CreatedAt      string  `json:"created_at"`
}

// MemoryEdge represents a directed relationship between two entities.
type MemoryEdge struct {
	SourceID        string  `json:"source_id"`
	TargetID        string  `json:"target_id"`
	Relation        string  `json:"relation"`
	ContextSnippet  *string `json:"context_snippet,omitempty"`
	SessionID       *string `json:"session_id,omitempty"`
	Weight          float64 `json:"weight"`
	ConfidenceScore float64 `json:"confidence_score"`
	LastObservedAt  string  `json:"last_observed_at"`
	CreatedAt       string  `json:"created_at"`
}

// MemoryGraphFilter defines parameters for filtering nodes and edges.
type MemoryGraphFilter struct {
	SessionID string   `json:"session_id"`
	Ticker    string   `json:"ticker"`
	Depth     int      `json:"depth"`
	NodeTypes []string `json:"node_types"`
	MinWeight float64  `json:"min_weight"`
}

// HubNode represents a node with its degree in the memory graph.
type HubNode struct {
	ID       string `json:"id"`
	Label    string `json:"label"`
	NodeType string `json:"node_type"`
	Degree   int    `json:"degree"`
}

// MemoryGraphStats provides high-level aggregation of memory nodes and edges.
type MemoryGraphStats struct {
	TotalNodes  int            `json:"total_nodes"`
	TotalEdges  int            `json:"total_edges"`
	NodeTypes   map[string]int `json:"node_types"`
	TopHubNodes []HubNode      `json:"top_hub_nodes"`
}

// SaveMemoryNode inserts or updates a memory node.
func (d *DB) SaveMemoryNode(node *MemoryNode) error {
	query := `
		INSERT INTO memory_nodes (id, label, node_type, metadata_json, last_observed_at)
		VALUES (?, ?, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET
			label = excluded.label,
			node_type = excluded.node_type,
			metadata_json = COALESCE(excluded.metadata_json, memory_nodes.metadata_json),
			last_observed_at = excluded.last_observed_at
	`
	_, err := d.conn.Exec(query, node.ID, node.Label, node.NodeType, node.MetadataJSON, node.LastObservedAt)
	if err != nil {
		return fmt.Errorf("failed to save memory node: %w", err)
	}
	return nil
}

// SaveMemoryEdge inserts or updates a memory edge.
func (d *DB) SaveMemoryEdge(edge *MemoryEdge) error {
	query := `
		INSERT INTO memory_edges (source_id, target_id, relation, context_snippet, session_id, weight, confidence_score, last_observed_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(source_id, target_id, relation) DO UPDATE SET
			context_snippet = COALESCE(excluded.context_snippet, memory_edges.context_snippet),
			session_id = COALESCE(excluded.session_id, memory_edges.session_id),
			weight = excluded.weight,
			confidence_score = excluded.confidence_score,
			last_observed_at = excluded.last_observed_at
	`
	_, err := d.conn.Exec(query, edge.SourceID, edge.TargetID, edge.Relation, edge.ContextSnippet, edge.SessionID, edge.Weight, edge.ConfidenceScore, edge.LastObservedAt)
	if err != nil {
		return fmt.Errorf("failed to save memory edge: %w", err)
	}
	return nil
}

// GetMemoryGraph retrieves nodes and edges, optionally filtered by sessionID.
func (d *DB) GetMemoryGraph(sessionID string) ([]MemoryNode, []MemoryEdge, error) {
	nodeQuery := `
		SELECT id, label, node_type, metadata_json, last_observed_at, created_at
		FROM memory_nodes
		ORDER BY last_observed_at DESC
	`
	nodeRows, err := d.conn.Query(nodeQuery)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to query memory nodes: %w", err)
	}
	defer nodeRows.Close()

	var nodes []MemoryNode
	for nodeRows.Next() {
		var n MemoryNode
		if err := nodeRows.Scan(&n.ID, &n.Label, &n.NodeType, &n.MetadataJSON, &n.LastObservedAt, &n.CreatedAt); err != nil {
			return nil, nil, fmt.Errorf("failed to scan memory node: %w", err)
		}
		nodes = append(nodes, n)
	}

	var (
		edgeQuery string
		edgeArgs  []interface{}
	)
	if sessionID != "" {
		edgeQuery = `
			SELECT source_id, target_id, relation, context_snippet, session_id,
			       weight, COALESCE(confidence_score, 1.0), last_observed_at, created_at
			FROM memory_edges
			WHERE session_id = ?
			ORDER BY weight DESC
		`
		edgeArgs = append(edgeArgs, sessionID)
	} else {
		edgeQuery = `
			SELECT source_id, target_id, relation, context_snippet, session_id,
			       weight, COALESCE(confidence_score, 1.0), last_observed_at, created_at
			FROM memory_edges
			ORDER BY weight DESC
		`
	}

	edgeRows, err := d.conn.Query(edgeQuery, edgeArgs...)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to query memory edges: %w", err)
	}
	defer edgeRows.Close()

	var edges []MemoryEdge
	for edgeRows.Next() {
		var e MemoryEdge
		if err := edgeRows.Scan(&e.SourceID, &e.TargetID, &e.Relation, &e.ContextSnippet, &e.SessionID, &e.Weight, &e.ConfidenceScore, &e.LastObservedAt, &e.CreatedAt); err != nil {
			return nil, nil, fmt.Errorf("failed to scan memory edge: %w", err)
		}
		edges = append(edges, e)
	}

	return nodes, edges, nil
}

// GetFilteredMemoryGraph retrieves memory nodes and edges filtered by ego-network, session, types, and weight.
func (d *DB) GetFilteredMemoryGraph(filter MemoryGraphFilter) ([]MemoryNode, []MemoryEdge, error) {
	// 1. Fetch all candidate edges
	edgeQuery := `
		SELECT source_id, target_id, relation, context_snippet, session_id,
		       weight, COALESCE(confidence_score, 1.0), last_observed_at, created_at
		FROM memory_edges
		WHERE (? = '' OR session_id = ?) AND weight >= ?
		ORDER BY weight DESC
	`
	edgeRows, err := d.conn.Query(edgeQuery, filter.SessionID, filter.SessionID, filter.MinWeight)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to query filtered memory edges: %w", err)
	}
	defer edgeRows.Close()

	var allEdges []MemoryEdge
	for edgeRows.Next() {
		var e MemoryEdge
		if err := edgeRows.Scan(&e.SourceID, &e.TargetID, &e.Relation, &e.ContextSnippet, &e.SessionID, &e.Weight, &e.ConfidenceScore, &e.LastObservedAt, &e.CreatedAt); err != nil {
			return nil, nil, fmt.Errorf("failed to scan memory edge: %w", err)
		}
		allEdges = append(allEdges, e)
	}

	// 2. Fetch all nodes
	nodeQuery := `
		SELECT id, label, node_type, metadata_json, last_observed_at, created_at
		FROM memory_nodes
		ORDER BY last_observed_at DESC
	`
	nodeRows, err := d.conn.Query(nodeQuery)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to query memory nodes: %w", err)
	}
	defer nodeRows.Close()

	nodesMap := make(map[string]MemoryNode)
	for nodeRows.Next() {
		var n MemoryNode
		if err := nodeRows.Scan(&n.ID, &n.Label, &n.NodeType, &n.MetadataJSON, &n.LastObservedAt, &n.CreatedAt); err != nil {
			return nil, nil, fmt.Errorf("failed to scan memory node: %w", err)
		}
		nodesMap[n.ID] = n
	}

	// 3. Filter by NodeTypes if provided
	typeSet := make(map[string]bool)
	for _, nt := range filter.NodeTypes {
		if trimmed := strings.TrimSpace(nt); trimmed != "" {
			typeSet[strings.ToUpper(trimmed)] = true
		}
	}

	if len(typeSet) > 0 {
		for id, n := range nodesMap {
			if !typeSet[strings.ToUpper(n.NodeType)] {
				delete(nodesMap, id)
			}
		}
	}

	// 4. If Ticker specified: Perform ego-network BFS expansion up to Depth
	if filter.Ticker != "" {
		tickerUpper := strings.ToUpper(strings.TrimSpace(filter.Ticker))
		depth := filter.Depth
		if depth <= 0 {
			depth = 1
		} else if depth > 2 {
			depth = 2
		}

		visitedNodes := make(map[string]bool)
		currentLevel := make(map[string]bool)

		// Seed initial nodes
		for id, n := range nodesMap {
			if strings.EqualFold(n.Label, tickerUpper) || strings.EqualFold(n.ID, tickerUpper) ||
				strings.HasSuffix(strings.ToUpper(n.ID), ":"+tickerUpper) || strings.HasPrefix(strings.ToUpper(n.ID), tickerUpper+":") {
				currentLevel[id] = true
				visitedNodes[id] = true
			}
		}

		// Traverse hops
		for hop := 0; hop < depth; hop++ {
			nextLevel := make(map[string]bool)
			for _, e := range allEdges {
				if currentLevel[e.SourceID] {
					if _, exists := nodesMap[e.TargetID]; exists && !visitedNodes[e.TargetID] {
						visitedNodes[e.TargetID] = true
						nextLevel[e.TargetID] = true
					}
				}
				if currentLevel[e.TargetID] {
					if _, exists := nodesMap[e.SourceID]; exists && !visitedNodes[e.SourceID] {
						visitedNodes[e.SourceID] = true
						nextLevel[e.SourceID] = true
					}
				}
			}
			currentLevel = nextLevel
			if len(currentLevel) == 0 {
				break
			}
		}

		// Filter edges to only those connecting visited nodes
		var resultEdges []MemoryEdge
		for _, e := range allEdges {
			if visitedNodes[e.SourceID] && visitedNodes[e.TargetID] {
				resultEdges = append(resultEdges, e)
			}
		}

		var resultNodes []MemoryNode
		for id := range visitedNodes {
			if n, ok := nodesMap[id]; ok {
				resultNodes = append(resultNodes, n)
			}
		}
		return resultNodes, resultEdges, nil
	}

	// 5. If SessionID specified without Ticker: return edges with endpoints
	if filter.SessionID != "" {
		activeNodeIDs := make(map[string]bool)
		var resultEdges []MemoryEdge
		for _, e := range allEdges {
			if _, sOk := nodesMap[e.SourceID]; sOk {
				if _, tOk := nodesMap[e.TargetID]; tOk {
					resultEdges = append(resultEdges, e)
					activeNodeIDs[e.SourceID] = true
					activeNodeIDs[e.TargetID] = true
				}
			}
		}
		var resultNodes []MemoryNode
		for id := range activeNodeIDs {
			resultNodes = append(resultNodes, nodesMap[id])
		}
		return resultNodes, resultEdges, nil
	}

	// 6. Global graph or filtered by types/weights
	var resultEdges []MemoryEdge
	activeNodeIDs := make(map[string]bool)
	for _, e := range allEdges {
		if _, sOk := nodesMap[e.SourceID]; sOk {
			if _, tOk := nodesMap[e.TargetID]; tOk {
				resultEdges = append(resultEdges, e)
				activeNodeIDs[e.SourceID] = true
				activeNodeIDs[e.TargetID] = true
			}
		}
	}
	var resultNodes []MemoryNode
	for id, n := range nodesMap {
		if len(allEdges) == 0 || activeNodeIDs[id] || len(typeSet) > 0 {
			resultNodes = append(resultNodes, n)
		}
	}
	return resultNodes, resultEdges, nil
}

// GetMemoryGraphStats computes aggregate metrics across the graph memory, optionally scoped by filter.
func (d *DB) GetMemoryGraphStats(filter ...MemoryGraphFilter) (*MemoryGraphStats, error) {
	if len(filter) > 0 && (filter[0].SessionID != "" || filter[0].Ticker != "" || len(filter[0].NodeTypes) > 0) {
		nodes, edges, err := d.GetFilteredMemoryGraph(filter[0])
		if err != nil {
			return nil, err
		}
		stats := &MemoryGraphStats{
			TotalNodes:  len(nodes),
			TotalEdges:  len(edges),
			NodeTypes:   make(map[string]int),
			TopHubNodes: []HubNode{},
		}
		degMap := make(map[string]int)
		for _, e := range edges {
			degMap[e.SourceID]++
			degMap[e.TargetID]++
		}
		for _, n := range nodes {
			stats.NodeTypes[n.NodeType]++
			stats.TopHubNodes = append(stats.TopHubNodes, HubNode{
				ID:       n.ID,
				Label:    n.Label,
				NodeType: n.NodeType,
				Degree:   degMap[n.ID],
			})
		}
		sort.Slice(stats.TopHubNodes, func(i, j int) bool {
			return stats.TopHubNodes[i].Degree > stats.TopHubNodes[j].Degree
		})
		if len(stats.TopHubNodes) > 10 {
			stats.TopHubNodes = stats.TopHubNodes[:10]
		}
		return stats, nil
	}

	stats := &MemoryGraphStats{
		NodeTypes:   make(map[string]int),
		TopHubNodes: []HubNode{},
	}

	// Total nodes
	_ = d.conn.QueryRow("SELECT COUNT(*) FROM memory_nodes").Scan(&stats.TotalNodes)
	// Total edges
	_ = d.conn.QueryRow("SELECT COUNT(*) FROM memory_edges").Scan(&stats.TotalEdges)

	// Node types breakdown
	rows, err := d.conn.Query("SELECT node_type, COUNT(*) FROM memory_nodes GROUP BY node_type")
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var nt string
			var count int
			if err := rows.Scan(&nt, &count); err == nil {
				stats.NodeTypes[nt] = count
			}
		}
	}

	// Hub nodes by degree
	hubQuery := `
		SELECT n.id, n.label, n.node_type, COUNT(e.node_id) as degree
		FROM memory_nodes n
		JOIN (
			SELECT source_id as node_id FROM memory_edges
			UNION ALL
			SELECT target_id as node_id FROM memory_edges
		) e ON n.id = e.node_id
		GROUP BY n.id, n.label, n.node_type
		ORDER BY degree DESC
		LIMIT 10
	`
	hubRows, err := d.conn.Query(hubQuery)
	if err == nil {
		defer hubRows.Close()
		for hubRows.Next() {
			var h HubNode
			if err := hubRows.Scan(&h.ID, &h.Label, &h.NodeType, &h.Degree); err == nil {
				stats.TopHubNodes = append(stats.TopHubNodes, h)
			}
		}
	}

	return stats, nil
}

// ClearMemoryGraph removes memory edges and nodes (optionally for a specific session).
func (d *DB) ClearMemoryGraph(sessionID ...string) error {
	if len(sessionID) > 0 && sessionID[0] != "" {
		_, err := d.conn.Exec("DELETE FROM memory_edges WHERE session_id = ?", sessionID[0])
		if err != nil {
			return fmt.Errorf("failed to delete memory edges for session %s: %w", sessionID[0], err)
		}
		_, _ = d.conn.Exec(`
			DELETE FROM memory_nodes 
			WHERE id NOT IN (SELECT source_id FROM memory_edges)
			  AND id NOT IN (SELECT target_id FROM memory_edges)
		`)
		return nil
	}

	if _, err := d.conn.Exec("DELETE FROM memory_edges"); err != nil {
		return fmt.Errorf("failed to clear memory edges: %w", err)
	}
	if _, err := d.conn.Exec("DELETE FROM memory_nodes"); err != nil {
		return fmt.Errorf("failed to clear memory nodes: %w", err)
	}
	return nil
}

// PruneMockTestData deletes test benchmark sessions (EVAL-*) and associated orphan nodes/edges.
func (d *DB) PruneMockTestData() (int64, error) {
	tx, err := d.conn.Begin()
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	// 1. Delete mock edges
	res, err := tx.Exec("DELETE FROM memory_edges WHERE session_id LIKE 'EVAL-%'")
	if err != nil {
		return 0, err
	}
	deletedEdges, _ := res.RowsAffected()

	// 2. Delete mock investigations & cascades
	_, _ = tx.Exec("DELETE FROM investigations WHERE id LIKE 'EVAL-%'")

	// 3. Delete orphan nodes that have no edges and are not user:default
	_, _ = tx.Exec(`
		DELETE FROM memory_nodes 
		WHERE id != 'user:default'
		  AND id NOT IN (SELECT source_id FROM memory_edges UNION SELECT target_id FROM memory_edges)
	`)

	if err := tx.Commit(); err != nil {
		return 0, err
	}
	return deletedEdges, nil
}

// TelegramChat represents a persistent mapping between a Telegram chat and a Niskava chat session.
type TelegramChat struct {
	ChatID           int64  `json:"chat_id"`
	CurrentSessionID string `json:"current_session_id"`
	UserID           int64  `json:"user_id"`
	Username         string `json:"username"`
	CreatedAt        string `json:"created_at"`
	UpdatedAt        string `json:"updated_at"`
}

// GetTelegramChat retrieves the telegram chat mapping by chatID.
func (d *DB) GetTelegramChat(chatID int64) (*TelegramChat, error) {
	query := `
		SELECT chat_id, current_session_id, user_id, username, created_at, updated_at
		FROM telegram_chats
		WHERE chat_id = ?
	`
	var chat TelegramChat
	err := d.conn.QueryRow(query, chatID).Scan(
		&chat.ChatID,
		&chat.CurrentSessionID,
		&chat.UserID,
		&chat.Username,
		&chat.CreatedAt,
		&chat.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get telegram chat %d: %w", chatID, err)
	}
	return &chat, nil
}

// GetOrCreateTelegramChatSession retrieves the active chat_sessions ID mapped to a Telegram chat,
// or creates both a new chat_session and the mapping record if none exists.
func (d *DB) GetOrCreateTelegramChatSession(chatID int64, userID int64, username string) (string, error) {
	existing, err := d.GetTelegramChat(chatID)
	if err != nil {
		return "", err
	}
	if existing != nil && existing.CurrentSessionID != "" {
		// Verify referenced session still exists
		sess, sessErr := d.GetChatSession(existing.CurrentSessionID)
		if sessErr == nil && sess != nil {
			return existing.CurrentSessionID, nil
		}
	}

	// Create new session
	return d.ResetTelegramChatSession(chatID, userID, username)
}

// ResetTelegramChatSession creates a fresh chat_sessions record and updates the telegram_chats mapping.
func (d *DB) ResetTelegramChatSession(chatID int64, userID int64, username string) (string, error) {
	now := time.Now().UTC().Format(time.RFC3339)
	newSessionID := fmt.Sprintf("TELE-%s-%04d", time.Now().Format("20060102"), time.Now().UnixNano()%10000)

	sessionTitle := "Sesi Telegram"
	if username != "" {
		sessionTitle = fmt.Sprintf("Telegram (@%s)", username)
	}

	sess := &ChatSession{
		ID:     newSessionID,
		Title:  sessionTitle,
		Model:  "hermes",
		Status: "IDLE",
	}
	if err := d.CreateChatSession(sess); err != nil {
		return "", fmt.Errorf("failed to create chat session for telegram %d: %w", chatID, err)
	}

	upsertQuery := `
		INSERT INTO telegram_chats (chat_id, current_session_id, user_id, username, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?)
		ON CONFLICT(chat_id) DO UPDATE SET
			current_session_id = excluded.current_session_id,
			user_id = excluded.user_id,
			username = excluded.username,
			updated_at = excluded.updated_at
	`
	if _, err := d.conn.Exec(upsertQuery, chatID, newSessionID, userID, username, now, now); err != nil {
		return "", fmt.Errorf("failed to map telegram chat %d to session %s: %w", chatID, newSessionID, err)
	}

	return newSessionID, nil
}

// SectorsCacheItem represents a cached HTTP response from the Sectors v2 API.
type SectorsCacheItem struct {
	CacheKey    string  `json:"cache_key"`
	Endpoint    string  `json:"endpoint"`
	PayloadJSON string  `json:"payload_json"`
	CreatedAt   string  `json:"created_at"`
	ExpiresAt   *string `json:"expires_at,omitempty"`
}

// SectorsCacheStats provides statistics on cache discipline for Law 5.
type SectorsCacheStats struct {
	TotalEntries     int `json:"total_entries"`
	ExpiredEntries   int `json:"expired_entries"`
	PermanentEntries int `json:"permanent_entries"`
}

// SetSectorsCache inserts or updates a cached Sectors API response.
func (d *DB) SetSectorsCache(cacheKey, endpoint, payloadJSON string, expiresAt *time.Time) error {
	query := `
		INSERT INTO sectors_cache (cache_key, endpoint, payload_json, expires_at)
		VALUES (?, ?, ?, ?)
		ON CONFLICT(cache_key) DO UPDATE SET
			endpoint = excluded.endpoint,
			payload_json = excluded.payload_json,
			expires_at = excluded.expires_at,
			created_at = CURRENT_TIMESTAMP
	`
	var expStr *string
	if expiresAt != nil {
		s := expiresAt.UTC().Format("2006-01-02 15:04:05")
		expStr = &s
	}
	_, err := d.conn.Exec(query, cacheKey, endpoint, payloadJSON, expStr)
	if err != nil {
		return fmt.Errorf("failed to set sectors cache: %w", err)
	}
	return nil
}

// GetSectorsCache retrieves an unexpired cached response by key.
func (d *DB) GetSectorsCache(cacheKey string) (string, error) {
	query := `
		SELECT payload_json FROM sectors_cache
		WHERE cache_key = ? AND (expires_at IS NULL OR expires_at > CURRENT_TIMESTAMP)
	`
	var payload string
	err := d.conn.QueryRow(query, cacheKey).Scan(&payload)
	if err != nil {
		return "", err
	}
	return payload, nil
}

// GetSectorsCacheStats returns aggregate cache numbers for credit discipline monitoring.
func (d *DB) GetSectorsCacheStats() (*SectorsCacheStats, error) {
	row := d.conn.QueryRow(`
		SELECT
			COUNT(*),
			COALESCE(SUM(CASE WHEN expires_at IS NOT NULL AND expires_at < CURRENT_TIMESTAMP THEN 1 ELSE 0 END), 0),
			COALESCE(SUM(CASE WHEN expires_at IS NULL THEN 1 ELSE 0 END), 0)
		FROM sectors_cache
	`)
	var stats SectorsCacheStats
	if err := row.Scan(&stats.TotalEntries, &stats.ExpiredEntries, &stats.PermanentEntries); err != nil {
		return nil, fmt.Errorf("failed to query sectors cache stats: %w", err)
	}
	return &stats, nil
}

// CleanExpiredCache deletes expired records from sectors_cache.
func (d *DB) CleanExpiredCache() (int64, error) {
	res, err := d.conn.Exec(`DELETE FROM sectors_cache WHERE expires_at IS NOT NULL AND expires_at < CURRENT_TIMESTAMP`)
	if err != nil {
		return 0, fmt.Errorf("failed to clean expired cache: %w", err)
	}
	return res.RowsAffected()
}

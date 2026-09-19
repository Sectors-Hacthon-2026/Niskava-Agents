// Package db provides SQLite persistence using pure-Go zero-CGO modernc.org/sqlite.
// Follows Local-First Data Sovereignty (Law 4) and schema specifications.
package db

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
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
    session_id TEXT REFERENCES investigations(id) ON DELETE SET NULL,
    weight REAL NOT NULL DEFAULT 1.0,
    last_observed_at TEXT NOT NULL,
    created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (source_id, target_id, relation)
);

CREATE TABLE IF NOT EXISTS chat_messages (
    id TEXT PRIMARY KEY,
    session_id TEXT NOT NULL,
    role TEXT NOT NULL,
    content TEXT NOT NULL,
    thought TEXT,
    tool_calls_json TEXT,
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
CREATE INDEX IF NOT EXISTS idx_chat_messages_session ON chat_messages(session_id);
`

// DB wraps the SQL database pool and provides high-level domain operations.
type DB struct {
	conn *sql.DB
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
	// Ensure parent directory exists
	dir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create db directory %s: %w", dir, err)
	}

	dsn := fmt.Sprintf("%s?_pragma=journal_mode(WAL)&_pragma=foreign_keys(ON)&_pragma=synchronous(NORMAL)", dbPath)
	conn, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open sqlite database at %s: %w", dbPath, err)
	}

	// Run migration DDL
	if _, err := conn.Exec(SchemaDDL); err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to apply database migrations: %w", err)
	}

	// Safe alter migrations for existing tables
	_, _ = conn.Exec("ALTER TABLE investigations ADD COLUMN market TEXT NOT NULL DEFAULT 'IDX';")
	_, _ = conn.Exec("ALTER TABLE investigations ADD COLUMN created_at TEXT DEFAULT '';")
	_, _ = conn.Exec("ALTER TABLE sectors_cache ADD COLUMN is_mock INTEGER NOT NULL DEFAULT 0;")

	return &DB{conn: conn}, nil
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
	return results, nil
}

// ChatMessage represents a single conversational turn in a research session.
type ChatMessage struct {
	ID            string  `json:"id"`
	SessionID     string  `json:"session_id"`
	Role          string  `json:"role"` // "user", "assistant", "system", "tool"
	Content       string  `json:"content"`
	Thought       *string `json:"thought,omitempty"`
	ToolCallsJSON *string `json:"tool_calls_json,omitempty"`
	CreatedAt     string  `json:"created_at"`
}

// SaveChatMessage records a user or assistant message to SQLite.
func (d *DB) SaveChatMessage(msg *ChatMessage) error {
	query := `
		INSERT INTO chat_messages (id, session_id, role, content, thought, tool_calls_json, created_at)
		VALUES (?, ?, ?, ?, ?, ?, COALESCE(?, CURRENT_TIMESTAMP))
	`
	var createdAt interface{} = msg.CreatedAt
	if msg.CreatedAt == "" {
		createdAt = nil
	}
	_, err := d.conn.Exec(query, msg.ID, msg.SessionID, msg.Role, msg.Content, msg.Thought, msg.ToolCallsJSON, createdAt)
	if err != nil {
		return fmt.Errorf("failed to save chat message: %w", err)
	}
	return nil
}

// GetChatHistory retrieves messages for a specific session ordered chronologically.
func (d *DB) GetChatHistory(sessionID string, limit int) ([]ChatMessage, error) {
	if limit <= 0 {
		limit = 50
	}
	query := `
		SELECT id, session_id, role, content, thought, tool_calls_json, created_at
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
		if err := rows.Scan(&m.ID, &m.SessionID, &m.Role, &m.Content, &m.Thought, &m.ToolCallsJSON, &m.CreatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan chat message: %w", err)
		}
		history = append(history, m)
	}
	return history, nil
}

// ChatSessionSummary represents a summary of a chat session.
type ChatSessionSummary struct {
	SessionID    string `json:"session_id"`
	FirstMessage string `json:"first_message"`
	LastMessage  string `json:"last_message"`
	MessageCount int    `json:"message_count"`
	UpdatedAt    string `json:"updated_at"`
}

// ListChatSessions retrieves unique chat sessions ordered by latest activity.
func (d *DB) ListChatSessions(limit int) ([]ChatSessionSummary, error) {
	if limit <= 0 {
		limit = 30
	}
	query := `
		SELECT 
			session_id,
			COALESCE((SELECT content FROM chat_messages m2 WHERE m2.session_id = m.session_id AND m2.role = 'user' ORDER BY m2.created_at ASC LIMIT 1), '') as first_msg,
			COALESCE((SELECT content FROM chat_messages m3 WHERE m3.session_id = m.session_id ORDER BY m3.created_at DESC LIMIT 1), '') as last_msg,
			COUNT(*) as msg_count,
			MAX(created_at) as updated_at
		FROM chat_messages m
		GROUP BY session_id
		ORDER BY updated_at DESC
		LIMIT ?
	`
	rows, err := d.conn.Query(query, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to list chat sessions: %w", err)
	}
	defer rows.Close()

	var sessions []ChatSessionSummary
	for rows.Next() {
		var s ChatSessionSummary
		if err := rows.Scan(&s.SessionID, &s.FirstMessage, &s.LastMessage, &s.MessageCount, &s.UpdatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan chat session: %w", err)
		}
		sessions = append(sessions, s)
	}
	return sessions, nil
}

// ClearAllChatHistory wipes all chat messages from the database.
func (d *DB) ClearAllChatHistory() error {
	_, err := d.conn.Exec("DELETE FROM chat_messages")
	return err
}

// DeleteChatSession deletes all chat messages for a specific session ID.
func (d *DB) DeleteChatSession(sessionID string) error {
	_, err := d.conn.Exec("DELETE FROM chat_messages WHERE session_id = ?", sessionID)
	return err
}

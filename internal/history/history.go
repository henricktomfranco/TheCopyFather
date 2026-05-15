package history

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

// HistoryEntry represents a single rewrite history entry
type HistoryEntry struct {
	ID            int64     `json:"id"`
	OriginalText  string    `json:"original_text"`
	RewrittenText string    `json:"rewritten_text"`
	Style         string    `json:"style"`
	TextType      string    `json:"text_type"`
	Provider      string    `json:"provider"`
	Confidence    float64   `json:"confidence"`
	IsFavorite    bool      `json:"is_favorite"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// HistoryDB manages the history database
type HistoryDB struct {
	db *sql.DB
}

// New creates a new HistoryDB instance
func New(dbPath string) (*HistoryDB, error) {
	// Create directory if it doesn't exist
	dir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create directory: %w", err)
	}

	// Open database
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Enable WAL mode for better performance
	_, err = db.Exec("PRAGMA journal_mode=WAL;")
	if err != nil {
		return nil, fmt.Errorf("failed to set WAL mode: %w", err)
	}

	// Create tables
	if err := createTables(db); err != nil {
		return nil, fmt.Errorf("failed to create tables: %w", err)
	}

	return &HistoryDB{db: db}, nil
}

// createTables creates the database tables
func createTables(db *sql.DB) error {
	// Create history table
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS history (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			original_text TEXT NOT NULL,
			rewritten_text TEXT NOT NULL,
			style TEXT NOT NULL,
			text_type TEXT DEFAULT '',
			provider TEXT DEFAULT 'ollama',
			confidence REAL DEFAULT 0.0,
			is_favorite BOOLEAN DEFAULT FALSE,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		return err
	}

	// Create index on created_at for faster queries
	_, err = db.Exec(`
		CREATE INDEX IF NOT EXISTS idx_history_created_at ON history(created_at)
	`)
	if err != nil {
		return err
	}

	// Create index on style
	_, err = db.Exec(`
		CREATE INDEX IF NOT EXISTS idx_history_style ON history(style)
	`)
	if err != nil {
		return err
	}

	// Create index on is_favorite
	_, err = db.Exec(`
		CREATE INDEX IF NOT EXISTS idx_history_is_favorite ON history(is_favorite)
	`)
	if err != nil {
		return err
	}

	return nil
}

// Close closes the database connection
func (h *HistoryDB) Close() error {
	return h.db.Close()
}

// Save saves a new history entry
func (h *HistoryDB) Save(ctx context.Context, entry HistoryEntry) (*HistoryEntry, error) {
	if entry.CreatedAt.IsZero() {
		entry.CreatedAt = time.Now()
	}
	entry.UpdatedAt = entry.CreatedAt

	query := `
		INSERT INTO history (original_text, rewritten_text, style, text_type, provider, confidence, is_favorite, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	result, err := h.db.ExecContext(ctx,
		query,
		entry.OriginalText,
		entry.RewrittenText,
		entry.Style,
		entry.TextType,
		entry.Provider,
		entry.Confidence,
		entry.IsFavorite,
		entry.CreatedAt,
		entry.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to save history: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("failed to get last insert id: %w", err)
	}

	entry.ID = id
	return &entry, nil
}

// GetByID retrieves a history entry by ID
func (h *HistoryDB) GetByID(ctx context.Context, id int64) (*HistoryEntry, error) {
	query := `
		SELECT id, original_text, rewritten_text, style, text_type, provider, confidence, is_favorite, created_at, updated_at
		FROM history WHERE id = ?
	`

	row := h.db.QueryRowContext(ctx, query, id)
	return scanRow(row)
}

// List retrieves all history entries, optionally filtered and paginated
func (h *HistoryDB) List(ctx context.Context, req ListRequest) ([]HistoryEntry, int64, error) {
	// Build query
	query := `
		SELECT id, original_text, rewritten_text, style, text_type, provider, confidence, is_favorite, created_at, updated_at
		FROM history
	`

	var args []interface{}

	// Add filters
	if req.Style != "" {
		query += " WHERE style = ?"
		args = append(args, req.Style)
	} else {
		// Add other filters if needed
		var conditions []string
		if req.IsFavorite {
			conditions = append(conditions, "is_favorite = TRUE")
		}
		if req.TextType != "" {
			conditions = append(conditions, "text_type = ?")
			args = append(args, req.TextType)
		}
		if req.Search != "" {
			conditions = append(conditions, "(original_text LIKE ? OR rewritten_text LIKE ?)")
			args = append(args, "%"+req.Search+"%", "%"+req.Search+"%")
		}
		if len(conditions) > 0 {
			query += " WHERE " + joinConditions(conditions)
		}
	}

	// Add ordering
	if req.OrderBy == "" {
		req.OrderBy = "created_at DESC"
	}
	query += " ORDER BY " + req.OrderBy

	// Add limit and offset
	if req.Limit > 0 {
		query += " LIMIT ?"
		args = append(args, req.Limit)
		if req.Offset > 0 {
			query += " OFFSET ?"
			args = append(args, req.Offset)
		}
	}

	// Execute query
	rows, err := h.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list history: %w", err)
	}
	defer rows.Close()

	var entries []HistoryEntry
	for rows.Next() {
		entry, err := scanRow(rows)
		if err != nil {
			return nil, 0, err
		}
		entries = append(entries, *entry)
	}

	// Get total count
	var total int64
	countQuery := "SELECT COUNT(*) FROM history"
	if len(args) > 0 && !req.CountAll {
		// Reuse the same conditions
		var countConditions []string
		var countArgs []interface{}
		if req.Style != "" {
			countConditions = append(countConditions, "style = ?")
			countArgs = append(countArgs, req.Style)
		} else {
			if req.IsFavorite {
				countConditions = append(countConditions, "is_favorite = TRUE")
			}
			if req.TextType != "" {
				countConditions = append(countConditions, "text_type = ?")
				countArgs = append(countArgs, req.TextType)
			}
			if req.Search != "" {
				countConditions = append(countConditions, "(original_text LIKE ? OR rewritten_text LIKE ?)")
				countArgs = append(countArgs, "%"+req.Search+"%", "%"+req.Search+"%")
			}
		}
		if len(countConditions) > 0 {
			countQuery += " WHERE " + joinConditions(countConditions)
		}
		err = h.db.QueryRowContext(ctx, countQuery, countArgs...).Scan(&total)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to get count: %w", err)
		}
	} else {
		// Count all
		err = h.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM history").Scan(&total)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to get count: %w", err)
		}
	}

	return entries, total, nil
}

// Update updates an existing history entry
func (h *HistoryDB) Update(ctx context.Context, entry HistoryEntry) error {
	entry.UpdatedAt = time.Now()

	query := `
		UPDATE history
		SET original_text = ?, rewritten_text = ?, style = ?, text_type = ?, provider = ?, confidence = ?, is_favorite = ?, updated_at = ?
		WHERE id = ?
	`

	_, err := h.db.ExecContext(ctx,
		query,
		entry.OriginalText,
		entry.RewrittenText,
		entry.Style,
		entry.TextType,
		entry.Provider,
		entry.Confidence,
		entry.IsFavorite,
		entry.UpdatedAt,
		entry.ID,
	)
	if err != nil {
		return fmt.Errorf("failed to update history: %w", err)
	}

	return nil
}

// Delete deletes a history entry
func (h *HistoryDB) Delete(ctx context.Context, id int64) error {
	query := "DELETE FROM history WHERE id = ?"
	_, err := h.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete history: %w", err)
	}
	return nil
}

// DeleteAll deletes all history entries
func (h *HistoryDB) DeleteAll(ctx context.Context) error {
	query := "DELETE FROM history"
	_, err := h.db.ExecContext(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to delete all history: %w", err)
	}
	return nil
}

// ToggleFavorite toggles the favorite status of an entry
func (h *HistoryDB) ToggleFavorite(ctx context.Context, id int64) error {
	query := "UPDATE history SET is_favorite = NOT is_favorite, updated_at = CURRENT_TIMESTAMP WHERE id = ?"
	_, err := h.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to toggle favorite: %w", err)
	}
	return nil
}

// ListRequest represents a request to list history entries
type ListRequest struct {
	Style      string `json:"style,omitempty"`
	TextType   string `json:"text_type,omitempty"`
	IsFavorite bool   `json:"is_favorite,omitempty"`
	Search     string `json:"search,omitempty"`
	OrderBy    string `json:"order_by,omitempty"`
	Limit      int    `json:"limit,omitempty"`
	Offset     int    `json:"offset,omitempty"`
	CountAll   bool   `json:"count_all,omitempty"`
}

// scanRow scans a database row into a HistoryEntry
func scanRow(row interface{ Scan(...) error }) (*HistoryEntry, error) {
	var entry HistoryEntry
	var createdAtStr, updatedAtStr string

	err := row.(*sql.Row).Scan(
		&entry.ID,
		&entry.OriginalText,
		&entry.RewrittenText,
		&entry.Style,
		&entry.TextType,
		&entry.Provider,
		&entry.Confidence,
		&entry.IsFavorite,
		&createdAtStr,
		&updatedAtStr,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to scan row: %w", err)
	}

	// Parse timestamps
	if createdAtStr != "" {
		entry.CreatedAt, err = time.Parse("2006-01-02 15:04:05", createdAtStr)
		if err != nil {
			return nil, fmt.Errorf("failed to parse created_at: %w", err)
		}
	}
	if updatedAtStr != "" {
		entry.UpdatedAt, err = time.Parse("2006-01-02 15:04:05", updatedAtStr)
		if err != nil {
			return nil, fmt.Errorf("failed to parse updated_at: %w", err)
		}
	}

	return &entry, nil
}

// joinConditions joins conditions with AND
func joinConditions(conditions []string) string {
	if len(conditions) == 0 {
		return ""
	}
	result := ""
	for i, cond := range conditions {
		if i > 0 {
			result += " AND "
		}
		result += cond
	}
	return result
}

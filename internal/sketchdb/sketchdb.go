package sketchdb

import (
	"database/sql"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	_ "modernc.org/sqlite"
)

type DB struct {
	sql  *sql.DB
	path string
	mu   sync.Mutex
}

func Open(dbPath string) (*DB, error) {
	if dir := filepath.Dir(dbPath); dir != "" && dir != "." {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return nil, err
		}
	}
	sqlDB, err := sql.Open("sqlite", dbPath+"?_pragma=busy_timeout(5000)")
	if err != nil {
		return nil, err
	}
	d := &DB{sql: sqlDB, path: dbPath}
	if err := d.migrate(); err != nil {
		_ = sqlDB.Close()
		return nil, err
	}
	return d, nil
}

func (d *DB) Close() error {
	return d.sql.Close()
}

func (d *DB) migrate() error {
	stmts := []string{
		`CREATE TABLE IF NOT EXISTS metadata (
			id INTEGER PRIMARY KEY CHECK (id = 1),
			sketch_name TEXT NOT NULL,
			created_at TEXT NOT NULL,
			sketch_dir_at_create TEXT NOT NULL,
			last_run_at TEXT NOT NULL
		);`,
		`CREATE TABLE IF NOT EXISTS saves (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			rel_path TEXT NOT NULL,
			format TEXT NOT NULL,
			created_at TEXT NOT NULL
		);`,
		`CREATE TABLE IF NOT EXISTS snapshots (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL UNIQUE,
			created_at TEXT NOT NULL,
			control_json TEXT NOT NULL,
			png_save_id INTEGER REFERENCES saves(id),
			svg_save_id INTEGER REFERENCES saves(id),
			description TEXT NOT NULL DEFAULT ''
		);`,
	}
	for _, s := range stmts {
		if _, err := d.sql.Exec(s); err != nil {
			return fmt.Errorf("migrate: %w", err)
		}
	}
	if err := d.ensureSnapshotDescriptionColumn(); err != nil {
		return err
	}
	return d.ensureSnapshotBuiltinJSONColumn()
}

func (d *DB) ensureSnapshotDescriptionColumn() error {
	rows, err := d.sql.Query(`PRAGMA table_info(snapshots)`)
	if err != nil {
		return err
	}
	defer rows.Close()
	var hasDesc bool
	for rows.Next() {
		var cid, notnull, pk int
		var name, ctype string
		var dflt any
		if err := rows.Scan(&cid, &name, &ctype, &notnull, &dflt, &pk); err != nil {
			return err
		}
		if name == "description" {
			hasDesc = true
		}
	}
	if err := rows.Err(); err != nil {
		return err
	}
	if hasDesc {
		return nil
	}
	_, err = d.sql.Exec(`ALTER TABLE snapshots ADD COLUMN description TEXT NOT NULL DEFAULT ''`)
	return err
}

func (d *DB) ensureSnapshotBuiltinJSONColumn() error {
	rows, err := d.sql.Query(`PRAGMA table_info(snapshots)`)
	if err != nil {
		return err
	}
	defer rows.Close()
	var hasBuiltin bool
	for rows.Next() {
		var cid, notnull, pk int
		var name, ctype string
		var dflt any
		if err := rows.Scan(&cid, &name, &ctype, &notnull, &dflt, &pk); err != nil {
			return err
		}
		if name == "builtin_json" {
			hasBuiltin = true
		}
	}
	if err := rows.Err(); err != nil {
		return err
	}
	if hasBuiltin {
		return nil
	}
	_, err = d.sql.Exec(`ALTER TABLE snapshots ADD COLUMN builtin_json TEXT NOT NULL DEFAULT ''`)
	return err
}

// InitMetadata ensures row id=1 exists and updates last_run_at.
func (d *DB) InitMetadata(sketchName, sketchDir string) error {
	d.mu.Lock()
	defer d.mu.Unlock()
	now := time.Now().UTC().Format(time.RFC3339Nano)
	var n int
	err := d.sql.QueryRow(`SELECT COUNT(*) FROM metadata WHERE id = 1`).Scan(&n)
	if err != nil {
		return err
	}
	if n == 0 {
		_, err = d.sql.Exec(
			`INSERT INTO metadata (id, sketch_name, created_at, sketch_dir_at_create, last_run_at)
			VALUES (1, ?, ?, ?, ?)`,
			sketchName, now, sketchDir, now,
		)
		return err
	}
	_, err = d.sql.Exec(`UPDATE metadata SET sketch_name = ?, last_run_at = ? WHERE id = 1`, sketchName, now)
	return err
}

func (d *DB) InsertSave(relPath, format string) (int64, error) {
	d.mu.Lock()
	defer d.mu.Unlock()
	now := time.Now().UTC().Format(time.RFC3339Nano)
	res, err := d.sql.Exec(`INSERT INTO saves (rel_path, format, created_at) VALUES (?, ?, ?)`, relPath, format, now)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

type SnapshotRow struct {
	ID          int64
	Name        string
	CreatedAt   string
	ControlJSON string
	BuiltinJSON string
	Description string
	PNGSaveID   sql.NullInt64
	SVGSaveID   sql.NullInt64
	PNGPath     string
	SVGPath     string
}

func (d *DB) ListSnapshotNames() ([]string, error) {
	d.mu.Lock()
	defer d.mu.Unlock()
	rows, err := d.sql.Query(`SELECT name FROM snapshots ORDER BY name COLLATE NOCASE ASC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var names []string
	for rows.Next() {
		var n string
		if err := rows.Scan(&n); err != nil {
			return nil, err
		}
		names = append(names, n)
	}
	return names, rows.Err()
}

func (d *DB) GetSnapshotByName(name string) (*SnapshotRow, error) {
	d.mu.Lock()
	defer d.mu.Unlock()
	var r SnapshotRow
	var pngPath, svgPath sql.NullString
	err := d.sql.QueryRow(`
		SELECT s.id, s.name, s.created_at, s.control_json, s.builtin_json, s.description, s.png_save_id, s.svg_save_id,
			p.rel_path, v.rel_path
		FROM snapshots s
		LEFT JOIN saves p ON s.png_save_id = p.id
		LEFT JOIN saves v ON s.svg_save_id = v.id
		WHERE s.name = ?`, name).Scan(
		&r.ID, &r.Name, &r.CreatedAt, &r.ControlJSON, &r.BuiltinJSON, &r.Description, &r.PNGSaveID, &r.SVGSaveID, &pngPath, &svgPath,
	)
	if err == nil {
		r.PNGPath = pngPath.String
		r.SVGPath = svgPath.String
	}
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &r, nil
}

func (d *DB) InsertSnapshot(name, description, controlJSON, builtinJSON string, pngSaveID, svgSaveID *int64) error {
	d.mu.Lock()
	defer d.mu.Unlock()
	now := time.Now().UTC().Format(time.RFC3339Nano)
	var png, svg any
	if pngSaveID != nil {
		png = *pngSaveID
	}
	if svgSaveID != nil {
		svg = *svgSaveID
	}
	_, err := d.sql.Exec(
		`INSERT INTO snapshots (name, created_at, control_json, builtin_json, png_save_id, svg_save_id, description) VALUES (?, ?, ?, ?, ?, ?, ?)`,
		name, now, controlJSON, builtinJSON, png, svg, description,
	)
	return err
}

func (d *DB) Path() string { return d.path }

package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"goshort/internal/domain"
	"goshort/internal/storage"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

type postgresRepository struct {
	db *sqlx.DB
}

// NewPostgresRepository creates a new PostgreSQL repository
func NewPostgresRepository(db *sqlx.DB) storage.URLRepository {
	return &postgresRepository{db: db}
}

// Connect creates a new database connection
func Connect(host string, port int, user, password, dbname, sslmode string) (*sqlx.DB, error) {
	dsn := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		host, port, user, password, dbname, sslmode,
	)

	db, err := sqlx.Connect("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Connection pool settings
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	// Verify connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return db, nil
}

func (r *postgresRepository) Create(ctx context.Context, url *domain.URL) error {
	// Generate UUID if not set
	if url.ID == "" {
		url.ID = uuid.New().String()
	}

	query := `
		INSERT INTO urls (id, original_url, short_code, created_at, expires_at, is_active, created_by_ip, user_agent, click_count)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`

	_, err := r.db.ExecContext(
		ctx,
		query,
		url.ID,
		url.OriginalURL,
		url.ShortCode,
		url.CreatedAt,
		url.ExpiresAt,
		url.IsActive,
		url.CreatedByIP,
		url.UserAgent,
		url.ClickCount,
	)

	if err != nil {
		// Check for unique constraint violation
		if isDuplicateKeyError(err) {
			return domain.ErrDuplicateShortCode
		}
		return fmt.Errorf("failed to create URL: %w", err)
	}

	return nil
}

func (r *postgresRepository) GetByShortCode(ctx context.Context, shortCode string) (*domain.URL, error) {
	var url domain.URL

	query := `
		SELECT id, original_url, short_code, created_at, expires_at, click_count, is_active, created_by_ip, user_agent
		FROM urls
		WHERE short_code = $1 AND is_active = true
	`

	err := r.db.GetContext(ctx, &url, query, shortCode)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrURLNotFound
		}
		return nil, fmt.Errorf("failed to get URL by short code: %w", err)
	}

	// Check expiration
	if url.IsExpired() {
		return nil, domain.ErrURLExpired
	}

	return &url, nil
}

func (r *postgresRepository) GetByID(ctx context.Context, id string) (*domain.URL, error) {
	var url domain.URL

	query := `
		SELECT id, original_url, short_code, created_at, expires_at, click_count, is_active, created_by_ip, user_agent
		FROM urls
		WHERE id = $1
	`

	err := r.db.GetContext(ctx, &url, query, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrURLNotFound
		}
		return nil, fmt.Errorf("failed to get URL by ID: %w", err)
	}

	return &url, nil
}

func (r *postgresRepository) Update(ctx context.Context, url *domain.URL) error {
	query := `
		UPDATE urls
		SET original_url = $1, expires_at = $2, is_active = $3, click_count = $4
		WHERE id = $5
	`

	result, err := r.db.ExecContext(
		ctx,
		query,
		url.OriginalURL,
		url.ExpiresAt,
		url.IsActive,
		url.ClickCount,
		url.ID,
	)

	if err != nil {
		return fmt.Errorf("failed to update URL: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return domain.ErrURLNotFound
	}

	return nil
}

func (r *postgresRepository) Delete(ctx context.Context, id string) error {
	query := `UPDATE urls SET is_active = false WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete URL: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return domain.ErrURLNotFound
	}

	return nil
}

func (r *postgresRepository) IncrementClickCount(ctx context.Context, shortCode string) error {
	query := `
		UPDATE urls
		SET click_count = click_count + 1
		WHERE short_code = $1 AND is_active = true
	`

	_, err := r.db.ExecContext(ctx, query, shortCode)
	if err != nil {
		return fmt.Errorf("failed to increment click count: %w", err)
	}

	return nil
}

func (r *postgresRepository) Exists(ctx context.Context, shortCode string) (bool, error) {
	var exists bool

	query := `SELECT EXISTS(SELECT 1 FROM urls WHERE short_code = $1)`

	err := r.db.GetContext(ctx, &exists, query, shortCode)
	if err != nil {
		return false, fmt.Errorf("failed to check if URL exists: %w", err)
	}

	return exists, nil
}

func (r *postgresRepository) List(ctx context.Context, limit, offset int) ([]*domain.URL, error) {
	var urls []*domain.URL

	query := `
		SELECT id, original_url, short_code, created_at, expires_at, click_count, is_active, created_by_ip, user_agent
		FROM urls
		WHERE is_active = true
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`

	err := r.db.SelectContext(ctx, &urls, query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list URLs: %w", err)
	}

	return urls, nil
}

// Helper function to check for duplicate key errors
func isDuplicateKeyError(err error) bool {
	return err != nil && (
		err.Error() == "pq: duplicate key value violates unique constraint \"urls_short_code_key\"" ||
		err.Error() == "UNIQUE constraint failed: urls.short_code")
}


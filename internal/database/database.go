package database

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/google/uuid"
	_ "github.com/joho/godotenv/autoload"
	_ "github.com/mattn/go-sqlite3"
)

type Service interface {
	Health() map[string]string
	Close() error
	GetShortenedUrlByShortURL(shortURL string) (string, error)
	CreateShortenedUrl(originalURL string) (ShortenedURL, error)
}

type service struct {
	db *sql.DB
}

var (
	dburl      = os.Getenv("DB_URL")
	dbInstance *service
)

func New() Service {
	if dbInstance != nil {
		return dbInstance
	}

	db, err := sql.Open("sqlite3", dburl)
	if err != nil {
		log.Fatal(err)
	}

	dbInstance = &service{
		db: db,
	}
	return dbInstance
}

func (s *service) Health() map[string]string {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	stats := make(map[string]string)

	err := s.db.PingContext(ctx)
	if err != nil {
		stats["status"] = "down"
		stats["error"] = fmt.Sprintf("db down: %v", err)
		log.Fatalf(fmt.Sprintf("db down: %v", err))
		return stats
	}

	stats["status"] = "up"
	stats["message"] = "It's healthy"

	dbStats := s.db.Stats()
	stats["open_connections"] = strconv.Itoa(dbStats.OpenConnections)
	stats["in_use"] = strconv.Itoa(dbStats.InUse)
	stats["idle"] = strconv.Itoa(dbStats.Idle)
	stats["wait_count"] = strconv.FormatInt(dbStats.WaitCount, 10)
	stats["wait_duration"] = dbStats.WaitDuration.String()
	stats["max_idle_closed"] = strconv.FormatInt(dbStats.MaxIdleClosed, 10)
	stats["max_lifetime_closed"] = strconv.FormatInt(dbStats.MaxLifetimeClosed, 10)

	if dbStats.OpenConnections > 40 {
		stats["message"] = "The database is experiencing heavy load."
	}

	if dbStats.WaitCount > 1000 {
		stats["message"] = "The database has a high number of wait events, indicating potential bottlenecks."
	}

	if dbStats.MaxIdleClosed > int64(dbStats.OpenConnections)/2 {
		stats["message"] = "Many idle connections are being closed, consider revising the connection pool settings."
	}

	if dbStats.MaxLifetimeClosed > int64(dbStats.OpenConnections)/2 {
		stats["message"] = "Many connections are being closed due to max lifetime, consider increasing max lifetime or revising the connection usage pattern."
	}

	return stats
}

func (s *service) Close() error {
	log.Printf("Disconnected from database: %s", dburl)
	return s.db.Close()
}

type ShortenedURL struct {
	ID          int
	ShortURL    string
	OriginalURL string
	CreatedAt   time.Time
}

func (s *service) GetShortenedUrlByShortURL(shortURL string) (string, error) {
	sqlStatement := `SELECT OriginalURL FROM ShortenedURL WHERE ShortURL = $1;`
	var originalUrl string

	row := s.db.QueryRow(sqlStatement, shortURL)
	switch err := row.Scan(&originalUrl); err {
	case sql.ErrNoRows:
		fmt.Println("No rows were returned!")
		log.Default().Println("No rows were returned!")
		return "", err
	}

	return originalUrl, nil
}

func (s *service) CreateShortenedUrl(originalURL string) (ShortenedURL, error) {
	shortURL := uuid.New().String()[:8]

	result, err := s.db.Exec("INSERT INTO ShortenedURL (shortUrl, originalUrl) VALUES (?, ?)", shortURL, originalURL)
	if err != nil {
		return ShortenedURL{}, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return ShortenedURL{}, err
	}

	return ShortenedURL{
		ID:          int(id),
		ShortURL:    shortURL,
		OriginalURL: originalURL,
		CreatedAt:   time.Now(),
	}, nil
}

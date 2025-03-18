package pg

// import (
// 	"context"
// 	"database/sql"
// 	"fmt"
// 	"log"
// 	"strconv"
// 	"sync"
// 	"time"

// 	"github.com/dockworks/dm-web-backend/internal/config"
// 	sqlc "github.com/dockworks/dm-web-backend/internal/db"
// 	_ "github.com/jackc/pgx/v5/stdlib"
// )

// // Service represents a service that interacts with a database.
// type DBService interface {
// 	// Health returns a map of health status information.
// 	// The keys and values in the map are service-specific.
// 	Health() map[string]string

// 	// Close terminates the database connection.
// 	// It returns an error if the connection cannot be closed.
// 	Close()

// 	// Queries returns the sqlc.Queries instance.
// 	Queries() *sqlc.Queries
// }

// type service struct {
// 	Pool  *sql.DB
// 	Query *sqlc.Queries
// }

// var (
// 	dbInstance *service
// 	pgOnce     sync.Once
// )

// func NewConnection(cfg *config.Config) DBService {
// 	// Reuse Connection
// 	// if dbInstance != nil {
// 	// 	return dbInstance
// 	// }
// 	pgOnce.Do(func() {
// 		dbCfg := cfg.DB
// 		connStr := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable&search_path=%s", dbCfg.User, dbCfg.Password, dbCfg.Host, dbCfg.Port, dbCfg.Name, dbCfg.Schema)

// 		db, err := sql.Open("pgx", connStr)
// 		if err != nil {
// 			log.Fatal(err)
// 		}

// 		query := sqlc.New(db)
// 		dbInstance = &service{
// 			Pool:  db,
// 			Query: query,
// 		}
// 	})
// 	return dbInstance
// }

// func (s *service) Queries() *sqlc.Queries {
// 	return s.Query
// }

// // Health checks the health of the database connection by pinging the database.
// // It returns a map with keys indicating various health statistics.
// func (s *service) Health() map[string]string {
// 	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
// 	defer cancel()

// 	stats := make(map[string]string)

// 	// Ping the database
// 	err := s.Pool.PingContext(ctx)
// 	if err != nil {
// 		stats["status"] = "down"
// 		stats["error"] = fmt.Sprintf("db down: %v", err)
// 		log.Fatalf(fmt.Sprintf("db down: %v", err)) // Log the error and terminate the program
// 		return stats
// 	}

// 	// Database is up, add more statistics
// 	stats["status"] = "up"
// 	stats["message"] = "It's healthy"

// 	// Get database stats (like open connections, in use, idle, etc.)
// 	dbStats := s.Pool.Stats()
// 	stats["open_connections"] = strconv.Itoa(int(dbStats.OpenConnections))
// 	stats["in_use"] = strconv.Itoa(int(dbStats.InUse))
// 	stats["idle"] = strconv.Itoa(int(dbStats.Idle))
// 	stats["wait_count"] = strconv.FormatInt(int64(dbStats.WaitCount), 10)
// 	stats["wait_duration"] = dbStats.WaitDuration.String()
// 	stats["max_idle_closed"] = strconv.FormatInt(int64(dbStats.MaxIdleClosed), 10)
// 	stats["max_lifetime_closed"] = strconv.FormatInt(int64(dbStats.MaxLifetimeClosed), 10)

// 	// Evaluate stats to provide a health message
// 	if dbStats.InUse > 40 { // Assuming 50 is the max for this example
// 		stats["message"] = "The database is experiencing heavy load."
// 	}

// 	// if dbStats.WaitCount > 1000 {
// 	// 	stats["message"] = "The database has a high number of wait events, indicating potential bottlenecks."
// 	// }

// 	// if dbStats.MaxIdleClosed > int64(dbStats.OpenConnections)/2 {
// 	// 	stats["message"] = "Many idle connections are being closed, consider revising the connection pool settings."
// 	// }

// 	// if dbStats.MaxLifetimeClosed > int64(dbStats.OpenConnections)/2 {
// 	// 	stats["message"] = "Many connections are being closed due to max lifetime, consider increasing max lifetime or revising the connection usage pattern."
// 	// }

// 	return stats
// }

// // Close closes the database connection.
// // It logs a message indicating the disconnection from the specific database.
// // If the connection is successfully closed, it returns nil.
// // If an error occurs while closing the connection, it returns the error.
// func (s *service) Close() {
// 	log.Printf("Disconnected from database")
// 	s.Pool.Close()
// }

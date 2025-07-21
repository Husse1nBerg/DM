package routes

import (
	s "github.com/dockworks/dm-web-backend/internal/server"
	h "github.com/dockworks/dm-web-backend/internal/server/handlers"
	"github.com/labstack/echo/v4"
)

// RegisterTestRoutes registers all testing and development routes
func RegisterTestRoutes(server *s.Server, permissionProtected *echo.Group) {
	permissionTestHandler := h.NewPermissionTestHandler(server)
	redisHandler := h.NewRedisHandler(server)

	// Test routes (for testing the permission system)
	test := permissionProtected.Group("/test")
	test.POST("/permissions", permissionTestHandler.TestPermission)
	test.GET("/permissions/user", permissionTestHandler.GetUserPermissions)
	test.GET("/permissions/routes", permissionTestHandler.GetRoutePermissions)

	// Redis test routes (for testing Redis connection and operations)
	redisTest := test.Group("/redis") // Redis tests under the test group
	redisTest.GET("/ping", redisHandler.Ping)
	redisTest.GET("/connection", redisHandler.TestRedisConnection)
	redisTest.POST("/operations", redisHandler.TestRedisOperations)
	redisTest.GET("/dual-databases", redisHandler.TestDualDatabaseConnection)

	// Cache operations (DB 0)
	redisCache := redisTest.Group("/cache")
	redisCache.POST("/operations", redisHandler.TestCacheOperations)
	redisCache.GET("/key/:key", redisHandler.GetCacheKey)
	redisCache.DELETE("/key/:key", redisHandler.DeleteCacheKey)

	// Task operations (DB 1)
	redisTask := redisTest.Group("/task")
	redisTask.POST("/operations", redisHandler.TestTaskOperations)
	redisTask.GET("/key/:key", redisHandler.GetTaskKey)
	redisTask.DELETE("/key/:key", redisHandler.DeleteTaskKey)

	// Queue operations (DB 1)
	redisQueue := redisTest.Group("/queue")
	redisQueue.POST("/operations", redisHandler.TestQueueOperations)
	redisQueue.GET("/:queue/length", redisHandler.GetQueueLength)
}

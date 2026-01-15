package routes

import (
	"github.com/DucLUT/goodstuff/controllers"
	"github.com/DucLUT/goodstuff/middleware"
	"github.com/DucLUT/goodstuff/models"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func SetupRouter() *gin.Engine {
	r := gin.Default()

	// CORS configuration
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))

	// Health check
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// API v1 routes
	v1 := r.Group("/api/v1")
	{
		// Public routes
		auth := v1.Group("/auth")
		{
			auth.POST("/register", controllers.Register)
			auth.POST("/login", controllers.Login)
		}

		// Public service/category routes
		v1.GET("/services", controllers.GetServices)
		v1.GET("/services/:id", controllers.GetServiceByID)
		v1.GET("/categories", controllers.GetCategories)
		v1.GET("/workers", controllers.GetWorkers)
		v1.GET("/workers/:id", controllers.GetWorkerByID)
		v1.GET("/workers/:id/reviews", controllers.GetWorkerReviews)

		// Protected routes
		protected := v1.Group("")
		protected.Use(middleware.AuthMiddleware())
		{
			// User routes
			users := protected.Group("/users")
			{
				users.GET("/profile", controllers.GetProfile)
				users.PUT("/profile", controllers.UpdateProfile)
				users.PUT("/password", controllers.ChangePassword)
			}

			// Booking routes (all authenticated users)
			bookings := protected.Group("/bookings")
			{
				bookings.POST("", controllers.CreateBooking)
				bookings.GET("", controllers.GetBookings)
				bookings.GET("/:id", controllers.GetBookingByID)
				bookings.PUT("/:id/cancel", controllers.CancelBooking)
			}

			// Worker-only routes
			worker := protected.Group("/worker")
			worker.Use(middleware.RoleMiddleware(string(models.RoleWorker)))
			{
				worker.GET("/profile", controllers.GetWorkerProfile)
				worker.PUT("/profile", controllers.UpdateWorkerProfile)
				worker.PUT("/availability", controllers.SetAvailability)
				worker.GET("/pending-bookings", controllers.GetPendingBookings)
				worker.PUT("/bookings/:id/accept", controllers.AcceptBooking)
				worker.PUT("/bookings/:id/start", controllers.StartBooking)
				worker.PUT("/bookings/:id/complete", controllers.CompleteBooking)
			}

			// Customer reviews
			reviews := protected.Group("/reviews")
			{
				reviews.POST("", controllers.CreateReview)
			}

			// Admin routes
			admin := protected.Group("/admin")
			admin.Use(middleware.RoleMiddleware(string(models.RoleAdmin)))
			{
				admin.POST("/services", controllers.CreateService)
				admin.POST("/categories", controllers.CreateCategory)
			}
		}
	}

	return r
}

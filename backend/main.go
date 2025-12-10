package main

import (
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/mtmh3135/honor/backend/config"
	"github.com/mtmh3135/honor/backend/handlers"
	"github.com/mtmh3135/honor/backend/middleware"
)

func main() {
	// API BACKEND
	// init DB & master
	config.InitDB()
	app := fiber.New(fiber.Config{
		ReadTimeout:  0,
		WriteTimeout: 0,
	})

	// Origin Deploy "http://localhost:8080, http://208.67.222.222:8080, http://127.0.0.1:8080  "
	app.Use(cors.New(cors.Config{
		AllowOrigins:     "http://localhost:5173",
		AllowMethods:     "GET,POST,PUT,DELETE,OPTIONS",
		AllowCredentials: true,
		AllowHeaders:     "Origin, Content-Type, Accept, Authorization",
	}))
	// Handling Login
	app.Post("/api/login", middleware.Login)

	// Auth
	api := app.Group("/api", middleware.AuthRequired)
	api.Get("/me", middleware.Me)

	// Handling Patient Bill
	api.Post("/upload-patientbill-chunk", handlers.UploadChunk)
	api.Post("/upload-patientbill-complete", handlers.UploadComplete)
	api.Get("/get-patientbills", handlers.GetPatientbills(config.DB))

	// Handling Data Perbandingan
	api.Post("/upload-comparisondata-chunk", handlers.UploadChunks)
	api.Post("/upload-comparisondata-complete", handlers.UploadCompletes)
	api.Get("/get-comparisondata", handlers.GetComparison(config.DB))

	// Handling Piutang
	api.Post("/upload-ais", handlers.HandleUploadAis)
	api.Get("/get-ais-data", handlers.GetPiutang(config.DB))

	// Handling Master TXN
	api.Post("/uploadtxn", handlers.HandleUploadtxn)
	api.Get("/get-txn-data", handlers.GetTxn(config.DB))
	api.Post("/create-txn", handlers.CreateTxn(config.DB))
	api.Put("/update-txn/:id", handlers.UpdateTxn(config.DB))
	api.Delete("/delete-txn/:id", handlers.DeleteTxn(config.DB))

	// Handling Master Doctor
	api.Post("/upload-doctor", handlers.HandleUploaddoctordata)
	api.Get("/get-doctor-data", handlers.GetDoctor(config.DB))
	api.Get("/get-doctor-list", handlers.GetDoctorList(config.DB))
	api.Post("/create-doctor", handlers.CreateDoctor(config.DB))
	api.Put("/update-doctor/:id", handlers.UpdateDoctor(config.DB))
	api.Delete("/delete-doctor/:id", handlers.DeleteDoctor(config.DB))

	// Handling Honor
	api.Post("/honor-count", handlers.HonorCountPatientBill(config.DB))
	api.Get("/get-honor-data", handlers.GetHonor(config.DB))
	api.Get("/get-doctor-honor", handlers.GetDoctorHonor(config.DB))
	api.Get("/get-doctor-honor-monthly", handlers.GetDoctorHonorMonthly(config.DB))
	api.Post("/add-honor-adjustment", handlers.AddHonorAdjustment(config.DB))
	api.Get("/honor-chart", handlers.GetHonorChart(config.DB))

	// Handling Update Billing
	api.Get("/get-update-billing", handlers.GetUpdateBillingData(config.DB))

	// Handling Users
	api.Get("/users", handlers.GetUsers)
	api.Post("/create-user", middleware.AdminOnly, handlers.CreateUser)
	api.Put("/edit-user/:id", middleware.AdminOnly, handlers.UpdateUser)
	api.Delete("/delete-user/:id", middleware.AdminOnly, handlers.DeleteUser)

	// Handling Request
	api.Get("/get-request-list", handlers.GetHonorRequests)
	api.Get("/request-list/:id", handlers.GetHonorRequestDetail)
	api.Post("/honor/submit-request", handlers.SubmitHonorRequest)
	api.Put("/honor-request/cancel/:id", handlers.CancelHonorRequest)
	api.Put("/honor-request/reject/:id", handlers.RejectHonorRequest)
	api.Put("/honor/approve/1/:id", handlers.ApproveLevel1)
	api.Put("/honor/approve/2/:id", handlers.ApproveLevel2)
	api.Put("/honor/approve/3/:id", handlers.ApproveLevel3)

	// //API FRONTEND
	// // Serve file static (CSS, JS, gambar)
	// app.Static("/", "./public")

	// // Fallback route untuk React Router (SPA)
	// app.Get("/*", func(c *fiber.Ctx) error {
	// 	if strings.HasPrefix(c.Path(), "/api") {
	// 		return fiber.ErrNotFound
	// 	}
	// 	return c.SendFile("./public/index.html")
	// })

	// Run Server
	log.Println("Listening: http://localhost:8080")
	log.Fatal(app.Listen(":8080"))
}

package main

// package main

// import (
// 	"database/sql"
// 	"encoding/csv"
// 	"fmt"
// 	"log"
// 	"strings"

// 	"github.com/gofiber/fiber/v2"
// 	_ "github.com/lib/pq" // contoh PostgreSQL
// )

// func maina() {
// 	app := fiber.New()

// 	db, err := sql.Open("postgres", "host=localhost user=postgres password=123 dbname=testdb sslmode=disable")
// 	if err != nil {
// 		log.Fatal(err)
// 	}
// 	defer db.Close()

// 	app.Post("/api/upload", func(c *fiber.Ctx) error {
// 		file, err := c.FormFile("file")
// 		if err != nil {
// 			return c.Status(400).SendString("File not found")
// 		}

// 		f, err := file.Open()
// 		if err != nil {
// 			return err
// 		}
// 		defer f.Close()

// 		reader := csv.NewReader(f)
// 		records, err := reader.ReadAll()
// 		if err != nil {
// 			return err
// 		}

// 		// asumsi baris pertama header
// 		for i, row := range records {
// 			if i == 0 {
// 				continue // skip header
// 			}

// 			if len(row) < 2 {
// 				continue // pastikan ada minimal 2 kolom
// 			}

// 			name := strings.TrimSpace(row[0])     // kolom 1 = nama user
// 			codeData := strings.TrimSpace(row[1]) // kolom 2 = kode campuran

// 			parts := strings.Split(codeData, "-")
// 			if len(parts) < 2 {
// 				continue // skip kalau format salah
// 			}

// 			userID := parts[0]
// 			role := parts[1]

// 			// Insert ke DB
// 			_, err := db.Exec(`
// 				INSERT INTO users (name, user_id, role)
// 				VALUES ($1, $2, $3)
// 			`, name, userID, role)
// 			if err != nil {
// 				fmt.Println("Insert error:", err)
// 				continue
// 			}
// 		}

// 		return c.JSON(fiber.Map{
// 			"message": "Data successfully processed and inserted",
// 		})
// 	})

// 	app.Listen(":8080")
// }

package processor

import (
	"database/sql"
	"log"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/mtmh3135/honor/backend/models"
)

// GET /api/comparisondata?page=1&visit_number=...&status=...
func GetComparison(db *sql.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// ambil query filter
		visitNumber := c.Query("visit_number")
		status := c.Query("status")

		// ambil query pagination
		pageStr := c.Query("page", "1")
		page, err := strconv.Atoi(pageStr)
		if err != nil || page < 1 {
			page = 1
		}
		limit := 10
		offset := (page - 1) * limit

		// query dasar
		query := `SELECT visit_number, status, visit_number_tujuan, tarif_ina_cbg 
		          FROM comparison_data WHERE 1=1`
		args := []interface{}{}

		// filter dinamis
		if visitNumber != "" {
			query += " AND visit_number LIKE ?"
			args = append(args, "%"+visitNumber+"%")
		}
		if status != "" {
			query += " AND status = ?"
			args = append(args, strings.ToUpper(status))
		}

		// tambahkan pagination
		query += " LIMIT ? OFFSET ?"
		args = append(args, limit, offset)

		// eksekusi query
		rows, err := db.Query(query, args...)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": err.Error()})
		}
		defer rows.Close()

		// mapping data
		var data []models.ComparisonData
		for rows.Next() {
			var d models.ComparisonData
			if err := rows.Scan(&d.VisitNumber, &d.Status, &d.VisitNumberTujuan, &d.TarifINACBG); err != nil {
				return c.Status(500).JSON(fiber.Map{"error": err.Error()})
			}
			data = append(data, d)
		}

		// ambil total data (tanpa limit)
		countQuery := `SELECT COUNT(*) FROM comparison_data WHERE 1=1`
		countArgs := []interface{}{}
		if visitNumber != "" {
			countQuery += " AND visit_number LIKE ?"
			countArgs = append(countArgs, "%"+visitNumber+"%")
		}
		if status != "" {
			countQuery += " AND status = ?"
			countArgs = append(countArgs, strings.ToUpper(status))
		}

		var total int
		if err := db.QueryRow(countQuery, countArgs...).Scan(&total); err != nil {
			return c.Status(500).JSON(fiber.Map{"error": err.Error()})
		}

		log.Printf("✅ Page %d, total %d data", page, total)

		return c.JSON(fiber.Map{
			"page":       page,
			"limit":      limit,
			"total":      total,
			"totalPages": (total + limit - 1) / limit,
			"data":       data,
		})
	}
}

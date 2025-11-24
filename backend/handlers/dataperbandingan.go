package handlers

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"io"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/mtmh3135/honor/backend/helpers"
	"github.com/mtmh3135/honor/backend/models"
	"github.com/mtmh3135/honor/backend/processor"
)

// GET /api/comparison
func GetComparison(db *sql.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {

		VisitNumber := c.Query("visit_number")
		Status := c.Query("status")

		// ambil query pagination
		pageStr := c.Query("page", "1") // default halaman 1
		limitStr := c.Query("limit", "10")

		page, err := strconv.Atoi(pageStr)
		if err != nil || page < 1 {
			page = 1
		}
		limit, err := strconv.Atoi(limitStr)
		if err != nil || limit < 1 {
			limit = 10
		}
		offset := (page - 1) * limit

		query := `SELECT visit_number,status,visit_number_tujuan,tarif_ina_cbg FROM comparison_data WHERE 1=1`
		args := []interface{}{}

		// hanya tambahkan filter jika user mengisi
		if VisitNumber != "" {
			query += " AND visit_number LIKE ?"
			args = append(args, "%"+VisitNumber+"%")
		}
		if Status != "" {
			query += " AND status = ?"
			args = append(args, strings.ToUpper(Status))
		}

		// tambahkan pagination
		query += " LIMIT ? OFFSET ?"
		args = append(args, limit, offset)

		rows, err := db.Query(query, args...)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": err.Error()})
		}
		defer rows.Close()

		var comparisondatas []models.ComparisonData
		for rows.Next() {
			var p models.ComparisonData
			if err := rows.Scan(&p.VisitNumber, &p.Status, &p.VisitNumberTujuan, &p.TarifINACBG); err != nil {
				return c.Status(500).JSON(fiber.Map{"error": err.Error()})
			}
			comparisondatas = append(comparisondatas, p)
		}

		// ambil total data (tanpa limit)
		countQuery := `SELECT COUNT(*) FROM comparison_data WHERE 1=1`
		countArgs := []interface{}{}
		if VisitNumber != "" {
			countQuery += " AND visit_number LIKE ?"
			countArgs = append(countArgs, "%"+VisitNumber+"%")
		}
		if Status != "" {
			countQuery += " AND status = ?"
			countArgs = append(countArgs, strings.ToUpper(Status))
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
			"data":       comparisondatas,
		})
	}
}

var uploadTempDirs = "./tmp_uploads"

func init() {
	os.MkdirAll(uploadTempDirs, 0755)
}

func UploadChunks(c *fiber.Ctx) error {
	fileId := c.FormValue("fileId")
	chunkIndex := c.FormValue("chunkIndex")
	file, err := c.FormFile("chunk")
	if err != nil {
		return c.Status(400).SendString("missing chunk")
	}
	outPath := filepath.Join(uploadTempDirs, fileId+"."+chunkIndex+".part")
	if err := c.SaveFile(file, outPath); err != nil {
		return c.Status(500).SendString("save failed")
	}
	return c.SendStatus(200)
	// simpan log aktivitas

}

func UploadCompletes(c *fiber.Ctx) error {
	var payload struct {
		FileId   string `json:"fileId"`
		FileName string `json:"fileName"`
	}
	if err := json.Unmarshal(c.Body(), &payload); err != nil {
		return c.Status(400).SendString("bad json")
	}

	parts, err := filepath.Glob(filepath.Join(uploadTempDirs, payload.FileId+".*.part"))
	if err != nil || len(parts) == 0 {
		return c.Status(400).SendString("no parts")
	}

	// Sort parts by chunk index
	sort.Slice(parts, func(i, j int) bool {
		baseI := filepath.Base(parts[i])
		baseJ := filepath.Base(parts[j])
		indexI := extractChunkIndexs(baseI)
		indexJ := extractChunkIndexs(baseJ)
		return indexI < indexJ
	})

	var buf bytes.Buffer
	for _, p := range parts {
		part, err := os.Open(p)
		if err != nil {
			return c.Status(500).SendString("open part failed")
		}
		if _, err := io.Copy(&buf, part); err != nil {
			part.Close()
			return c.Status(500).SendString("copy part failed")
		}
		part.Close()
		os.Remove(p)
	}

	data := buf.Bytes()

	// ✅ Ambil user_id dari JWT
	userIDValue := c.Locals("user_id")
	UserID, ok := userIDValue.(int64)
	if !ok {
		return c.Status(401).JSON(fiber.Map{"error": "Unauthorized"})
	}
	// ✅ Log aktivitas
	if err := helpers.LogActivity(c, UserID, "Upload Data Perbandingan", "Upload file "+payload.FileName); err != nil {
		log.Println("❌ Gagal simpan log activity:", err)
	} else {
		log.Println("✅ Activity logged untuk user", UserID)
	}

	// Proses file di background
	if err := processor.ComparisonDataUp(bytes.NewReader(data)); err != nil {
		log.Println("process error:", err)
		return c.Status(400).JSON(fiber.Map{
			"status":  "error",
			"message": err.Error(),
		})
	}

	return c.JSON(fiber.Map{"status": "processing_started"})
}

func extractChunkIndexs(base string) int {
	// base = fileId.chunkIndex.part
	partIndex := strings.LastIndex(base, ".part")
	if partIndex == -1 {
		return 0
	}
	before := base[:partIndex]
	dotIndex := strings.LastIndex(before, ".")
	if dotIndex == -1 {
		return 0
	}
	indexStr := before[dotIndex+1:]
	index, _ := strconv.Atoi(indexStr)
	return index
}

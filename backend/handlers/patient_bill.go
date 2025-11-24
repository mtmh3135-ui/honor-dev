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

// GET /api/patients
func GetPatientbills(db *sql.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		PatientName := c.Query("patient_name")
		PatientClass := c.Query("patient_class")
		VisitNo := c.Query("visit_no")

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

		query := `SELECT id,
		visit_no, 
		patient_name, 
		patient_type, 
		patient_class, 
		txn_code, 
		txn_category, 
		txn_desc, 
		txn_doctor, 
		regn_doctor 
		FROM patient_bill WHERE 1=1`
		args := []interface{}{}

		// hanya tambahkan filter jika user mengisi
		if PatientName != "" {
			query += " AND patient_name LIKE ?"
			args = append(args, "%"+PatientName+"%")
		}

		if PatientClass != "" {
			query += " AND patient_class = ?"
			args = append(args, strings.ToUpper(PatientClass))
		}

		if VisitNo != "" {
			query += " AND visit_no LIKE ?"
			args = append(args, "%"+VisitNo+"%")
		}

		// tambahkan pagination
		query += " LIMIT ? OFFSET ?"
		args = append(args, limit, offset)

		rows, err := db.Query(query, args...)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": err.Error()})
		}
		defer rows.Close()

		var patientbills []models.Patientbill
		for rows.Next() {
			var p models.Patientbill
			if err := rows.Scan(&p.ID,
				&p.VisitNo,
				&p.PatientName,
				&p.PatientType,
				&p.PatientClass,
				&p.TxnCode,
				&p.TxnCategory,
				&p.TxnDesc,
				&p.TxnDoctor,
				&p.RegnDoctor); err != nil {
				return c.Status(500).JSON(fiber.Map{"error": err.Error()})
			}
			patientbills = append(patientbills, p)
		}

		// ambil total data (tanpa limit)
		countQuery := `SELECT COUNT(*) FROM patient_bill WHERE 1=1`
		countArgs := []interface{}{}
		if PatientName != "" {
			countQuery += " AND patient_name LIKE ?"
			countArgs = append(countArgs, "%"+PatientName+"%")
		}

		if PatientClass != "" {
			countQuery += " AND patient_class = ?"
			countArgs = append(countArgs, strings.ToUpper(PatientClass))
		}

		if VisitNo != "" {
			countQuery += " AND visit_no LIKE ?"
			countArgs = append(countArgs, "%"+VisitNo+"%")
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
			"data":       patientbills,
		})
	}
}

var uploadTempDir = "./tmp_uploads"

func init() {
	os.MkdirAll(uploadTempDir, 0755)
}

func UploadChunk(c *fiber.Ctx) error {
	fileId := c.FormValue("fileId")
	chunkIndex := c.FormValue("chunkIndex")
	file, err := c.FormFile("chunk")
	if err != nil {
		return c.Status(400).SendString("missing chunk")
	}
	outPath := filepath.Join(uploadTempDir, fileId+"."+chunkIndex+".part")
	if err := c.SaveFile(file, outPath); err != nil {
		return c.Status(500).SendString("save failed")
	}
	return c.SendStatus(200)
}

func UploadComplete(c *fiber.Ctx) error {
	var payload struct {
		FileId   string `json:"fileId"`
		FileName string `json:"fileName"`
	}
	if err := json.Unmarshal(c.Body(), &payload); err != nil {
		return c.Status(400).SendString("bad json")
	}

	parts, err := filepath.Glob(filepath.Join(uploadTempDir, payload.FileId+".*.part"))
	if err != nil || len(parts) == 0 {
		return c.Status(400).SendString("no parts")
	}

	// Sort parts by chunk index
	sort.Slice(parts, func(i, j int) bool {
		baseI := filepath.Base(parts[i])
		baseJ := filepath.Base(parts[j])
		indexI := extractChunkIndex(baseI)
		indexJ := extractChunkIndex(baseJ)
		return indexI < indexJ
	})

	// Gabung part satu per satu
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

	// Ambil userID dari JWT
	userIDValue := c.Locals("user_id")
	UserID, ok := userIDValue.(int64)
	if !ok {
		return c.Status(401).JSON(fiber.Map{"error": "Unauthorized"})
	}
	// ✅ Log aktivitas
	if err := helpers.LogActivity(c, UserID, "Upload Patient Bill", "Upload file "+payload.FileName); err != nil {
		log.Println("❌ Gagal simpan log activity:", err)
	} else {
		log.Println("✅ Activity logged untuk user", UserID)
	}

	if err := processor.ProcessXLSX(bytes.NewReader(data)); err != nil {
		log.Println("process error:", err)
		return c.Status(400).JSON(fiber.Map{
			"success": false,
			"message": err.Error(),
		})
	}

	return c.JSON(fiber.Map{"status": "processing_started"})
}

func extractChunkIndex(base string) int {
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

package handlers

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/mtmh3135/honor/backend/config"
	"github.com/mtmh3135/honor/backend/helpers"
	"github.com/xuri/excelize/v2"
)

type Insurance struct {
	Nama            string
	VisitNo         string
	TotalPembayaran float64
	SisaPembayaran  float64
}

func GetPiutang(db *sql.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
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

		query := `
		SELECT 
			visit_no,
			nama,
			total,
			sisa
		FROM piutang
		WHERE 1=1`
		args := []interface{}{}

		//filter
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

		var result []Insurance
		for rows.Next() {
			var p Insurance
			if err := rows.Scan(&p.VisitNo, &p.Nama, &p.TotalPembayaran, &p.SisaPembayaran); err != nil {
				return fiber.NewError(fiber.StatusInternalServerError, err.Error())
			}
			result = append(result, p)
		}
		countQuery := `SELECT COUNT(*) FROM piutang WHERE 1=1`
		countArgs := []interface{}{}
		//filter
		if VisitNo != "" {
			countQuery += " AND visit_no LIKE ?"
			countArgs = append(countArgs, "%"+VisitNo+"%")
		}

		var total int
		if err := db.QueryRow(countQuery, countArgs...).Scan(&total); err != nil {
			return c.Status(500).JSON(fiber.Map{"error": err.Error()})
		}

		return c.JSON(fiber.Map{
			"page":       page,
			"total":      len(result),
			"totalPages": (total + limit - 1) / limit,
			"data":       result,
		})
	}
}
func HandleUploadAis(c *fiber.Ctx) error {
	if db == nil {
		db = config.DB
	}

	// ---------------------------
	// Ambil file upload
	// ---------------------------
	fileHeader, err := c.FormFile("file")
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "file tidak ditemukan")
	}

	tmpPath := fmt.Sprintf("./tmp_%s", fileHeader.Filename)
	if err := c.SaveFile(fileHeader, tmpPath); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "gagal menyimpan file")
	}
	defer os.Remove(tmpPath)

	// ---------------------------
	// Baca pakai Excelize
	// ---------------------------
	f, err := excelize.OpenFile(tmpPath)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "gagal membuka file xlsx")
	}

	sheet := f.GetSheetName(0)
	rows, _ := f.GetRows(sheet)

	headerMap := map[string]int{}
	for i, h := range rows[0] {
		headerMap[strings.ToLower(strings.TrimSpace(h))] = i
	}

	get := func(row []string, key string) string {
		key = strings.ToLower(strings.TrimSpace(key))
		idx, ok := headerMap[key]
		if !ok || idx >= len(row) {
			return ""
		}
		return strings.TrimSpace(row[idx])
	}

	// --------------------------------------
	// 1) Filter pertama → hanya ambil kthis
	// --------------------------------------

	var Filtered [][]string
	for i := 1; i < len(rows); i++ {
		row := rows[i]
		// log.Printf("HEADER = %#v", rows[0])
		// log.Printf("Data : %s", row)
		sumber := strings.ToUpper(get(row, "sumber data"))

		void := get(row, "void")

		if sumber == "KTHIS" {
			if void == "FALSE" {
				Filtered = append(Filtered, row)
			}

		}
	}

	// for _, row := range Filtered {
	// 	log.Printf("Data : %v", row)
	// }
	// --------------------------------------
	// 2) Filter kedua → keterangan valid (- dan tidak ada /)
	// --------------------------------------

	var keteranganFiltered []Insurance

	for _, row := range Filtered {

		keterangan := get(row, "keterangan")

		parts := strings.Split(keterangan, "-")
		if len(parts) >= 4 {
			nama := strings.Join(parts[1:len(parts)-2], " ")
			visit := parts[len(parts)-2]

			sisaStr := get(row, "sisa")
			clean := strings.ReplaceAll(sisaStr, ",", "")
			sisa, _ := strconv.ParseFloat(clean, 64)

			totallunas := get(row, "total lunas")
			clean2 := strings.ReplaceAll(totallunas, ",", "")
			total, _ := strconv.ParseFloat(clean2, 64)

			keteranganFiltered = append(keteranganFiltered, Insurance{
				Nama:            nama,
				VisitNo:         visit,
				TotalPembayaran: total,
				SisaPembayaran:  sisa,
			})
		} else {
			log.Printf("Data ini tidak sesuai %v", row)
		}

	}
	// for i := range Filtered {
	// log.Printf("Data : %v", i)
	// }
	// for _, row := range keteranganFiltered {
	// 	log.Printf("Data : %v", row)
	// }
	log.Printf("Total data setelah filter keterangan = %d", len(keteranganFiltered))

	// --------------------------------------
	// 3) Insert ke Database
	// --------------------------------------

	inserted, err := insert(keteranganFiltered)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}
	// ✅ Ambil user_id dari JWT
	userIDValue := c.Locals("user_id")
	UserID, ok := userIDValue.(int64)
	if !ok {
		return c.Status(401).JSON(fiber.Map{"error": "Unauthorized"})
	}
	// ✅ Log aktivitas
	if err := helpers.LogActivity(c, UserID, "Upload Data Piutang", "Upload file"); err != nil {
		log.Println("❌ Gagal simpan log activity:", err)
	} else {
		log.Println("✅ Activity logged untuk user", UserID)
	}
	return c.JSON(fiber.Map{
		"from_file":        len(rows) - 1,
		"after_status":     len(Filtered),
		"after_keterangan": len(keteranganFiltered),
		"inserted":         inserted,
	})
}

func insert(insurance []Insurance) (int, error) {
	tx, err := db.Begin()
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare(`
		INSERT INTO piutang (visit_no, nama,total, sisa)
		VALUES (?, ?, ?, ?)
		ON DUPLICATE KEY UPDATE
			total = VALUES(total),	
			sisa = VALUES(sisa)
	`)
	if err != nil {
		return 0, err
	}
	defer stmt.Close()

	count := 0
	for _, r := range insurance {
		_, err := stmt.Exec(r.VisitNo, r.Nama, r.TotalPembayaran, r.SisaPembayaran)
		if err != nil {
			return count, fmt.Errorf("error pada row %d: %v", count+1, err)
		}
		count++
	}

	if err := tx.Commit(); err != nil {
		return 0, err
	}

	return count, nil
}

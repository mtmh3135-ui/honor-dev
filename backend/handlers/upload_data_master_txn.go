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

var db *sql.DB

type Mastertxn struct {
	TxnId        int
	TxnCode      string
	TxnDesc      string
	TxnCategory  string
	TxnType      string
	BPJSIP       string
	BPJSOP       string
	RumusGeneral string
}

func GetTxn(db *sql.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		Txn_desc := c.Query("txn_desc")
		Txn_code := c.Query("txn_code")
		Txn_category := c.Query("txn_category")
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
			txn_id,txn_code, txn_desc, txn_category,txn_type,bpjs_ip,bpjs_op,rumus_general
		FROM master_txn
		WHERE 1=1`
		args := []interface{}{}

		//filter
		if Txn_desc != "" {
			query += " AND txn_desc LIKE ?"
			args = append(args, "%"+Txn_desc+"%")
		}

		if Txn_code != "" {
			query += " AND txn_code LIKE ?"
			args = append(args, "%"+Txn_code+"%")
		}

		if Txn_category != "" {
			query += " AND txn_category = ?"
			args = append(args, strings.ToUpper(Txn_category))
		}
		// tambahkan pagination
		query += " LIMIT ? OFFSET ?"
		args = append(args, limit, offset)

		rows, err := db.Query(query, args...)

		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": err.Error()})
		}
		defer rows.Close()

		var result []Mastertxn
		for rows.Next() {
			var p Mastertxn
			if err := rows.Scan(&p.TxnId, &p.TxnCode, &p.TxnDesc, &p.TxnCategory, &p.TxnType, &p.BPJSIP, &p.BPJSOP, &p.RumusGeneral); err != nil {
				return fiber.NewError(fiber.StatusInternalServerError, err.Error())
			}
			result = append(result, p)
		}
		countQuery := `SELECT COUNT(*) FROM master_txn WHERE 1=1`
		countArgs := []interface{}{}
		//filter
		if Txn_desc != "" {
			countQuery += " AND txn_desc LIKE ?"
			countArgs = append(countArgs, "%"+Txn_desc+"%")
		}

		if Txn_code != "" {
			countQuery += " AND txn_code LIKE ?"
			countArgs = append(countArgs, "%"+Txn_code+"%")
		}

		if Txn_category != "" {
			countQuery += " AND txn_category = ?"
			countArgs = append(countArgs, strings.ToUpper(Txn_category))
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
func HandleUploadtxn(c *fiber.Ctx) error {
	if db == nil {
		db = config.DB
	}
	fileHeader, err := c.FormFile("file")
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "file tidak ditemukan")
	}

	// Simpan file sementara
	tmpPath := fmt.Sprintf("./tmp_%s", fileHeader.Filename)
	if err := c.SaveFile(fileHeader, tmpPath); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "gagal menyimpan file")
	}
	defer os.Remove(tmpPath)

	// Buka excel
	f, err := excelize.OpenFile(tmpPath)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "gagal membuka file excel")
	}
	defer f.Close()

	sheets := f.GetSheetList()
	if len(sheets) == 0 {
		return fiber.NewError(fiber.StatusBadRequest, "file kosong")
	}
	rows, err := f.GetRows(sheets[0])
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "gagal membaca sheet")
	}

	// Parsing rows
	var data []Mastertxn
	for i, row := range rows {
		if i == 0 {
			continue // skip header
		}
		if len(row) < 3 {
			continue
		}

		data = append(data, Mastertxn{

			TxnCode:      row[0],
			TxnDesc:      row[1],
			TxnCategory:  row[2],
			TxnType:      row[3],
			BPJSIP:       row[4],
			BPJSOP:       row[5],
			RumusGeneral: row[6],
		})
	}

	// Insert batch
	inserted, err := insertBatch(data)
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
	if err := helpers.LogActivity(c, UserID, "Upload Data Txn", "Upload file"); err != nil {
		log.Println("❌ Gagal simpan log activity:", err)
	} else {
		log.Println("✅ Activity logged untuk user", UserID)
	}
	return c.JSON(fiber.Map{
		"total_rows": len(data),
		"inserted":   inserted,
	})
}

func insertBatch(mastertxn []Mastertxn) (int, error) {
	tx, err := db.Begin()
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare(`
		INSERT INTO master_txn ( txn_code, txn_desc, txn_category,txn_type,bpjs_ip,bpjs_op,rumus_general)
		VALUES (?, ?, ?,?,?,?,?)
		ON DUPLICATE KEY UPDATE
			txn_desc = VALUES(txn_desc),
			txn_category = VALUES(txn_category),
			txn_type = VALUES(txn_type),
			bpjs_ip = VALUES(bpjs_ip),
			bpjs_op = VALUES(bpjs_op),
			rumus_general = VALUES(rumus_general),
	`)
	if err != nil {
		return 0, err
	}
	defer stmt.Close()

	count := 0
	for _, r := range mastertxn {
		_, err := stmt.Exec(r.TxnId, r.TxnCode, r.TxnDesc, r.TxnCategory, r.TxnType, r.BPJSIP, r.BPJSOP, r.RumusGeneral)
		if err != nil {
			return count, fmt.Errorf("insert gagal di row %d: %v", count+1, err)
		}
		count++
	}

	if err := tx.Commit(); err != nil {
		return 0, err
	}
	return count, nil
}

func CreateTxn(db *sql.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var d Mastertxn

		if err := c.BodyParser(&d); err != nil {
			return fiber.NewError(400, "Invalid body")
		}
		log.Println("RAW BODY:", string(c.Body()))
		log.Printf("Parsed Data: %+v\n", d)
		// Cek duplikasi txn desc
		var count int
		err := db.QueryRow(`
    	SELECT COUNT(*) FROM master_txn 
    		WHERE txn_desc = ?
		`, d.TxnDesc).Scan(&count)

		if err != nil {
			return fiber.NewError(500, err.Error())
		}
		if count > 0 {
			return fiber.NewError(400, "Data dengan Txn Desc ini sudah digunakan")
		}

		// Cek duplikasi txn_code
		err = db.QueryRow(`
    	SELECT COUNT(*) FROM master_txn 
    		WHERE txn_code = ?
		`, d.TxnCode).Scan(&count)

		if err != nil {
			return fiber.NewError(500, err.Error())
		}
		if count > 0 {
			return fiber.NewError(400, "Txn Code ini sudah terdaftar")
		}

		if err := c.BodyParser(&d); err != nil {
			return fiber.NewError(400, "Invalid body")
		}

		_, err = db.Exec(`
            INSERT INTO master_txn ( txn_code, txn_desc, txn_category,txn_type,bpjs_ip,bpjs_op,rumus_general)
		VALUES (?, ?, ?,?,?,?,?)
        `, d.TxnCode, d.TxnDesc, d.TxnCategory, d.TxnType, d.BPJSIP, d.BPJSOP, d.RumusGeneral)

		if err != nil {
			return fiber.NewError(500, err.Error())
		}
		// ✅ Ambil user_id dari JWT
		userIDValue := c.Locals("user_id")
		UserID, ok := userIDValue.(int64)
		if !ok {
			return c.Status(401).JSON(fiber.Map{"error": "Unauthorized"})
		}
		// ✅ Log aktivitas
		if err := helpers.LogActivity(c, UserID, "Add Txn", "Add "+d.TxnDesc); err != nil {
			log.Println("❌ Gagal simpan log activity:", err)
		} else {
			log.Println("✅ Activity logged untuk user", UserID)
		}
		return c.JSON(fiber.Map{
			"message": "Data Txn berhasil ditambahkan",
		})
	}
}

func UpdateTxn(db *sql.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {

		id := c.Params("id")
		var d Mastertxn

		if err := c.BodyParser(&d); err != nil {
			return fiber.NewError(400, "Invalid body")
		}
		log.Println("RAW BODY:", string(c.Body()))
		log.Printf("Parsed Data: %+v\n", d)
		// Cek duplikasi nama
		var count int
		err := db.QueryRow(`
			SELECT COUNT(*) FROM master_txn
			WHERE txn_desc = ?
			  AND txn_id != ?
		`, d.TxnDesc, id).Scan(&count)

		if err != nil {
			return fiber.NewError(500, err.Error())
		}
		if count > 0 {
			return fiber.NewError(400, "Data dengan Txn Desc ini sudah digunakan")
		}

		// Cek duplikasi careprovider id
		err = db.QueryRow(`
			SELECT COUNT(*) FROM master_txn
			WHERE txn_code = ?
				AND txn_id = ?
		`, d.TxnCode, id).Scan(&count)

		if err != nil {
			return fiber.NewError(500, err.Error())
		}
		if count > 0 {
			return fiber.NewError(400, "Txn Code ini sudah terdaftar")
		}

		// Update data
		_, err = db.Exec(`
			UPDATE master_txn
			SET txn_code = ?, txn_desc = ?, txn_category = ?,txn_type = ?,bpjs_ip = ?,bpjs_op = ?,rumus_general = ?
			WHERE txn_id = ?
		`, d.TxnCode, d.TxnDesc, d.TxnCategory, d.TxnType, d.BPJSIP, d.BPJSOP, d.RumusGeneral, id)

		if err != nil {
			return fiber.NewError(500, err.Error())
		}
		// ✅ Ambil user_id dari JWT
		userIDValue := c.Locals("user_id")
		UserID, ok := userIDValue.(int64)
		if !ok {
			return c.Status(401).JSON(fiber.Map{"error": "Unauthorized"})
		}
		// ✅ Log aktivitas
		if err := helpers.LogActivity(c, UserID, "Update Txn", "Update "+d.TxnDesc); err != nil {
			log.Println("❌ Gagal simpan log activity:", err)
		} else {
			log.Println("✅ Activity logged untuk user", UserID)
		}
		return c.JSON(fiber.Map{"message": "Update berhasil"})
	}
}

func DeleteTxn(db *sql.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		id := c.Params("id")
		log.Println("RAW BODY:", string(c.Body()))
		_, err := db.Exec(`
            DELETE FROM master_txn WHERE txn_id = ?
        `, id)

		if err != nil {
			return fiber.NewError(500, err.Error())
		}
		// ✅ Ambil user_id dari JWT
		userIDValue := c.Locals("user_id")
		UserID, ok := userIDValue.(int64)
		if !ok {
			return c.Status(401).JSON(fiber.Map{"error": "Unauthorized"})
		}
		// ✅ Log aktivitas
		if err := helpers.LogActivity(c, UserID, "Delete Txn", "Delete "+id); err != nil {
			log.Println("❌ Gagal simpan log activity:", err)
		} else {
			log.Println("✅ Activity logged untuk user", UserID)
		}
		return c.JSON(fiber.Map{"message": "Delete berhasil"})
	}
}

package handlers

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/mtmh3135/honor/backend/config"
	"github.com/mtmh3135/honor/backend/helpers"
	"github.com/xuri/excelize/v2"
)

var db1 *sql.DB

// upload data dokter
type Doctor struct {
	IdDoctor                int    `json:"IdDoctor"`
	DoctorName              string `json:"DoctorName"`
	CareproviderTxnDoctorId int64  `json:"CareproviderTxnDoctorId"`
	Description             string `json:"Description"`
}

func GetDoctor(db *sql.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		DoctorName := c.Query("doctor_name")
		Description := c.Query("description")
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
			id_doctor,
			doctor_name,
			description,
			careprovider_txn_doctor_id
		FROM doctor_data
		WHERE 1=1`
		args := []interface{}{}

		//filter
		if DoctorName != "" {
			query += " AND doctor_name LIKE ?"
			args = append(args, "%"+DoctorName+"%")
		}

		if Description != "" {
			query += " AND description LIKE ?"
			args = append(args, "%"+Description+"%")
		}

		// tambahkan pagination
		query += " LIMIT ? OFFSET ?"
		args = append(args, limit, offset)

		rows, err := db.Query(query, args...)

		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": err.Error()})
		}
		defer rows.Close()

		var result []Doctor
		for rows.Next() {
			var p Doctor
			if err := rows.Scan(&p.IdDoctor, &p.DoctorName, &p.Description, &p.CareproviderTxnDoctorId); err != nil {
				return fiber.NewError(fiber.StatusInternalServerError, err.Error())
			}
			result = append(result, p)
		}
		countQuery := `SELECT COUNT(*) FROM doctor_data WHERE 1=1`
		countArgs := []interface{}{}
		//filter
		if DoctorName != "" {
			countQuery += " AND doctor_name LIKE ?"
			countArgs = append(countArgs, "%"+DoctorName+"%")
		}

		if Description != "" {
			countQuery += " AND description LIKE ?"
			countArgs = append(countArgs, "%"+Description+"%")
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

func GetDoctorList(db *sql.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		DoctorName := c.Query("doctor_name")
		query := `
		SELECT 
			id_doctor,
			doctor_name,
			description,
			careprovider_txn_doctor_id
		FROM doctor_data
		WHERE 1=1`
		args := []interface{}{}

		//filter
		if DoctorName != "" {
			query += " AND doctor_name LIKE ?"
			args = append(args, "%"+DoctorName+"%")
		}
		rows, err := db.Query(query, args...)

		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": err.Error()})
		}
		defer rows.Close()

		var result []Doctor
		for rows.Next() {
			var p Doctor
			if err := rows.Scan(&p.IdDoctor, &p.DoctorName, &p.Description, &p.CareproviderTxnDoctorId); err != nil {
				return fiber.NewError(fiber.StatusInternalServerError, err.Error())
			}
			result = append(result, p)

		}
		log.Printf("data:%v", result)
		return c.JSON(fiber.Map{

			"data": result,
		})
	}
}

func HandleUploaddoctordata(c *fiber.Ctx) error {
	if db1 == nil {
		db1 = config.DB
	}
	fileHeader, err := c.FormFile("file")
	if err != nil {
		log.Println("file tidak ditemukan")
		return fiber.NewError(fiber.StatusBadRequest, "file tidak ditemukan")
	}

	// Simpan file sementara
	tmpPath := fmt.Sprintf("./tmp_%s", fileHeader.Filename)
	if err := c.SaveFile(fileHeader, tmpPath); err != nil {
		log.Println("gagal menyimpan file")
		return fiber.NewError(fiber.StatusInternalServerError, "gagal menyimpan file")
	}
	defer os.Remove(tmpPath)

	// Buka excel
	f, err := excelize.OpenFile(tmpPath)
	if err != nil {
		log.Println("gagal membuka file excel")
		return fiber.NewError(fiber.StatusBadRequest, "gagal membuka file excel")
	}
	defer f.Close()

	sheets := f.GetSheetList()
	if len(sheets) == 0 {
		log.Println("file kosong")
		return fiber.NewError(fiber.StatusBadRequest, "file kosong")
	}
	rows, err := f.GetRows(sheets[0])
	if err != nil {
		log.Println("gagal membaca sheet")
		return fiber.NewError(fiber.StatusInternalServerError, "gagal membaca sheet")
	}

	// Parsing rows
	var data []Doctor
	for i, row := range rows {
		if i == 0 {
			continue // skip header
		}
		if len(row) < 1 {
			continue
		}
		cpId, _ := strconv.ParseInt(row[2], 10, 64)
		data = append(data, Doctor{
			DoctorName:              row[0],
			Description:             row[1],
			CareproviderTxnDoctorId: cpId,
		})
	}

	// Insert batch
	inserted, err := insertBatch1(data)
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
	if err := helpers.LogActivity(c, UserID, "Upload Doctor Data", "Upload file "+fileHeader.Filename); err != nil {
		log.Println("❌ Gagal simpan log activity:", err)
	} else {
		log.Println("✅ Activity logged untuk user", UserID)
	}
	return c.JSON(fiber.Map{
		"total_rows": len(data),
		"inserted":   inserted,
	})
}

func insertBatch1(mastertxn []Doctor) (int, error) {
	tx, err := db1.Begin()
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare(`
		INSERT INTO doctor_data (doctor_name,description,careprovider_txn_doctor_id)
		VALUES (?,?,?)
		ON DUPLICATE KEY UPDATE
			doctor_name = VALUES(doctor_name),
			description = VALUES(description)
	`)
	if err != nil {
		return 0, err
	}
	defer stmt.Close()

	count := 0
	for _, r := range mastertxn {
		_, err := stmt.Exec(r.DoctorName, r.Description, r.CareproviderTxnDoctorId)
		if err != nil {
			log.Printf("insert gagal di row %d: %v", count+1, err)
			return count, fmt.Errorf("insert gagal di row %d: %v", count+1, err)
		}
		count++
	}

	if err := tx.Commit(); err != nil {
		return 0, err
	}
	return count, nil
}

func CreateDoctor(db *sql.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var d Doctor

		if err := c.BodyParser(&d); err != nil {
			return fiber.NewError(400, "Invalid body")
		}
		log.Println("RAW BODY:", string(c.Body()))
		log.Printf("Parsed Data: %+v\n", d)
		// Cek duplikasi nama
		var count int
		err := db.QueryRow(`
    	SELECT COUNT(*) FROM doctor_data 
    		WHERE doctor_name = ?
		`, d.DoctorName).Scan(&count)

		if err != nil {
			return fiber.NewError(500, err.Error())
		}
		if count > 0 {
			return fiber.NewError(400, "Nama dokter sudah digunakan")
		}

		// Cek duplikasi careprovider_txn_doctor_id
		err = db.QueryRow(`
    	SELECT COUNT(*) FROM doctor_data 
    		WHERE careprovider_txn_doctor_id = ?
		`, d.CareproviderTxnDoctorId).Scan(&count)

		if err != nil {
			return fiber.NewError(500, err.Error())
		}
		if count > 0 {
			return fiber.NewError(400, "Careprovider ID sudah terdaftar")
		}

		if err := c.BodyParser(&d); err != nil {
			return fiber.NewError(400, "Invalid body")
		}

		_, err = db.Exec(`
            INSERT INTO doctor_data (doctor_name, description, careprovider_txn_doctor_id)
            VALUES (?, ?, ?)
        `, d.DoctorName, d.Description, d.CareproviderTxnDoctorId)

		if err != nil {
			return fiber.NewError(500, err.Error())
		}
		// Ambil userID dari JWT
		userIDValue := c.Locals("user_id")
		UserID, ok := userIDValue.(int64)
		if !ok {
			return c.Status(401).JSON(fiber.Map{"error": "Unauthorized"})
		}
		// ✅ Log aktivitas
		if err := helpers.LogActivity(c, UserID, "Add Doctor", "Add Doctor "+d.DoctorName); err != nil {
			log.Println("❌ Gagal simpan log activity:", err)
		} else {
			log.Println("✅ Activity logged untuk user", UserID)
		}
		return c.JSON(fiber.Map{
			"message": "Data dokter berhasil ditambahkan",
		})
	}
}

func UpdateDoctor(db *sql.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {

		id := c.Params("id")
		var d Doctor

		if err := c.BodyParser(&d); err != nil {
			return fiber.NewError(400, "Invalid body")
		}
		log.Println("RAW BODY:", string(c.Body()))
		log.Printf("Parsed Data: %+v\n", d)
		// Cek duplikasi nama
		var count int
		err := db.QueryRow(`
			SELECT COUNT(*) FROM doctor_data
			WHERE doctor_name = ?
			  AND id_doctor != ?
		`, d.DoctorName, id).Scan(&count)

		log.Printf("nama: %s", d.DoctorName)
		log.Printf("nama: %d", d.CareproviderTxnDoctorId)
		log.Printf("nama: %s", d.Description)
		log.Printf("count : %v", count)
		if err != nil {
			return fiber.NewError(500, err.Error())
		}
		if count > 0 {
			return fiber.NewError(400, "Nama dokter sudah digunakan")
		}

		// Cek duplikasi careprovider id
		err = db.QueryRow(`
			SELECT COUNT(*) FROM doctor_data
			WHERE careprovider_txn_doctor_id = ?
			  AND id_doctor != ?
		`, d.CareproviderTxnDoctorId, id).Scan(&count)

		if err != nil {
			return fiber.NewError(500, err.Error())
		}
		if count > 0 {
			return fiber.NewError(400, "Careprovider ID sudah digunakan dokter lain")
		}

		// Update data
		_, err = db.Exec(`
			UPDATE doctor_data
			SET doctor_name = ?, description = ?, careprovider_txn_doctor_id = ?
			WHERE id_doctor = ?
		`, d.DoctorName, d.Description, d.CareproviderTxnDoctorId, id)

		if err != nil {
			return fiber.NewError(500, err.Error())
		}
		// Ambil userID dari JWT
		userIDValue := c.Locals("user_id")
		UserID, ok := userIDValue.(int64)
		if !ok {
			return c.Status(401).JSON(fiber.Map{"error": "Unauthorized"})
		}
		// ✅ Log aktivitas
		if err := helpers.LogActivity(c, UserID, "Update Doctor", "Update Doctor "+d.DoctorName); err != nil {
			log.Println("❌ Gagal simpan log activity:", err)
		} else {
			log.Println("✅ Activity logged untuk user", UserID)
		}
		return c.JSON(fiber.Map{"message": "Update berhasil"})
	}
}

func DeleteDoctor(db *sql.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		id := c.Params("id")

		_, err := db.Exec(`
            DELETE FROM doctor_data WHERE id_doctor = ?
        `, id)

		if err != nil {
			return fiber.NewError(500, err.Error())
		}
		// Ambil userID dari JWT
		userIDValue := c.Locals("user_id")
		UserID, ok := userIDValue.(int64)
		if !ok {
			return c.Status(401).JSON(fiber.Map{"error": "Unauthorized"})
		}
		// ✅ Log aktivitas
		if err := helpers.LogActivity(c, UserID, "Delete Doctor", "Delete Doctor "+id); err != nil {
			log.Println("❌ Gagal simpan log activity:", err)
		} else {
			log.Println("✅ Activity logged untuk user", UserID)
		}
		return c.JSON(fiber.Map{"message": "Delete berhasil"})
	}
}

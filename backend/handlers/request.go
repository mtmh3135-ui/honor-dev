package handlers

import (
	"database/sql"
	"log"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/mtmh3135/honor/backend/config"
	"github.com/mtmh3135/honor/backend/helpers"
	"github.com/mtmh3135/honor/backend/models"
)

// Get semua permohonan (filter by role)
func GetHonorRequests(c *fiber.Ctx) error {
	role := c.Locals("role")
	userId := c.Locals("user_id").(int64)
	log.Printf("user : %d dengan role %s", userId, role)
	// Pagination
	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "10"))
	offset := (page - 1) * limit

	// Filters
	month := c.Query("month")
	year := c.Query("year")
	status := c.Query("status")

	// Query SQL dengan created_by disertakan
	query := `
    SELECT h.id, h.description, h.counted_month, h.counted_year, h.status,
           h.created_by, u.username, h.created_at, 
           h.approved_lvl1, h.approved_lvl2, h.cancelled_at
    FROM honor_request h
    LEFT JOIN user u ON h.created_by = u.user_id
    WHERE 1=1
	`

	args := []interface{}{}

	// Requester hanya lihat permohonan mereka
	if role == "User" {
		query += " AND created_by = ?"
		args = append(args, userId)
	}
	// Approver 1 lihat permohonan pending approval 1
	if role == "Approver_1" {
		query += " AND status = 'PENDING_APPROVAL_1'"
	}
	// Approver 2 lihat permohonan pending approval 2
	// yang sudah diapprove oleh Approver 1
	if role == "Approver_2" {
		query += " AND status = 'PENDING_APPROVAL_2' AND approved_lvl1 IS NOT NULL"
	}
	if role == "Admin" {
	}
	if month != "" {
		query += " AND counted_month = ?"
		args = append(args, month)
	}
	if year != "" {
		query += " AND counted_year = ?"
		args = append(args, year)
	}
	if status != "" {
		query += " AND status = ?"
		args = append(args, status)
	}

	query += " ORDER BY created_at DESC LIMIT ? OFFSET ?"
	args = append(args, limit, offset)

	rows, err := config.DB.Query(query, args...)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	defer rows.Close()

	var result []models.HonorRequest

	for rows.Next() {
		var r models.HonorRequest
		var description sql.NullString
		var approvedLvl1, approvedLvl2, cancelledAt sql.NullString

		err := rows.Scan(
			&r.ID, &description, &r.CountedMonth, &r.CountedYear,
			&r.Status, &r.CreatedBy, &r.Username, &r.CreatedAt,
			&approvedLvl1, &approvedLvl2, &cancelledAt,
		)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": err.Error()})
		}

		// Convert nullable ke pointer
		if description.Valid {
			r.Description = description.String
		} else {
			r.Description = ""
		}

		if approvedLvl1.Valid {
			r.ApprovedLvl1 = &approvedLvl1.String
		} else {
			r.ApprovedLvl1 = nil
		}

		if approvedLvl2.Valid {
			r.ApprovedLvl2 = &approvedLvl2.String
		} else {
			r.ApprovedLvl2 = nil
		}

		if cancelledAt.Valid {
			r.CancelledAt = &cancelledAt.String
		} else {
			r.CancelledAt = nil
		}

		result = append(result, r)
	}
	// Ambil nama user

	return c.JSON(fiber.Map{
		"role":  role,
		"page":  page,
		"limit": limit,
		"data":  result,
	})
}

func GetHonorRequestDetail(c *fiber.Ctx) error {
	requestId := c.Params("id")

	// --- GET MAIN REQUEST DATA ---
	var request models.HonorRequest
	err := config.DB.QueryRow(`
        SELECT id, description, counted_month, counted_year, created_by, status
        FROM honor_request
        WHERE id = ?
    `, requestId).Scan(
		&request.ID,
		&request.Description,
		&request.CountedMonth,
		&request.CountedYear,
		&request.CreatedBy,
		&request.Status,
	)

	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Request not found"})
	}

	// --- GET DOCTOR LIST FROM DETAIL ---
	rows, err := config.DB.Query(`
        SELECT doctor_name, SUM(total_honor) AS total_honor
        FROM honor_request_detail
        WHERE request_id = ?
        GROUP BY doctor_name
    `, requestId)

	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	defer rows.Close()

	var doctorData []fiber.Map

	for rows.Next() {
		var doctorName string
		var totalHonor float64

		rows.Scan(&doctorName, &totalHonor)

		// GET visit + txn_code details
		visitRows, err := config.DB.Query(`
            SELECT DISTINCT visit_no
            FROM honor_request_detail
            WHERE request_id = ? AND doctor_name = ?
        `, requestId, doctorName)

		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": err.Error()})
		}

		var visitDetails []fiber.Map

		for visitRows.Next() {
			var visitNo string
			visitRows.Scan(&visitNo)

			// GET TXN DETAIL FROM patient_bill
			txnRows, _ := config.DB.Query(`
                SELECT txn_code, honor_final
                FROM patient_bill
                WHERE visit_no = ?
                  AND LOWER(txn_doctor) = LOWER(?)
            `, visitNo, doctorName)

			var items []fiber.Map
			for txnRows.Next() {
				var txnCode string
				var amount float64
				txnRows.Scan(&txnCode, &amount)

				items = append(items, fiber.Map{
					"txn_code": txnCode,
					"honor":    amount,
				})
			}
			txnRows.Close()

			visitDetails = append(visitDetails, fiber.Map{
				"visit_no": visitNo,
				"items":    items,
			})
		}
		visitRows.Close()

		doctorData = append(doctorData, fiber.Map{
			"doctor_name": doctorName,
			"total_honor": totalHonor,
			"details":     visitDetails,
		})
	}

	return c.JSON(fiber.Map{
		"request": request,
		"doctors": doctorData,
	})
}

func SubmitHonorRequest(c *fiber.Ctx) error {
	var body struct {
		CountedMonth *int                 `json:"counted_month"`
		CountedYear  *int                 `json:"counted_year"`
		Data         []models.HonorDoctor `json:"data"`
	}

	if err := c.BodyParser(&body); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid JSON body"})
	}

	uID := c.Locals("user_id").(int64)

	// Validasi agar tidak nil
	if body.CountedMonth == nil || body.CountedYear == nil {
		return c.Status(400).JSON(fiber.Map{"error": "counted_month and counted_year are required"})
	}
	month := *body.CountedMonth
	year := *body.CountedYear

	// Cek duplicate
	var existing int
	err := config.DB.QueryRow(`
        SELECT COUNT(*) 
        FROM honor_request
        WHERE counted_month = ? 
          AND counted_year = ?
          AND status IN ('PENDING_APPROVAL_1','PENDING_APPROVAL_2','APPROVED')
    `, month, year).Scan(&existing)

	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	if existing > 0 {
		return c.Status(400).JSON(fiber.Map{"error": "Duplicate request"})
	}

	if len(body.Data) == 0 {
		return c.Status(400).JSON(fiber.Map{"error": "No doctor data provided"})
	}

	tx, err := config.DB.Begin()
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	// INSERT header
	res, err := tx.Exec(`
		INSERT INTO honor_request 
		(description, counted_month, counted_year, created_by, status)
		VALUES (CONCAT('Honor Dokter MTSW Bulan ', ?, ' Tahun ', ?), ?, ?, ?, 'PENDING_APPROVAL_1')
	`, month, year, month, year, uID)

	if err != nil {
		tx.Rollback()
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	requestID, err := res.LastInsertId()
	if err != nil {
		tx.Rollback()
		return c.Status(500).JSON(fiber.Map{"error": "last insert id: " + err.Error()})
	}

	// Prepare SELECT
	rows, err := tx.Query(`
        SELECT txn_doctor,visit_no_fix 
        FROM patient_bill
        WHERE  honor_status = 'COUNTED'
          AND MONTH(counted_date) = ?
          AND YEAR(counted_date) = ?
    `, month, year)

	if err != nil {
		tx.Rollback()
		return c.Status(500).JSON(fiber.Map{"error": "Prepare SELECT gagal: " + err.Error()})
	}
	defer rows.Close()
	// build map doctor(lower) -> []visit_no_fix
	doctorVisits := make(map[string][]string)
	for rows.Next() {
		var doc sql.NullString
		var visit sql.NullString
		if err := rows.Scan(&doc, &visit); err != nil {

			rows.Close()
			tx.Rollback()
			return c.Status(500).JSON(fiber.Map{"error": "scan patient_bill: " + err.Error()})
		}
		if !doc.Valid || !visit.Valid {
			continue
		}
		key := strings.ToLower(strings.TrimSpace(doc.String))
		doctorVisits[key] = append(doctorVisits[key], visit.String)
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		tx.Rollback()
		return c.Status(500).JSON(fiber.Map{"error": "rows error: " + err.Error()})
	}

	// prepare insert detail
	stmtInsert, err := tx.Prepare(`
        INSERT INTO honor_request_detail
        (request_id, doctor_name, doctor_id, visit_no, total_honor)
        VALUES (?, ?, ?, ?, ?)
    `)
	if err != nil {
		tx.Rollback()
		return c.Status(500).JSON(fiber.Map{"error": "prepare insert detail: " + err.Error()})
	}
	defer stmtInsert.Close()

	// loop dokter dari payload, cari visit list di map, insert per visit
	totalInserted := 0
	for _, d := range body.Data {
		key := strings.ToLower(strings.TrimSpace(d.DoctorName))
		visits := doctorVisits[key]
		if len(visits) == 0 {
			// untuk debugging: tidak ada visit untuk nama dokter ini
			log.Println("SubmitHonorRequest: no visits found for doctor:", d.DoctorName)
			continue
		}

		var doctorID interface{}
		if d.CareproviderTxnDoctorId != 0 {
			doctorID = d.CareproviderTxnDoctorId
		} else {
			doctorID = nil
		}

		// Jika total honor pada payload adalah total per dokter (bukan per visit)
		for _, v := range visits {
			if _, err := stmtInsert.Exec(requestID, d.DoctorName, doctorID, v, d.TotalHonor); err != nil {
				tx.Rollback()
				return c.Status(500).JSON(fiber.Map{"error": "insert detail failed: " + err.Error()})
			}
			totalInserted++
		}
	}

	// Update patient_bill
	_, err = tx.Exec(`
        UPDATE patient_bill pb
        JOIN honor_request_detail hrd ON pb.visit_no_fix = hrd.visit_no
        SET pb.honor_status = 'ON_PROGRESS'
        WHERE hrd.request_id = ?
    `, requestID)

	if err != nil {
		tx.Rollback()
		return c.Status(500).JSON(fiber.Map{"error": "Update bill gagal: " + err.Error()})
	}

	if err := tx.Commit(); err != nil {
		tx.Rollback()
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	// Ambil userID dari JWT
	userIDValue := c.Locals("user_id")
	UserID, ok := userIDValue.(int64)
	if !ok {
		return c.Status(401).JSON(fiber.Map{"error": "Unauthorized"})
	}
	// ✅ Log aktivitas
	if err := helpers.LogActivity(c, UserID, "Input Permohonan", "Permohonan Honor Bulan "); err != nil {
		log.Println("❌ Gagal simpan log activity:", err)
	} else {
		log.Println("✅ Activity logged untuk user", UserID)
	}

	return c.JSON(fiber.Map{
		"message":        "Permohonan honor berhasil dikirim",
		"request_id":     requestID,
		"total_dokter":   len(body.Data),
		"total_inserted": totalInserted,
	})
}

func ApproveLevel1(c *fiber.Ctx) error {
	id := c.Params("id")
	log.Printf("Honor req Id: %s", id)

	tx, err := config.DB.Begin()
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	res, err := tx.Exec(`
        UPDATE honor_request 
        SET status = 'PENDING_APPROVAL_2', approved_lvl1 = NOW() 
        WHERE id = ? AND status = 'PENDING_APPROVAL_1'
    `, id)
	if err != nil {
		tx.Rollback()
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	rows, _ := res.RowsAffected()
	if rows == 0 {
		tx.Rollback()
		return c.Status(400).JSON(fiber.Map{
			"error": "Data tidak ditemukan atau status tidak valid",
		})
	}

	// 🔥 WAJIB agar perubahan masuk DB
	if err := tx.Commit(); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"message": "Approved by Approver 1"})
}

func ApproveLevel2(c *fiber.Ctx) error {
	id := c.Params("id")
	log.Printf("Honor req Id: %s", id)

	tx, err := config.DB.Begin()
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	_, err = tx.Exec(`
        UPDATE honor_request 
		SET status = 'APPROVED', approved_lvl2 =NOW() 
        WHERE id = ? AND status = 'PENDING_APPROVAL_2'
    `, id)
	if err != nil {
		tx.Rollback()
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	// Update patient_bill
	_, err = tx.Exec(`
        UPDATE patient_bill pb
        JOIN honor_request_detail hrd ON pb.visit_no_fix = hrd.visit_no
        SET pb.honor_status = 'APPROVED'
        WHERE hrd.request_id = ?
    `, id)

	if err != nil {
		tx.Rollback()
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	// 🔥 WAJIB agar perubahan masuk DB
	if err := tx.Commit(); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"message": "Approved by Approver 1"})
}

func CancelHonorRequest(c *fiber.Ctx) error {
	requestId := c.Params("id")

	tx, err := config.DB.Begin()
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "failed to start transaction"})
	}

	// 1. Ambil semua visit_no dalam request detail
	rows, err := tx.Query(`
        SELECT visit_no 
        FROM honor_request_detail 
        WHERE request_id = ?`,
		requestId)
	if err != nil {
		tx.Rollback()
		return c.Status(500).JSON(fiber.Map{"error": "failed to fetch request details"})
	}
	defer rows.Close()

	var visitNos []string
	for rows.Next() {
		var visit string
		rows.Scan(&visit)
		visitNos = append(visitNos, visit)
	}

	if len(visitNos) == 0 {
		tx.Rollback()
		return c.Status(400).JSON(fiber.Map{"error": "no detail found for this request"})
	}

	// Convert visit_no array to args
	args := make([]interface{}, len(visitNos))
	for i, v := range visitNos {
		args[i] = v
	}

	// 2. Update patient_bill → COUNTED kembali
	query := `
        UPDATE patient_bill 
        SET honor_status = 'COUNTED' 
        WHERE visit_no IN (` + strings.Repeat("?,", len(visitNos)-1) + `?) 
          AND honor_status = 'ON_PROGRESS'
    `

	_, err = tx.Exec(query, args...)
	if err != nil {
		tx.Rollback()
		return c.Status(500).JSON(fiber.Map{"error": "failed to update patient bill"})
	}

	// 3. Update honor_request → CANCELLED
	_, err = tx.Exec(`
        UPDATE honor_request 
        SET status = 'CANCELLED', cancelled_at = NOW()
        WHERE id = ?`, requestId)
	if err != nil {
		tx.Rollback()
		return c.Status(500).JSON(fiber.Map{"error": "failed to update honor request"})
	}

	// Commit transaction
	err = tx.Commit()
	if err != nil {
		tx.Rollback()
		return c.Status(500).JSON(fiber.Map{"error": "failed to commit transaction"})
	}

	return c.JSON(fiber.Map{
		"message": "Request cancelled successfully",
	})
}

func RejectHonorRequest(c *fiber.Ctx) error {
	requestId := c.Params("id")

	tx, err := config.DB.Begin()
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "failed to start transaction"})
	}

	// 1. Ambil semua visit_no dalam request detail
	rows, err := tx.Query(`
        SELECT visit_no 
        FROM honor_request_detail 
        WHERE request_id = ?`,
		requestId)
	if err != nil {
		tx.Rollback()
		return c.Status(500).JSON(fiber.Map{"error": "failed to fetch request details"})
	}
	defer rows.Close()

	var visitNos []string
	for rows.Next() {
		var visit string
		rows.Scan(&visit)
		visitNos = append(visitNos, visit)
	}

	if len(visitNos) == 0 {
		tx.Rollback()
		return c.Status(400).JSON(fiber.Map{"error": "no detail found for this request"})
	}

	// Convert visit_no array to args
	args := make([]interface{}, len(visitNos))
	for i, v := range visitNos {
		args[i] = v
	}

	// 2. Update patient_bill → COUNTED kembali
	query := `
        UPDATE patient_bill 
        SET honor_status = 'COUNTED' 
        WHERE visit_no IN (` + strings.Repeat("?,", len(visitNos)-1) + `?) 
          AND honor_status = 'ON_PROGRESS'
    `

	_, err = tx.Exec(query, args...)
	if err != nil {
		tx.Rollback()
		return c.Status(500).JSON(fiber.Map{"error": "failed to update patient bill"})
	}

	// 3. Update honor_request → REJECTED
	_, err = tx.Exec(`
        UPDATE honor_request 
        SET status = 'REJECTED', rejected_at = NOW()
        WHERE id = ? AND status IN ('PENDING_APPROVAL_1', 'PENDING_APPROVAL_2')`, requestId)
	if err != nil {
		tx.Rollback()
		return c.Status(500).JSON(fiber.Map{"error": "failed to update honor request"})
	}

	// Commit transaction
	err = tx.Commit()
	if err != nil {
		tx.Rollback()
		return c.Status(500).JSON(fiber.Map{"error": "failed to commit transaction"})
	}

	return c.JSON(fiber.Map{
		"message": "Request rejected successfully",
	})
}

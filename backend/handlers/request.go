package handlers

import (
	"database/sql"
	"fmt"
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
           h.created_by, u.username,
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
	// Approver 3 lihat permohonan pending approval 3
	// yang sudah diapprove oleh Approver 1
	if role == "Approver_3" {
		query += " AND status = 'PENDING_APPROVAL_3' AND approved_lvl2 IS NOT NULL"
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

	query += " ORDER BY h.created_at DESC LIMIT ? OFFSET ?"
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
			&r.Status, &r.CreatedBy, &r.Username,
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
	role := c.Locals("role")
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

		// === 1. GET VISIT DETAIL (BILL) ===
		visitRows, err := config.DB.Query(`
    		SELECT DISTINCT hrdt.visit_no
   			FROM honor_request_detail_doctor hrdt
    		JOIN honor_request_detail hrd ON hrdt.honor_request_detail_id = hrd.id
    		WHERE hrd.request_id = ? AND hrd.doctor_name = ?
    		AND hrdt.type = 'BILL'
		`, requestId, doctorName)

		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": err.Error()})
		}

		var visitDetails []fiber.Map

		for visitRows.Next() {
			var visitNo string
			visitRows.Scan(&visitNo)

			// GET BILL ITEMS
			txnRows, _ := config.DB.Query(`
        		SELECT patient_type,
               DATE_FORMAT(admission_date_time,'%d %b %Y'),
               DATE_FORMAT(discharge_date_time,'%d %b %Y'),
               card_no,
               patient_name,
               organisation_name,
               txn_desc,
               honor_final
        	FROM patient_bill
       		WHERE visit_no = ?
          	AND LOWER(txn_doctor) = LOWER(?)
          	AND honor_final != 0
    		`, visitNo, doctorName)

			var items []fiber.Map

			for txnRows.Next() {
				var patientType, inDate, outDate, nama, company, desc string
				var nrm int64
				var amount float64

				txnRows.Scan(&patientType, &inDate, &outDate, &nrm, &nama, &company, &desc, &amount)

				items = append(items, fiber.Map{
					"patient_type": patientType,
					"masuk":        inDate,
					"keluar":       outDate,
					"nrm":          nrm,
					"pasien":       nama,
					"company":      company,
					"txn_desc":     desc,
					"honor":        amount,
				})
			}
			txnRows.Close()

			visitDetails = append(visitDetails, fiber.Map{
				"visit_no": visitNo,
				"items":    items,
			})
		}
		visitRows.Close()

		// === 2. GET ADJUSTMENT (TANPA VISIT NO) ===
		adjRows, _ := config.DB.Query(`
    			SELECT hrdt.description, hrdt.amount
    			FROM honor_request_detail_doctor hrdt
    			JOIN honor_request_detail hrd ON hrdt.honor_request_detail_id = hrd.id
    			WHERE hrd.request_id = ? AND hrd.doctor_name = ? AND hrdt.type = 'ADJUSTMENT'
			`, requestId, doctorName)

		var adjItems []fiber.Map

		for adjRows.Next() {
			var desc string
			var amount float64

			adjRows.Scan(&desc, &amount)

			adjItems = append(adjItems, fiber.Map{
				"patient_type": nil,
				"masuk":        nil,
				"keluar":       nil,
				"nrm":          nil,
				"pasien":       nil,
				"company":      nil,
				"txn_desc":     desc, // DESKRIPSI ADJUSTMENT
				"honor":        amount,
			})
		}
		adjRows.Close()

		// Jika ada adjustment → masukkan sebagai visit terpisah dengan visit_no = null
		if len(adjItems) > 0 {
			visitDetails = append(visitDetails, fiber.Map{
				"visit_no": nil,
				"items":    adjItems,
			})
		}
		// === FIX PENTING ===
		// Masukkan seluruh visitDetails ke doctorData
		doctorData = append(doctorData, fiber.Map{
			"doctor_name": doctorName,
			"total_honor": totalHonor,
			"details":     visitDetails,
		})
	}
	return c.JSON(fiber.Map{
		"role":    role,
		"request": request,
		"doctors": doctorData,
	})
}

// Submit Request Sebelumnya
func SubmitHonorRequestLate(c *fiber.Ctx) error {
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
        SELECT txn_doctor,visit_no,visit_no_fix, honor_final, txn_desc,careprovider_txn_doctor_id
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
	doctorVisits := make(map[string][]struct {
		VisitNo                 string
		VisitNoFix              string
		Honor                   float64
		Desc                    string
		CareproviderTxnDoctorId int64
	})
	for rows.Next() {
		var doc, visit, visitfix, desc string
		var honor float64
		var careproviderid int64
		if err := rows.Scan(&doc, &visit, &visitfix, &honor, &desc, &careproviderid); err != nil {
			tx.Rollback()
			return c.Status(500).JSON(fiber.Map{"error": "scan patient_bill: " + err.Error()})
		}

		key := strings.ToLower(strings.TrimSpace(doc))
		doctorVisits[key] = append(doctorVisits[key], struct {
			VisitNo                 string
			VisitNoFix              string
			Honor                   float64
			Desc                    string
			CareproviderTxnDoctorId int64
		}{visit, visitfix, honor, desc, careproviderid})
	}

	if err := rows.Err(); err != nil {
		tx.Rollback()
		return c.Status(500).JSON(fiber.Map{"error": "rows error: " + err.Error()})
	}

	// prepare insert detail
	stmtInsert, err := tx.Prepare(`
        INSERT INTO honor_request_detail
        (request_id, doctor_name, doctor_id, total_honor)
        VALUES (?, ?, ?, ?)
    `)
	if err != nil {
		tx.Rollback()
		return c.Status(500).JSON(fiber.Map{"error": "prepare insert detail: " + err.Error()})
	}
	defer stmtInsert.Close()

	// Insert Detail doctor (BILL+ADJUSTMENT)
	stmtInsertItems, err := tx.Prepare(`
    	INSERT INTO honor_request_detail_doctor
		(honor_request_detail_id, doctor_name, visit_no, description, amount, type, careprovider_txn_doctor_id)
		VALUES (?, ?, ?, ?, ?, ?, ?)
		`)
	if err != nil {
		tx.Rollback()
		return c.Status(500).JSON(fiber.Map{"error": "prepare insert detail items: " + err.Error()})
	}
	defer stmtInsertItems.Close()

	// loop dokter dari payload, cari visit list di map, insert per visit
	totalInserted := 0

	for _, d := range body.Data {
		key := strings.ToLower(strings.TrimSpace(d.DoctorName))
		bills := doctorVisits[key]

		// insert detail header
		resDetail, err := stmtInsert.Exec(requestID, d.DoctorName, d.CareproviderTxnDoctorId, d.TotalHonor)
		if err != nil {
			tx.Rollback()
			return c.Status(500).JSON(fiber.Map{"error": "insert detail gagal: " + err.Error()})
		}

		detailID, err := resDetail.LastInsertId()
		if err != nil {
			tx.Rollback()
			return c.Status(500).JSON(fiber.Map{"error": "last insert detail error: " + err.Error()})
		}

		// insert BILL
		for _, b := range bills {
			_, err := stmtInsertItems.Exec(detailID, d.DoctorName, b.VisitNo, b.Desc, b.Honor, "BILL", b.CareproviderTxnDoctorId)
			if err != nil {
				tx.Rollback()
				return c.Status(500).JSON(fiber.Map{"error": err.Error()})
			}
			totalInserted++
		}

		// insert ADJUSTMENT
		rowsAdj, err := tx.Query(`
			SELECT notes, amount
			FROM honor_adjustment
			WHERE careprovider_txn_doctor_id=? AND counted_month=? AND counted_year=?
		`, d.CareproviderTxnDoctorId, month, year)

		if err != nil {
			tx.Rollback()
			return c.Status(500).JSON(fiber.Map{
				"error": "Query adjustment gagal: " + err.Error(),
			})
		}
		defer rowsAdj.Close()

		for rowsAdj.Next() {
			var note string
			var amt float64
			if err := rowsAdj.Scan(&note, &amt); err != nil {
				tx.Rollback()
				return c.Status(500).JSON(fiber.Map{
					"error": "scan adjustment: " + err.Error(),
				})
			}

			_, err := stmtInsertItems.Exec(
				detailID,
				d.DoctorName,
				nil, // visit_no = NULL
				note,
				amt,
				"ADJUSTMENT",
				d.CareproviderTxnDoctorId,
			)
			if err != nil {
				tx.Rollback()
				return c.Status(500).JSON(fiber.Map{"error": err.Error()})
			}
			totalInserted++
		}
		if err := rowsAdj.Err(); err != nil {
			tx.Rollback()
			return c.Status(500).JSON(fiber.Map{"error": "rowsAdj error: " + err.Error()})
		}
		rowsAdj.Close()
	}

	// Update patient_bill
	_, err = tx.Exec(`
	    UPDATE patient_bill pb
	JOIN honor_request_detail_doctor hrd 
      ON pb.visit_no = hrd.visit_no
	JOIN honor_request_detail hrdt
      ON hrd.honor_request_detail_id = hrdt.id
	SET pb.honor_status = 'ON_PROGRESS'
	WHERE hrdt.request_id = ?
	`, requestID)

	if err != nil {
		tx.Rollback()
		return c.Status(500).JSON(fiber.Map{"error": "Update bill gagal: " + err.Error()})
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

	if err := tx.Commit(); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{
		"message":        "Permohonan honor berhasil dikirim",
		"request_id":     requestID,
		"total_dokter":   len(body.Data),
		"total_inserted": totalInserted,
	})
}

func SubmitHonorRequest(c *fiber.Ctx) error {
	// Parse body
	var body struct {
		CountedMonth *int                 `json:"counted_month"`
		CountedYear  *int                 `json:"counted_year"`
		Data         []models.HonorDoctor `json:"data"`
	}

	if err := c.BodyParser(&body); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid JSON body"})
	}

	uID, ok := c.Locals("user_id").(int64)
	if !ok {
		return c.Status(401).JSON(fiber.Map{"error": "Unauthorized"})
	}

	if body.CountedMonth == nil || body.CountedYear == nil {
		return c.Status(400).JSON(fiber.Map{"error": "counted_month and counted_year are required"})
	}
	month := *body.CountedMonth
	year := *body.CountedYear

	if len(body.Data) == 0 {
		return c.Status(400).JSON(fiber.Map{"error": "No doctor data provided"})
	}

	// Start transaction
	tx, err := config.DB.Begin()
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	defer func() {
		if p := recover(); p != nil {
			tx.Rollback()
			panic(p)
		}
	}()

	// Cek duplicate request
	if err := checkDuplicateRequest(tx, month, year); err != nil {
		tx.Rollback()
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	}

	// Insert header
	requestID, err := insertHonorRequestHeader(tx, uID, month, year)
	if err != nil {
		tx.Rollback()
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	// Ambil semua visit per dokter dari patient_bill
	doctorVisits, err := getDoctorVisitsMap(tx, month, year)
	if err != nil {
		tx.Rollback()
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	// Ambil semua adjustment sekaligus (1 query)
	allAdjustments, err := getAllAdjustments(tx, month, year)
	if err != nil {
		tx.Rollback()
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	// Insert detail & items
	totalInserted, err := insertHonorDetailsBatch(tx, requestID, body.Data, doctorVisits, allAdjustments)
	if err != nil {
		tx.Rollback()
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	// Update patient_bill
	if err := updatePatientBill(tx, requestID); err != nil {
		tx.Rollback()
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	// Log activity
	if err := helpers.LogActivity(c, uID, "Input Permohonan", "Permohonan Honor Bulan"); err != nil {
		log.Println("❌ Gagal simpan log activity:", err)
	} else {
		log.Println("✅ Activity logged untuk user", uID)
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{
		"message":        "Permohonan honor berhasil dikirim",
		"request_id":     requestID,
		"total_dokter":   len(body.Data),
		"total_inserted": totalInserted,
	})
}

func checkDuplicateRequest(tx *sql.Tx, month, year int) error {
	var existing int
	err := tx.QueryRow(`
        SELECT COUNT(*) 
        FROM honor_request
        WHERE counted_month = ? AND counted_year = ? 
          AND status IN ('PENDING_APPROVAL_1','PENDING_APPROVAL_2','APPROVED')
    `, month, year).Scan(&existing)
	if err != nil {
		return err
	}
	if existing > 0 {
		return fmt.Errorf("Duplicate request")
	}
	return nil
}

func insertHonorRequestHeader(tx *sql.Tx, userID int64, month, year int) (int64, error) {
	res, err := tx.Exec(`
        INSERT INTO honor_request 
        (description, counted_month, counted_year, created_by, status)
        VALUES (CONCAT('Honor Dokter MTSW Bulan ', ?, ' Tahun ', ?), ?, ?, ?, 'PENDING_APPROVAL_1')
    `, month, year, month, year, userID)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

// Ambil semua adjustment sekaligus
func getAllAdjustments(tx *sql.Tx, month, year int) (map[int64][]struct {
	Note   string
	Amount float64
}, error) {
	rows, err := tx.Query(`
        SELECT careprovider_txn_doctor_id, notes, amount
        FROM honor_adjustment
        WHERE counted_month=? AND counted_year=?
    `, month, year)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	adjustments := make(map[int64][]struct {
		Note   string
		Amount float64
	})

	for rows.Next() {
		var doctorID int64
		var note string
		var amt float64
		if err := rows.Scan(&doctorID, &note, &amt); err != nil {
			return nil, err
		}
		adjustments[doctorID] = append(adjustments[doctorID], struct {
			Note   string
			Amount float64
		}{note, amt})
	}
	return adjustments, rows.Err()
}

// Batch insert detail doctor
func insertHonorDetailsBatch(tx *sql.Tx, requestID int64, doctors []models.HonorDoctor,
	doctorVisits map[string][]struct {
		VisitNo, VisitNoFix, Desc string
		Honor                     float64
		CareproviderTxnDoctorId   int64
	},
	allAdjustments map[int64][]struct {
		Note   string
		Amount float64
	}) (int, error) {

	totalInserted := 0

	stmtDetail, err := tx.Prepare(`
        INSERT INTO honor_request_detail
        (request_id, doctor_name, doctor_id, total_honor)
        VALUES (?, ?, ?, ?)
    `)
	if err != nil {
		return 0, err
	}
	defer stmtDetail.Close()

	for _, d := range doctors {
		key := strings.ToLower(strings.TrimSpace(d.DoctorName))
		bills := doctorVisits[key]

		resDetail, err := stmtDetail.Exec(requestID, d.DoctorName, d.CareproviderTxnDoctorId, d.TotalHonor)
		if err != nil {
			return totalInserted, err
		}
		detailID, _ := resDetail.LastInsertId()

		// Prepare batch insert for bills + adjustments
		var args []interface{}
		for _, b := range bills {
			args = append(args, detailID, d.DoctorName, b.VisitNo, b.Desc, b.Honor, "BILL", b.CareproviderTxnDoctorId)
			totalInserted++
		}

		for _, adj := range allAdjustments[d.CareproviderTxnDoctorId] {
			args = append(args, detailID, d.DoctorName, nil, adj.Note, adj.Amount, "ADJUSTMENT", d.CareproviderTxnDoctorId)
			totalInserted++
		}

		// Execute batch insert per 50 row
		for i := 0; i < len(args); i += 7 * 50 {
			end := i + 7*50
			if end > len(args) {
				end = len(args)
			}
			batchArgs := args[i:end]
			vals := make([]string, len(batchArgs)/7)
			for j := range vals {
				vals[j] = "(?, ?, ?, ?, ?, ?, ?)"
			}
			query := "INSERT INTO honor_request_detail_doctor (honor_request_detail_id, doctor_name, visit_no, description, amount, type, careprovider_txn_doctor_id) VALUES " + strings.Join(vals, ",")
			if _, err := tx.Exec(query, batchArgs...); err != nil {
				return totalInserted, err
			}
		}
	}

	return totalInserted, nil
}

func getDoctorVisitsMap(tx *sql.Tx, month, year int) (map[string][]struct {
	VisitNo, VisitNoFix, Desc string
	Honor                     float64
	CareproviderTxnDoctorId   int64
}, error) {
	rows, err := tx.Query(`
        SELECT txn_doctor, visit_no, visit_no_fix, honor_final, txn_desc, careprovider_txn_doctor_id
        FROM patient_bill
        WHERE honor_status = 'COUNTED' AND MONTH(counted_date) = ? AND YEAR(counted_date) = ?
    `, month, year)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	doctorVisits := make(map[string][]struct {
		VisitNo, VisitNoFix, Desc string
		Honor                     float64
		CareproviderTxnDoctorId   int64
	})

	for rows.Next() {
		var doc, visit, visitfix, desc string
		var honor float64
		var careproviderid int64
		if err := rows.Scan(&doc, &visit, &visitfix, &honor, &desc, &careproviderid); err != nil {
			return nil, err
		}
		key := strings.ToLower(strings.TrimSpace(doc))
		doctorVisits[key] = append(doctorVisits[key], struct {
			VisitNo, VisitNoFix, Desc string
			Honor                     float64
			CareproviderTxnDoctorId   int64
		}{visit, visitfix, desc, honor, careproviderid})
	}

	return doctorVisits, rows.Err()
}

func updatePatientBill(tx *sql.Tx, requestID int64) error {
	_, err := tx.Exec(`
        UPDATE patient_bill pb
        JOIN honor_request_detail_doctor hrd ON pb.visit_no = hrd.visit_no
        JOIN honor_request_detail hrdt ON hrd.honor_request_detail_id = hrdt.id
        SET pb.honor_status = 'ON_PROGRESS'
        WHERE hrdt.request_id = ?
    `, requestID)
	return err
}

func ApproveLevel1(c *fiber.Ctx) error {
	id := c.Params("id")
	log.Printf("Honor req Id: %v", id)

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

	res, err := tx.Exec(`
        UPDATE honor_request 
        SET status = 'PENDING_APPROVAL_3', approved_lvl2 = NOW() 
        WHERE id = ? AND status = 'PENDING_APPROVAL_2'
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

func ApproveLevel3(c *fiber.Ctx) error {
	id := c.Params("id")
	log.Printf("Honor req Id: %s", id)

	tx, err := config.DB.Begin()
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	_, err = tx.Exec(`
        UPDATE honor_request 
		SET status = 'APPROVED', approved_lvl3 =NOW() 
        WHERE id = ? AND status = 'PENDING_APPROVAL_3'
    `, id)
	if err != nil {
		tx.Rollback()
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	// Update patient_bill
	_, err = tx.Exec(`
         UPDATE patient_bill pb
	JOIN honor_request_detail_doctor hrd 
      ON pb.visit_no_fix = hrd.visit_no
	JOIN honor_request_detail hrdt
      ON hrd.honor_request_detail_id = hrdt.id
	SET pb.honor_status = 'APPROVED'
	WHERE hrdt.request_id = ?
    `, id)

	if err != nil {
		tx.Rollback()
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	// 🔥 WAJIB agar perubahan masuk DB
	if err := tx.Commit(); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"message": "Approved by Approver 3"})
}

func CancelHonorRequest(c *fiber.Ctx) error {
	requestId := c.Params("id")

	tx, err := config.DB.Begin()
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "failed to start transaction"})
	}

	// 1. Ambil semua visit_no dalam request detail doctor
	rows, err := tx.Query(`
        SELECT hrd.visit_no 
        FROM honor_request_detail_doctor hrd
		JOIN honor_request_detail hrdt
      	 	ON hrd.honor_request_detail_id = hrdt.id
        WHERE hrdt.request_id = ?`,
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
        SELECT hrd.visit_no 
        FROM honor_request_detail_doctor hrd
		JOIN honor_request_detail hrdt
      	 	ON hrd.honor_request_detail_id = hrdt.id
        WHERE hrdt.request_id = ?`,
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

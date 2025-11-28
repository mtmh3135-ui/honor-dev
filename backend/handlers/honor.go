package handlers

import (
	"database/sql"
	"fmt"
	"log"
	"math"
	"sort"
	"strconv"
	"strings"

	"github.com/Knetic/govaluate"
	"github.com/gofiber/fiber/v2"
	"github.com/mtmh3135/honor/backend/models"
)

// ✅ Struct data utama
type Honorfull struct {
	ID              int
	VisitNo         string
	VisitNoFix      string
	TxnCode         string
	TxnCategory     string
	TxnDesc         string
	PatientType     string
	PatientClass    string // "bpjs" atau "general"
	TxnType         string // "tindakan", "visit", "fix"
	Qty             float64
	NetAmount       float64
	Inacbg          float64
	TxnDoctor       string
	Bpjs_ip         string
	Bpjs_op         string
	RumusGeneral    string
	HonorMaster     float64
	HonorProp       float64
	HonorFinal      float64
	HonorStatus     string
	BPJSClass       string
	Status          string
	Description     string
	PreviousTxnType string
}

type DoctorSummary struct {
	DoctorID   int
	DoctorName string
	TotalHonor float64
}

func GetHonor(db *sql.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		PatientName := c.Query("patient_name")
		VisitNo := c.Query("visit_no")
		PatientClass := c.Query("patient_class")
		all := c.Query("all", "false")

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

		query := `SELECT t.id, t.patient_name, t.card_no, t.visit_no, t.regn_dept, t.ward_desc, t.patient_class, t.txn_category, t.txn_code, t.gl_account,
		t.txn_desc, t.careprovider_txn_doctor_id, t.txn_doctor, t.regn_doctor, t.ref_doctor,IFNULL(t.base_price,0) AS base_price, t.qty, IFNULL(t.txn_amount,0) AS txn_amount,
		IFNULL(t.margin_amount,0) AS margin_amount, IFNULL(t.claim_amount,0) AS claim_amount, IFNULL(t.discount_visit,0) AS discount_visit,
		IFNULL(t.honor_master,0) AS honor_master,IFNULL(t.honor_prop,0) AS honor_prop,IFNULL(t.honor_final,0) AS honor_final,IFNULL(c.tarif_ina_cbg,0) AS tarif_ina_cbg,    
		t.net_amount, 
		
		CASE
    	-- 1. Jika kelas BPJS
    		WHEN t.patient_class = 'BPJS' THEN 
        	CASE
				WHEN c.status!= 'OFFER (PENDING)' THEN 'LUNAS'
				ELSE 'BELUM LUNAS'
			END
    	-- 2. Jika Insurance atau Corporate
    		WHEN t.patient_class IN ('INSURANCE', 'CORPORATE') THEN
        	CASE
            	WHEN p.sisa = 0 THEN 'LUNAS'
            	ELSE 'BELUM LUNAS'
        	END
   		 -- 3. Jika General atau Hospital Staff
    		WHEN t.patient_class IN ('GENERAL', 'HOSPITAL STAFF') THEN
        	CASE
            	WHEN t.bill_status = 'PAID' THEN 'LUNAS'
            	ELSE 'BELUM LUNAS'
        	END
    	ELSE 'BELUM LUNAS'
		END AS status,

		t.bill_datetime, t.bill_status, t.organisation_name, t.admission_date_time, IFNULL(t.discharge_date_time,'-') AS discharge_date_time,
		t.patient_type
		FROM patient_bill t
		LEFT JOIN piutang p ON t.visit_no_fix = p.visit_no
		LEFT JOIN comparison_data c ON t.visit_no_fix = c.visit_number		
		WHERE 1=1`
		args := []interface{}{}

		// hanya tambahkan filter jika user mengisi
		if PatientName != "" {
			query += " AND t.patient_name LIKE ?"
			args = append(args, "%"+PatientName+"%")
		}

		if VisitNo != "" {
			query += " AND t.visit_no LIKE ?"
			args = append(args, "%"+VisitNo+"%")
		}
		if PatientClass != "" {
			query += " AND t.patient_class = ?"
			args = append(args, strings.ToUpper(PatientClass))
		}
		// 🟢 Jika export semua data (tanpa LIMIT)
		if all != "true" {
			query += " LIMIT ? OFFSET ?"
			args = append(args, limit, offset)
		}

		rows, err := db.Query(query, args...)
		if err != nil {
			log.Printf("error:%v", err)
			return c.Status(500).JSON(fiber.Map{"error": err.Error()})
		}

		defer rows.Close()

		var data []models.Honor
		for rows.Next() {
			var p models.Honor
			if err := rows.Scan(&p.ID,
				&p.PatientName, &p.CardNo, &p.VisitNo, &p.RegnDept, &p.WardDesc, &p.PatientClass, &p.TxnCategory, &p.TxnCode, &p.GlAccount, &p.TxnDesc, &p.CareproviderTxnDoctorId, &p.TxnDoctor, &p.RegnDoctor,
				&p.RefDoctor, &p.BasePrice, &p.Qty, &p.TxnAmount, &p.MarginAmount, &p.ClaimAmount, &p.DiscountVisit, &p.HonorMaster, &p.HonorProp, &p.HonorFinal, &p.TarifINACBG, &p.NetAmount,
				&p.Status, &p.BillDateTime, &p.BillStatus, &p.OrganisationName, &p.AdmissionDateTime, &p.DischargeDateTime, &p.PatientType,
			); err != nil {
				log.Printf("error:%v", err)
				return c.Status(500).JSON(fiber.Map{"error": err.Error()})
			}
			data = append(data, p)
		}
		// jika all=true, tidak perlu pagination info
		if all == "true" {
			log.Printf("📦 Exporting %d data (all=true)", len(data))
			return c.JSON(fiber.Map{
				"data": data,
			})
		}
		// ambil total data (tanpa limit)
		countQuery := `SELECT COUNT(*) FROM patient_bill t WHERE 1=1`
		countArgs := []interface{}{}
		if PatientName != "" {
			countQuery += " AND t.patient_name LIKE ?"
			countArgs = append(countArgs, "%"+PatientName+"%")
		}

		if VisitNo != "" {
			countQuery += " AND t.visit_no LIKE ?"
			countArgs = append(countArgs, "%"+VisitNo+"%")
		}
		if PatientClass != "" {
			countQuery += " AND t.patient_class = ?"
			countArgs = append(countArgs, strings.ToUpper(PatientClass))
		}

		var total int
		if err := db.QueryRow(countQuery, countArgs...).Scan(&total); err != nil {
			log.Printf("error:%v", err)
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

func GetDoctorHonor(db *sql.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		DoctorName := c.Query("txn_doctor")
		Month := c.Query("month")
		Year := c.Query("year")

		query := `
			SELECT 
				t.txn_doctor, 
				t.honor_final,
				d.careprovider_txn_doctor_id
			FROM patient_bill t
			LEFT JOIN doctor_data d ON t.txn_doctor = d.doctor_name
			WHERE honor_status = 'COUNTED'
			AND txn_doctor IN (SELECT DISTINCT doctor_name FROM doctor_data)
		`

		args := []interface{}{}

		if DoctorName != "" {
			query += " AND txn_doctor LIKE ?"
			args = append(args, "%"+DoctorName+"%")
		}
		if Month != "" {
			query += " AND MONTH(counted_date) = ?"
			args = append(args, Month)
		}
		if Year != "" {
			query += " AND YEAR(counted_date) = ?"
			args = append(args, Year)
		}

		rows, err := db.Query(query, args...)
		if err != nil {
			log.Printf("query error: %v", err)
			return c.Status(500).JSON(fiber.Map{"error": err.Error()})
		}
		defer rows.Close()

		var data []models.HonorDoctor
		for rows.Next() {
			var r models.HonorDoctor
			if err := rows.Scan(&r.DoctorName, &r.HonorFinal, &r.CareproviderTxnDoctorId); err != nil {
				log.Printf("scan error: %v", err)
				return c.Status(500).JSON(fiber.Map{"error": err.Error()})
			}
			data = append(data, r)
		}

		// ✅ Grouping per dokter
		grouped := make(map[string]*models.HonorDoctor)
		for _, d := range data {
			if _, exists := grouped[d.DoctorName]; !exists {
				grouped[d.DoctorName] = &models.HonorDoctor{
					CareproviderTxnDoctorId: d.CareproviderTxnDoctorId,
					DoctorName:              d.DoctorName,
					TotalHonor:              0,
				}
			}
			grouped[d.DoctorName].TotalHonor += d.HonorFinal
		}

		// ✅ Konversi hasilnya ke slice agar bisa dikirim ke frontend
		var result []models.HonorDoctor
		for _, g := range grouped {
			result = append(result, models.HonorDoctor{
				DoctorName:              g.DoctorName,
				TotalHonor:              g.TotalHonor,
				CareproviderTxnDoctorId: g.CareproviderTxnDoctorId,
			})
		}

		// Optional: urutkan dari honor terbesar ke terkecil
		sort.Slice(result, func(i, j int) bool {
			return result[i].TotalHonor > result[j].TotalHonor
		})
		log.Printf("Total Dokter: %d", len(result))

		return c.JSON(fiber.Map{

			"total_dokter": len(result),
			"data":         result,
		})
	}
}

func GetDoctorHonorMonthly(db *sql.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		doctorName := c.Query("txn_doctor")
		month := c.Query("month")
		year := c.Query("year")

		if year == "" {
			return c.Status(400).JSON(fiber.Map{"error": "year is required"})
		}

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
                hrd.doctor_name,
                SUM(hrd.total_honor) AS total_honor				
            FROM honor_request_detail hrd
            JOIN honor_request hr ON hrd.request_id = hr.id
            WHERE YEAR(hr.approved_lvl2) = ?
        `

		args := []interface{}{year}

		// Jika bulan dipilih
		if month != "" {
			query += " AND MONTH(approved_lvl2) = ?"
			args = append(args, month)
		}

		// Jika nama dokter dipilih
		if doctorName != "" {
			query += " AND doctor_name LIKE ?"
			args = append(args, "%"+doctorName+"%")
		}

		query += `
            AND hr.status = 'APPROVED'			
            GROUP BY hrd.doctor_name
            ORDER BY total_honor DESC
        `
		query += " LIMIT ? OFFSET ?"
		args = append(args, limit, offset)

		rows, err := db.Query(query, args...)
		if err != nil {
			log.Println(err)
			return c.Status(500).JSON(fiber.Map{"error": "query error"})
		}
		defer rows.Close()

		var result []fiber.Map

		for rows.Next() {
			var doctor string
			var total float64
			rows.Scan(&doctor, &total)

			result = append(result, fiber.Map{
				"DoctorName": doctor,
				"TotalHonor": total,
			})
		}

		// ambil total data (tanpa limit)
		countQuery := `SELECT COUNT(*) FROM honor_request_detail hrd JOIN honor_request hr ON hrd.request_id = hr.id WHERE 1=1`
		countArgs := []interface{}{}
		// Jika bulan dipilih
		if month != "" {
			countQuery += " AND MONTH(approved_lvl2) = ?"
			countArgs = append(countArgs, month)
		}

		// Jika nama dokter dipilih
		if doctorName != "" {
			countQuery += " AND doctor_name LIKE ?"
			countArgs = append(countArgs, "%"+doctorName+"%")
		}

		var total int
		if err := db.QueryRow(countQuery, countArgs...).Scan(&total); err != nil {
			return c.Status(500).JSON(fiber.Map{"error": err.Error()})
		}
		return c.JSON(fiber.Map{
			"page":       page,
			"limit":      limit,
			"total":      total,
			"totalPages": (total + limit - 1) / limit,
			"data":       result,
		})
	}
}

// ✅ Fungsi utama pemrosesan honor
func HonorCount(db *sql.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// 1️⃣ Ambil data gabungan txn + master_txn
		rows, err := db.Query(`
		SELECT
  				t.id,		
				t.visit_no,		
  				t.visit_no_fix,
  				t.txn_code,
				t.txn_category,
				t.txn_desc,
  				t.patient_type,
  				t.patient_class,
  				t.qty,
  				t.net_amount,
  				t.txn_doctor,
  				m.txn_type,
  				m.bpjs_ip,
  				m.bpjs_op,
  				m.rumus_general,
  				IFNULL(e.status, '-') AS status,
  				IFNULL(c.tarif_ina_cbg, 0) AS tarif_ina_cbg,
				IFNULL(c.kelas_bpjs,'-') AS kelas_bpjs,
				IFNULL(d.description,'-') as description
				FROM patient_bill t
				JOIN master_txn m ON t.txn_code = m.txn_code
				LEFT JOIN comparison_data c ON t.visit_no_fix = c.visit_number
				LEFT JOIN comparison_data e ON t.visit_no = e.visit_number
				LEFT JOIN doctor_data d on t.txn_doctor = d.doctor_name
		WHERE
		(
  			-- 🔹 CASE 1: Pasien BPJS
  			(
    			(t.honor_status IS NULL OR t.honor_status NOT IN ('FINISH', 'ON PROGRESS'))
    			AND t.patient_class = 'BPJS'
    			AND DATE_FORMAT(t.bill_datetime, '%Y-%m') < DATE_FORMAT(DATE_SUB(NOW(), INTERVAL 2 MONTH), '%Y-%m')
  			)
  
  			-- 🔹 CASE 2: Pasien NON BPJS
  			OR
  			(
    			(t.honor_status IS NULL OR t.honor_status NOT IN ('FINISH', 'ON PROGRESS'))
    			AND t.patient_class <> 'BPJS'
    			AND DATE_FORMAT(t.bill_datetime, '%Y-%m') < DATE_FORMAT(DATE_SUB(NOW(), INTERVAL 1 MONTH), '%Y-%m')
  			)
		);
	`)
		if err != nil {
			log.Printf("gagal query data:%v", err)
			return fmt.Errorf("gagal query data: %v", err)
		}
		defer rows.Close()

		var allHonor []Honorfull
		for rows.Next() {
			var t Honorfull
			if err := rows.Scan(
				&t.ID, &t.VisitNo, &t.VisitNoFix, &t.TxnCode, &t.TxnCategory, &t.TxnDesc, &t.PatientType, &t.PatientClass, &t.Qty,
				&t.NetAmount, &t.TxnDoctor, &t.TxnType, &t.Bpjs_ip, &t.Bpjs_op, &t.RumusGeneral,
				&t.Status, &t.Inacbg, &t.BPJSClass, &t.Description,
			); err != nil {
				log.Printf("gagal scan:%v", err)
				return fmt.Errorf("gagal scan: %v", err)
			}
			t.PreviousTxnType = t.TxnType
			allHonor = append(allHonor, t)
		}

		// 2️⃣ Hitung honor master (dari rumus)
		for i := range allHonor {

			rumus := allHonor[i].RumusGeneral
			if allHonor[i].PatientClass == "BPJS" && allHonor[i].PatientType == "INPATIENTS" {
				rumus = allHonor[i].Bpjs_ip
			} else if allHonor[i].PatientClass == "BPJS" && allHonor[i].PatientType == "OUTPATIENTS" {
				rumus = allHonor[i].Bpjs_op
			}
			honor, err := evalRumus(rumus, allHonor[i])
			if err != nil {
				log.Printf("⚠️ Gagal evaluasi %s %s: %v", allHonor[i].VisitNoFix, allHonor[i].TxnCode, err)
				continue
			}

			// Set default dulu
			allHonor[i].HonorMaster = honor

			//  Kondisi jika ada transaksi yang tidak berisikan class BPJS
			if allHonor[i].TxnDesc == "HONOR DOKTER" {
				switch allHonor[i].BPJSClass {
				case "I":
					allHonor[i].HonorMaster = 75000 * allHonor[i].Qty
				case "II":
					allHonor[i].HonorMaster = 60000 * allHonor[i].Qty
				case "III":
					allHonor[i].HonorMaster = 50000 * allHonor[i].Qty
				}

			}
			if allHonor[i].TxnDesc == "KONSULTASI DOKTER" && allHonor[i].PatientType == "INPATIENT" {
				switch allHonor[i].BPJSClass {
				case "I":
					allHonor[i].HonorMaster = 75000 * allHonor[i].Qty
				case "II":
					allHonor[i].HonorMaster = 60000 * allHonor[i].Qty
				case "III":
					allHonor[i].HonorMaster = 50000 * allHonor[i].Qty
				}

			}

			//  Kondisi khusus dr. MUTIARA MARGARETHA
			if allHonor[i].TxnDoctor == "dr. MUTIARA MARGARETHA, SpJP" {
				// log.Println("Ada spesialis jantung")
				switch allHonor[i].TxnDesc {
				case "HONOR DOKTER SPESIALIS VISITE (CLASS I)":
					allHonor[i].HonorMaster = 100000 * allHonor[i].Qty
					// log.Printf(" Dokter %s (Spesialis JANTUNG) - honor disesuaikan menjadi %v pada visit number %s", allHonor[i].TxnDoctor, allHonor[i].HonorMaster, allHonor[i].VisitNo)
				case "HONOR DOKTER SPESIALIS VISITE (CLASS II)":
					allHonor[i].HonorMaster = 75000 * allHonor[i].Qty
					// log.Printf(" Dokter %s (Spesialis JANTUNG) - honor disesuaikan pada visit number %s", allHonor[i].TxnDoctor, allHonor[i].VisitNo)
				}
				// log.Printf(" Dokter %s (Spesialis JANTUNG) - honor disesuaikan pada visit number %s", allHonor[i].TxnDoctor, allHonor[i].VisitNo)
			}

			//  Validasi dinamis jika rumus mengandung fungsi min()
			if strings.Contains(strings.ToLower(rumus), "min(") {
				isNetChosen, err := evalMinChoice(rumus, allHonor[i])
				if err != nil {
					log.Printf("⚠️ Gagal deteksi min() %s %s: %v", allHonor[i].VisitNoFix, allHonor[i].TxnCode, err)
				} else if isNetChosen {
					allHonor[i].TxnType = "fix"

					_, err := db.Exec(`
				UPDATE patient_bill
				SET txn_type = 'fix'
				WHERE id = ?`,
						allHonor[i].ID)
					if err != nil {
						// log.Printf("🔒 %s (%s) berubah jadi FIX karena min() memilih net_amount", allHonor[i].TxnCode, allHonor[i].VisitNo)
					}
				}
			}

			//  FIX RULE
			if allHonor[i].TxnType == "fix" && allHonor[i].PatientClass == "BPJS" {
				// log.Printf("%s pada visit number %s tidak ikut di proporsional karena masih <0.2 inacbg", allHonor[i].TxnCode, allHonor[i].VisitNo)
				allHonor[i].HonorProp = 0
				allHonor[i].HonorFinal = honor

			}
			// ECHOCARDIOGRAPHY OP
			if allHonor[i].PatientType == "OUTPATIENTS" {
				if allHonor[i].TxnDesc == "ECHOCARDIOGRAPHY" {
					allHonor[i].TxnType = "tindakan"
				}
			}
			// log.Printf("data : %v", allHonor[i])

		}

		// 3️⃣ Group per visit_no untuk perhitungan proporsional BPJS
		grouped := make(map[string][]Honorfull)
		for _, t := range allHonor {
			grouped[t.VisitNoFix] = append(grouped[t.VisitNoFix], t)
		}

		for visitNo, list := range grouped {

			// log.Printf("Data : %v", list)
			// ✅ Non-BPJS: langsung simpan hasil master
			if list[0].PatientClass != "BPJS" {
				for i := range list {
					// log.Printf("Data Non BPJS: %v", list)
					list[i].HonorFinal = list[i].HonorMaster
				}
				grouped[visitNo] = list

				continue
			}

			// ✅ Step 1: Cek apakah visit mengandung tindakan
			hasTindakan := false
			for _, tx := range list {
				// log.Printf("%s dokter %s tipe %s", tx.TxnCode, tx.TxnDoctor, tx.TxnType)
				if tx.PreviousTxnType == "tindakan" {
					if tx.HonorMaster != 0 {
						hasTindakan = true
					}
					// log.Printf("ada tindakan pada visit number %v", tx.VisitNo)
					break
				}

				// log.Printf("tidak ada tindakan pada visit number %v", tx.VisitNo)
			}

			// ✅ Step 2: Tentukan batas honor visit_no
			limit := 0.0
			for _, tx := range list {

				switch tx.PatientType {
				case "INPATIENTS":
					limit = 0.2 * list[0].Inacbg
					// log.Printf("Inacbg:%v", list[0].Inacbg)

					if hasTindakan {
						limit = 0.4 * list[0].Inacbg
					}

				case "OUTPATIENTS":
					limit = list[0].Inacbg
					// log.Printf("Inacbg:%v", list[0].Inacbg)
				}

			}
			// log.Printf("Limit awal %v", limit)

			// ✅ Step 3b honor dr. Wihan
			drwilhan := make(map[string]bool)
			for _, tx := range list {

				if !hasTindakan && tx.PatientType == "INPATIENTS" {
					if tx.TxnDoctor == "dr.WILHAN,SP.PD" {
						drwilhan[tx.TxnDoctor] = true
						// log.Println("ada dokter wilhan")
					}

				}
			}

			for i := range list {
				tx := &list[i]
				if tx.PreviousTxnType == "visit" {
					if drwilhan[tx.TxnDoctor] {
						tx.HonorMaster = tx.Inacbg * 0.15
						if tx.HonorMaster > tx.NetAmount {
							tx.HonorMaster = tx.NetAmount
						}
					}

				}

			}

			// log.Printf("Data:%v", list)
			// ✅ Step 3b Nolkan honor visit dan fix untuk dokter yang punya tindakan
			doctorHasTindakan := make(map[string]bool)
			visithastindakan := make(map[string]bool)
			for _, tx := range list {
				if tx.TxnType == "tindakan" && tx.HonorMaster != 0 {
					// log.Printf("dokter %v memiliki tindakan berbayar dan visit pada visit number %v", tx.TxnDoctor, tx.VisitNo)
					doctorHasTindakan[tx.TxnDoctor] = true
					visithastindakan[tx.TxnCode] = true

				}
			}
			doctorChangedToTindakan := make(map[string]bool)
			for _, tx := range list {
				if tx.PreviousTxnType == "fix" && tx.TxnType == "tindakan" {
					doctorChangedToTindakan[tx.TxnDoctor] = true
					// log.Printf("dokter %v memiliki tindakan berbayar dan visit pada visit number %v", tx.TxnDoctor, tx.VisitNoFix)
				}
			}
			doctorChangedToFix := make(map[string]bool)
			for _, tx := range list {
				// misal kamu punya field PreviousTxnType dari query awal
				if tx.PreviousTxnType == "tindakan" && tx.TxnType == "fix" {
					doctorChangedToFix[tx.TxnDoctor] = true
					// log.Printf("⚠️ Dokter %s berubah dari tindakan → fix pada visit %s", tx.TxnDoctor, tx.VisitNo)
				}
			}
			// Pengecekan Colono/Gastro
			colonoorgastro := false
			for _, tx := range list {
				if tx.PatientType == "OUTPATIENTS" {
					if tx.TxnCode == "ENDOS0000005" || tx.TxnCode == "ENDOS0000006" {
						colonoorgastro = true
						// log.Println("tes1")
					}
				}
			}

			// Pengecekan apakah ada anastesi
			anastesi := false
			for _, tx := range list {
				if tx.PatientType == "OUTPATIENTS" {
					if tx.TxnCode == "CFEEENDOS007" || tx.TxnCode == "CFEEENDOS008" {
						anastesi = true
						// log.Println("tes2")
					}
				}
			}

			// Pengecekan apakah ada SATU SEP
			sep := false
			for _, tx := range list {
				if tx.Status == "SATU SEP" {
					sep = true
					// log.Println("tes3")
				}
			}
			visittindakan := make(map[string]bool)
			for _, tx := range list {
				if visithastindakan[tx.TxnCode] {
					// log.Printf("pada visit no %s txn code %s adalah tindakan", tx.VisitNo, tx.TxnCode)
					visittindakan[tx.VisitNo] = true

				}
			}
			// Pengecekan apakah ada honor visit di luar dari visit no yang ada tindakan
			for _, tx := range list {
				if visittindakan[tx.VisitNo] {
					// log.Printf("txn code %s memiliki visit no yang sama dengan tindakan", tx.TxnCode)
				} else {
					// log.Printf("txn code %s tidak memiliki visit no yang sama dengan tindakan", tx.TxnCode)
				}
			}
			for i := range list {
				tx := &list[i]
				if colonoorgastro && anastesi {
					if tx.TxnCode == "ENDOS0000005" || tx.TxnCode == "ENDOS0000006" {
						tx.HonorMaster = tx.Inacbg * 0.35
						// log.Println("tes")
					}
					if tx.TxnCode == "CFEEENDOS007" || tx.TxnCode == "CFEEENDOS008" {
						tx.HonorMaster = tx.Inacbg * 0.10
					}
				}
				if colonoorgastro && anastesi && sep {
					limit = tx.Inacbg * 0.45
					if tx.TxnCode == "ENDOS0000005" || tx.TxnCode == "ENDOS0000006" {
						tx.HonorMaster = limit * 3 / 4
						// log.Println("tes")
					}
					if tx.TxnCode == "CFEEENDOS007" || tx.TxnCode == "CFEEENDOS008" {
						tx.HonorMaster = limit * 1 / 4
					}
					if tx.TxnType == "visit" && !visittindakan[tx.VisitNo] {
						tx.TxnType = "fix"
						log.Println("di satu sep kan ")
					}
				}
				// log.Printf("Limit Baru:%v", limit)

			}

			for i := range list {
				tx := &list[i]
				if tx.PreviousTxnType != "tindakan" && tx.TxnType != "tindakan" && !sep && tx.Inacbg < 200000 {
					if doctorHasTindakan[tx.TxnDoctor] || doctorChangedToFix[tx.TxnDoctor] || doctorChangedToTindakan[tx.TxnDoctor] {
						// log.Println("honor visit dan fix di 0 kan ")
						tx.HonorMaster = 0
						tx.HonorFinal = 0
					}
				}
				// log.Printf("Txn %s dokter %s honor master = %v", tx.TxnDesc, tx.TxnDoctor, tx.HonorMaster)
			}
			// ENT Clinic cek untuk hanya memberikan honor visit max 50.000
			for _, tx := range list {
				if tx.TxnCategory == "ENT CLINIC" && tx.Inacbg > 200000 {

				}
			}
			// Jika ada 2 tindakan oleh dokter yang sama → hanya bayar 1 (net tertinggi)
			tindakanByDoctor := make(map[string][]*Honorfull)
			// Kumpulkan semua tindakan per dokter
			for i := range list {
				tx := &list[i]
				if tx.PreviousTxnType == "tindakan" && tx.HonorMaster != 0 {
					tindakanByDoctor[tx.TxnDoctor] = append(tindakanByDoctor[tx.TxnDoctor], tx)
				}
			}

			// Cek per dokter
			for _, tindakanList := range tindakanByDoctor {
				if len(tindakanList) > 1 {
					// Cari tindakan dengan net_amount tertinggi
					maxIdx := 0
					maxNet := tindakanList[0].NetAmount
					for i, t := range tindakanList {
						if t.NetAmount > maxNet {
							maxNet = t.NetAmount
							maxIdx = i
						}
					}

					// Nolkan semua tindakan kecuali yang tertinggi
					for i, t := range tindakanList {
						if i != maxIdx {
							// log.Printf("❌ Nolkan tindakan dokter %s pada visit %s (lebih dari 1 tindakan, hanya ambil net tertinggi %.2f)", doctor, t.VisitNo, maxNet)
							t.HonorMaster = 0
							t.HonorFinal = 0
						}
					}
				}
			}

			// ✅ Step 4: Hitung totalMaster (skip fix)
			totalMaster := 0.0
			for _, tx := range list {
				if tx.TxnType == "fix" {
					limit -= tx.HonorMaster //pengurangan limit dengan honor fix
					continue
				}
				totalMaster += tx.HonorMaster
				// log.Printf("total master %v", totalMaster)
				// log.Printf(" limit : %v", limit)

			}

			if totalMaster == 0 {
				grouped[visitNo] = list
				continue
			}

			// ✅ Step 5: Proporsional scaling jika melebihi batas
			scale := 1.0
			if totalMaster > limit {
				// log.Printf("total honor %v melebihi limit %v ", totalMaster, limit)
				scale = limit / totalMaster
				// log.Printf("visit number %s Scale:%v", visitNo, scale)
			}

			for i := range list {
				tx := &list[i]
				if tx.TxnType == "fix" {
					tx.HonorFinal = tx.HonorMaster
					continue
				}
				tx.HonorProp = tx.HonorMaster * scale
				tx.HonorFinal = tx.HonorProp
			}

			// ✅ Step 8: Cek ulang jika proporsional honor visit <50% honor master
			includeFix := false
			for _, tx := range list {
				if tx.TxnType == "visit" && tx.HonorProp < 0.5*tx.HonorMaster {
					includeFix = true
					// log.Printf("honor proporsional dari %s %v <0.5 honor master %v", tx.TxnDesc, tx.HonorProp, tx.HonorMaster)
					// log.Println("proporsional honor visit melebihi 1/2 dari master, honor fix ikut di proporsionalkan")
					break
				}
			}
			//kembalikan nominal limit yang sebelumnya sudah dikurangi
			// log.Printf("Limit sebelumnya : %v", limit)
			for _, tx := range list {
				if tx.TxnType == "fix" {
					limit += tx.HonorMaster
					// log.Printf("Limit : %v", limit)
				}
			}
			//perhitungan proporsional kembali include txn type fix
			if includeFix {
				totalMasterbaru := 0.0
				for _, tx := range list {
					totalMasterbaru += tx.HonorMaster
					// log.Printf("%v + honor master %v", totalMasterbaru, tx.HonorMaster)
				}
				// log.Printf("Limit akhir %v", limit)
				newScale := limit / totalMasterbaru
				// log.Printf("total master baru: %v scale baru: %v", totalMasterbaru, newScale)
				for i := range list {
					tx := &list[i]
					tx.HonorProp = tx.HonorMaster * newScale
					tx.HonorFinal = tx.HonorProp
					// log.Printf("Honor Final : %v", tx.HonorFinal)
				}

			}

			// ✅ Step 8: Final check — bulatkan & pastikan tidak melebihi limit total
			totalFinal := 0.0
			for _, tx := range list {
				totalFinal += tx.HonorFinal
			}

			// Jika total melebihi limit (karena pembulatan bisa bikin sedikit lewat)
			if totalFinal > limit {
				// Hitung ulang scale kecil agar totalFinal == limit
				scaleAdjust := limit / totalFinal
				for i := range list {
					list[i].HonorFinal = math.Floor(list[i].HonorFinal * scaleAdjust)
				}
			}

			// Bulatkan semua nilai final agar tidak ada desimal
			for i := range list {
				list[i].HonorProp = math.Floor(list[i].HonorProp)
				list[i].HonorFinal = math.Floor(list[i].HonorFinal)

			}
			// ✅ Jika memang dilakukan proporsionalisasi (scale < 1) → pastikan totalFinal == limit
			if scale < 1.0 {
				totalFinal = 0
				for _, tx := range list {
					totalFinal += tx.HonorFinal
				}

				diff := math.Round(limit - totalFinal)

				if diff > 0 {
					// Kumpulkan pecahan sebelum floor (hanya untuk transaksi non-fix dengan HonorMaster > 0)
					type remainder struct {
						idx  int
						frac float64
					}
					var remainders []remainder

					for i, tx := range list {
						if tx.HonorMaster == 0 || tx.TxnType == "fix" {
							continue
						}
						frac := tx.HonorProp - math.Floor(tx.HonorProp)
						remainders = append(remainders, remainder{i, frac})
					}

					// Urutkan berdasarkan pecahan terbesar
					sort.Slice(remainders, func(i, j int) bool {
						return remainders[i].frac > remainders[j].frac
					})

					// Tambahkan +1 ke yang pecahannya paling besar sampai diff habis
					for i := 0; i < int(diff) && i < len(remainders); i++ {
						list[remainders[i].idx].HonorFinal += 1
					}
				}

				// Recalculate total untuk memastikan hasil akhir pas
				totalFinal = 0
				for _, tx := range list {
					totalFinal += tx.HonorFinal
					// log.Printf("txn %s honor final = %v", tx.TxnDesc, tx.HonorFinal)
				}
				// log.Printf("✅ Visit %s: total honor proporsional akhir = %.0f (limit = %.0f)", visitNo, totalFinal, limit)
			} else {
				// Non-proporsional (totalMaster <= limit) — tidak perlu redistribusi
				totalFinal = 0
				for _, tx := range list {
					totalFinal += tx.HonorFinal
				}
				// log.Printf("🟢 Visit %s: tidak proporsional (total %.0f <= limit %.0f)", visitNo, totalFinal, limit)
			}

			grouped[visitNo] = list
		}

		// Step 8: Simpan kembali hasilnya ke database
		tx, err := db.Begin()
		if err != nil {
			log.Println("gagal mulai transaksi:", err)
			return err
		}

		for _, list := range grouped {
			for _, t := range list {
				_, err := tx.Exec(`
						UPDATE patient_bill t
						LEFT JOIN comparison_data c ON t.visit_no_fix = c.visit_number
						LEFT JOIN piutang p ON t.visit_no_fix = p.visit_no
						SET t.honor_master = ?,t.honor_prop=?, t.honor_final = ?,
						t.honor_status = CASE
						WHEN (
							(t.patient_class='BPJS' AND c.status <> 'OFFER (PENDING)') 
							OR 
							(t.patient_class IN ('INSURANCE','CORPORATE') AND COALESCE(p.sisa,1)=0)
							OR
							(t.patient_class IN ('GENERAL','HOSPITAL STAFF'))
						)
							THEN 'COUNTED'
							ELSE t.honor_status
						END,
						t.counted_date= CASE
						WHEN(
							(t.patient_class='BPJS' AND c.status <> 'OFFER (PENDING)') 
							OR 
							(t.patient_class IN ('INSURANCE','CORPORATE') AND COALESCE(p.sisa,1)=0)
							OR
							(t.patient_class IN ('GENERAL','HOSPITAL STAFF') AND t.bill_status = 'PAID')
						)
							THEN NOW()
							ELSE t.counted_date
						END
						WHERE t.id = ?;						
						`,
					(t.HonorMaster), (t.HonorProp), (t.HonorFinal), t.ID)
				if err != nil {
					log.Printf("❌ Gagal update %d: %v", t.ID, err)
				}
			}
		}
		if err := tx.Commit(); err != nil {
			log.Println("❌ Gagal commit transaksi:", err)
			return err
		}

		log.Println("✅ Proses perhitungan honor selesai.")
		return nil
	}
}

// Chart Pada Dashboard
func GetHonorChart(db *sql.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		doctorName := c.Query("doctor_name")
		year := c.Query("year")
		if year == "" {
			return c.Status(400).JSON(fiber.Map{"error": "year is required"})
		}

		// 🔥 Base query
		query := `
			SELECT 
				MONTH(hrd.created_at) AS bulan,
				SUM(hrd.total_honor) AS total
				FROM honor_request_detail hrd
				JOIN honor_request hr ON hrd.request_id = hr.id
			WHERE YEAR(hrd.created_at) = ? 
		`

		args := []interface{}{year}

		// 🔥 Jika dokter diberikan → filter
		if doctorName != "" {
			// log.Printf("Dokter:%s", doctorName)
			query += " AND doctor_name LIKE ?"
			args = append(args, "%"+doctorName+"%")
		}

		query += `			
			AND hr.status = 'APPROVED' 
			GROUP BY MONTH(created_at)
			ORDER BY bulan ASC
		`

		rows, err := db.Query(query, args...)
		if err != nil {
			log.Printf("query error: %v", err)
			return c.Status(500).JSON(fiber.Map{"error": err.Error()})
		}
		defer rows.Close()

		monthlyData := make(map[int]float64)
		for rows.Next() {
			var month int
			var total float64
			if err := rows.Scan(&month, &total); err != nil {
				log.Printf("scan error: %v", err)
				return c.Status(500).JSON(fiber.Map{"error": err.Error()})
			}
			monthlyData[month] = total
		}

		monthNames := []string{
			"Jan", "Feb", "Mar", "Apr", "May", "Jun",
			"Jul", "Aug", "Sep", "Oct", "Nov", "Dec",
		}

		var result []fiber.Map
		for i := 1; i <= 12; i++ {
			result = append(result, fiber.Map{
				"month": monthNames[i-1],
				"Total": monthlyData[i], // default 0
			})
		}

		return c.JSON(fiber.Map{
			"filter_doctor": doctorName,
			"year":          year,
			"monthlyHonor":  result,
		})
	}
}

// 📘 Fungsi bantu evaluasi rumus menggunakan govaluate
func evalRumus(rumus string, t Honorfull) (float64, error) {
	if rumus == "" {
		return 0, nil
	}

	functions := map[string]govaluate.ExpressionFunction{
		"min": func(args ...interface{}) (interface{}, error) {
			if len(args) < 2 {
				return nil, fmt.Errorf("min butuh 2 argumen")
			}
			a := toFloat(args[0])
			b := toFloat(args[1])
			return math.Min(a, b), nil
		},
		"max": func(args ...interface{}) (interface{}, error) {
			if len(args) < 2 {
				return nil, fmt.Errorf("max butuh 2 argumen")
			}
			a := toFloat(args[0])
			b := toFloat(args[1])
			return math.Max(a, b), nil
		},
	}

	expr, err := govaluate.NewEvaluableExpressionWithFunctions(rumus, functions)
	if err != nil {
		return 0, fmt.Errorf("rumus tidak valid: %v", err)
	}

	params := map[string]interface{}{
		"qty":           t.Qty,
		"net_amount":    t.NetAmount,
		"tarif_ina_cbg": t.Inacbg,
	}

	result, err := expr.Evaluate(params)
	if err != nil {
		return 0, fmt.Errorf("gagal evaluasi: %v", err)
	}

	val, ok := result.(float64)
	if !ok {
		return 0, fmt.Errorf("hasil bukan angka")
	}
	// log.Printf("Hasil perhitungan: %v", val)
	return val, nil
}

func evalMinChoice(rumus string, t Honorfull) (bool, error) {
	// Cek apakah rumusnya memang mengandung "min("
	if !strings.Contains(strings.ToLower(rumus), "min(") {
		return false, nil
	}

	// Ambil isi dalam tanda kurung, contoh: min(a,b) → a,b
	start := strings.Index(rumus, "(")
	end := strings.LastIndex(rumus, ")")
	if start == -1 || end == -1 || end <= start {
		return false, fmt.Errorf("format min() tidak valid: %s", rumus)
	}
	args := rumus[start+1 : end]
	parts := strings.Split(args, ",")
	if len(parts) < 2 {
		return false, fmt.Errorf("min() harus punya 2 argumen")
	}

	// Evaluasi kedua argumen
	leftExpr, _ := govaluate.NewEvaluableExpression(parts[0])
	rightExpr, _ := govaluate.NewEvaluableExpression(parts[1])

	params := map[string]interface{}{
		"qty":           t.Qty,
		"net_amount":    t.NetAmount,
		"tarif_ina_cbg": t.Inacbg,
	}

	leftVal, err1 := leftExpr.Evaluate(params)
	rightVal, err2 := rightExpr.Evaluate(params)
	if err1 != nil || err2 != nil {
		return false, fmt.Errorf("gagal evaluasi argumen min(): %v, %v", err1, err2)
	}

	left := toFloat(leftVal)
	right := toFloat(rightVal)

	// 🧩 Jika nilai net_amount lebih kecil dari argumen lainnya → berarti min memilih net_amount
	// tapi kita cek bukan dari nama variabel, melainkan dari nilainya
	if right < left {
		return true, nil // true artinya: hasil min berasal dari perhitungan net_amount
	}
	return false, nil
}

func toFloat(v interface{}) float64 {
	switch val := v.(type) {
	case float64:
		return val
	case int:
		return float64(val)
	case int64:
		return float64(val)
	default:
		return 0
	}
}

func HonorCountHandler(db *sql.DB) fiber.Handler {
	return HonorCount(db)
}

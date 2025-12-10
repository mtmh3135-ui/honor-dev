package main

// func LoadPatientBillUpdate(db *sql.DB) ([]Honorfull, error) {
// 	rows, err := db.Query(`
//         SELECT
//   				t.id,
// 				t.visit_no,
//   				t.visit_no_fix,
//   				t.txn_code,
// 				t.txn_category,
// 				t.txn_desc,
//   				t.patient_type,
//   				t.patient_class,
//   				t.qty,
//   				t.net_amount,
//   				t.txn_doctor,
//   				m.txn_type,
//   				m.bpjs_ip,
//   				m.bpjs_op,
//   				m.rumus_general,
//   				IFNULL(e.status, '-') AS status,
//   				IFNULL(c.tarif_ina_cbg, 0) AS tarif_ina_cbg,
// 				IFNULL(c.kelas_bpjs,'-') AS kelas_bpjs,
// 				IFNULL(d.description,'-') as description
// 				FROM patient_bill t
// 				JOIN master_txn m ON t.txn_code = m.txn_code
// 				LEFT JOIN comparison_data c ON t.visit_no_fix = c.visit_number
// 				LEFT JOIN comparison_data e ON t.visit_no = e.visit_number
// 				LEFT JOIN doctor_data d on t.txn_doctor = d.doctor_name
// 		WHERE
// 		(
//   			-- 🔹 CASE 1: Pasien BPJS
//   			(
//     			(t.honor_status IS NULL OR t.honor_status NOT IN ('FINISH', 'ON PROGRESS'))
//     			AND t.patient_class = 'BPJS'
//     			AND DATE_FORMAT(t.bill_datetime, '%Y-%m') < DATE_FORMAT(DATE_SUB(NOW(), INTERVAL 2 MONTH), '%Y-%m')
//   			)

//   			-- 🔹 CASE 2: Pasien NON BPJS
//   			OR
//   			(
//     			(t.honor_status IS NULL OR t.honor_status NOT IN ('FINISH', 'ON PROGRESS'))
//     			AND t.patient_class <> 'BPJS'
//     			AND DATE_FORMAT(t.bill_datetime, '%Y-%m') < DATE_FORMAT(DATE_SUB(NOW(), INTERVAL 1 MONTH), '%Y-%m')
//   			)
// 		);
//     `)
// 	if err != nil {
// 		return nil, err
// 	}
// 	defer rows.Close()

// 	var result []Honorfull

// 	for rows.Next() {
// 		var item Honorfull

// 		rows.Scan(
// 			&item.ID, &item.VisitNo, &item.VisitNoFix, &item.TxnCode, &item.TxnCategory, &item.TxnDesc, &item.PatientType, &item.PatientClass, &item.Qty,
// 			&item.NetAmount, &item.TxnDoctor, &item.TxnType, &item.Bpjs_ip, &item.Bpjs_op, &item.RumusGeneral,
// 			&item.Status, &item.Inacbg, &item.BPJSClass, &item.Description,
// 		)

// 		result = append(result, item)
// 	}
// 	return result, nil
// }

// func SaveToPatientBill(db *sql.DB, grouped map[string][]Honorfull) error {
// 	//  Simpan kembali hasilnya ke database
// 	tx, err := db.Begin()
// 	if err != nil {
// 		log.Println("gagal mulai transaksi:", err)
// 		return err
// 	}

// 	for _, list := range grouped {
// 		for _, t := range list {
// 			_, err := tx.Exec(`
// 						UPDATE patient_bill t
// 						LEFT JOIN comparison_data c ON t.visit_no_fix = c.visit_number
// 						LEFT JOIN piutang p ON t.visit_no_fix = p.visit_no
// 						SET t.honor_master = ?,t.honor_prop=?, t.honor_final = ?,
// 						t.honor_status = CASE
// 						WHEN (
// 							(t.patient_class='BPJS' AND c.status <> 'OFFER (PENDING)')
// 							OR
// 							(t.patient_class IN ('INSURANCE','CORPORATE') AND COALESCE(p.sisa,1)=0)
// 							OR
// 							(t.patient_class IN ('GENERAL','HOSPITAL STAFF'))
// 						)
// 							THEN 'COUNTED'
// 							ELSE t.honor_status
// 						END,
// 						t.counted_date= CASE
// 						WHEN(
// 							(t.patient_class='BPJS' AND c.status <> 'OFFER (PENDING)')
// 							OR
// 							(t.patient_class IN ('INSURANCE','CORPORATE') AND COALESCE(p.sisa,1)=0)
// 							OR
// 							(t.patient_class IN ('GENERAL','HOSPITAL STAFF') AND t.bill_status = 'PAID')
// 						)
// 							THEN NOW()
// 							ELSE t.counted_date
// 						END
// 						WHERE t.id = ?;
// 						`,
// 				(t.HonorMaster), (t.HonorProp), (t.HonorFinal), t.ID)
// 			if err != nil {
// 				log.Printf("❌ Gagal update %d: %v", t.ID, err)
// 			}
// 		}
// 	}
// 	if err := tx.Commit(); err != nil {
// 		log.Println("❌ Gagal commit transaksi:", err)
// 		return err
// 	}

// 	log.Println("✅ Proses perhitungan honor selesai.")
// 	return err
// }

// // ✅ Fungsi utama pemrosesan honor
// func TesHonorCount(db *sql.DB, items []Honorfull) []Honorfull {
// 	result := make([]Honorfull, len(items))
// 	// 2️⃣ Hitung honor master (dari rumus)
// 	for i := range items {

// 		rumus := items[i].RumusGeneral
// 		if items[i].PatientClass == "BPJS" && items[i].PatientType == "INPATIENTS" {
// 			rumus = items[i].Bpjs_ip
// 		} else if items[i].PatientClass == "BPJS" && items[i].PatientType == "OUTPATIENTS" {
// 			rumus = items[i].Bpjs_op
// 		}
// 		honor, err := evalRumus(rumus, items[i])
// 		if err != nil {
// 			log.Printf("⚠️ Gagal evaluasi %s %s: %v", items[i].VisitNoFix, items[i].TxnCode, err)
// 			continue
// 		}

// 		// Set default dulu
// 		items[i].HonorMaster = honor

// 		//  Kondisi jika ada transaksi yang tidak berisikan class BPJS
// 		if items[i].TxnDesc == "HONOR DOKTER" {
// 			switch items[i].BPJSClass {
// 			case "I":
// 				items[i].HonorMaster = 75000 * items[i].Qty
// 			case "II":
// 				items[i].HonorMaster = 60000 * items[i].Qty
// 			case "III":
// 				items[i].HonorMaster = 50000 * items[i].Qty
// 			}

// 		}
// 		if items[i].TxnDesc == "KONSULTASI DOKTER" && items[i].PatientType == "INPATIENT" {
// 			switch items[i].BPJSClass {
// 			case "I":
// 				items[i].HonorMaster = 75000 * items[i].Qty
// 			case "II":
// 				items[i].HonorMaster = 60000 * items[i].Qty
// 			case "III":
// 				items[i].HonorMaster = 50000 * items[i].Qty
// 			}

// 		}

// 		//  Kondisi khusus dr. MUTIARA MARGARETHA
// 		if items[i].TxnDoctor == "dr. MUTIARA MARGARETHA, SpJP" {
// 			// log.Println("Ada spesialis jantung")
// 			switch items[i].TxnDesc {
// 			case "HONOR DOKTER SPESIALIS VISITE (CLASS I)":
// 				items[i].HonorMaster = 100000 * items[i].Qty
// 				// log.Printf(" Dokter %s (Spesialis JANTUNG) - honor disesuaikan menjadi %v pada visit number %s", items[i].TxnDoctor, items[i].HonorMaster, items[i].VisitNo)
// 			case "HONOR DOKTER SPESIALIS VISITE (CLASS II)":
// 				items[i].HonorMaster = 75000 * items[i].Qty
// 				// log.Printf(" Dokter %s (Spesialis JANTUNG) - honor disesuaikan pada visit number %s", items[i].TxnDoctor, items[i].VisitNo)
// 			}
// 			// log.Printf(" Dokter %s (Spesialis JANTUNG) - honor disesuaikan pada visit number %s", items[i].TxnDoctor, items[i].VisitNo)
// 		}

// 		//  Validasi dinamis jika rumus mengandung fungsi min()
// 		if strings.Contains(strings.ToLower(rumus), "min(") {
// 			isNetChosen, err := evalMinChoice(rumus, items[i])
// 			if err != nil {
// 				log.Printf("⚠️ Gagal deteksi min() %s %s: %v", items[i].VisitNoFix, items[i].TxnCode, err)
// 			} else if isNetChosen {
// 				items[i].TxnType = "fix"

// 				_, err := db.Exec(`
// 				UPDATE patient_bill
// 				SET txn_type = 'fix'
// 				WHERE id = ?`,
// 					items[i].ID)
// 				if err != nil {
// 					// log.Printf("🔒 %s (%s) berubah jadi FIX karena min() memilih net_amount", items[i].TxnCode, items[i].VisitNo)
// 				}
// 			}
// 		}

// 		//  FIX RULE
// 		if items[i].TxnType == "fix" && items[i].PatientClass == "BPJS" {
// 			// log.Printf("%s pada visit number %s tidak ikut di proporsional karena masih <0.2 inacbg", items[i].TxnCode, items[i].VisitNo)
// 			items[i].HonorProp = 0
// 			items[i].HonorFinal = honor

// 		}
// 		// ECHOCARDIOGRAPHY OP
// 		if items[i].PatientType == "OUTPATIENTS" {
// 			if items[i].TxnDesc == "ECHOCARDIOGRAPHY" {
// 				items[i].TxnType = "tindakan"
// 			}
// 		}
// 		// log.Printf("data : %v", items[i])

// 	}

// 	// 3️⃣ Group per visit_no untuk perhitungan proporsional BPJS
// 	grouped := make(map[string][]Honorfull)
// 	for _, t := range items {
// 		grouped[t.VisitNoFix] = append(grouped[t.VisitNoFix], t)
// 	}

// 	for visitNo, list := range grouped {

// 		// log.Printf("Data : %v", list)
// 		// ✅ Non-BPJS: langsung simpan hasil master
// 		if list[0].PatientClass != "BPJS" {
// 			for i := range list {
// 				// log.Printf("Data Non BPJS: %v", list)
// 				list[i].HonorFinal = list[i].HonorMaster
// 			}
// 			grouped[visitNo] = list

// 			continue
// 		}

// 		// ✅ Step 1: Cek apakah visit mengandung tindakan
// 		hasTindakan := false
// 		for _, tx := range list {
// 			// log.Printf("%s dokter %s tipe %s", tx.TxnCode, tx.TxnDoctor, tx.TxnType)
// 			if tx.PreviousTxnType == "tindakan" {
// 				if tx.HonorMaster != 0 {
// 					hasTindakan = true
// 				}
// 				// log.Printf("ada tindakan pada visit number %v", tx.VisitNo)
// 				break
// 			}

// 			// log.Printf("tidak ada tindakan pada visit number %v", tx.VisitNo)
// 		}

// 		// ✅ Step 2: Tentukan batas honor visit_no
// 		limit := 0.0
// 		for _, tx := range list {

// 			switch tx.PatientType {
// 			case "INPATIENTS":
// 				limit = 0.2 * list[0].Inacbg
// 				// log.Printf("Inacbg:%v", list[0].Inacbg)

// 				if hasTindakan {
// 					limit = 0.4 * list[0].Inacbg
// 				}

// 			case "OUTPATIENTS":
// 				limit = list[0].Inacbg
// 				// log.Printf("Inacbg:%v", list[0].Inacbg)
// 			}

// 		}
// 		// log.Printf("Limit awal %v", limit)

// 		// ✅ Step 3b honor dr. Wihan
// 		drwilhan := make(map[string]bool)
// 		for _, tx := range list {

// 			if !hasTindakan && tx.PatientType == "INPATIENTS" {
// 				if tx.TxnDoctor == "dr.WILHAN,SP.PD" {
// 					drwilhan[tx.TxnDoctor] = true
// 					// log.Println("ada dokter wilhan")
// 				}

// 			}
// 		}

// 		for i := range list {
// 			tx := &list[i]
// 			if tx.PreviousTxnType == "visit" {
// 				if drwilhan[tx.TxnDoctor] {
// 					tx.HonorMaster = tx.Inacbg * 0.15
// 					if tx.HonorMaster > tx.NetAmount {
// 						tx.HonorMaster = tx.NetAmount
// 					}
// 				}

// 			}

// 		}

// 		// log.Printf("Data:%v", list)
// 		// ✅ Step 3b Nolkan honor visit dan fix untuk dokter yang punya tindakan
// 		doctorHasTindakan := make(map[string]bool)
// 		visithastindakan := make(map[string]bool)
// 		for _, tx := range list {
// 			if tx.TxnType == "tindakan" && tx.HonorMaster != 0 {
// 				// log.Printf("dokter %v memiliki tindakan berbayar dan visit pada visit number %v", tx.TxnDoctor, tx.VisitNo)
// 				doctorHasTindakan[tx.TxnDoctor] = true
// 				visithastindakan[tx.TxnCode] = true

// 			}
// 		}
// 		doctorChangedToTindakan := make(map[string]bool)
// 		for _, tx := range list {
// 			if tx.PreviousTxnType == "fix" && tx.TxnType == "tindakan" {
// 				doctorChangedToTindakan[tx.TxnDoctor] = true
// 				// log.Printf("dokter %v memiliki tindakan berbayar dan visit pada visit number %v", tx.TxnDoctor, tx.VisitNoFix)
// 			}
// 		}
// 		doctorChangedToFix := make(map[string]bool)
// 		for _, tx := range list {
// 			// misal kamu punya field PreviousTxnType dari query awal
// 			if tx.PreviousTxnType == "tindakan" && tx.TxnType == "fix" {
// 				doctorChangedToFix[tx.TxnDoctor] = true
// 				// log.Printf("⚠️ Dokter %s berubah dari tindakan → fix pada visit %s", tx.TxnDoctor, tx.VisitNo)
// 			}
// 		}
// 		// Pengecekan Colono/Gastro
// 		colonoorgastro := false
// 		for _, tx := range list {
// 			if tx.PatientType == "OUTPATIENTS" {
// 				if tx.TxnCode == "ENDOS0000005" || tx.TxnCode == "ENDOS0000006" {
// 					colonoorgastro = true
// 					// log.Println("tes1")
// 				}
// 			}
// 		}

// 		// Pengecekan apakah ada anastesi
// 		anastesi := false
// 		for _, tx := range list {
// 			if tx.PatientType == "OUTPATIENTS" {
// 				if tx.TxnCode == "CFEEENDOS007" || tx.TxnCode == "CFEEENDOS008" {
// 					anastesi = true
// 					// log.Println("tes2")
// 				}
// 			}
// 		}

// 		// Pengecekan apakah ada SATU SEP
// 		sep := false
// 		for _, tx := range list {
// 			if tx.Status == "SATU SEP" {
// 				sep = true
// 				// log.Println("tes3")
// 			}
// 		}
// 		visittindakan := make(map[string]bool)
// 		for _, tx := range list {
// 			if visithastindakan[tx.TxnCode] {
// 				// log.Printf("pada visit no %s txn code %s adalah tindakan", tx.VisitNo, tx.TxnCode)
// 				visittindakan[tx.VisitNo] = true

// 			}
// 		}
// 		// Pengecekan apakah ada honor visit di luar dari visit no yang ada tindakan
// 		for _, tx := range list {
// 			if visittindakan[tx.VisitNo] {
// 				// log.Printf("txn code %s memiliki visit no yang sama dengan tindakan", tx.TxnCode)
// 			} else {
// 				// log.Printf("txn code %s tidak memiliki visit no yang sama dengan tindakan", tx.TxnCode)
// 			}
// 		}
// 		for i := range list {
// 			tx := &list[i]
// 			if colonoorgastro && anastesi {
// 				if tx.TxnCode == "ENDOS0000005" || tx.TxnCode == "ENDOS0000006" {
// 					tx.HonorMaster = tx.Inacbg * 0.35
// 					// log.Println("tes")
// 				}
// 				if tx.TxnCode == "CFEEENDOS007" || tx.TxnCode == "CFEEENDOS008" {
// 					tx.HonorMaster = tx.Inacbg * 0.10
// 				}
// 			}
// 			if colonoorgastro && anastesi && sep {
// 				limit = tx.Inacbg * 0.45
// 				if tx.TxnCode == "ENDOS0000005" || tx.TxnCode == "ENDOS0000006" {
// 					tx.HonorMaster = limit * 3 / 4
// 					// log.Println("tes")
// 				}
// 				if tx.TxnCode == "CFEEENDOS007" || tx.TxnCode == "CFEEENDOS008" {
// 					tx.HonorMaster = limit * 1 / 4
// 				}
// 				if tx.TxnType == "visit" && !visittindakan[tx.VisitNo] {
// 					tx.TxnType = "fix"
// 					log.Println("di satu sep kan ")
// 				}
// 			}
// 			// log.Printf("Limit Baru:%v", limit)

// 		}

// 		for i := range list {
// 			tx := &list[i]
// 			if tx.PreviousTxnType != "tindakan" && tx.TxnType != "tindakan" && !sep && tx.Inacbg < 200000 {
// 				if doctorHasTindakan[tx.TxnDoctor] || doctorChangedToFix[tx.TxnDoctor] || doctorChangedToTindakan[tx.TxnDoctor] {
// 					// log.Println("honor visit dan fix di 0 kan ")
// 					tx.HonorMaster = 0
// 					tx.HonorFinal = 0
// 				}
// 			}
// 			// log.Printf("Txn %s dokter %s honor master = %v", tx.TxnDesc, tx.TxnDoctor, tx.HonorMaster)
// 		}
// 		// ENT Clinic cek untuk hanya memberikan honor visit max 50.000
// 		for _, tx := range list {
// 			if tx.TxnCategory == "ENT CLINIC" && tx.Inacbg > 200000 {

// 			}
// 		}
// 		// Jika ada 2 tindakan oleh dokter yang sama → hanya bayar 1 (net tertinggi)
// 		tindakanByDoctor := make(map[string][]*Honorfull)
// 		// Kumpulkan semua tindakan per dokter
// 		for i := range list {
// 			tx := &list[i]
// 			if tx.PreviousTxnType == "tindakan" && tx.HonorMaster != 0 {
// 				tindakanByDoctor[tx.TxnDoctor] = append(tindakanByDoctor[tx.TxnDoctor], tx)
// 			}
// 		}

// 		// Cek per dokter
// 		for _, tindakanList := range tindakanByDoctor {
// 			if len(tindakanList) > 1 {
// 				// Cari tindakan dengan net_amount tertinggi
// 				maxIdx := 0
// 				maxNet := tindakanList[0].NetAmount
// 				for i, t := range tindakanList {
// 					if t.NetAmount > maxNet {
// 						maxNet = t.NetAmount
// 						maxIdx = i
// 					}
// 				}

// 				// Nolkan semua tindakan kecuali yang tertinggi
// 				for i, t := range tindakanList {
// 					if i != maxIdx {
// 						// log.Printf("❌ Nolkan tindakan dokter %s pada visit %s (lebih dari 1 tindakan, hanya ambil net tertinggi %.2f)", doctor, t.VisitNo, maxNet)
// 						t.HonorMaster = 0
// 						t.HonorFinal = 0
// 					}
// 				}
// 			}
// 		}

// 		// ✅ Step 4: Hitung totalMaster (skip fix)
// 		totalMaster := 0.0
// 		for _, tx := range list {
// 			if tx.TxnType == "fix" {
// 				limit -= tx.HonorMaster //pengurangan limit dengan honor fix
// 				continue
// 			}
// 			totalMaster += tx.HonorMaster
// 			// log.Printf("total master %v", totalMaster)
// 			// log.Printf(" limit : %v", limit)

// 		}

// 		if totalMaster == 0 {
// 			grouped[visitNo] = list
// 			continue
// 		}

// 		// ✅ Step 5: Proporsional scaling jika melebihi batas
// 		scale := 1.0
// 		if totalMaster > limit {
// 			// log.Printf("total honor %v melebihi limit %v ", totalMaster, limit)
// 			scale = limit / totalMaster
// 			// log.Printf("visit number %s Scale:%v", visitNo, scale)
// 		}

// 		for i := range list {
// 			tx := &list[i]
// 			if tx.TxnType == "fix" {
// 				tx.HonorFinal = tx.HonorMaster
// 				continue
// 			}
// 			tx.HonorProp = tx.HonorMaster * scale
// 			tx.HonorFinal = tx.HonorProp
// 		}

// 		// ✅ Step 8: Cek ulang jika proporsional honor visit <50% honor master
// 		includeFix := false
// 		for _, tx := range list {
// 			if tx.TxnType == "visit" && tx.HonorProp < 0.5*tx.HonorMaster {
// 				includeFix = true
// 				// log.Printf("honor proporsional dari %s %v <0.5 honor master %v", tx.TxnDesc, tx.HonorProp, tx.HonorMaster)
// 				// log.Println("proporsional honor visit melebihi 1/2 dari master, honor fix ikut di proporsionalkan")
// 				break
// 			}
// 		}
// 		//kembalikan nominal limit yang sebelumnya sudah dikurangi
// 		// log.Printf("Limit sebelumnya : %v", limit)
// 		for _, tx := range list {
// 			if tx.TxnType == "fix" {
// 				limit += tx.HonorMaster
// 				// log.Printf("Limit : %v", limit)
// 			}
// 		}
// 		//perhitungan proporsional kembali include txn type fix
// 		if includeFix {
// 			totalMasterbaru := 0.0
// 			for _, tx := range list {
// 				totalMasterbaru += tx.HonorMaster
// 				// log.Printf("%v + honor master %v", totalMasterbaru, tx.HonorMaster)
// 			}
// 			// log.Printf("Limit akhir %v", limit)
// 			newScale := limit / totalMasterbaru
// 			// log.Printf("total master baru: %v scale baru: %v", totalMasterbaru, newScale)
// 			for i := range list {
// 				tx := &list[i]
// 				tx.HonorProp = tx.HonorMaster * newScale
// 				tx.HonorFinal = tx.HonorProp
// 				// log.Printf("Honor Final : %v", tx.HonorFinal)
// 			}

// 		}

// 		// ✅ Step 8: Final check — bulatkan & pastikan tidak melebihi limit total
// 		totalFinal := 0.0
// 		for _, tx := range list {
// 			totalFinal += tx.HonorFinal
// 		}

// 		// Jika total melebihi limit (karena pembulatan bisa bikin sedikit lewat)
// 		if totalFinal > limit {
// 			// Hitung ulang scale kecil agar totalFinal == limit
// 			scaleAdjust := limit / totalFinal
// 			for i := range list {
// 				list[i].HonorFinal = math.Floor(list[i].HonorFinal * scaleAdjust)
// 			}
// 		}

// 		// Bulatkan semua nilai final agar tidak ada desimal
// 		for i := range list {
// 			list[i].HonorProp = math.Floor(list[i].HonorProp)
// 			list[i].HonorFinal = math.Floor(list[i].HonorFinal)

// 		}
// 		// ✅ Jika memang dilakukan proporsionalisasi (scale < 1) → pastikan totalFinal == limit
// 		if scale < 1.0 {
// 			totalFinal = 0
// 			for _, tx := range list {
// 				totalFinal += tx.HonorFinal
// 			}

// 			diff := math.Round(limit - totalFinal)

// 			if diff > 0 {
// 				// Kumpulkan pecahan sebelum floor (hanya untuk transaksi non-fix dengan HonorMaster > 0)
// 				type remainder struct {
// 					idx  int
// 					frac float64
// 				}
// 				var remainders []remainder

// 				for i, tx := range list {
// 					if tx.HonorMaster == 0 || tx.TxnType == "fix" {
// 						continue
// 					}
// 					frac := tx.HonorProp - math.Floor(tx.HonorProp)
// 					remainders = append(remainders, remainder{i, frac})
// 				}

// 				// Urutkan berdasarkan pecahan terbesar
// 				sort.Slice(remainders, func(i, j int) bool {
// 					return remainders[i].frac > remainders[j].frac
// 				})

// 				// Tambahkan +1 ke yang pecahannya paling besar sampai diff habis
// 				for i := 0; i < int(diff) && i < len(remainders); i++ {
// 					list[remainders[i].idx].HonorFinal += 1
// 				}
// 			}

// 			// Recalculate total untuk memastikan hasil akhir pas
// 			totalFinal = 0
// 			for _, tx := range list {
// 				totalFinal += tx.HonorFinal
// 				// log.Printf("txn %s honor final = %v", tx.TxnDesc, tx.HonorFinal)
// 			}
// 			// log.Printf("✅ Visit %s: total honor proporsional akhir = %.0f (limit = %.0f)", visitNo, totalFinal, limit)
// 		} else {
// 			// Non-proporsional (totalMaster <= limit) — tidak perlu redistribusi
// 			totalFinal = 0
// 			for _, tx := range list {
// 				totalFinal += tx.HonorFinal
// 			}
// 			// log.Printf("🟢 Visit %s: tidak proporsional (total %.0f <= limit %.0f)", visitNo, totalFinal, limit)
// 		}

// 		grouped[visitNo] = list
// 		result = grouped[visitNo]
// 	}

// 	return result

// }

// func ProcessHonor(db *sql.DB) error {
// 	// 1. Load data
// 	items, err := LoadPatientBillUpdate(db)
// 	if err != nil {
// 		return err
// 	}

// 	// 2. Hitung honor
// 	processed := TesHonorCount(db, items)
// 	if err != nil {
// 		return err
// 	}

// 	// 4. Simpan hasil perhitungan balik ke database
// 	return SaveToPatientBill(db, processed,)
// }

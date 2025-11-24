package processor

import (
	"database/sql"
	"fmt"
	"io"
	"log"
	"strings"

	"github.com/mtmh3135/honor/backend/config"
	"github.com/xuri/excelize/v2"
)

func ProcessXLSX(r io.Reader) error {
	log.Println("Start processing")

	f, err := excelize.OpenReader(r)
	if err != nil {
		return err
	}
	defer f.Close()

	sheets := f.GetSheetList()
	log.Println("Sheets:", sheets)
	sheet := sheets[0]
	rows, err := f.Rows(sheet)
	if err != nil {
		return err
	}
	defer rows.Close()

	// Skip first row if it's title
	rows.Next()

	// Header
	headers, _ := rows.Columns()
	log.Println("Headers:", headers)

	// Validasi header
	expectedHeaders := []string{"patient name", "card no", "patient type", "visit no", "regn dept", "ward desc", "patient class", "txn category", "txn code", "gl account", "txn desc", "careprovider txn doctor id", "txn doctor", "regn doctor", "ref doctor", "base price", "qty", "txn amount", "margin amount", "claim amount", "discount visit", "net amount", "bill datetime", "bill status", "organisation name", "admission date time", "discharge date time"}
	if err := validateHeader(headers, expectedHeaders); err != nil {
		return fmt.Errorf("header pada file yang di upload tidak sesuai\nheader: %v \nexpected:%v", headers, expectedHeaders)
	}

	// Batch insert
	batch := make([][]string, 0, 5000)

	processed := 0
	skipped := 0
	for rows.Next() {
		values, _ := rows.Columns()
		record := mapRow(headers, values)
		txn_code := record["Txn Code"]

		if config.Mastertxn[txn_code] {
			processed++
			batch = append(batch, values)
			if len(batch) >= 5000 {
				insertBatchDynamic(headers, batch)
				batch = batch[:0]
			}
		}
	}
	if len(batch) > 0 {
		insertBatchDynamic(headers, batch)
	}

	log.Printf("Processing done \n processed: %d\n skipped: %d", processed, skipped)
	return nil
}
func validateHeader(actual, expected []string) error {
	if len(actual) < len(expected) {
		return fmt.Errorf("header tidak lengkap: ditemukan %d kolom, seharusnya %d", len(actual), len(expected))
	}
	for i, exp := range expected {
		act := strings.TrimSpace(strings.ToLower(actual[i]))
		if act != strings.ToLower(exp) {
			return fmt.Errorf("kolom header ke-%d tidak sesuai, diharapkan '%s', ditemukan '%s'", i+1, exp, act)
		}
	}
	return nil
}
func mapRow(headers, values []string) map[string]string {
	m := make(map[string]string)
	for i, h := range headers {
		if i < len(values) {
			m[h] = values[i]
		}
	}
	return m
}
func toSnakeCase(s string) string {
	return strings.ToLower(strings.ReplaceAll(s, " ", "_"))
}

func insertBatchDynamic(headers []string, batch [][]string) {
	// Convert headers to snake_case assuming DB columns are snake_case
	quotedHeaders := make([]string, len(headers))
	for i, h := range headers {
		quotedHeaders[i] = "`" + toSnakeCase(h) + "`"
	}

	// Buat placeholder: (?, ?, ?, …)
	placeholders := make([]string, len(headers))
	for i := range headers {
		placeholders[i] = "?"
	}

	// === Ambil semua visit_no unik dari batch ===
	visitNoIndex := -1
	for i, h := range headers {
		if strings.ToLower(strings.TrimSpace(h)) == "visit no" {
			visitNoIndex = i
			break
		}
	}

	if visitNoIndex == -1 {
		log.Println("Kolom 'visit no' tidak ditemukan di header")
		return
	}

	visitNos := make(map[string]bool)
	for _, row := range batch {
		if visitNoIndex < len(row) {
			v := strings.TrimSpace(row[visitNoIndex])
			if v != "" {
				visitNos[v] = true
			}
		}
	}

	// === Hapus data lama berdasarkan visit_no ===

	for visit := range visitNos {
		tx, err := config.DB.Begin()
		if err != nil {
			log.Println("Tx begin error:", err)
			return
		}
		var status sql.NullString
		err = tx.QueryRow("SELECT honor_status FROM patient_bill WHERE visit_no = ? LIMIT 1", visit).Scan(&status)
		// log.Println("status", status)
		if err != nil && err != sql.ErrNoRows {
			log.Println("Error cek status visit_no", visit, ":", err)
			tx.Rollback()
			return
		}

		switch {
		case strings.ToUpper(status.String) == "FIX" || strings.ToUpper(status.String) == "ON_PROGRESS":
			// ---- FIX / ON_PROGRESS ----//
			insertQuery := fmt.Sprintf(
				"INSERT INTO patient_bill_update_billing (%s) VALUES (%s)",
				strings.Join(quotedHeaders, ", "),
				strings.Join(placeholders, ", "),
			)

			stmtUpdate, err := tx.Prepare(insertQuery)
			if err != nil {
				tx.Rollback()
				return
			}

			insertedUpdate := 0
			for _, row := range batch {
				if visitNoIndex < len(row) && strings.TrimSpace(row[visitNoIndex]) == visit {
					args := make([]interface{}, len(headers))
					for i := range headers {
						if i < len(row) {
							args[i] = row[i]
						} else {
							args[i] = nil
						}
					}
					if _, err := stmtUpdate.Exec(args...); err != nil {
						insertedUpdate++
						log.Println("Insert UPDATE_BILLING error:", err)
						tx.Rollback()
						return
					}
				}
			}
			tx.Commit()
			continue

		case !status.Valid || strings.ToUpper(status.String) == "COUNTED":
			// ---- COUNTED ATAU NULL ----
			_, err := tx.Exec("DELETE FROM patient_bill WHERE visit_no = ?", visit)
			if err != nil {
				log.Println("DELETE error:", err)
				tx.Rollback()
				return
			}
			// Query dinamis
			query := fmt.Sprintf(
				"INSERT INTO patient_bill (%s) VALUES (%s)",
				strings.Join(quotedHeaders, ", "),
				strings.Join(placeholders, ", "),
			)

			stmt, err := tx.Prepare(query)
			if err != nil {
				log.Println("Prepare stmt error:", err)
				tx.Rollback()
				return
			}

			inserted := 0
			for _, row := range batch {
				if visitNoIndex < len(row) && strings.TrimSpace(row[visitNoIndex]) == visit {
					// Jika row lebih pendek dari header, pad dengan nil
					args := make([]interface{}, len(headers))
					for i := range headers {
						if i < len(row) {
							args[i] = row[i]
						} else {
							args[i] = nil
						}
					}
					_, err := stmt.Exec(args...)
					if err != nil {
						log.Println("Insert error:", err)
					} else {
						inserted++
					}
				}
			}

			if err := tx.Commit(); err != nil {
				log.Println("Commit error:", err)
			} else {
				// log.Printf("Committed %d rows", inserted)
			}
		}

	}

}

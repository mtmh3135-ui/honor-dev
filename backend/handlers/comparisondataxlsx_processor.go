package handlers

import (
	"fmt"
	"io"
	"log"
	"strings"

	"github.com/mtmh3135/honor/backend/config"
	"github.com/xuri/excelize/v2"
)

func ComparisonDataUp(r io.Reader) error {
	log.Println("Start processing")

	f, err := excelize.OpenReader(r)
	if err != nil {
		return fmt.Errorf("gagal membaca file excel: %v", err)
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
	rowIndex := 0
	// Skip first row if it's title
	rows.Next()
	rowIndex++
	// Header

	headers, _ := rows.Columns()
	log.Println("Headers:", headers)

	//validasi header
	expectedHeaders := []string{"visit number", "status", "visit number tujuan", "tarif ina cbg", "kelas bpjs", "tarif sebelum topup"}
	if err := validateHeaders(headers, expectedHeaders); err != nil {
		return fmt.Errorf("header pada file yang di upload tidak sesuai\nheader: %v", headers)
	}
	headerCount := len(headers)
	headers = append(headers, "tkr_status")

	// Batch insert
	batch := make([][]string, 0, 1000)
	uniqueKeys := []string{
		"visit_number",
	}

	processed := 0

	for rows.Next() {
		rowIndex++
		values, _ := rows.Columns()
		processed++

		if len(values) < headerCount {
			// pad dengan string kosong sampai sama panjang
			for len(values) < headerCount {
				values = append(values, "")
			}
		} else if len(values) > headerCount {
			// trim jika ada kolom berlebih (misal ada trailing accidental columns)
			values = values[:headerCount]
		}

		if len(values) < len(expectedHeaders) {
			return fmt.Errorf("baris %d memiliki kolom kurang dari %d", processed+2, len(expectedHeaders))
		}
		if values[0] == "" {
			return fmt.Errorf("baris %d tidak memiliki visit_number", processed+2)
		}

		// Cek TKR
		IsTkr := 0
		if IsTKR(f, sheet, rowIndex) { // processed+1 karena row index di Excel mulai dari 2 (header di row 1)
			IsTkr = 1
			log.Println("visit number ini bedah TKR")
		}

		values = append(values, fmt.Sprintf("%d", IsTkr))
		batch = append(batch, values)
		if len(batch) >= 1000 {
			insertBatchDynamics(headers, batch, uniqueKeys)
			batch = batch[:0]
		}

	}

	if len(batch) > 0 {
		insertBatchDynamics(headers, batch, uniqueKeys)

	}
	log.Printf("Processing done\n processed:%d", processed)

	return nil
}
func validateHeaders(actual, expected []string) error {
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
func ToSnakeCases(s string) string {
	return strings.ToLower(strings.ReplaceAll(s, " ", "_"))
}

func insertBatchDynamics(headers []string, batch [][]string, uniqueKeys []string) {
	// Convert headers to snake_case assuming DB columns are snake_case
	quotedHeaders := make([]string, len(headers))
	for i, h := range headers {
		quotedHeaders[i] = "`" + ToSnakeCases(h) + "`"

	}

	// Buat placeholder: (?, ?, ?, …)
	placeholders := make([]string, len(headers))
	for i := range headers {
		placeholders[i] = "?"
	}

	// --- Build bagian ON DUPLICATE KEY UPDATE ---
	updateParts := []string{}
	uniqueKeySet := make(map[string]bool)
	for _, k := range uniqueKeys {
		uniqueKeySet[ToSnakeCases(k)] = true
	}

	for _, h := range headers {
		col := ToSnakeCases(h)
		if !uniqueKeySet[col] { // kolom unik jangan ikut diupdate
			updateParts = append(updateParts, fmt.Sprintf("`%s` = VALUES(`%s`)", col, col))
		}
	}

	onDuplicate := ""
	if len(updateParts) > 0 {
		onDuplicate = " ON DUPLICATE KEY UPDATE " + strings.Join(updateParts, ", ")
	}

	// Query dinamis
	query := fmt.Sprintf(
		"INSERT INTO comparison_data (%s) VALUES (%s) %s",
		strings.Join(quotedHeaders, ", "),
		strings.Join(placeholders, ", "),
		onDuplicate,
	)
	log.Println("Query:", query)

	tx, err := config.DB.Begin()
	if err != nil {
		log.Println("Gagal Mulai Transaksi:", err)
		return
	}
	stmt, err := tx.Prepare(query)
	if err != nil {
		log.Println("Gagal Mulai Statement:", err)
		tx.Rollback()
		return
	}
	defer stmt.Close()

	inserted := 0
	for _, row := range batch {
		// Jika row lebih pendek dari header, pad dengan nil
		args := make([]interface{}, len(headers))
		for i := range headers {
			if i < len(row) {
				v := strings.TrimSpace(row[i])
				// jika kolom kosong → jadikan nil supaya MySQL menerima NULL
				if v == "" {
					args[i] = nil
				} else {
					args[i] = v
				}
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

	if err := tx.Commit(); err != nil {
		log.Println("Commit error:", err)
	} else {
		log.Printf("Committed %d rows", inserted)
	}
}

func HasBackgroundFill(f *excelize.File, sheet, cell string) (bool, string) {
	styleID, err := f.GetCellStyle(sheet, cell)
	if err != nil {
		return false, ""
	}

	style, err := f.GetStyle(styleID)
	if err != nil {
		return false, ""
	}

	fill := style.Fill

	// Jika tidak ada pola warna → tidak ada warna
	if fill.Pattern == 0 && len(fill.Color) == 0 {
		return false, ""
	}

	// Jika Excel mengembalikan banyak warna, ambil yang pertama saja
	if len(fill.Color) > 0 {
		color := strings.ToUpper(fill.Color[0])
		color = strings.TrimPrefix(color, "FF") // normalisasi

		return true, color
	}

	return false, ""
}

func IsBlueColor(hex string) bool {
	hex = strings.ToUpper(strings.ReplaceAll(hex, "FF", "")) // normalisasi

	bluePallet := []string{
		"0070C0",
		"5B9BD5",
		"2F5597",
		"0000FF",
	}

	for _, b := range bluePallet {
		if strings.Contains(hex, b) {
			return true
		}
	}
	return false
}

func IsTKR(f *excelize.File, sheet string, rowIndex int) bool {
	// cek warna hanya pada kolom A
	cell, _ := excelize.CoordinatesToCellName(1, rowIndex)
	hasFill, color := HasBackgroundFill(f, sheet, cell)
	if !hasFill {
		// log.Println("Cell ini tidak memiliki warna background")
		return false
	}

	return IsBlueColor(color)
}

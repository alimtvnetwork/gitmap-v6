package formatter

import (
	"encoding/csv"
	"io"

	"github.com/user/gitmap/constants"
	"github.com/user/gitmap/model"
)

// WriteCSV writes records to the given writer in CSV format.
func WriteCSV(w io.Writer, records []model.ScanRecord) error {
	cw := csv.NewWriter(w)
	err := cw.Write(constants.ScanCSVHeaders)
	if err != nil {
		return err
	}

	return writeCSVRows(cw, records)
}

// writeCSVRows writes each record as a CSV row and flushes.
func writeCSVRows(cw *csv.Writer, records []model.ScanRecord) error {
	for _, r := range records {
		err := writeCSVRow(cw, r)
		if err != nil {
			return err
		}
	}
	cw.Flush()

	return cw.Error()
}

// writeCSVRow converts a single record to a CSV row.
func writeCSVRow(cw *csv.Writer, r model.ScanRecord) error {
	row := []string{
		r.RepoName, r.HTTPSUrl, r.SSHUrl, r.Branch,
		r.RelativePath, r.AbsolutePath, r.CloneInstruction, r.Notes,
	}

	return cw.Write(row)
}

// ParseCSV reads records from a CSV reader.
func ParseCSV(reader io.Reader) ([]model.ScanRecord, error) {
	cr := csv.NewReader(reader)
	rows, err := cr.ReadAll()
	if err != nil {
		return nil, err
	}

	return parseCSVRows(rows), nil
}

// parseCSVRows converts raw CSV rows (skipping header) into records.
func parseCSVRows(rows [][]string) []model.ScanRecord {
	records := make([]model.ScanRecord, 0, len(rows))
	for i, row := range rows {
		if i == 0 {
			continue // skip header
		}
		if len(row) >= 8 {
			records = append(records, rowToRecord(row))
		}
	}

	return records
}

// rowToRecord maps a CSV row to a ScanRecord.
func rowToRecord(row []string) model.ScanRecord {
	notes := ""
	if len(row) > 7 {
		notes = row[7]
	}

	return model.ScanRecord{
		RepoName: row[0], HTTPSUrl: row[1], SSHUrl: row[2],
		Branch: row[3], RelativePath: row[4], AbsolutePath: row[5],
		CloneInstruction: row[6], Notes: notes,
	}
}

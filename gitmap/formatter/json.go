package formatter

import (
	"encoding/json"
	"io"

	"github.com/user/gitmap/constants"
	"github.com/user/gitmap/model"
)

// WriteJSON writes records to the given writer as a JSON array.
func WriteJSON(w io.Writer, records []model.ScanRecord) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", constants.JSONIndent)

	return enc.Encode(records)
}

// ParseJSON reads records from a JSON reader.
func ParseJSON(reader io.Reader) ([]model.ScanRecord, error) {
	var records []model.ScanRecord
	dec := json.NewDecoder(reader)
	err := dec.Decode(&records)
	if err != nil {
		return nil, err
	}

	return records, nil
}

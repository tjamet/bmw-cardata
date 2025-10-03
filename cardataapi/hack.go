package cardataapi

import (
	"encoding/json"
	"time"
)

// Hack around time de-serialization as the server returns different formats
func (e *BusinessErrorDto) UnmarshalJSON(data []byte) error {
	unstructured := map[string]string{}
	if err := json.Unmarshal(data, &unstructured); err != nil {
		return err
	}

	if val, ok := unstructured["hint"]; ok {
		e.Hint = &val
	}

	if val, ok := unstructured["creationTime"]; ok {
		for _, format := range []string{"2006-01-02T15:04:05", "2006-01-02T15:04:05.000Z", "2006-01-02T15:04:05Z", "2006-01-02T15:04:05Z07:00"} {
			if t, err := time.Parse(format, val); err == nil {
				e.CreationTime = &t
				break
			}
		}
	}

	return nil
}

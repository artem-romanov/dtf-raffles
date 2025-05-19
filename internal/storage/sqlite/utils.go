package sqlite

import "time"

func ToDbTime(t time.Time) string {
	return t.UTC().Format(time.RFC3339)
}

func FromDbTime(t string) (time.Time, error) {
	result, err := time.Parse(time.RFC3339, t)
	if err != nil {
		return time.Time{}, err
	}

	return result, nil
}

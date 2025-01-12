package utils

import (
	"encoding/json"
	"fmt"
	"log"
	"time"
)

func DateFormat(date string, format string) string {
	// validate is correct date
	const dateFormat = "2006-01-02"
	parsedDate, err := time.Parse(dateFormat, date)
	if err != nil {
		log.Println("Error parsing date:", err)
		return ""
	}

	return parsedDate.Format(format)
}

func CalculateAge(date string) (int, error) {
	birthDate, err := time.Parse("2006-01-02", date)
	if err != nil {
		return 0, fmt.Errorf("invalid date format: %v", err)
	}

	currentDate := time.Now()
	age := currentDate.Year() - birthDate.Year()
	if currentDate.YearDay() < birthDate.YearDay() {
		age--
	}

	return age, nil
}

func IsJson(str string) bool {
	var js map[string]interface{}
	return json.Unmarshal([]byte(str), &js) == nil
}

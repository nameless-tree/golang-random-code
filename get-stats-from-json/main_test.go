package main

import (
	"get-stats-from-json/jsonparse"
	"os"
	"testing"
)

func createAndWriteToTemp(t *testing.T, data string) *os.File {
	// Create a temporary file
	tmpfile, err := os.CreateTemp("", "example")
	if err != nil {
		t.Fatal(err)
	}

	// Write JSON data to the temporary file
	if _, err := tmpfile.Write([]byte(data)); err != nil {
		t.Fatal(err)
	}
	if err := tmpfile.Close(); err != nil {
		t.Fatal(err)
	}

	file, err := os.Open(tmpfile.Name())
	if err != nil {
		t.Fatal(err)
	}

	return file
}

func TestMatchedRecipes(t *testing.T) {
	// JSON data as a string
	jsonData := `
	[
		{
			"postcode": "1",
			"recipe": "A",
			"delivery": "Wednesday 1AM - 7PM"
		},
		{
			"postcode": "2",
			"recipe": "A",
			"delivery": "Thursday 7AM - 5PM"
		},
		{
			"postcode": "3",
			"recipe": "B",
			"delivery": "Thursday 7AM - 9PM"
		},
		{
			"postcode": "4",
			"recipe": "B",
			"delivery": "Saturday 2AM - 8PM"
		},
		{
			"postcode": "4",
			"recipe": "C",
			"delivery": "Wednesday 7AM - 4PM"
		},
		{
			"postcode": "4",
			"recipe": "A",
			"delivery": "Wednesday 2AM - 5PM"
		},
		{
			"postcode": "4",
			"recipe": "A",
			"delivery": "Wednesday 12PM - 5PM"
		}
	]
	`

	var p = jsonparse.JsonPostcodeCountPerTimeRange{
		Name:  "4",
		Count: 0,
		From:  "2AM",
		To:    "5PM",
	}

	file := createAndWriteToTemp(t, jsonData)
	defer os.Remove(file.Name())
	defer file.Close()

	var r = new(jsonparse.JsonRawData)
	r.Init(file, RECIPES_MATCH[:], p)

	parse(r)

}

package main

import (
	"fmt"
	"os"

	"get-stats-from-json/jsonparse"
)

var (
	// need to count number of deliveries to this postcode by the time
	POSTCODE_MATCH           = "10120"
	POSTCODE_MATCH_TIME_FROM = "10AM"
	POSTCODE_MATCH_TIME_TO   = "3PM"

	// recipe names patterns
	RECIPES_MATCH = [...]string{"Potato", "Veggie", "Mushroom"}
)

var (
	DIST_RECIPES   = 2000
	DIST_POSTCODES = 1000000
)

func main() {
	file, err := os.Open("test.json")

	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	defer file.Close()

	var p = jsonparse.JsonPostcodeCountPerTimeRange{
		Name:  POSTCODE_MATCH,
		Count: 0,
		From:  POSTCODE_MATCH_TIME_FROM,
		To:    POSTCODE_MATCH_TIME_TO,
	}

	var r = new(jsonparse.JsonRawData)
	r.Init(file, RECIPES_MATCH[:], p)

	parse(r)
}

func parse(r *jsonparse.JsonRawData) {
	if err := r.Branches(); err != nil {
		return
	}

	for r.Decoder.More() {
		var item jsonparse.JsonItem

		if err := r.Decoder.Decode(&item); err != nil {
			return
		}

		r.AddRecipe(&item)
		r.AddPostcode(&item)
	}

	if err := r.Branches(); err != nil {
		return
	}

	jsonstats, err := r.Results()
	if err != nil {
		return
	}

	fmt.Println(string(jsonstats))
}

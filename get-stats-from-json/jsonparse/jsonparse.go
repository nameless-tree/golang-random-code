package jsonparse

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"get-stats-from-json/extra"
)

var (
	// The number of distinct recipe names is lower than `2K`, one recipe name is not longer than `100` chars.
	DIST_RECIPES = 2000

	// The number of distinct postcodes is lower than `1M`, one postcode is not longer than `10` chars.
	DIST_POSTCODES = 1000000
)

type JsonItem struct {
	Postcode string `json:"postcode"`
	Recipe   string `json:"recipe"`
	Delivery string `json:"delivery"`
}

type JsonRawData struct {
	Decoder *json.Decoder

	Recipes   map[string]int // unique recipes with total count
	Postcodes map[string]int // unique postcodes with total count

	postcodeBusiest           JsonPostcodeCount             // postcode with most delivered recipes.
	postcodeCountPerTimeRange JsonPostcodeCountPerTimeRange // postcode that need to count the number of deliveries in timerange
	postcodeTimeRange         extra.TimeRange               // given timerange

	recipesMatched      map[string]bool // unique recipe names that contain in their name one of the matching words
	recipeMatchPatterns []string        // given words
}

func (r *JsonRawData) Init(file *os.File, matches []string, p JsonPostcodeCountPerTimeRange) {
	// new decoder to a given file
	r.Decoder = json.NewDecoder(file)

	r.Postcodes = make(map[string]int, DIST_POSTCODES)
	r.Recipes = make(map[string]int, DIST_RECIPES)
	r.recipesMatched = make(map[string]bool, DIST_RECIPES)

	r.recipeMatchPatterns = matches
	r.postcodeCountPerTimeRange = p

	start, _ := extra.Time12Parse(p.From)
	end, _ := extra.Time12Parse(p.To)

	r.postcodeTimeRange = extra.TimeRange{
		Start: start,
		End:   end,
	}
}

// skip [, ], {, }
func (r *JsonRawData) Branches() error {
	if _, err := r.Decoder.Token(); err != nil {
		return err
	}
	return nil
}

// add unique recipe
func (r *JsonRawData) AddRecipe(item *JsonItem) {
	if val, ok := r.Recipes[item.Recipe]; ok {
		r.Recipes[item.Recipe] = val + 1
	} else {
		r.Recipes[item.Recipe] = 1
		r.updateMatchRecipes(item)
	}

}

// add unique postcode and update for busiest one
func (r *JsonRawData) AddPostcode(item *JsonItem) {
	if val, ok := r.Postcodes[item.Postcode]; ok {
		r.Postcodes[item.Postcode] = val + 1
		r.updateBusiestPostcode(item.Postcode, val+1)
	} else {
		r.Postcodes[item.Postcode] = 1
		r.updateBusiestPostcode(item.Postcode, 1)
	}

	r.updateTimeRange(item)
}

// update busiest postcode
func (r *JsonRawData) updateBusiestPostcode(name string, count int) {
	if r.postcodeBusiest.Count < count {
		r.postcodeBusiest.Count = count
		r.postcodeBusiest.Name = name
	}
}

// add unique recipe that matches
func (r *JsonRawData) updateMatchRecipes(item *JsonItem) {
	for _, sub := range r.recipeMatchPatterns {
		if strings.Contains(item.Recipe, sub) {
			if _, ok := r.recipesMatched[item.Recipe]; !ok {
				r.recipesMatched[item.Recipe] = true
			}
			break
		}
	}
}

func (r *JsonRawData) updateTimeRange(item *JsonItem) {
	if r.postcodeCountPerTimeRange.Name == item.Postcode {

		timerange, err := deliveryToTime(item.Delivery)

		if err != nil {
			return
		}

		if extra.IsTimeRangeWithin(timerange, r.postcodeTimeRange) {
			r.postcodeCountPerTimeRange.Count += 1
		}
	}

}

func (r *JsonRawData) Results() ([]byte, error) {
	var j JsonStatsData

	return j.results(r)
}

// format:
// Wednesday 8AM - 4PM
func deliveryToTime(str string) (extra.TimeRange, error) {
	parts := strings.Split(str, " ")

	if len(parts) != 4 {
		return extra.TimeRange{}, fmt.Errorf("invalid format")
	}

	start, err := extra.Time12Parse(parts[1])
	if err != nil {
		return extra.TimeRange{}, err
	}

	end, err := extra.Time12Parse(parts[3])
	if err != nil {
		return extra.TimeRange{}, err
	}

	return extra.TimeRange{Start: start, End: end}, nil
}

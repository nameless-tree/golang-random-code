package jsonparse

import (
	"encoding/json"
	"sort"
	"sync"
)

type JsonRecipeCount struct {
	Name  string `json:"recipe"`
	Count int    `json:"count"`
}

type JsonPostcodeCount struct {
	Name  string `json:"postcode"`
	Count int    `json:"delivery_count"`
}

type JsonPostcodeCountPerTimeRange struct {
	Name  string `json:"postcode"`
	From  string `json:"from"`
	To    string `json:"to"`
	Count int    `json:"delivery_count"`
}

type JsonStatsData struct {
	UniqueRecipeCount         int                           `json:"unique_recipe_count"`
	CountPerRecipe            []JsonRecipeCount             `json:"count_per_recipe"`
	PostcodeBusiest           JsonPostcodeCount             `json:"busiest_postcode"`
	PostcodeCountPerTimeRange JsonPostcodeCountPerTimeRange `json:"count_per_postcode_and_time"`
	MatchByName               []string                      `json:"match_by_name"`
}

func (j *JsonStatsData) results(r *JsonRawData) ([]byte, error) {
	var wg sync.WaitGroup

	// Sort distinct recipes
	wg.Add(1)
	go func() {
		j.CountPerRecipe = make([]JsonRecipeCount, 0, 2000)
		for val, key := range r.Recipes {
			j.CountPerRecipe = append(j.CountPerRecipe, JsonRecipeCount{Name: val, Count: key})
		}

		sort.Slice(j.CountPerRecipe, func(p, q int) bool {
			return j.CountPerRecipe[p].Name < j.CountPerRecipe[q].Name
		})

		wg.Done()

	}()

	// Sort distinct recipes that contain pattern words
	wg.Add(1)
	go func() {
		j.MatchByName = make([]string, 0, 2000)
		for key := range r.recipesMatched {
			j.MatchByName = append(j.MatchByName, key)
		}

		sort.Strings(j.MatchByName)

		wg.Done()
	}()

	// Distinct recipe count
	j.UniqueRecipeCount = len(r.Recipes)

	// Postcode with most delivered recipes
	j.PostcodeBusiest = r.postcodeBusiest

	// Count the number of deliveries to given postcode that lie within the delivery time
	j.PostcodeCountPerTimeRange = r.postcodeCountPerTimeRange

	wg.Wait()

	return json.MarshalIndent(j, "", "    ")
}

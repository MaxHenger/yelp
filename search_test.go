package yelp

import (
	"testing"
)

func TestArgumentRepetition(t *testing.T) {
	//create a list of possible arguments
	listPosition := []SearchQuerier{SearchLocation("Delft"),
		SearchLocationCoordinates{"Delft", 0, 0},
		SearchCoordinates{0, 0},
		SearchBounds{0, 0, 1, 1}}
	listRemaining := []SearchQuerier{SearchTerms([]string{"bar"}),
		SearchLimit(1),
		SearchOffset(1),
		SearchSort(SearchSortDistance),
		SearchCategories([]SearchCategory{SearchCategoryBars}),
		SearchRadius(20000),
		SearchDeals(false)}

	//first test all possible combinations of positions
	for _, v := range listPosition {
		for _, w := range listPosition {
			var q SearchQuery
			err := v.Query(&q)
			if err != nil {
				t.Errorf("Expected position first call with '%v' to succeed", v)
			}

			err = w.Query(&q)
			if err == nil {
				t.Errorf("Expected position second call with '%v' to fail", w)
			}
		}
	}

	//test all possible combinations of the remaining elements
	for i, v := range listRemaining {
		for j, w := range listRemaining {
			var q SearchQuery
			err := v.Query(&q)

			if err != nil {
				t.Errorf("Expected remaining first call with '%v' to succeed", v)
			}

			err = w.Query(&q)

			switch {
			case i == j && err == nil:
				t.Errorf("Expected two similar search options ('%v' and '%v') to fail", v, w)
			case i != j && err != nil:
				t.Errorf("Expected two dissimilar search options ('%v' and '%v') to succeed", v, w)
			}
		}
	}
}

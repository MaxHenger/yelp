package yelp

import (
	"bytes"
	"fmt"
	"sort"
	"strconv"
	"strings"
)

//The searchIdentifierXXX terms constants are the names of the query elements that
//can be used in the Yelp API to specify what kind of businesses are requested
//from the API
const (
	searchIdentifierTerm            = "term"
	searchIdentifierLimit           = "limit"
	searchIdentifierOffset          = "offset"
	searchIdentifierSort            = "sort"
	searchIdentifierCategory        = "category_filter"
	searchIdentifierRadius          = "radius_filter"
	searchIdentifierDeals           = "deals_filter"
	searchIdentifierLocation        = "location"
	searchIdentifierCoordinates     = "ll"
	searchIdentifierCoordinatesHint = "cll"
	searchIdentifierBounds          = "bounds"
)

//The Yelp query bitmask. This bitmask is used when asking the client to perform
//a search query on the basis of specified options to make sure options do not
//appear twice in the total query.
type searchBitMask uint8

//The searchBitMaskXXX terms constants are the binary masks that are used by the
//SearchQuery structure to keep track of which query elements have already been
//added to it when the search options are added to the SearchQuery class by
//implementations of the SearchQuerier interface.
const (
	//The bitmask values
	searchBitMaskTerm searchBitMask = 1 << iota
	searchBitMaskLimit
	searchBitMaskOffset
	searchBitMaskSort
	searchBitMaskCategory
	searchBitMaskRadius
	searchBitMaskDeals
	searchBitMaskLocation
	//Note: 8 values are specified, when this list is extended please update the
	//searchBitMask to use a larger number of bits
)

//The searchQueryElement represents an element in a SearchQuery. It contains a
//name and a value
type searchQueryElement struct {
	Name  string
	Value string
}

//The SearchQuery structure contains a list of search query elements and a bit
//mask. The bit mask is used when the Yelp client is creating a search query
//from specified options (implementing the SearchQuerier interface) with the
//purpose of not performing the same search twice
type SearchQuery struct {
	queries []searchQueryElement
	mask    searchBitMask
}

//The Sort function will sort all query elements in the SearchQuery by name. It
//does this by implementing the sort.Interface methods Len(), Less() and Swap()
func (q *SearchQuery) Sort() {
	sort.Sort(q)
}

func (q *SearchQuery) Len() int {
	return len(q.queries)
}

func (q *SearchQuery) Less(i, j int) bool {
	return q.queries[i].Name < q.queries[j].Name
}

func (q *SearchQuery) Swap(i, j int) {
	q.queries[i], q.queries[j] = q.queries[j], q.queries[i]
}

func (q *SearchQuery) String() string {
	//return all queries, each seperated by a "&"
	length := len(q.queries)

	if length == 0 {
		return ""
	}

	var buffer bytes.Buffer
	buffer.WriteString(q.queries[0].Name)
	buffer.WriteString("=")
	buffer.WriteString(q.queries[0].Value)

	for i := 1; i < length; i++ {
		buffer.WriteString("&")
		buffer.WriteString(q.queries[i].Name)
		buffer.WriteString("=")
		buffer.WriteString(q.queries[i].Value)
	}

	return buffer.String()
}

//Append simply addes a new query element, defined by its name and value, to
//the SearchQuery.
func (q *SearchQuery) Append(name, value string) {
	q.queries = append(q.queries, searchQueryElement{name, value})
}

//The SearchQuerier interface provides a method for search options to translate
//their option into a search query element.
type SearchQuerier interface {
	Query(*SearchQuery) error
}

//SearchCoordinates is a search option specifying a location in terms of
//latitude and longitude coordinates
type SearchCoordinates struct {
	Latitude  float64
	Longitude float64
}

func (sl SearchCoordinates) Query(sq *SearchQuery) error {
	//make sure the variable has not been set already
	if sq.mask&searchBitMaskLocation != 0 {
		return Error{"SearchCoordinates", "Attempting to set location for a second time"}
	}

	//check if the latitude and longitude have correct values
	if validLatitudeLongitude(sl.Latitude, sl.Longitude) == false {
		return Error{"SearchCoordinates", fmt.Sprintf("Invalid latitude and/or longitude: %f, %f", sl.Latitude, sl.Longitude)}
	}

	//add to the query
	var buffer bytes.Buffer
	buffer.WriteString(strconv.FormatFloat(sl.Latitude, 'f', -1, 64))
	buffer.WriteString(",")
	buffer.WriteString(strconv.FormatFloat(sl.Longitude, 'f', -1, 64))
	sq.Append(searchIdentifierCoordinates, buffer.String())

	//modify the mask and return
	sq.mask |= searchBitMaskLocation
	return nil
}

//SearchLocation is a search option specifying a location in terms of its name
type SearchLocation string

func (sl SearchLocation) Query(sq *SearchQuery) error {
	//make sure the location has not been set already
	if sq.mask&searchBitMaskLocation != 0 {
		return Error{"SearchLocation", "Attempting to set location for a second time"}
	}

	//ensure there are no spaces in the location name and add the result to the query
	sq.Append(searchIdentifierLocation, strings.Replace(string(sl), " ", "+", -1))

	//modify the mask and return
	sq.mask |= searchBitMaskLocation
	return nil
}

//SearchLocationCoordinates is a search option specifying a location in terms of
//both a location name as latitude and longitude coordinates. The coordinates
//are used in case the location name results in an ambiguous specification of
//location
type SearchLocationCoordinates struct {
	Location  string
	Latitude  float64
	Longitude float64
}

func (slc SearchLocationCoordinates) Query(sq *SearchQuery) error {
	//make sure the location has not been set already
	if sq.mask&searchBitMaskLocation != 0 {
		return Error{"SearchLocationCoordinates", "Attempting to set location for a second time"}
	}

	//ensure there are no spaces in the location name
	sq.Append(searchIdentifierLocation, strings.Replace(slc.Location, " ", "+", -1))

	//ensure the provided latitude and longitude are correct
	if validLatitudeLongitude(slc.Latitude, slc.Longitude) == false {
		return Error{"SearchLocationCoordinates", fmt.Sprintf("Invalid latitude and/or longitude: %f, $f", slc.Latitude, slc.Longitude)}
	}

	//convert float latitude and longitude to string
	var buffer bytes.Buffer
	buffer.WriteString(strconv.FormatFloat(slc.Latitude, 'f', -1, 64))
	buffer.WriteString(",")
	buffer.WriteString(strconv.FormatFloat(slc.Longitude, 'f', -1, 64))

	sq.Append(searchIdentifierCoordinatesHint, buffer.String())

	//modify the mask and return
	sq.mask |= searchBitMaskLocation
	return nil
}

//SearchBounds is a search option specifying a location range in terms of a
//bounding box created by two pairs of longitude and latitude coordinates
type SearchBounds struct {
	SWLatitude  float64
	SWLongitude float64
	NELatitude  float64
	NELongitude float64
}

func (sb SearchBounds) Query(sq *SearchQuery) error {
	//make sure the location has not been set already
	if sq.mask&searchBitMaskLocation != 0 {
		return Error{"SearchBounds", "Attempting to set location for a second time"}
	}

	//check the validity of the arguments
	if validLatitudeLongitude(sb.SWLatitude, sb.SWLongitude) == false {
		return Error{"SearchBounds", fmt.Sprintf("Invalid southwest latitude and/or longitude: %f, %f", sb.SWLatitude, sb.SWLongitude)}
	}

	if validLatitudeLongitude(sb.NELatitude, sb.NELongitude) == false {
		return Error{"SearchBounds", fmt.Sprintf("Invalid northeast latitude and/or longitude: %f, %f", sb.NELatitude, sb.NELongitude)}
	}

	//convert float latitudes and longitudes to the required format
	var buffer bytes.Buffer
	buffer.WriteString(strconv.FormatFloat(sb.SWLatitude, 'f', -1, 64))
	buffer.WriteString(",")
	buffer.WriteString(strconv.FormatFloat(sb.SWLongitude, 'f', -1, 64))
	buffer.WriteString("|")
	buffer.WriteString(strconv.FormatFloat(sb.NELatitude, 'f', -1, 64))
	buffer.WriteString(",")
	buffer.WriteString(strconv.FormatFloat(sb.NELongitude, 'f', -1, 64))

	//add the query
	sq.Append(searchIdentifierBounds, buffer.String())

	//modify the mask and return
	sq.mask |= searchBitMaskLocation
	return nil
}

//SearchTerms is a search option specifying yelp which terms should be included
//in the found businesses
type SearchTerms []string

func (st SearchTerms) Query(sq *SearchQuery) error {
	//make sure the search terms havent already been set
	if sq.mask&searchBitMaskTerm != 0 {
		return Error{"SearchTerms", "Attempting to set search terms a second time"}
	}

	//replace all terms with a space by a plus-sign
	for i, v := range st {
		st[i] = strings.Replace(v, " ", "+", -1)
	}

	//set the terms by joining all terms with a comma
	sq.Append(searchIdentifierTerm, strings.Join(st, ","))

	//set the mask and return
	sq.mask |= searchBitMaskTerm
	return nil
}

//SearchLimit is a search option specifying the maximum number of businesses
//that should be retrieved
type SearchLimit int

func (sl SearchLimit) Query(sq *SearchQuery) error {
	//make sure the search limit has not already been set
	if sq.mask&searchBitMaskLimit != 0 {
		return Error{"SearchLimit", "Attempting to set search limit a second time"}
	}

	//make sure the limit is valid
	if sl < 0 || sl > 20 {
		return Error{"SearchLimit", fmt.Sprintf("Invalid search limit: %d", int(sl))}
	}

	//Set the query and mask, and return
	sq.Append(searchIdentifierLimit, strconv.Itoa(int(sl)))

	sq.mask |= searchBitMaskLimit
	return nil
}

//SearchOffset is a search option specifying by which offset businesses should
//be obtained from Yelp (e.g. If one would like to retrieve businesses 6 through
//10, then set the limit to 5 and the offset to 5)
type SearchOffset int

func (so SearchOffset) Query(sq *SearchQuery) error {
	//make sure the search offset hasn't already been set
	if sq.mask&searchBitMaskOffset != 0 {
		return Error{"SearchOffset", "Attempting to set search offset a second time"}
	}

	//make sure the offset is valid
	if so < 0 {
		return Error{"SearchOffset", fmt.Sprintf("Invalid search offsets: %d", int(so))}
	}

	//set the query and mask, and return
	sq.Append(searchIdentifierOffset, strconv.Itoa(int(so)))

	sq.mask |= searchBitMaskOffset
	return nil
}

//SearchSort is a search option specifying by which sorting method the results
//from Yelp should be returned.
type SearchSort int

const (
	SearchSortBestMatched  SearchSort = 0
	SearchSortDistance                = 1
	SearchSortHighestRated            = 2
)

func (ss SearchSort) Query(sq *SearchQuery) error {
	//make sure the sorting method hasn't already been set
	if sq.mask&searchBitMaskSort != 0 {
		return Error{"SearchSort", "Attempting to set sorting method a second time"}
	}

	//make sure the sorting method is valid
	if ss < SearchSortBestMatched || ss > SearchSortHighestRated {
		return Error{"SearchSort", fmt.Sprintf("Invalid sorting method: %v", ss)}
	}

	//set the sorting method, update the mask and return
	sq.Append(searchIdentifierSort, strconv.Itoa(int(ss)))

	sq.mask |= searchBitMaskSort
	return nil
}

//SearchCategory is a typedefinition to be used in combination with the following
//go-style enumeration. It is used in the SearchCategories search option to
//specify multiple type-safe categories
type SearchCategory int

const (
	//if needed more can be added, if so, don't forget to update the constant
	//string array as well
	SearchCategoryActive SearchCategory = iota
	SearchCategoryArtsEntertainment
	SearchCategoryAutomotive
	SearchCategoryBeautySpas
	SearchCategoryGolf
	SearchCategoryNightlife
	SearchCategoryBars
	SearchCategoryRestaurants
	SearchCategoryTotal
)

var searchCategoryNames = [...]string{"active", "arts", "auto", "beautysvc",
	"golf", "nightlife", "bars", "restaurants"}

//SearchCategories is a search option to tell Yelp to only return businesses
//belonging to a certain set of categories
type SearchCategories []SearchCategory

func (sc SearchCategories) Query(sq *SearchQuery) error {
	//check if the categories aren't already set
	if sq.mask&searchBitMaskCategory != 0 {
		return Error{"SearchCategories", "Attempting to set the category filter a second time"}
	}

	//append all search categories in a single string, it should be comma-seperated
	length := len([]SearchCategory(sc))

	if length == 0 {
		//no search categories specified
		return Error{"SearchCategories", "No search categories are specified"}
	}

	var buffer bytes.Buffer
	for i, v := range sc {
		if i != 0 {
			buffer.WriteString(",")
		}

		if v < SearchCategoryActive || v >= SearchCategoryTotal {
			//invalid category specified
			return Error{"SearchCategory", "Invalid search category specified"}
		}

		buffer.WriteString(searchCategoryNames[v])
	}

	//write query, update mask and return
	sq.Append(searchIdentifierCategory, buffer.String())

	sq.mask |= searchBitMaskCategory
	return nil
}

//SearchRadius is a search option specifying the maximum range wherein businesses
//should be found by Yelp. The maximum allowed range is 40000 meters.
type SearchRadius int

func (sr SearchRadius) Query(sq *SearchQuery) error {
	//make sure the radius isnt already set
	if sq.mask&searchBitMaskRadius != 0 {
		return Error{"SearchRadius", "Attempting to set the radius filter a second time"}
	}

	//make sure the specified value is valid
	if sr < 0 || sr > 40000 {
		return Error{"SearchRadius", fmt.Sprintf("Invalid radius specified: %d", int(sr))}
	}

	//write query, update mask and return
	sq.Append(searchIdentifierRadius, strconv.Itoa(int(sr)))

	sq.mask |= searchBitMaskRadius
	return nil
}

//SearchDeals is a boolean search option whether Yelp should only return
//businesses with deals.
type SearchDeals bool

func (sd SearchDeals) Query(sq *SearchQuery) error {
	//make sure the deals option isn't already set
	if sq.mask&searchBitMaskDeals != 0 {
		return Error{"SearchDeals", "Attempting to set the deals search option a second time"}
	}

	//add query, update mask and return
	if sd == true {
		sq.Append(searchIdentifierDeals, "true")
	} else {
		sq.Append(searchIdentifierDeals, "false")
	}

	sq.mask |= searchBitMaskDeals
	return nil
}

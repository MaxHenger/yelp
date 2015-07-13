package yelp

//The Coordinates structure represents latitude and longitude coordinates and has
//JSON tags such that it can be read from the returned yelp data
type Coordinates struct {
	Latitude  float64 `json:"latitude" json:"latitude_delta"`
	Longitude float64 `json:"longitude" json:"longitude_delta"`
}

//The BusinessLocation represents the location of a yelp business. It is embedded
//within the Business structure.
type BusinessLocation struct {
	Address        []string    `json:"address"`
	City           string      `json:"city"`
	Position       Coordinates `json:"coordinate"`
	CountryCode    string      `json:"country_code"`
	DisplayAddress []string    `json:"display_address"`
	PostalCode     string      `json:"postal_code"`
	StateCode      string      `json:"state_code"`
}

//The Business structure is the complete description of a business as provided
//by yelp.
type Business struct {
	DisplayPhone string            `json:"display_phone"`
	Distance     float64           `json:"distance"`
	IsClosed     bool              `json:"is_closed"`
	Location     *BusinessLocation `json:"location"`
	Name         string            `json:"name"`
	Phone        string            `json:"phone"`
	Rating       float64           `json:"rating"`
}

//The BusinessRegion structure specifies the center of the region which is
//considered in the query and the distance (span) it is from the provided
//coordinates
type BusinessRegion struct {
	Center Coordinates `json:"center"`
	Span   Coordinates `json:"span"`
}

//The Businesses structure acts as a container for the JSON data that yelp
//returns upon a successfull Yelp API call. This is the structure in which the
//result will be unmarshalled.
type Businesses struct {
	Businesses []*Business     `json:"businesses"`
	Region     *BusinessRegion `json:"region"`
	Total      int             `json:"total"`
}

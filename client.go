package yelp

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"math/rand"
	"time"
)

func init() {
	rand.Seed(time.Now().Unix())
}

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
	DisplayPhone string           `json:"display_phone"`
	Distance     float64          `json:"distance"`
	IsClosed     bool             `json:"is_closed"`
	Location     *BusinessLocation `json:"location"`
	Name         string           `json:"name"`
	Phone        string           `json:"phone"`
	Rating       float64          `json:"rating"`
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
	Total      int            `json:"total"`
}

//The responseError struct is used when the data request was not successfully
//handled by the Yelp API and an error is occurred in JSON format. This error
//will then be unmarshalled into this structure.
type responseError struct {
	Text        string `json:"text"`
	ID          string `json:"id"`
	Description string `json:"description"`
}

//The responseErrorContainer is the container structure of the responseError
//structure used for unmarshalling JSON data returned from the Yelp API.
type responseErrorContainer struct {
	Error responseError `json:"error"`
}

//The Client structure is the interface through which all the Yelp API
//functionality can be accessed. The structure can be best created through the
//provided New(...) method. Once created, queries to Yelp can be performed in
//two ways:
// 1. Through SearchQuery(...), specifying the query elements manually with the
//		chance of getting them wrong.
// 2. Through SearchOptions(...), specifying the various query elements through
//		components implementing the SearchQuerier interface. This will be
//		computationally more intensive but safer.
type Client struct {
	url    string
	signer oauth
}

//New will create a new client from the provided arguments.
func New(URL, consumerKey, consumerSecret, token, tokenSecret string) (c *Client) {
	c = &Client{}
	c.url = URL
	c.signer.ConsumerKey = consumerKey
	c.signer.Token = token
	c.signer.SetHashKey(consumerSecret, tokenSecret)

	return
}

//validateResponse is a function that will accept the data retrieved from the
//body of the retrieved HTML page and scan it for a default error message. If
//the default error message exists this function will return a non-nil error
//detailing the contents of the error message. If the default error message is
//not on the page, then this function will attempt to unmarshal the page
//assuming it contains businesses. If this assumption is invalid and/or the
//data is incorrect, this function will return an error
func (c Client) validateResponse(response []byte) (businesses *Businesses, err error) {
	//peek ahead in the reponse to see if the first found text is 'error'
	found := false
	var text string

	for iStart, vStart := range response {
		if vStart == '"' {
			//found the first bracket, search for the second one
			for iEnd := iStart + 1; iEnd < len(response); iEnd++ {
				if response[iEnd] == '"' {
					//found the second bracket
					found = true
					text = string(response[iStart+1 : iEnd])
					break
				}
			}

			if !found {
				//did not find a second bracket
				return nil, Error{"Client", "Could not find a matching closing '\"' bracket while searching for the first JSON entry"}
			}

			break
		}
	}

	if !found {
		//did not find an opening bracket
		return nil, Error{"Client", "Could not find an opening '\"' bracket while search for the first JSON entry"}
	}

	//check if the returned data contained an error
	if text == "error" {
		//Yelp has returned an error, process it and return as a go error
		var yelpError responseErrorContainer
		e := json.Unmarshal(response, &yelpError)

		if e != nil {
			//Unmarshaling into the error structure also yielded problems
			return nil, ErrorEmbedded{"Client", "Error retrieved from Yelp, could not unmarshal it", e}
		}

		//Return error information
		return nil, Error{"Client", fmt.Sprintf("Retrieved error from Yelp:\n\tText:%v\n\tID:%v\n\tDescription:%v",
			yelpError.Error.Text, yelpError.Error.ID, yelpError.Error.Description)}
	}

	//no error, response is highly probable to be correct. If not then json
	//unmarshaling will get the error
	businesses = &Businesses{}
	err = json.Unmarshal(response, businesses)
	return
}

//SearchQuery allows performing a search on the Yelp API by specifying the
//query elements manually. The SearchQuery object is passed to this function by
//copy such that the original query will not be altered after this function is
//completed (succesfully or otherwise). The reason being that OAuth query
//elements have to be added to the query.
func (c Client) SearchQuery(q SearchQuery) (*Businesses, error) {
	//The Query is intentionally passed to the function by value such that the
	//original query is not changed when the oauth-elements get added to the
	//query
	qp := &q

	//sign the current query
	c.signer.Sign("GET", c.url, qp)

	//create the URL from which to request data
	data, err := http.Get(strings.Join([]string{c.url, qp.String()}, "?"))
	defer data.Body.Close()

	if err != nil {
		return nil, ErrorEmbedded{"Client", "Failed to perform HTTP request", err}
	}

	//read the body of the html page
	body, err := ioutil.ReadAll(data.Body)

	if err != nil {
		return nil, ErrorEmbedded{"Client", "Failed to read entire HTML body", err}
	}

	//validate the response
	businesses, err := c.validateResponse(body)

	if err == nil {
		//correctly retrieved
		return businesses, nil
	}

	return nil, ErrorEmbedded{"Client", "The retrieved yelp data is invalid", err}
}

//SearchOptions allows performing a search using the Yelp API by options
//implementing the SearchQuerier interface.
func (c Client) SearchOptions(options ...SearchQuerier) (*Businesses, error) {
	//This version will create a query from the provided options using the
	//SearchQuerier interface
	var qp SearchQuery

	for _, v := range options {
		e := v.Query(&qp)

		if e != nil {
			//an error ocurred while querying the option
			return nil, ErrorEmbedded{"Client", "Failed to create query from options", e}
		}
	}

	return c.SearchQuery(qp)
}

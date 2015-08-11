/*
This package provides an interface to the functionality the Yelp 2.0 API offers.
Its use is entirely facilitated by instantiating a Client type through the
New(...) function. Here one has to specify consecutively:
1) The Yelp API URL (at the moment of writing this is http://api.yelp.com/v2/search
2) The consumer key
3) The consumer key secret
4) The token
5) The token secret
The last four values are provided by Yelp by creating an account.

After the Client is created, one can perform queries in two manners:

	1) Using the Query structure
Creating an instance of the Query structure. After which consecutive
calls to the Append(name, value) function will add a new query element.
Each consists of a element name and a corresponding value. These values
must follow the rules provided by the Yelp API documentation.

The advantage of this method is that the queries are all performed
slightly faster. However, these are more error-prone and, in the case
that Yelp decides to change their API, harder to maintain.

Once the Query type is fully initialized, one can call the
Client.SearchQuery(...) function to retrieve the businesses

	2) Using the SearchOptions(...) function with SearchQuerier implementations
Calling Client.SearchOptions(...) with the dedicated option types, all
implementing a SearchQuerier interface. The currently available search
options are:
- SearchLocation
- SearchCoordinates
- SearchLocationCoordinates
- SearchBounds
- SearchTerms
- SearchLimit
- SearchOffset
- SearchSort
- SearchCategory
- SearchRadius
- SearchDeals

This method if searching is slightly slower, but the resulting code is
more easily maintainable and will check for the possibility of multiple
defined search options.

The query will still have to be checked for possible errors. In case the error
originated from within the Yelp API this error can be displayed.
*/
package yelp

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
)

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
//
//1. Through SearchQuery(...), specifying the query elements manually with the
//chance of getting them wrong.
//
//2. Through SearchOptions(...), specifying the various query elements through
//components implementing the SearchQuerier interface. This will be
//computationally more intensive but safer.
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
				return nil, Error{ErrorTypeInvalidYelpResponse, "Client", "Could not find a matching closing '\"' bracket while searching for the first JSON entry"}
			}

			break
		}
	}

	if !found {
		//did not find an opening bracket
		return nil, Error{ErrorTypeInvalidYelpResponse, "Client", "Could not find an opening '\"' bracket while search for the first JSON entry"}
	}

	//check if the returned data contained an error
	if text == "error" {
		//Yelp has returned an error, process it and return as a go error
		var yelpError responseErrorContainer
		e := json.Unmarshal(response, &yelpError)

		if e != nil {
			//Unmarshaling into the error structure also yielded problems
			return nil, Error{ErrorTypeInvalidYelpResponse, "Client", "Error retrieved from Yelp, could not unmarshal it"}
		}

		//Return error information
		return nil, Error{ErrorTypeInvalidYelpResponse, "Client", fmt.Sprintf("Retrieved error from Yelp:\n\tText:%v\n\tID:%v\n\tDescription:%v",
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
	err := c.signer.Sign("GET", c.url, qp)

	if err != nil {
		return nil, Error{ErrorTypeOAuthFailure, "Client", "Failed to sign request"}
	}

	//create the URL from which to request data
	data, err := http.Get(strings.Join([]string{c.url, qp.String()}, "?"))
	defer data.Body.Close()

	if err != nil {
		return nil, Error{ErrorTypeHTTPFailure, "Client", "Failed to perform HTTP request"}
	}

	//read the body of the html page
	body, err := ioutil.ReadAll(data.Body)

	if err != nil {
		return nil, Error{ErrorTypeReadFailure, "Client", "Failed to read entire HTML body"}
	}

	//validate the response and return businesses and possible error
	return c.validateResponse(body)
}

//SearchOptions allows performing a search using the Yelp API by options
//implementing the SearchQuerier interface.
func (c Client) SearchOptions(options ...SearchQuerier) (*Businesses, error) {
	//This version will create a query from the provided options using the
	//SearchQuerier interface
	var qp SearchQuery

	for _, v := range options {
		err := v.Query(&qp)

		if err != nil {
			//an error ocurred while querying the option
			return nil, err
		}
	}

	return c.SearchQuery(qp)
}

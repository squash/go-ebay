package ebay

import (
	"encoding/xml"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

const (
	// GlobalIDEbayUS is the global eBay ID for US
	GlobalIDEbayUS = "EBAY-US"
	// GlobalIDEbayFR is the global eBay ID for France
	GlobalIDEbayFR = "EBAY-FR"
	// GlobalIDEbayDE is the global eBay ID for Germany
	GlobalIDEbayDE = "EBAY-DE"
	// GlobalIDEbayIT is the global eBay ID for Italy
	GlobalIDEbayIT = "EBAY-IT"
	// GlobalIDEbayES is the global eBay ID for Spain
	GlobalIDEbayES = "EBAY-ES"
)

// Item is an eBay auction/sale item
type Item struct {
	ItemID        string    `xml:"itemId"`
	Title         string    `xml:"title"`
	Location      string    `xml:"location"`
	CurrentPrice  float64   `xml:"sellingStatus>currentPrice"`
	ShippingPrice float64   `xml:"shippingInfo>shippingServiceCost"`
	BinPrice      float64   `xml:"listingInfo>buyItNowPrice"`
	ShipsTo       []string  `xml:"shippingInfo>shipToLocations"`
	ListingURL    string    `xml:"viewItemURL"`
	ImageURL      string    `xml:"galleryURL"`
	Site          string    `xml:"globalId"`
	EndTime       time.Time `xml:"listingInfo>endTime"`
}

// FindItemsResponse is a response to a findItemsByKeywords search
type FindItemsResponse struct {
	Name      xml.Name `xml:"findItemsByKeywordsResponse"`
	Items     []Item   `xml:"searchResult>item"`
	Timestamp string   `xml:"timestamp"`
}

// ErrorMessage is an eBay error message
type ErrorMessage struct {
	Name  xml.Name `xml:"errorMessage"`
	Error Error    `xml:"error"`
}

// Error is an eBay error response
type Error struct {
	ErrorID   string `xml:"errorId"`
	Domain    string `xml:"domain"`
	Severity  string `xml:"severity"`
	Category  string `xml:"category"`
	Message   string `xml:"message"`
	SubDomain string `xml:"subdomain"`
}

// Session is an eBay searches session
type Session struct {
	ApplicationID string
}

type getURL func(string, string, int) (string, error)

//New creates a new eBay Session
func New(applicationID string) *Session {
	e := Session{}
	e.ApplicationID = applicationID
	return &e
}

var transport http.RoundTripper

func getTransport() http.RoundTripper {
	if transport != nil {
		return transport
	}
	return http.DefaultTransport
}

// SetTransport overrides the HTTP transport for test purposes
func SetTransport(t http.RoundTripper) {
	transport = t
}

func get(url string, headers map[string]string) (resp *http.Response, error error) {
	client := &http.Client{Transport: getTransport()}
	req, _ := http.NewRequest("GET", url, nil)
	if headers != nil {
		for key, val := range headers {
			req.Header.Set(key, val)
		}
	}
	response, error := client.Do(req)

	return response, error
}

func (e *Session) buildSoldURL(globalID string, keywords string, entriesPerPage int) (string, error) {
	filters := url.Values{}
	filters.Add("itemFilter(0).name", "Condition")
	filters.Add("itemFilter(0).value(0)", "Used")
	filters.Add("itemFilter(0).value(1)", "Unspecified")
	filters.Add("itemFilter(1).name", "SoldItemsOnly")
	filters.Add("itemFilter(1).value(0)", "true")
	return e.buildURL(globalID, keywords, "findCompletedItems", entriesPerPage, filters)
}

func (e *Session) buildSearchURL(globalID string, keywords string, entriesPerPage int) (string, error) {
	filters := url.Values{}
	filters.Add("itemFilter(0).name", "ListingType")
	filters.Add("itemFilter(0).value(0)", "FixedPrice")
	filters.Add("itemFilter(0).value(1)", "AuctionWithBIN")
	filters.Add("itemFilter(0).value(2)", "Auction")
	return e.buildURL(globalID, keywords, "findItemsByKeywords", entriesPerPage, filters)
}

func (e *Session) buildURL(globalID string, keywords string, operationName string, entriesPerPage int, filters url.Values) (string, error) {
	var u *url.URL
	u, err := url.Parse("http://svcs.ebay.com/services/search/FindingService/v1")
	if err != nil {
		return "", err
	}
	params := url.Values{}
	params.Add("OPERATION-NAME", operationName)
	params.Add("SERVICE-VERSION", "1.0.0")
	params.Add("SECURITY-APPNAME", e.ApplicationID)
	params.Add("GLOBAL-ID", globalID)
	params.Add("RESPONSE-DATA-FORMAT", "XML")
	params.Add("REST-PAYLOAD", "")
	params.Add("keywords", keywords)
	params.Add("paginationInput.entriesPerPage", strconv.Itoa(entriesPerPage))
	for key := range filters {
		for _, val := range filters[key] {
			params.Add(key, val)
		}
	}
	u.RawQuery = params.Encode()
	//fmt.Println(u.String())
	return u.String(), err
}

func (e *Session) findItems(globalID string, keywords string, entriesPerPage int, getURL getURL) (FindItemsResponse, error) {
	var response FindItemsResponse
	url, err := getURL(globalID, keywords, entriesPerPage)
	if err != nil {
		return response, err
	}
	headers := make(map[string]string)
	headers["User-Agent"] = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_7_3) AppleWebKit/535.11 (KHTML, like Gecko) Chrome/17.0.963.56 Safari/535.11"
	httpResp, err := get(url, headers)
	if err != nil {
		return response, err
	}

	defer httpResp.Body.Close()
	body, err := ioutil.ReadAll(httpResp.Body)

	if httpResp.StatusCode != 200 {
		var em ErrorMessage
		err = xml.Unmarshal([]byte(body), &em)
		if err != nil {
			return response, err
		}
		return response, errors.New(em.Error.Message)
	}
	err = xml.Unmarshal([]byte(body), &response)
	if err != nil {
		return response, err
	}
	return response, err
}

// FindItemsByKeywords searches for ebay items by keywords.  It returns the items found and any error encountered.
func (e *Session) FindItemsByKeywords(globalID string, keywords string, entriesPerPage int) (FindItemsResponse, error) {
	return e.findItems(globalID, keywords, entriesPerPage, e.buildSearchURL)
}

// FindSoldItems searches for sold ebay items by keywords.  It returns the items found and any error encountered.
func (e *Session) FindSoldItems(globalID string, keywords string, entriesPerPage int) (FindItemsResponse, error) {
	return e.findItems(globalID, keywords, entriesPerPage, e.buildSoldURL)
}

// Dump writes out the contents of the FindItemsResponse into a formatted string for debugging purposes.
func (r *FindItemsResponse) Dump() {
	fmt.Println("FindItemsResponse")
	fmt.Println("--------------------------")
	fmt.Println("Timestamp: ", r.Timestamp)
	fmt.Println("Items:")
	fmt.Println("------")
	for _, i := range r.Items {
		fmt.Println("Title: ", i.Title)
		fmt.Println("------")
		fmt.Println("\tListing Url:     ", i.ListingURL)
		fmt.Println("\tBin Price:       ", i.BinPrice)
		fmt.Println("\tCurrent Price:   ", i.CurrentPrice)
		fmt.Println("\tShipping Price:  ", i.ShippingPrice)
		fmt.Println("\tShips To:        ", i.ShipsTo)
		fmt.Println("\tSeller Location: ", i.Location)
		fmt.Println()
	}
}

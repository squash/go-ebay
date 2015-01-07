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
	GLOBAL_ID_EBAY_US = "EBAY-US"
	GLOBAL_ID_EBAY_FR = "EBAY-FR"
	GLOBAL_ID_EBAY_DE = "EBAY-DE"
	GLOBAL_ID_EBAY_IT = "EBAY-IT"
	GLOBAL_ID_EBAY_ES = "EBAY-ES"
)

type Item struct {
	ItemId        string    `xml:"itemId"`
	Title         string    `xml:"title"`
	Location      string    `xml:"location"`
	CurrentPrice  float64   `xml:"sellingStatus>currentPrice"`
	ShippingPrice float64   `xml:"shippingInfo>shippingServiceCost"`
	BinPrice      float64   `xml:"listingInfo>buyItNowPrice"`
	ShipsTo       []string  `xml:"shippingInfo>shipToLocations"`
	ListingUrl    string    `xml:"viewItemURL"`
	ImageUrl      string    `xml:"galleryURL"`
	Site          string    `xml:"globalId"`
	EndTime       time.Time `xml:"listingInfo>endTime"`
}

type FindItemsResponse struct {
	XmlName   xml.Name `xml:"findItemsByKeywordsResponse"`
	Items     []Item   `xml:"searchResult>item"`
	Timestamp string   `xml:"timestamp"`
}

type ErrorMessage struct {
	XmlName xml.Name `xml:"errorMessage"`
	Error   Error    `xml:"error"`
}

type Error struct {
	ErrorId   string `xml:"errorId"`
	Domain    string `xml:"domain"`
	Severity  string `xml:"severity"`
	Category  string `xml:"category"`
	Message   string `xml:"message"`
	SubDomain string `xml:"subdomain"`
}

type Session struct {
	ApplicationId string
	//HttpRequest   *httprequest.HttpRequest
}

type getUrl func(string, string, int) (string, error)

func New(application_id string) *Session {
	e := Session{}
	e.ApplicationId = application_id
	return &e
}

var transport http.RoundTripper

func getTransport() http.RoundTripper {
	if transport != nil {
		return transport
	}
	return http.DefaultTransport
}

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

func (e *Session) build_sold_url(global_id string, keywords string, entries_per_page int) (string, error) {
	filters := url.Values{}
	filters.Add("itemFilter(0).name", "Condition")
	filters.Add("itemFilter(0).value(0)", "Used")
	filters.Add("itemFilter(0).value(1)", "Unspecified")
	filters.Add("itemFilter(1).name", "SoldItemsOnly")
	filters.Add("itemFilter(1).value(0)", "true")
	return e.build_url(global_id, keywords, "findCompletedItems", entries_per_page, filters)
}

func (e *Session) build_search_url(global_id string, keywords string, entries_per_page int) (string, error) {
	filters := url.Values{}
	filters.Add("itemFilter(0).name", "ListingType")
	filters.Add("itemFilter(0).value(0)", "FixedPrice")
	filters.Add("itemFilter(0).value(1)", "AuctionWithBIN")
	filters.Add("itemFilter(0).value(2)", "Auction")
	return e.build_url(global_id, keywords, "findItemsByKeywords", entries_per_page, filters)
}

func (e *Session) build_url(global_id string, keywords string, operationName string, entries_per_page int, filters url.Values) (string, error) {
	var u *url.URL
	u, err := url.Parse("http://svcs.ebay.com/services/search/FindingService/v1")
	if err != nil {
		return "", err
	}
	params := url.Values{}
	params.Add("OPERATION-NAME", operationName)
	params.Add("SERVICE-VERSION", "1.0.0")
	params.Add("SECURITY-APPNAME", e.ApplicationId)
	params.Add("GLOBAL-ID", global_id)
	params.Add("RESPONSE-DATA-FORMAT", "XML")
	params.Add("REST-PAYLOAD", "")
	params.Add("keywords", keywords)
	params.Add("paginationInput.entriesPerPage", strconv.Itoa(entries_per_page))
	for key := range filters {
		for _, val := range filters[key] {
			params.Add(key, val)
		}
	}
	u.RawQuery = params.Encode()
	return u.String(), err
}

func (e *Session) findItems(global_id string, keywords string, entries_per_page int, getUrl getUrl) (FindItemsResponse, error) {
	var response FindItemsResponse
	url, err := getUrl(global_id, keywords, entries_per_page)
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
	} else {
		err = xml.Unmarshal([]byte(body), &response)
		if err != nil {
			return response, err
		}
	}
	return response, err
}

func (e *Session) FindItemsByKeywords(global_id string, keywords string, entries_per_page int) (FindItemsResponse, error) {
	return e.findItems(global_id, keywords, entries_per_page, e.build_search_url)
}

func (e *Session) FindSoldItems(global_id string, keywords string, entries_per_page int) (FindItemsResponse, error) {
	return e.findItems(global_id, keywords, entries_per_page, e.build_sold_url)
}

func (r *FindItemsResponse) Dump() {
	fmt.Println("FindItemsResponse")
	fmt.Println("--------------------------")
	fmt.Println("Timestamp: ", r.Timestamp)
	fmt.Println("Items:")
	fmt.Println("------")
	for _, i := range r.Items {
		fmt.Println("Title: ", i.Title)
		fmt.Println("------")
		fmt.Println("\tListing Url:     ", i.ListingUrl)
		fmt.Println("\tBin Price:       ", i.BinPrice)
		fmt.Println("\tCurrent Price:   ", i.CurrentPrice)
		fmt.Println("\tShipping Price:  ", i.ShippingPrice)
		fmt.Println("\tShips To:        ", i.ShipsTo)
		fmt.Println("\tSeller Location: ", i.Location)
		fmt.Println()
	}
}

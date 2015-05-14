package ebay

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
)

var (
	testApplicationID = "your_application_id_here"

	emptySoldItemsResponse = `<?xml version='1.0' encoding='UTF-8'?>
<findCompletedItemsResponse xmlns="http://www.ebay.com/marketplace/search/v1/services">
   <ack>Success</ack>
   <version>1.13.0</version>
   <timestamp>2015-05-14T02:30:45.558Z</timestamp>
   <searchResult count="0"/>
   <paginationOutput>
      <pageNumber>0</pageNumber>
      <entriesPerPage>100</entriesPerPage>
      <totalPages>0</totalPages>
      <totalEntries>0</totalEntries>
   </paginationOutput>
</findCompletedItemsResponse>`

	successfulSoldItemsResponse = `<findCompletedItemsResponse xmlns="http://www.ebay.com/marketplace/search/v1/services">
   <ack>Success</ack>
   <version>1.11.0</version>
   <timestamp>2011-12-17T22:55:44.894Z</timestamp>
   <searchResult count="2">
      <item>
         <itemId>260912534793</itemId>
         <title>Garmin nuvi 1300 Automotive GPS Receiver HAS PIN PLEASE READ</title>
         <globalId>EBAY-US</globalId>
         <primaryCategory>
            <categoryId>156955</categoryId>
            <categoryName>GPS Systems</categoryName>
         </primaryCategory>
         <galleryURL>http://thumbs2.ebaystatic.com/pict/2609125347934040_1.jpg</galleryURL>
         <viewItemURL>http://www.ebay.com/itm/Garmin-nuvi-1300-Automotive-GPS-Receiver-HAS-PIN-PLEASE-READ-/260912534793?pt=GPS_Devices</viewItemURL>
         <productId type="ReferenceID">72068366</productId>
         <paymentMethod>PayPal</paymentMethod>
         <autoPay>false</autoPay>
         <postalCode>56303</postalCode>
         <location>Saint Cloud,MN,USA</location>
         <country>US</country>
         <shippingInfo>
            <shippingServiceCost currencyId="USD">0.0</shippingServiceCost>
            <shippingType>Free</shippingType>
            <expeditedShipping>true</expeditedShipping>
            <oneDayShippingAvailable>false</oneDayShippingAvailable>
            <handlingTime>3</handlingTime>
            <shipToLocations>US</shipToLocations>
         </shippingInfo>
         <sellingStatus>
            <currentPrice currencyId="USD">25.0</currentPrice>
            <convertedCurrentPrice currencyId="USD">25.0</convertedCurrentPrice>
            <bidCount>1</bidCount>
            <sellingState>EndedWithSales</sellingState>
         </sellingStatus>
         <listingInfo>
            <bestOfferEnabled>false</bestOfferEnabled>
            <buyItNowAvailable>false</buyItNowAvailable>
            <startTime>2011-12-08T17:02:40.000Z</startTime>
            <endTime>2011-12-08T19:17:28.000Z</endTime>
            <listingType>Auction</listingType>
            <gift>false</gift>
         </listingInfo>
         <returnsAccepted>false</returnsAccepted>
         <condition>
            <conditionId>3000</conditionId>
            <conditionDisplayName>Used</conditionDisplayName>
         </condition>
         <isMultiVariationListing>false</isMultiVariationListing>
      </item>
      <item>
         <itemId>220913028191</itemId>
         <title>Garmin nuvi 1300 Automotive GPS Receiver</title>
         <globalId>EBAY-US</globalId>
         <primaryCategory>
            <categoryId>156955</categoryId>
            <categoryName>GPS Systems</categoryName>
         </primaryCategory>
         <galleryURL>http://thumbs4.ebaystatic.com/pict/2209130281914040_1.jpg</galleryURL>
         <viewItemURL>http://www.ebay.com/itm/Garmin-nuvi-1300-Automotive-GPS-Receiver-/220913028191?pt=GPS_Devices</viewItemURL>
         <productId type="ReferenceID">72068366</productId>
         <paymentMethod>PayPal</paymentMethod>
         <autoPay>false</autoPay>
         <postalCode>98065</postalCode>
         <location>Snoqualmie,WA,USA</location>
         <country>US</country>
         <shippingInfo>
            <shippingServiceCost currencyId="USD">0.0</shippingServiceCost>
            <shippingType>Free</shippingType>
            <expeditedShipping>true</expeditedShipping>
            <oneDayShippingAvailable>false</oneDayShippingAvailable>
            <handlingTime>2</handlingTime>
            <shipToLocations>US</shipToLocations>
         </shippingInfo>
         <sellingStatus>
            <currentPrice currencyId="USD">44.0</currentPrice>
            <convertedCurrentPrice currencyId="USD">44.0</convertedCurrentPrice>
            <bidCount>1</bidCount>
            <sellingState>EndedWithSales</sellingState>
         </sellingStatus>
         <listingInfo>
            <bestOfferEnabled>false</bestOfferEnabled>
            <buyItNowAvailable>false</buyItNowAvailable>
            <startTime>2011-12-11T21:25:47.000Z</startTime>
            <endTime>2011-12-11T22:45:55.000Z</endTime>
            <listingType>Auction</listingType>
            <gift>false</gift>
         </listingInfo>
         <returnsAccepted>false</returnsAccepted>
         <condition>
            <conditionId>3000</conditionId>
            <conditionDisplayName>Used</conditionDisplayName>
         </condition>
         <isMultiVariationListing>false</isMultiVariationListing>
      </item>
   </searchResult>
   <paginationOutput>
      <pageNumber>1</pageNumber>
      <entriesPerPage>2</entriesPerPage>
      <totalPages>28</totalPages>
      <totalEntries>55</totalEntries>
   </paginationOutput>
</findCompletedItemsResponse>`
)

// func TestFindItemsByKeywords(t *testing.T) {
// 	fmt.Println("ebay.FindItemsByKeywords")
// 	e := New(testApplicationID)
// 	response, err := e.FindItemsByKeywords(GLOBAL_ID_EBAY_US, "DJM 900, DJM 850", 10)
// 	if err != nil {
// 		t.Errorf("ERROR: ", err)
// 	} else {
// 		fmt.Println("Timestamp: ", response.Timestamp)
// 		fmt.Println("Items:")
// 		fmt.Println("------")
// 		for _, i := range response.Items {
// 			fmt.Println("Title: ", i.Title)
// 			fmt.Println("------")
// 			fmt.Println("\tListing Url:     ", i.ListingUrl)
// 			fmt.Println("\tBin Price:       ", i.BinPrice)
// 			fmt.Println("\tCurrent Price:   ", i.CurrentPrice)
// 			fmt.Println("\tShipping Price:  ", i.ShippingPrice)
// 			fmt.Println("\tShips To:        ", i.ShipsTo)
// 			fmt.Println("\tSeller Location: ", i.Location)
// 			fmt.Println()
// 		}
// 	}
// }

func setupServerWithSuccessResponse(response string) *httptest.Server {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.Header().Set("Content-Type", "application/xml")
		fmt.Fprintln(w, response)
	}))
	// Make a transport that reroutes all traffic to the example server
	SetTransport(&http.Transport{
		Proxy: func(req *http.Request) (*url.URL, error) {
			return url.Parse(server.URL)
		},
	})
	return server
}

func Test_FindSoldItems_ItemsReturned_Success(t *testing.T) {
	server := setupServerWithSuccessResponse(successfulSoldItemsResponse)
	defer server.Close()

	s := Session{testApplicationID}
	resp, e := s.FindSoldItems(GLOBAL_ID_EBAY_US, "something", 100)
	assert.Nil(t, e)
	assert.Equal(t, 2, len(resp.Items))
	assert.Equal(t, "Garmin nuvi 1300 Automotive GPS Receiver HAS PIN PLEASE READ", resp.Items[0].Title)
	assert.Equal(t, "Garmin nuvi 1300 Automotive GPS Receiver", resp.Items[1].Title)
}

func Test_FindSoldItems_NoItemsReturned_Success(t *testing.T) {
	server := setupServerWithSuccessResponse(emptySoldItemsResponse)
	defer server.Close()

	s := Session{testApplicationID}
	resp, e := s.FindSoldItems(GLOBAL_ID_EBAY_US, "something", 100)
	assert.Nil(t, e)
	assert.Equal(t, 0, len(resp.Items))
}

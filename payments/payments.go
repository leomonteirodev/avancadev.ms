package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"

	"github.com/hashicorp/go-retryablehttp"
)

type Result struct {
	Status string
}

func main() {
	fmt.Println("Payments MS running on port :9091")

	http.HandleFunc("/", home)
	http.ListenAndServe(":9091", nil)
}

func home(w http.ResponseWriter, r *http.Request) {
	coupon := r.FormValue("coupon")
	ccNumber := r.FormValue("ccNumber")

	resultCoupon := makeHttpCall("http://localhost:9092", coupon)

	result := Result{Status: "Payment declined"}

	if ccNumber == "1" {
		result.Status = "Payment approved"
	}

	if resultCoupon.Status == "Sorry, service unavailable temporarily" {
		result.Status = "Sorry, service unavailable temporarily"
	}

	if resultCoupon.Status == "invalid" {
		result.Status = "Invalid coupon"
	}

	jsonData, err := json.Marshal(result)

	if err != nil {
		log.Fatal("Error converting data to JSON")
	}

	fmt.Fprintf(w, string(jsonData))
}

func makeHttpCall(urlMicroservice string, coupon string) Result {
	values := url.Values{}

	values.Add("coupon", coupon)

	retryClient := retryablehttp.NewClient()

	// Retries amount
	retryClient.RetryMax = 5

	res, err := retryClient.PostForm(urlMicroservice, values)

	if err != nil {
		result := Result{Status: "Sorry, service unavailable temporarily"}

		return result
	}

	defer res.Body.Close()

	data, err := ioutil.ReadAll(res.Body)

	if err != nil {
		log.Fatal("Error processing the result")
	}

	result := Result{}

	json.Unmarshal(data, &result)

	return result
}

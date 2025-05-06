package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
)

func main() {
	apiKey, err := getAPIKey("api_key.txt")
	if err != nil {
		fmt.Println("Error reading API key:", err)
		return
	}

	location := "50.061947,19.936856" // Example: Rynek Główny w Krakowie
	radius := "2000"                  // Search within 2km
	placeTypes := []string{           // Example: Search for restaurants
		"restaurant",
		"food",
		"bar",
		"coffee",
		"point_of_interest"}

	places, err := getHighReviewPlaces(apiKey, location, radius, placeTypes)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	for _, place := range places {
		fmt.Printf("Name: %s, PlaceID: %s, Address: %s, Phone: %s, Website: %s, URL: %s, Opening Hours: %s, Rating: %.1f, User Ratings Total: %d, Bayesian Rating: %.2f\n",
			place.Name, place.PlaceID, place.FormattedAddress, place.PhoneNumber, place.Website, place.URL, place.OpeningHours, place.Rating, place.UserRatingsTotal, place.BayesianRating)
	}
}

const minReviews = 50 // Minimum number of reviews required
const minRating = 4.0 // Minimum rating required

type Place struct {
	Name             string  `json:"name"`
	PlaceID          string  `json:"place_id"`
	FormattedAddress string  `json:"formatted_address"`
	PhoneNumber      string  `json:"formatted_phone_number"`
	Website          string  `json:"website"`
	URL              string  `json:"url"`
	OpeningHours     string  `json:"opening_hours"`
	Rating           float64 `json:"rating"`
	UserRatingsTotal int     `json:"user_ratings_total"`
	BayesianRating   float64
}

func getAPIKey(filePath string) (string, error) {
	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func getPlaceDetails(apiKey, placeID string) (*Place, error) {
	baseURL := "https://maps.googleapis.com/maps/api/place/details/json"
	params := url.Values{}
	params.Add("place_id", placeID)
	params.Add("key", apiKey)

	resp, err := http.Get(fmt.Sprintf("%s?%s", baseURL, params.Encode()))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API request failed with status: %s", resp.Status)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var result struct {
		Result struct {
			Name             string `json:"name"`
			FormattedAddress string `json:"formatted_address"`
			FormattedPhone   string `json:"formatted_phone_number"`
			Website          string `json:"website"`
			URL              string `json:"url"`
			OpeningHours     struct {
				WeekdayText []string `json:"weekday_text"`
			} `json:"opening_hours"`
			Rating           float64 `json:"rating"`
			UserRatingsTotal int     `json:"user_ratings_total"`
		} `json:"result"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}

	// Constants for Bayesian Average calculation
	m := float64(minReviews) // Minimum number of reviews required
	C := minRating           // Use minRating as the baseline for Bayesian Average

	// Calculate Bayesian Average
	v := float64(result.Result.UserRatingsTotal)
	R := result.Result.Rating
	bayesianRating := (v/(v+m))*R + (m/(v+m))*C

	// Construct the Place object
	place := &Place{
		Name:             result.Result.Name,
		FormattedAddress: result.Result.FormattedAddress,
		PhoneNumber:      result.Result.FormattedPhone,
		Website:          result.Result.Website,
		URL:              result.Result.URL,
		Rating:           R,
		UserRatingsTotal: int(v),
		BayesianRating:   bayesianRating,
	}

	// Combine opening hours into a single string
	if len(result.Result.OpeningHours.WeekdayText) > 0 {
		place.OpeningHours = fmt.Sprintf("%v", result.Result.OpeningHours.WeekdayText)
	}

	return place, nil
}

func getHighReviewPlaces(apiKey, location, radius string, placeTypes []string) ([]Place, error) {
	highReviewPlaces := []Place{}
	for _, placeType := range placeTypes {
		baseURL := "https://maps.googleapis.com/maps/api/place/nearbysearch/json"
		fmt.Println("Fetching places of type:", placeType)
		params := url.Values{}
		params.Add("location", location)
		params.Add("radius", radius)
		params.Add("type", placeType)
		params.Add("key", apiKey)

		resp, err := http.Get(fmt.Sprintf("%s?%s", baseURL, params.Encode()))
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("API request failed with status: %s", resp.Status)
		}

		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}

		var result struct {
			Results []struct {
				PlaceID          string  `json:"place_id"`
				Rating           float64 `json:"rating"`
				UserRatingsTotal int     `json:"user_ratings_total"`
			} `json:"results"`
		}
		if err := json.Unmarshal(body, &result); err != nil {
			return nil, err
		}

		for _, basicPlace := range result.Results {
			// Filter places based on rating and user reviews
			if basicPlace.Rating >= minRating && basicPlace.UserRatingsTotal >= minReviews {
				// Fetch additional details
				details, err := getPlaceDetails(apiKey, basicPlace.PlaceID)
				if err != nil {
					fmt.Println("Error fetching details for place with PlaceID:", basicPlace.PlaceID, err)
					continue
				}

				// Add the detailed place to the list
				highReviewPlaces = append(highReviewPlaces, *details)
			}
		}
	}
	return highReviewPlaces, nil
}

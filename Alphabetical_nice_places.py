import requests
import pandas as pd
import time

with open("api_key.txt", "r") as f:
    API_KEY = f.read().strip()

LOCATION = "50.061947,19.936856"  # Rynek Główny w Krakowie
RADIUS = 2000  # w metrach
KEYWORD = [
    "restaurant",
    "food",
    "bar",
    "coffee",
    "point_of_interest"
]
MIN_RATING = 4.5
place_id = ""
url = "https://maps.googleapis.com/maps/api/place/nearbysearch/json"
details_url = "https://maps.googleapis.com/maps/api/place/details/json"


params = {
    "location": LOCATION,
    "radius": RADIUS,
    "keyword": "",
    "key": API_KEY
}

detailed_params = {
    "place_id": place_id,
    "fields": "name,formatted_address,formatted_phone_number,website,opening_hours,url",
    "key": API_KEY
}
places_id = set()
results = []
for key in KEYWORD:
    params.pop("pagetoken", None)
    while True:
        params["keyword"] = key
        response = requests.get(url, params=params)
        data = response.json()
        for place in data.get("results", []):
            place_id = place.get("place_id")
            if place_id not in places_id:
                places_id.add(place_id)
                rating = place.get("rating")
                review_number = place.get("user_ratings_total")
                name = place.get("name")
                types = place.get("types")
                if rating and rating >= MIN_RATING:
                    results.append({
                        "ID": place_id,
                        "Nazwa": name,
                        "Ocena": rating,
                        "Liczba Ocen": review_number,
                        "Typy": types
                    })

        next_page_token = data.get("next_page_token")
        if next_page_token:
            time.sleep(1)
            params["pagetoken"] = next_page_token
        else:
            break

for i in range(len(results)):
    while True:
        detailed_params.pop("pagetoken", None)
        detailed_params["place_id"] = results[i]["ID"]
        response = requests.get(details_url, params=detailed_params)
        data = response.json()
        places = data.get("result", {})
        open_hours = places.get("opening_hours", {}).get("weekday_text")
        hours = ""
        if open_hours:
            for day_hours in open_hours:
                hours = hours + day_hours + "  "
            results[i]["Godziny Otwarcia"] = hours
            results[i]["Strona"] = places.get("website")
            results[i]["Mapy"] = places.get("url")
        next_page_token = data.get("next_page_token")
        if next_page_token:
            time.sleep(1)
            params["pagetoken"] = next_page_token
        else:
            break

# print(results)
# Zapisz do Excela
df = pd.DataFrame(results)
df.to_excel("miejsca_krakow.xlsx", index=False)

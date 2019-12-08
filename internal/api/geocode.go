package api

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

var netClient = &http.Client{
	Timeout: time.Second * 10,
}

func geocode(query string) (Coordinates, error) {
	var coords Coordinates

	url := fmt.Sprintf("https://nominatim.openstreetmap.org/search/%s?format=json&limit=1", url.QueryEscape(query))
	resp, err := netClient.Get(url)
	if err != nil {
		return coords, err
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return coords, err
	}

	var places []interface{}
	err = json.Unmarshal(body, &places)
	if err != nil {
		return coords, err
	}
	if len(places) == 0 {
		return coords, fmt.Errorf("no places found for \"%s\"", query)
	}

	var placeInfo = places[0].(map[string]interface{})

	parse := func(in interface{}) (float64, error) { return strconv.ParseFloat(in.(string), 64) }
	lon, err := parse(placeInfo["lon"])
	lat, err := parse(placeInfo["lat"])
	coords = Coordinates{lon, lat}

	return coords, err
}

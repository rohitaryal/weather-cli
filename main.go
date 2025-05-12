package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

// Location name in string format. eg California
type locationName string

// Pin-pointed coordinate for a location
type coordinate struct {
	Lat float64
	Lon float64
}

// Each matching location in search
type location struct {
	Coord       coordinate `json:"coord"`
	Name        string     `json:"name"`
	FullName    string     `json:"full_name"`
	CompactName string     `json:"compact_name"`
	Country     string     `json:"country"`
}

// These define schema for a searched response for a location
type locationSearchResult struct {
	Message string     `json:"message"`
	Cod     string     `json:"cod"`
	Count   int        `json:"count"`
	Lists   []location `json:"list"`
}

type weatherCondition struct {
	ID          int64  `json:"id"`
	Main        string `json:"main"`
	Description string `json:"description"`
	Icon        string `json:"icon"`
}

type forecast struct {
	Dt            int64   `json:"dt"`
	Temp          float64 `json:"temp"`
	Precipitation float64 `json:"precipitation"`
}

type rainInfo struct {
	OneH float64 `json:"1h"`
}

type currentWeather struct {
	Dt         int64              `json:"dt"`
	Sunrise    int64              `json:"sunrise"`
	Sunset     int64              `json:"sunset"`
	Temp       float64            `json:"temp"`
	FeelsLike  float64            `json:"feels_like"`
	Pressure   int64              `json:"pressure"`
	Humidity   int64              `json:"humidity"`
	DewPoint   float64            `json:"dew_point"`
	UVI        float64            `json:"uvi"`
	Clouds     int64              `json:"clouds"`
	Visibility int64              `json:"visibility"`
	WindSpeed  float64            `json:"wind_speed"`
	WindDeg    int64              `json:"wind_deg"`
	WindGust   float64            `json:"wind_gust"`
	Weather    []weatherCondition `json:"weather"`
}

type minutelyForecast struct {
	Dt            int64   `json:"dt"`
	Precipitation float64 `json:"precipitation"`
}

type hourlyForecast struct {
	Dt         int64              `json:"dt"`
	Temp       float64            `json:"temp"`
	FeelsLike  float64            `json:"feels_like"`
	Pressure   int64              `json:"pressure"`
	Humidity   int64              `json:"humidity"`
	DewPoint   float64            `json:"dew_point"`
	UVI        float64            `json:"uvi"`
	Clouds     int64              `json:"clouds"`
	Visibility int64              `json:"visibility"`
	WindSpeed  float64            `json:"wind_speed"`
	WindDeg    int64              `json:"wind_deg"`
	WindGust   float64            `json:"wind_gust"`
	Weather    []weatherCondition `json:"weather"`
	Pop        float64            `json:"pop"`
	Rain       *rainInfo          `json:"rain,omitempty"`
}

type dailyForecast struct {
	Dt            int64              `json:"dt"`
	Sunrise       int64              `json:"sunrise"`
	Sunset        int64              `json:"sunset"`
	TempMax       float64            `json:"temp_max"`
	TempMin       float64            `json:"temp_min"`
	Pressure      int64              `json:"pressure"`
	Humidity      int64              `json:"humidity"`
	WindSpeed     float64            `json:"wind_speed"`
	WindDeg       int64              `json:"wind_deg"`
	WindGust      float64            `json:"wind_gust"`
	Weather       []weatherCondition `json:"weather"`
	Clouds        int64              `json:"clouds"`
	Precipitation float64            `json:"precipitation"`
	Pop           float64            `json:"pop"`
	UVI           float64            `json:"uvi"`
	Forecast      []forecast         `json:"forecast"`
}

type weatherData struct {
	Lat            float64            `json:"lat"`
	Lon            float64            `json:"lon"`
	Timezone       string             `json:"timezone"`
	TimezoneOffset float64            `json:"timezone_offset"`
	Current        currentWeather     `json:"current"`
	Minutely       []minutelyForecast `json:"minutely"`
	Hourly         []hourlyForecast   `json:"hourly"`
	Daily          []dailyForecast    `json:"daily"`
}

type IPInfo struct {
	IP          string  `json:"ip"`
	Country     string  `json:"country"`
	CountryCode string  `json:"country_code"`
	Region      string  `json:"region"`
	ZipCode     string  `json:"zip_code"`
	City        string  `json:"city"`
	StateCode   string  `json:"state_code"`
	Longitude   float64 `json:"longitude"`
	Latitude    float64 `json:"latitude"`
	ISP         string  `json:"isp"`
	ISPASN      int     `json:"isp_asn"`
	GDPR        bool    `json:"gdpr"`
	Protected   bool    `json:"protected"`
}

const URL = "https://app.owm.io/app"

// These are specific API keys
const DEVICE_ID = "e13401912dbaf7cc"
const APP_ID = "e0c56f6c3cee94d1a83f36043ff1ce5b"
const TOKEN = DEVICE_ID + ":APA91bGAmF46L0bGb2jVYVfVKNpWePUqWdgoo4hz8_LLkfECQ8qw8JdcA-8hsJ6WSgjfEY5CvgjNoYMYF8PLvGlJ9GFM2ERKnKWjBR_Hq2tjsuZABJ_io3c"

var weatherIconEmojis = map[string]string{
	"01d": "☀️",
	"01n": "🌙",
	"02d": "🌤️",
	"02n": "🌥️",
	"03d": "🌥️",
	"03n": "☁️",
	"04d": "☁️",
	"04n": "☁️",
	"09d": "🌦️",
	"09n": "🌧️",
	"10d": "🌦️",
	"10n": "🌧️",
	"11d": "⛈️",
	"11n": "⛈️",
	"13d": "🌨️",
	"13n": "🌨️",
	"50d": "🌫️",
	"50n": "🌫️",
}

func fetch(url string) []byte {
	// Create a client
	client := http.Client{Timeout: time.Second * 10}

	// Defer the connections closing part
	defer client.CloseIdleConnections()

	// Create a request
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		fmt.Println("Failed to create a new request.")
		fmt.Println(err)
		os.Exit(1)
	}

	// Make the request
	res, err := client.Do(req)
	if err != nil {
		fmt.Println("Failed to send request to " + URL)
		fmt.Println(err)
		os.Exit(2)
	}

	// Defer the body (stream) closing part
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		fmt.Println("Failed to read response body")
		fmt.Println(err)
		os.Exit(3)
	}

	return body
}

func (l locationName) findCoordinate() locationSearchResult {
	fmt.Println("[@] Searching for " + string(l))

	// URL to be used to make request
	TARGET_URL := fmt.Sprintf("%s/1.1/find/?q=%s&appid=%s&deviceid=%s", URL, string(l), APP_ID, DEVICE_ID)

	body := fetch(TARGET_URL)

	// Parse the response to json
	var parsedResponse locationSearchResult
	err := json.Unmarshal(body, &parsedResponse)
	if err != nil {
		fmt.Println("Failed to marshal response to JSON")
		fmt.Println(err)
		fmt.Println(string(body))
		os.Exit(4)
	}

	return parsedResponse
}

func (l locationSearchResult) print() {
	fmt.Printf("Total available locations: %d\n", l.Count)
	for index, value := range l.Lists {
		fmt.Printf("---------------[%d]----------------\n", index+1)

		fmt.Println("Country: " + value.Country)
		fmt.Println("Location: " + value.CompactName)
		fmt.Printf("Latitude: %f\n", value.Coord.Lat)
		fmt.Printf("Longitude: %f\n\n", value.Coord.Lon)
	}
}

func (c coordinate) findWeather() weatherData {
	fmt.Println("[@] Searching for weather")

	UNIT := "metric" // or "imperial"

	TARGET_URL := fmt.Sprintf("%s/1.0/weather/?lat=%f&lon=%f&units=%s&appid=%s&deviceid=%s&token=%s", URL, c.Lat, c.Lon, UNIT, APP_ID, DEVICE_ID, TOKEN)

	body := fetch(TARGET_URL)

	var parsedResponse weatherData
	err := json.Unmarshal(body, &parsedResponse)
	if err != nil {
		fmt.Println("Failed to marshal response to JSON")
		fmt.Println(err)
		fmt.Println(string(body))
		os.Exit(4)
	}

	return parsedResponse
}

func (w weatherData) print() {
	// Create location from timezone info
	location := time.FixedZone(w.Timezone, int(w.TimezoneOffset))

	fmt.Printf("\nLocation: %s (Lat: %.4f, Lon: %.4f)\n", w.Timezone, w.Lat, w.Lon)
	fmt.Printf("Timezone Offset: %d seconds\n\n", int(w.TimezoneOffset))

	timeFormat := "15:04:05 MST" // HH:MM:SS Timezone
	dateFormat := "2006-01-02"   // YYYY-MM-DD

	current := w.Current

	dtTime := time.Unix(current.Dt, 0).In(location)
	sunriseTime := time.Unix(current.Sunrise, 0).In(location)
	sunsetTime := time.Unix(current.Sunset, 0).In(location)

	fmt.Printf("%s  Current Weather: \n", weatherIconEmojis[current.Weather[0].Icon])
	fmt.Printf("Time:                %s %s\n", dtTime.Format(dateFormat), dtTime.Format(timeFormat))
	fmt.Printf("Sunrise:             %s\n", sunriseTime.Format(timeFormat))
	fmt.Printf("Sunset:              %s\n", sunsetTime.Format(timeFormat))
	fmt.Printf("Temperature:         %.2f°C\n", current.Temp)
	fmt.Printf("Feels Like:          %.2f°C\n", current.FeelsLike)
	fmt.Printf("Pressure:            %d hPa\n", current.Pressure)
	fmt.Printf("Humidity:            %d%%\n", current.Humidity)
	fmt.Printf("Dew Point:           %.2f°C\n", current.DewPoint)
	fmt.Printf("UV Index:            %.2f\n", current.UVI)
	fmt.Printf("Clouds:              %d%%\n", current.Clouds)
	fmt.Printf("Visibility:          %d m\n", current.Visibility)
	fmt.Printf("Wind Speed:          %.2f m/s\n", current.WindSpeed)
	fmt.Printf("Wind Degrees:        %d°\n", current.WindDeg)
	if current.WindGust > 0 {
		fmt.Printf("Wind Gust:           %.2f m/s\n", current.WindGust)
	}

	fmt.Println("-----------------------")
}

func fetchUserCoordinates() coordinate {
	fmt.Println("[@] Fetching your coordinates")

	body := fetch("https://web-api.nordvpn.com/v1/ips/info")

	var parsedResponse IPInfo
	err := json.Unmarshal(body, &parsedResponse)
	if err != nil {
		fmt.Println("Failed to parse IP info")
		fmt.Println(err)
		os.Exit(10)
	}

	return coordinate{Lat: parsedResponse.Latitude, Lon: parsedResponse.Longitude}
}

func main() {
	flag.Usage = func() {
		fmt.Printf("🌤️  weather: Know the weather from your command-line\n")

		flag.PrintDefaults()
	}

	search := flag.String("search", "", "Search for a location")
	lat := flag.Float64("lat", 0.0, "Latitude of the location")
	lon := flag.Float64("lon", 0.0, "Longitude of the location")
	auto := flag.Bool("auto", false, "Automatically fetch your weather")

	flag.Parse()

	if *auto {
		fetchUserCoordinates().findWeather().print()
	} else if *search != "" {
		searchedLocations := locationName(*search).findCoordinate()

		searchedLocations.print()

		reader := bufio.NewReader(os.Stdin)
		fmt.Print("\nChoose searched index: ")

		text, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println("Failed to read from stdin")
			fmt.Println(err)
			os.Exit(7)
		}

		text = strings.TrimSpace(text)

		chosenIndex, err := strconv.Atoi(text)
		if err != nil || chosenIndex > len(searchedLocations.Lists) || chosenIndex <= 0 {
			fmt.Println("Provided index is invalid or out of bounds.")
			os.Exit(8)
		}

		searchedLocations.Lists[chosenIndex-1].Coord.findWeather().print()
	} else if *lat != 0.0 && *lon != 0.0 {
		newCoordinate := coordinate{Lat: *lat, Lon: *lon}
		newCoordinate.findWeather().print()
	} else {
		flag.Usage()
	}
}

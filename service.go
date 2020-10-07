package main

import (
	"database/sql"
	"github.com/gin-gonic/gin"
	_ "github.com/mattn/go-sqlite3"
	"io"
	"log"
	"net/http"
	"os"
	"github.com/gin-gonic/autotls"
)

type City struct {
	Zip        string `json:"zip"`
	City       string `json:"city"`
	State      string `json:"state"`
	County     string `json:"county"`
	Timezone   string `json:"timezone"`
	Latitude   string `json:"latitude"`
	Longitude  string `json:"longitude"`
	Population string `json:"population"`
}

func main() {

	gin.SetMode(gin.ReleaseMode)

	// Logging to a file.
	f, _ := os.Create("zipcode.log")
	gin.DefaultWriter = io.MultiWriter(f)

	router := gin.Default()

	v1 := router.Group("/v1")
	{
		v1.GET("/city/:zip", cityLookup)
		v1.GET("/zip", zipLookup)
		v1.GET("/cityonly/:zip", quickCityLookup)
	}
	autotls.Run(router, "zippy-zmqa4.ondigitalocean.app")
	//router.RunTLS() // listen and serve on 0.0.0.0:8080 (for windows "localhost:8080")
}

func zipLookup(c *gin.Context){
	validResult := true
	// variables
	var cities []City
	// from param
	city := c.Query("city")

	db, err := sql.Open("sqlite3", "./zipcode.db")
	if err != nil {
		log.Fatal(err)
	}

	rows, err := db.Query("select zip, primaryCity, state, county, timezone, latitude, longitude, irsEstimatedPopulation2015 from zip_code_database where primaryCity like" + city + " and type = 'STANDARD'")
	if err != nil {
		validResult = false
	}
	defer rows.Close()

	for rows.Next() {
		city := City{}
		err = rows.Scan(&city.Zip, &city.City, &city.State, &city.County, &city.Timezone, &city.Latitude, &city.Longitude, &city.Population)
		cities = append(cities, city)
		if err != nil {
			log.Fatal(err)
		}
	}

	err = rows.Err()
	if err != nil {
		log.Fatal(err)
	}

	if validResult && len(cities) > 0{
		c.JSON(http.StatusOK, cities)
	}else {
		c.JSON(404, gin.H{"code": "NO_RESULTS", "message": "No Results were found"})
	}
}

func cityLookup(c *gin.Context) {

	validResult := true

	// variables
	var city City

	// from URL
	zip := c.Param("zip")

	// connect database
	db, err := sql.Open("sqlite3", "./zipcode.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// prepare query
	stmt, err := db.Prepare("select primaryCity, state, county, timezone, latitude, longitude, irsEstimatedPopulation2015 from zip_code_database where zip = ? and type = 'STANDARD'")
	if err != nil {
		validResult = false
	}
	defer stmt.Close()

	err = stmt.QueryRow(zip).Scan(&city.City, &city.State, &city.County, &city.Timezone, &city.Latitude, &city.Longitude, &city.Population)
	if err != nil {
		validResult = false
	}

	if len(city.City) < 0 {
		validResult = false
	}

	if validResult {
		c.JSON(http.StatusOK, city)
	}else {
		c.JSON(404, gin.H{"code": "NO_RESULTS", "message": "No Results were found"})
	}
}

func quickCityLookup(c *gin.Context) {

	validResult := true

	// variables
	var cityName string
	
	// from URL
	zip := c.Param("zip")

	// connect database
	db, err := sql.Open("sqlite3", "./zipcode.db")
	if err != nil {
		validResult = false
		c.Err()
	}
	defer db.Close()

	// prepare query
	stmt, err := db.Prepare("select primaryCity from zip_code_database where zip = ? and type = 'STANDARD'")
	if err != nil {
		validResult = false
		c.Err()
		//log.Fatal(err)
	}
	defer stmt.Close()

	err = stmt.QueryRow(zip).Scan(&cityName)
	if err != nil {
		validResult = false
		c.Err()
	}

	if validResult {
		c.JSON(http.StatusOK, gin.H{
			"city": cityName,
		})
	}else {
		c.JSON(404, gin.H{"code": "NO_RESULTS", "message": "No Results were found"})
	}
}
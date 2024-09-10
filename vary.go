package main

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"sort"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
)

const (
	appID       = "11638"
	appSecurity = "6675394085603"
)

func generateToken(params map[string]string) string {
	// Sort the parameters by key
	keys := make([]string, 0, len(params))
	for k := range params {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	// Create the string to be hashed
	var strBuilder strings.Builder
	for _, k := range keys {
		strBuilder.WriteString(fmt.Sprintf("%s=%s&", k, params[k]))
	}
	str := strings.TrimRight(strBuilder.String(), "&")

	// Append the app security to the string
	strWithAppSecurity := str + appSecurity

	// Generate the first MD5 hash
	firstHash := md5.Sum([]byte(strWithAppSecurity))
	firstHashStr := hex.EncodeToString(firstHash[:])

	// Generate the second MD5 hash
	secondHash := md5.Sum([]byte(firstHashStr))
	token := hex.EncodeToString(secondHash[:])

	return token
}

func callVariflightAPI(url string) (interface{}, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to call Variflight API: %w", err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response from Variflight API: %w", err)
	}

	// Try to unmarshal into an array
	var responseArray []interface{}
	if err := json.Unmarshal(body, &responseArray); err == nil {
		return responseArray, nil
	}

	// If it fails, try to unmarshal into a map
	var responseMap map[string]interface{}
	if err := json.Unmarshal(body, &responseMap); err == nil {
		return responseMap, nil
	}

	return nil, fmt.Errorf("failed to parse response from Variflight API")
}

func vary() {
	app := fiber.New()

	// Middleware
	app.Use(logger.New())
	app.Use(recover.New())

	// Route to handle token generation and API call by flight number
	app.Get("/api/flight", func(c *fiber.Ctx) error {
		// Extract query parameters
		params := map[string]string{
			"appid": appID,
			"fnum":  c.Query("fnum"),
			"date":  c.Query("date"),
			"lang":  c.Query("lang"),
		}

		// Generate token
		token := generateToken(params)

		// Construct the URL with the token
		urlWithToken := fmt.Sprintf("http://open-al.variflight.com/api/flight?appid=%s&fnum=%s&date=%s&lang=%s&token=%s",
			appID, c.Query("fnum"), c.Query("date"), c.Query("lang"), token)

		// Call the Variflight API
		responseJSON, err := callVariflightAPI(urlWithToken)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error":   "Failed to call Variflight API",
				"details": err.Error(),
			})
		}

		// Return the JSON response
		return c.JSON(responseJSON)
	})

	// Route to handle token generation and API call by departure and arrival codes
	app.Get("/api/flight-by-route", func(c *fiber.Ctx) error {
		// Extract query parameters
		params := map[string]string{
			"appid": appID,
			"date":  c.Query("date"),
			"dep":   c.Query("dep"),
			"arr":   c.Query("arr"),
			"lang":  c.Query("lang"),
		}

		// Generate token
		token := generateToken(params)

		// Construct the URL with the token
		urlWithToken := fmt.Sprintf("http://open-al.variflight.com/api/flight?appid=%s&date=%s&dep=%s&arr=%s&lang=%s&token=%s",
			appID, c.Query("date"), c.Query("dep"), c.Query("arr"), c.Query("lang"), token)

		// Call the Variflight API
		responseJSON, err := callVariflightAPI(urlWithToken)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error":   "Failed to call Variflight API",
				"details": err.Error(),
			})
		}

		// Return the JSON response
		return c.JSON(responseJSON)
	})

	log.Fatal(app.Listen(":3000"))
}

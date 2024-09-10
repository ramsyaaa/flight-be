package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/gofiber/fiber/v2"
)

type Meta struct {
	Message string `json:"message"`
	Code    int    `json:"code"`
	Status  string `json:"status"`
}

type Response struct {
	Meta Meta        `json:"meta"`
	Data interface{} `json:"data"`
}

type FlightRouteRequest struct {
	Dep  string `json:"dep"`
	Arr  string `json:"arr"`
	Date string `json:"date"`
}

type FlightNumberRequest struct {
	Fnum string `json:"fnum"`
	Date string `json:"date"`
}

func formatResponse(message string, code int, status string, data interface{}) Response {
	meta := Meta{
		Message: message,
		Code:    code,
		Status:  status,
	}

	return Response{
		Meta: meta,
		Data: data,
	}
}

func getVariflightData(url string) (interface{}, error) {
	response, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("Gagal melakukan permintaan ke API Variflight: %v", err)
	}
	defer response.Body.Close()

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("Gagal membaca respons dari API Variflight: %v", err)
	}

	var result interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("Respons dari API Variflight bukan JSON yang valid: %v", err)
	}

	return result, nil
}

func handleFlightRoute(c *fiber.Ctx) error {
	var req FlightRouteRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(formatResponse("Format JSON tidak valid", fiber.StatusBadRequest, "error", nil))
	}

	url := fmt.Sprintf("https://www.variflight.com/en/api/uniapi/flightroute?dep=%s&arr=%s&date=%s&type=api", req.Dep, req.Arr, req.Date)

	result, err := getVariflightData(url)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(formatResponse("Terjadi kesalahan", fiber.StatusInternalServerError, "error", fiber.Map{
			"error": err.Error(),
		}))
	}

	return c.JSON(formatResponse("Berhasil mendapatkan data rute penerbangan", fiber.StatusOK, "success", result))
}

func handleFlightNumber(c *fiber.Ctx) error {
	var req FlightNumberRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(formatResponse("Format JSON tidak valid", fiber.StatusBadRequest, "error", nil))
	}

	url := fmt.Sprintf("https://www.variflight.com/en/api/uniapi/flightnum?fnum=%s&date=%s&type=api", req.Fnum, req.Date)

	result, err := getVariflightData(url)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(formatResponse("Terjadi kesalahan", fiber.StatusInternalServerError, "error", fiber.Map{
			"error": err.Error(),
		}))
	}

	return c.JSON(formatResponse("Berhasil mendapatkan data nomor penerbangan", fiber.StatusOK, "success", result))
}

func main() {
	app := fiber.New()

	app.Post("/api/flightroute", handleFlightRoute)
	app.Post("/api/flightnum", handleFlightNumber)

	log.Fatal(app.Listen(":7000"))
}

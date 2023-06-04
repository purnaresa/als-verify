package main

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/config"
	location "github.com/aws/aws-sdk-go-v2/service/location"
)

var (
	appConfig     AppConfig
	locClient     *location.Client
	outputFile    string
	outputFileErr string
)

func init() {
	t := time.Now()
	appConfig = loadConfig("config.json")
	outputFile = fmt.Sprintf("%s_%d.csv", appConfig.OutputFile, t.Unix())
	outputFileErr = fmt.Sprintf("err_%s_%d.csv", appConfig.OutputFile, t.Unix())
	locClient = initALSClient()
}

func main() {
	input, err := readInput(appConfig.InputFile)
	if err != nil {
		log.Fatalln(err)
	}

	output := []dataOut{}
	outputErr := []dataErr{}
	for i, v := range input {
		data := dataOut{
			Index:     i,
			InputText: v.Text,
			InputLat:  v.Lat,
			InputLong: v.Long,
			Status:    "OK",
		}

		if err == nil {
			label, confidence, lat, long, err := searchPlace(v.Text)
			if err != nil {
				dataErr := dataErr{
					Index: i,
					Text:  v.Text,
					Error: err.Error(),
				}
				outputErr = append(outputErr, dataErr)
				continue
			}
			if confidence < appConfig.ConfidenceThreshold {
				data.Status = "LOW CONFIDENCE"
			}
			distance := calculateDistance(
				v.Lat,
				v.Long,
				lat,
				long)

			distanceFormated := fmt.Sprintf("%.3f", distance) //format to km with three decimal

			data.OutputText = label
			data.OutputLat = lat
			data.OutputLong = long
			data.Confidence = confidence
			data.Distance = distanceFormated

		}
		output = append(output, data)
	}

	writeOutput(output, outputFile)
	writeOutputErr(outputErr, outputFileErr)
}

// read input from csv file
func readInput(filename string) (data []dataIn, err error) {
	file, err := os.Open(filename)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	defer file.Close()

	reader := csv.NewReader(file)

	lines, err := reader.ReadAll()
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	for i, line := range lines {
		// skip header
		if i == 0 {
			continue
		}

		lat, _ := strconv.ParseFloat(strings.TrimSpace(line[1]), 64)
		long, _ := strconv.ParseFloat(strings.TrimSpace(line[2]), 64)
		d := dataIn{
			Text: line[0],
			Lat:  lat,
			Long: long,
		}
		data = append(data, d)
	}
	return
}

// write output to csv file
func writeOutput(data []dataOut, filename string) (err error) {
	file, err := os.Create(filename)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	header := []string{"Index", "InputText", "InputLat", "InputLong", "OutputText", "OutputLat", "OutputLong", "Confidence", "Distance", "Status"}
	err = writer.Write(header) // writing header
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	for _, value := range data {
		err = writer.Write([]string{
			fmt.Sprintf("%d", value.Index),
			value.InputText,
			fmt.Sprintf("%f", value.InputLat),
			fmt.Sprintf("%f", value.InputLong),
			value.OutputText,
			fmt.Sprintf("%f", value.OutputLat),
			fmt.Sprintf("%f", value.OutputLong),
			fmt.Sprintf("%f", value.Confidence),
			value.Distance,
			value.Status,
		})
		if err != nil {
			fmt.Println("Error:", err)
			return
		}
	}
	return
}

func writeOutputErr(data []dataErr, filename string) (err error) {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	header := []string{"Index", "Text", "Error"}
	if err := writer.Write(header); err != nil { // writing header
		return err
	}

	for _, value := range data {
		strData := []string{
			fmt.Sprintf("%d", value.Index),
			value.Text,
			value.Error,
		}

		if err := writer.Write(strData); err != nil {
			return err
		}
	}
	return nil
}

// get place geometry based on input
func searchPlace(text string) (label string, confidence, lat, long float64, err error) {

	input := location.SearchPlaceIndexForTextInput{
		IndexName:       &appConfig.MapIndex,
		Text:            &text,
		FilterCountries: appConfig.Countries,
	}

	output, err := locClient.SearchPlaceIndexForText(context.Background(), &input)
	if err != nil {
		log.Println(err)
		return
	}
	if len(output.Results) == 0 {
		err = fmt.Errorf("no result found")
		log.Println(err)
		return
	}
	confidence = *output.Results[0].Relevance
	label = *output.Results[0].Place.Label
	lat = output.Results[0].Place.Geometry.Point[1]
	long = output.Results[0].Place.Geometry.Point[0]
	return

}

func initALSClient() (client *location.Client) {
	cfg, err := config.LoadDefaultConfig(context.Background(),
		config.WithRegion(appConfig.ALSRegion))
	if err != nil {
		log.Fatalf("unable to load SDK config, %v", err)
	}

	client = location.NewFromConfig(cfg)
	return
}

// Types start
type AppConfig struct {
	ALSRegion           string
	MapIndex            string
	Countries           []string
	ConfidenceThreshold float64
	InputFile           string
	OutputFile          string
}

type dataIn struct {
	Text string
	Lat  float64
	Long float64
}
type dataOut struct {
	Index      int
	InputText  string
	InputLat   float64
	InputLong  float64
	OutputText string
	OutputLat  float64
	OutputLong float64
	Confidence float64
	Distance   string
	Status     string
}
type dataErr struct {
	Index int
	Text  string
	Error string
}

// Types end

// Utils start

func loadConfig(file string) (appcfg AppConfig) {

	jsonFile, err := os.Open(file)
	if err != nil {
		log.Fatalln(err)
	}
	defer jsonFile.Close()

	byteValue, _ := io.ReadAll(jsonFile)

	err = json.Unmarshal(byteValue, &appcfg)
	if err != nil {
		log.Fatalln(err)
	}

	return appcfg
}

const earthRadius = 6371

func calculateDistance(lat1, lon1, lat2, lon2 float64) float64 {
	// Convert latitude and longitude from degrees to radians
	lat1Rad := degreesToRadians(lat1)
	lon1Rad := degreesToRadians(lon1)
	lat2Rad := degreesToRadians(lat2)
	lon2Rad := degreesToRadians(lon2)

	// Haversine formula to calculate the distance
	deltaLat := lat2Rad - lat1Rad
	deltaLon := lon2Rad - lon1Rad

	a := math.Sin(deltaLat/2)*math.Sin(deltaLat/2) +
		math.Cos(lat1Rad)*math.Cos(lat2Rad)*
			math.Sin(deltaLon/2)*math.Sin(deltaLon/2)

	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
	distance := earthRadius * c

	return distance
}
func degreesToRadians(degrees float64) float64 {
	return degrees * math.Pi / 180
}

// Utils end

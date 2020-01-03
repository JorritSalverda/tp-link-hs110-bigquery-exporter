package main

import (
	"fmt"
	stdlog "log"
	"os"
	"runtime"
	"time"

	"github.com/alecthomas/kingpin"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

var (
	// set when building the application
	app       string
	version   string
	branch    string
	revision  string
	buildDate string
	goVersion = runtime.Version()

	// application specific config
	bigqueryProjectID = kingpin.Flag("bigquery-project-id", "Google Cloud project id that contains the BigQuery dataset").Envar("BQ_PROJECT_ID").Required().String()
	bigqueryDataset   = kingpin.Flag("bigquery-dataset", "Name of the BigQuery dataset").Envar("BQ_DATASET").Required().String()
	bigqueryTable     = kingpin.Flag("bigquery-table", "Name of the BigQuery table").Envar("BQ_TABLE").Required().String()
	timeoutSeconds    = kingpin.Flag("timeout-seconds", "Timeout in seconds waiting for responses from devices").Envar("TIMEOUT_SECONDS").Required().Int()
	intervalSeconds   = kingpin.Flag("interval-seconds", "Interval in seconds between 2 measurements").Envar("INTERVAL_SECONDS").Required().Int()
)

func main() {

	// parse command line parameters
	kingpin.Parse()

	initLogging()

	// log startup message
	log.Info().
		Str("branch", branch).
		Str("revision", revision).
		Str("buildDate", buildDate).
		Str("goVersion", goVersion).
		Msgf("Starting %v version %v...", app, version)

	if *timeoutSeconds >= *intervalSeconds {
		log.Fatal().Msgf("Timeout of %v seconds should be less than interval of %v seconds", *timeoutSeconds, *intervalSeconds)
	}

	bigqueryClient, err := NewBigQueryClient(*bigqueryProjectID)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed creating bigquery client")
	}

	initBigqueryTable(bigqueryClient)

	client, err := NewTPLinkClient()
	if err != nil {
		log.Fatal().Err(err).Msg("Failed creating TP-link client")
	}

	// request smart home devices every minute
	for {
		log.Info().Msg("Discovering devices...")
		devices, err := client.DiscoverDevices(*timeoutSeconds)
		if err != nil {
			log.Warn().Err(err).Msg("Failed discovering devices")
		} else {
			log.Info().Interface("devices", devices).Msg("Retrieved devices...")

			devices, err = client.GetUsageForAllDevices(devices, *timeoutSeconds)
			if err != nil {
				log.Warn().Err(err).Msg("Failed retrieving metrics for devices")
			} else {
				measurement := mapDevicesToBigQueryMeasurement(devices)
				if measurement != nil {
					log.Info().Interface("measurement", measurement).Msg("Inserting measurements into bigquery...")
					err = bigqueryClient.InsertMeasurements(*bigqueryDataset, *bigqueryTable, []BigQueryMeasurement{*measurement})
					if err != nil {
						log.Fatal().Err(err).Msg("Failed inserting measurements into bigquery table")
					}
				} else {
					log.Warn().Msg("No measurement has been recorded...")
				}
			}
		}

		sleep := *intervalSeconds - *timeoutSeconds

		log.Info().Msgf("Sleeping for %v seconds...", sleep)
		time.Sleep(time.Duration(sleep) * time.Second)
	}
}

func initLogging() {
	// log as severity for stackdriver logging to recognize the level
	zerolog.LevelFieldName = "severity"

	output := zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.RFC3339}
	output.FormatLevel = func(i interface{}) string {
		return ""
	}
	output.FormatMessage = func(i interface{}) string {
		return fmt.Sprintf("%s", i)
	}
	output.FormatFieldName = func(i interface{}) string {
		return fmt.Sprintf("| %s: ", i)
	}
	output.FormatFieldValue = func(i interface{}) string {
		return fmt.Sprintf("%s", i)
	}

	log.Logger = zerolog.New(output).With().Timestamp().Logger()

	// use zerolog for any logs sent via standard log library
	stdlog.SetFlags(0)
	stdlog.SetOutput(log.Logger)

}

func initBigqueryTable(bigqueryClient BigQueryClient) {

	log.Debug().Msgf("Checking if table %v.%v.%v exists...", *bigqueryProjectID, *bigqueryDataset, *bigqueryTable)
	tableExist := bigqueryClient.CheckIfTableExists(*bigqueryDataset, *bigqueryTable)
	if !tableExist {
		log.Debug().Msgf("Creating table %v.%v.%v...", *bigqueryProjectID, *bigqueryDataset, *bigqueryTable)
		err := bigqueryClient.CreateTable(*bigqueryDataset, *bigqueryTable, BigQueryMeasurement{}, "inserted_at", true)
		if err != nil {
			log.Fatal().Err(err).Msg("Failed creating bigquery table")
		}
	} else {
		log.Debug().Msgf("Trying to update table %v.%v.%v schema...", *bigqueryProjectID, *bigqueryDataset, *bigqueryTable)
		err := bigqueryClient.UpdateTableSchema(*bigqueryDataset, *bigqueryTable, BigQueryMeasurement{})
		if err != nil {
			log.Fatal().Err(err).Msg("Failed updating bigquery table schema")
		}
	}
}

package main

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"
)

const (
	JWT_TOKEN    = "your-jwt-token-here"
	CSV_FILENAME = "AtomicRescheduleimpact.csv"
)

type ImpactedAppointment struct {
	AppointmentID  string
	RequestID      int64
	EventTimestamp string
}

type RescheduleEventPayload struct {
	RequestID      int64  `json:"request_id"`
	EventTimestamp string `json:"event_timestamp"`
}

func readAtomicRescheduleImpact() []ImpactedAppointment {
	// Open the CSV file
	file, err := os.Open(CSV_FILENAME)
	if err != nil {
		log.Fatal("Error opening file:", err)
	}
	defer file.Close()

	// Create a new CSV reader
	reader := csv.NewReader(file)

	// Read all records
	records, err := reader.ReadAll()
	if err != nil {
		log.Fatal("Error reading CSV:", err)
	}

	if len(records) < 1 {
		log.Fatal("CSV file is empty")
	}

	// Get headers from first row
	headers := records[0]

	// Create slice of maps for the data
	var result []ImpactedAppointment

	// Convert each record to a map using headers as keys
	for _, record := range records[1:] {
		recordMap := make(map[string]string)
		for i, value := range record {
			if i < len(headers) {
				recordMap[headers[i]] = value
			}
		}
		result = append(result, createImpactedAppointment(recordMap))
	}

	return result
}

func createImpactedAppointment(record map[string]string) ImpactedAppointment {
	// Check for APPOINTMENT_ID
	appointmentID, ok := record["APPOINTMENT_ID"]
	if !ok {
		log.Fatalf("APPOINTMENT_ID field missing from record")
	}

	// Check for REQUEST_ID
	id, ok := record["REQUEST_ID"]
	if !ok {
		log.Fatalf("REQUEST_ID field missing from record")
	}

	var requestID int64
	if _, err := fmt.Sscanf(id, "%d", &requestID); err != nil {
		log.Fatalf("invalid REQUEST_ID format: %v", err)
	}

	// Check for TIMESTAMP_UTC
	timestamp, ok := record["TIMESTAMP_UTC"]
	if !ok {
		log.Fatalf("TIMESTAMP_UTC field missing from record")
	}

	layout := "2006-01-02 15:04:05"
	t, err := time.Parse(layout, timestamp)
	if err != nil {
		log.Fatalf("error parsing timestamp: %v", err)
	}

	return ImpactedAppointment{
		AppointmentID:  appointmentID,
		RequestID:      requestID,
		EventTimestamp: t.Format(time.RFC3339),
	}
}

func createPayload(impactedAppointment ImpactedAppointment) RescheduleEventPayload {
	return RescheduleEventPayload{
		RequestID:      impactedAppointment.RequestID,
		EventTimestamp: impactedAppointment.EventTimestamp,
	}
}

func makeRequest(payload RescheduleEventPayload) (*http.Request, error) {
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("error marshaling JSON: %v", err)
	}

	url := "https://api2-private.east.zocdoccloud.com/synchronizer-updates/v1/reschedule-events"
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("error creating request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+JWT_TOKEN)
	return req, nil
}

func logRequestBody(req *http.Request) error {
	bodyBytes, err := io.ReadAll(req.Body)
	if err != nil {
		return fmt.Errorf("error reading request body: %v", err)
	}
	// Restore the body for the actual request
	req.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
	log.Printf("Sending request body: %s", string(bodyBytes))
	return nil
}

func executeRequest(req *http.Request) error {
	if err := logRequestBody(req); err != nil {
		return err
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("error sending request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("request failed with status: %d", resp.StatusCode)
	}
	return nil
}

func sendRescheduleEvent(impactedAppointment ImpactedAppointment) error {
	payload := createPayload(impactedAppointment)

	req, err := makeRequest(payload)
	if err != nil {
		return fmt.Errorf("error preparing request: %v", err)
	}

	if err := executeRequest(req); err != nil {
		return fmt.Errorf("error executing request: %v", err)
	}
	return nil
}

func setupLogging() *os.File {
	logFile, err := os.OpenFile("reschedule_backfill.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("Failed to setup logging: %v", err)
	}

	// Create a multi writer that writes to both stdout and the file
	multiWriter := io.MultiWriter(os.Stdout, logFile)
	log.SetOutput(multiWriter)

	return logFile
}

func main() {
	logFile := setupLogging()
	defer logFile.Close()

	records := readAtomicRescheduleImpact()
	for i, record := range records {
		if err := sendRescheduleEvent(record); err != nil {
			log.Fatalf("Error processing record %d (Appointment ID: %s): %v", i+1, record.AppointmentID, err)
		}
		log.Printf("Successfully processed record %d (Appointment ID: %s)", i+1, record.AppointmentID)
	}
}

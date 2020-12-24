package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/robfig/cron"
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	//"github.com/robfig/cron"
	"github.com/gorilla/mux"
	"net/http"
)

func main() {
	fmt.Println("Starting Reservation Application")
	l, _ := time.LoadLocation("America/Denver")
	cronTimer := cron.New(cron.WithLocation(l), cron.WithSeconds())
	cronTimer.AddFunc("5 0 11 * * *", func() {
		ProcessValidRequests()
	})

	cronTimer.Start()

	r := mux.NewRouter()
	r.HandleFunc("/et/reservation", processReservationRequest).Methods(http.MethodPost)
	r.HandleFunc("/test", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Received Reservation Request"))
	})

	http.Handle("/", r)
	log.Fatalln(http.ListenAndServe(":8080", nil))
	//Set Cron job to run every day 1 in the afternoon

}

func parseCurrentTime() (string, string) {
	t := time.Now()
	fmt.Println("Executing job at " + t.String())
	mst, err := time.LoadLocation("America/Denver")
	//Check if valid location
	if err != nil {
		fmt.Println(err)
	}
	localTime := t.In(mst)
	fmt.Println(localTime.Format("2006-01-02 3:4:5 PM"))
	hour := localTime.Hour()
	//Account for military time which website doesn't use
	if hour != 12 {
		hour = hour % 12
	}
	var amPm string
	formattedTime := localTime.Format("2006-01-02 3:4:5 PM")
	if strings.Contains(formattedTime, "PM") {
		amPm = "PM"
	} else {
		amPm = "AM"
	}
	stringHour := strconv.Itoa(hour)
	return stringHour, amPm
}

func processReservationRequest(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Received Reservation Request")
	var reservationRequest ReservationRequest
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&reservationRequest); err != nil {
		w.Write([]byte(err.Error()))
	}

	defer r.Body.Close()
	writeRequestToFile(reservationRequest)
	w.Write([]byte("Reservation Successfully Processed"))
}

func writeRequestToFile(request ReservationRequest) {
	f, err := os.OpenFile("reservations.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Println(err)
	}

	defer f.Close()
	stringRequest, err := json.Marshal(request)
	if err != nil {
		log.Println("Error Marshalling Request to String to write to file")
		return
	}
	if _, err := f.WriteString((string)(stringRequest) + "\n"); err != nil {
		log.Println(err)
	}
}

func ProcessValidRequests() {

	f, err := os.Open("reservations.txt")
	var newFile string
	if err != nil {
		fmt.Println(err)
	}

	scanner := bufio.NewScanner(f)
	var currentRequest ReservationRequest
	for scanner.Scan() {
		line := scanner.Text()
		err = json.Unmarshal([]byte(line), &currentRequest)
		if err != nil {
			fmt.Println("Error Unmarhsalling line in file to Reservation Request Struct")
		}

		fmt.Println("Going to submit " + line + " for processing")
		//Hopefully send it over as a post to Java container
		//resp, err := http.Post("http://localhost:8082/attemptreservation", "application/json", bytes.NewBufferString(line))
		resp, err := http.Post("http://172.17.0.2:8082/attemptreservation", "application/json", bytes.NewBufferString(line))
		if err != nil {
			log.Println(err)
		}
		responseBody, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			fmt.Println("Error reading response body from server")
		}
		fmt.Println("Response code from Reservation request was " + resp.Status)
		fmt.Println("Response message from reservation request was \n" + (string)(responseBody))
		fmt.Println("And then remove entry " + line + " from file")
	}

	fmt.Println("Contents of New File are " + newFile)
	f.Close()
	replaceFile, err := os.OpenFile("reservations.txt", os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
	replaceFile.Write([]byte(newFile))
	replaceFile.Close()
}

type ReservationRequest struct {
	Time        string `json:"time"`
	Ampm        string `json:"ampm"`
	Week        string `json:"week"`
	Day         string `json:"day"`
	UserDetails struct {
		FirstName   string `json:"firstName"`
		MiddleName  string `json:"middleName"`
		LastName    string `json:"lastName"`
		BirthYear   string `json:"birthYear"`
		BirthMonth  string `json:"birthMonth"`
		BirthDay    string `json:"birthDay"`
		Email       string `json:"email"`
		PhoneNumber string `json:"phoneNumber"`
	} `json:"userDetails"`
}

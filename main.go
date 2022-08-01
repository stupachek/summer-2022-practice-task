package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"
)

const (
	file              = "./data.json"
	priceCriteria     = "price"
	arrivalCriteria   = "arrival-time"
	departureCriteria = "departure-time"
)

type Trains []Train

type Train struct {
	TrainID            int
	DepartureStationID int
	ArrivalStationID   int
	Price              float32
	ArrivalTime        time.Time
	DepartureTime      time.Time
}

func prompt(label string) string {
	r := bufio.NewReader(os.Stdin)
	fmt.Print(label + " ")
	s, err := r.ReadString('\n')
	if err != nil {
		handleError(err)
	}
	return strings.TrimSpace(s)
}

func handleError(err error) {
	if err != nil {
		fmt.Printf("error: %s\n", err)
		os.Exit(1)
	}
}

func main() {
	departure := prompt("enter depature station:")
	arrival := prompt("enter arrival station:")
	criteria := prompt("enter criteria:")
	trains, err := FindTrains(departure, arrival, criteria)
	handleError(err)
	for _, j := range trains {
		fmt.Printf("%+v,\n", j)
	}
}

func FindTrains(departureStation, arrivalStation, criteria string) (Trains, error) {
	trains, err := readTrains(file)
	if err != nil {
		return nil, err
	}
	departure, arrival, crit, err := parseParams(departureStation, arrivalStation, criteria)
	if err != nil {
		return nil, err
	}
	filtered := filterTrains(trains, departure, arrival)
	switch crit {
	case priceCriteria:
		sort.Slice(filtered, filtered.byPrice)
	case arrivalCriteria:
		sort.Slice(filtered, filtered.byArrival)
	case departureCriteria:
		sort.Slice(filtered, filtered.byDeparture)
	}
	if len(filtered) >= 3 {
		return filtered[:3], nil
	}
	return filtered, nil // маєте повернути правильні значення
}

func (t Trains) byPrice(i, j int) bool {
	return t[i].Price < t[j].Price
}
func (t Trains) byArrival(i, j int) bool {
	return t[i].ArrivalTime.Before(t[j].ArrivalTime)
}
func (t Trains) byDeparture(i, j int) bool {
	return t[i].DepartureTime.Before(t[j].DepartureTime)
}

func filterTrains(trains Trains, departure int, arrival int) Trains {
	var filtered Trains
	for _, train := range trains {
		if train.DepartureStationID == departure && train.ArrivalStationID == arrival {
			filtered = append(filtered, train)
		}
	}
	return filtered
}

func readTrains(file string) (Trains, error) {
	jsonFile, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer jsonFile.Close()
	decoder := json.NewDecoder(jsonFile)
	var trains Trains
	err = decoder.Decode(&trains)
	if err != nil {
		return nil, err
	}
	return trains, nil
}

func parseParams(departureStation, arrivalStation, criteria string) (departure, arrival int, crit string, err error) {
	if departureStation == "" {
		return 0, 0, "", errors.New("empty departure station")
	}
	if arrivalStation == "" {
		return 0, 0, "", errors.New("empty arrival station")
	}
	departure, err = strconv.Atoi(departureStation)
	if err != nil {
		return 0, 0, "", err
	}
	arrival, err = strconv.Atoi(arrivalStation)
	if err != nil {
		return 0, 0, "", err
	}
	if departure <= 0 {
		return 0, 0, "", errors.New("bad departure station input")
	}
	if arrival <= 0 {
		return 0, 0, "", errors.New("bad arrival station input")
	}
	switch criteria {
	case priceCriteria, arrivalCriteria, departureCriteria:
		crit = criteria
	default:
		return 0, 0, "", errors.New("unsupported criteria")
	}
	return departure, arrival, crit, nil
}

type timeJSON time.Time

func (t *timeJSON) UnmarshalJSON(b []byte) error {
	var s string
	err := json.Unmarshal(b, &s)
	if err != nil {
		return err
	}
	t2, err := time.Parse("15:04:05", s)
	if err != nil {
		return err
	}
	*t = timeJSON(t2)
	return nil
}

func (t *Train) UnmarshalJSON(b []byte) error {
	type trainJSON struct {
		TrainID            int
		DepartureStationID int
		ArrivalStationID   int
		Price              float32
		ArrivalTime        timeJSON
		DepartureTime      timeJSON
	}

	var tjson trainJSON
	err := json.Unmarshal(b, &tjson)
	if err != nil {
		return err
	}
	t.TrainID = tjson.TrainID
	t.DepartureStationID = tjson.DepartureStationID
	t.ArrivalStationID = tjson.ArrivalStationID
	t.Price = tjson.Price
	t.ArrivalTime = time.Time(tjson.ArrivalTime)
	t.DepartureTime = time.Time(tjson.DepartureTime)
	return nil
}

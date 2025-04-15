package testdata

import (
	"fmt"
	"maryan_api/config"
	"maryan_api/internal/entity"
	"math/rand"
	"time"

	"github.com/d3code/uuid"
	"gorm.io/gorm"
)

func CreateTestData(db *gorm.DB) {
	drivers := entity.TestDrivers()
	err := db.Create(drivers).Error
	if err != nil {
		panic(err)
	}

	buses := entity.TestBuses()

	for i, _ := range buses {
		buses[i].LeadDriverID = uuid.NullUUID{drivers[i*2].ID, true}
		buses[i].AssistantDriverID = uuid.NullUUID{drivers[(i*2)+1].ID, true}
	}

	err = db.Create(buses).Error
	if err != nil {
		panic(err)
	}

	var line = 100
	countries, ukraineID := config.GetCountries()
	var busIndex int
	fmt.Println(len(countries))
	for _, countryID := range countries {
		if countryID == ukraineID {
			continue
		}

		departureTime := time.Now()
		var trips = make([]entity.Trip, 50)
		for i := 0; i < 50; i++ {

			tripiID := uuid.New()
			outboundConnectionID := uuid.New()
			returnConnectionID := uuid.New()
			trips[i] = entity.Trip{
				ID: tripiID,
				OutboundConnection: entity.Connection{
					ID:                   outboundConnectionID,
					Line:                 line,
					Price:                (rand.Intn(250) + 100) * 100,
					DepartureCountryID:   ukraineID,
					DestinationCountryID: countryID,
					DepartureTime:        departureTime,
					ArrivalTime:          departureTime.Add(time.Hour*15 + time.Hour*time.Duration(rand.Intn(20))),
					BusID:                buses[busIndex].ID,
					Updates: []entity.ConnectionUpdate{{
						ConnectionID: outboundConnectionID,
						Status:       entity.RegisteredConnectionStatus,
					}},
					Type: entity.ComertialConnectionType,
				},
				ReturnConnection: entity.Connection{
					ID:                   returnConnectionID,
					Line:                 line,
					Price:                (rand.Intn(250) + 100) * 100,
					DepartureCountryID:   countryID,
					DestinationCountryID: ukraineID,
					DepartureTime:        departureTime.Add(time.Hour * 60),
					ArrivalTime:          departureTime.Add(time.Hour*60 + time.Hour*15 + time.Hour*time.Duration(rand.Intn(20))),
					BusID:                buses[busIndex].ID,
					Updates: []entity.ConnectionUpdate{{
						ConnectionID: returnConnectionID,
						Status:       entity.RegisteredConnectionStatus,
					}},
					Type: entity.ComertialConnectionType,
				},
				Updates: []entity.TripUpdate{{
					TripID: tripiID,
					Status: entity.TripStatusRegistered,
				}},
			}

			departureTime = departureTime.Add(time.Hour * 24 * 10)
		}
		err = db.Create(trips).Error
		if err != nil {
			panic(err)
		}

		line += 100
		busIndex++
		departureTime = time.Now()
	}
}

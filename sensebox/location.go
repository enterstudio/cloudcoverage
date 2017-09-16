package sensebox

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
)

type Location struct {
	Longitude, Latitude float64
}

type loc struct {
	Geom struct {
		Coordinates []float64 `json:"coordinates"`
	} `json:"geometry"`
}

type boxStub struct {
	Locs []loc `json:"loc"`
	swag string
}

func (l *loc) toLocation() *Location {
	return &Location{Longitude: l.Geom.Coordinates[0], Latitude: l.Geom.Coordinates[1]}
}

// QueryLocation requests the location of the senseBox from the openSenseMap API
func (b *Sensebox) QueryLocation() (location *Location, err error) {
	res, err := http.Get(b.boxURL)
	if err != nil {
		log.Fatal(err)
	}
	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return
	}
	var jsonBox boxStub
	err = json.Unmarshal(body, &jsonBox)
	if err != nil {
		return
	}

	location = jsonBox.Locs[0].toLocation()

	return
}

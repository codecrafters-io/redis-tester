package data_structures

import (
	"math"

	"github.com/codecrafters-io/tester-utils/random"
)

const (
	LATITUDE_MAX           = 85.05112878
	LONGITUDE_MAX          = 180.0
	LATITUDE_MIN           = -LATITUDE_MAX
	LONGITUDE_MIN          = -LONGITUDE_MAX
	EARTH_RADIUS_IN_METERS = 6372797.560856
)

type Coordinates struct {
	Latitude  float64
	Longitude float64
}

// Location : I'll remove this comment later
// I'm not sure if making a strucutre like `LocationSet` is a good idea
// Given that the data will not encapsulate any new information regarding the locations (unlike ZSET where order, etc is preserved)
// I didn't find it necessary to create one
// However, I found myself repeating the same logic {}Location[] -> {}string[] for each location
// so, i'm not sure if a utility function will suffice or we should create a datastructure like `LocationSet`
type Location struct {
	Coordinates *Coordinates
	Name        string
}

func NewLocation(name string, coordinates Coordinates) *Location {
	return &Location{
		Coordinates: &coordinates,
		Name:        name,
	}
}

func (l *Location) GetLatitude() float64 {
	return l.Coordinates.Latitude
}

func (l *Location) GetLongitude() float64 {
	return l.Coordinates.Longitude
}

// GetGeoGridCenterCoordinates returns the coordiantes of the center of the geogrid that
// the location falls in
func (l *Location) GetGeoGridCenterCoordinates() Coordinates {
	geoCode := l.GetGeoCode()
	return decodeGeoCodeToCoordinates(geoCode)
}

// GetGeoCode returns the [WGS84](https://en.wikipedia.org/wiki/World_Geodetic_System) code of a location
// This is the same geocode used by Redis
func (l *Location) GetGeoCode() uint64 {
	// Normalize to the range 0-2^26
	latitudeOffset := (l.GetLatitude() - LATITUDE_MIN) / (LATITUDE_MAX - LATITUDE_MIN)
	longitudeOffset := (l.GetLongitude() - LONGITUDE_MIN) / (LONGITUDE_MAX - LONGITUDE_MIN)

	latitudeOffset *= (1 << 26)
	longitudeOffset *= (1 << 26)

	nLatitude := uint64(latitudeOffset)
	nLongitude := uint64(longitudeOffset)

	// Spread latitude bits
	x := nLatitude
	x = (x | (x << 16)) & 0x0000FFFF0000FFFF
	x = (x | (x << 8)) & 0x00FF00FF00FF00FF
	x = (x | (x << 4)) & 0x0F0F0F0F0F0F0F0F
	x = (x | (x << 2)) & 0x3333333333333333
	x = (x | (x << 1)) & 0x5555555555555555

	// Spread longitude bits
	y := nLongitude
	y = (y | (y << 16)) & 0x0000FFFF0000FFFF
	y = (y | (y << 8)) & 0x00FF00FF00FF00FF
	y = (y | (y << 4)) & 0x0F0F0F0F0F0F0F0F
	y = (y | (y << 2)) & 0x3333333333333333
	y = (y | (y << 1)) & 0x5555555555555555

	return x | (y << 1)
}

func (l *Location) CalculateDistance(location *Location) float64 {
	l1 := l.GetGeoGridCenterCoordinates()
	l2 := location.GetGeoGridCenterCoordinates()

	lat1radians := degToRad(l1.Latitude)
	lat2radians := degToRad(l2.Latitude)
	lon1radians := degToRad(l1.Longitude)
	lon2radians := degToRad(l2.Longitude)

	v := math.Sin((lon2radians - lon1radians) / 2)
	u := math.Sin((lat2radians - lat1radians) / 2)

	a := u*u + math.Cos(lat1radians)*math.Cos(lat2radians)*v*v
	return 2.0 * EARTH_RADIUS_IN_METERS * math.Asin(math.Sqrt(a))
}

// GenerateRandomLocations generates 'count' number of locations
func GenerateRandomLocations(count int) []*Location {
	result := make([]*Location, count)
	locationNames := random.RandomWords(count)
	for i := range count {
		result[i] = &Location{
			Name: locationNames[i],
			Coordinates: &Coordinates{
				Latitude:  random.RandomFloat64(LATITUDE_MIN, LATITUDE_MAX),
				Longitude: random.RandomFloat64(LONGITUDE_MIN, LONGITUDE_MAX),
			},
		}
	}
	return result
}

// decodeGeoCodeToCoordinates decodes a geocode and returns the coordinates of
// the center of the geocode's decoded area
func decodeGeoCodeToCoordinates(geoCode uint64) Coordinates {
	y := geoCode >> 1
	x := geoCode

	// Compact bits back to 32-bit ints
	x = geoCode & 0x5555555555555555
	x = (x | (x >> 1)) & 0x3333333333333333
	x = (x | (x >> 2)) & 0x0F0F0F0F0F0F0F0F
	x = (x | (x >> 4)) & 0x00FF00FF00FF00FF
	x = (x | (x >> 8)) & 0x0000FFFF0000FFFF
	x = (x | (x >> 16)) & 0x00000000FFFFFFFF

	y = y & 0x5555555555555555
	y = (y | (y >> 1)) & 0x3333333333333333
	y = (y | (y >> 2)) & 0x0F0F0F0F0F0F0F0F
	y = (y | (y >> 4)) & 0x00FF00FF00FF00FF
	y = (y | (y >> 8)) & 0x0000FFFF0000FFFF
	y = (y | (y >> 16)) & 0x00000000FFFFFFFF

	lat_scale := LATITUDE_MAX - LATITUDE_MIN
	lon_scale := LONGITUDE_MAX - LONGITUDE_MIN

	ilato := uint32(x)
	ilono := uint32(y)

	gridLatitudeMin := LATITUDE_MIN + lat_scale*(float64(ilato)*1.0/(1<<26))
	gridLatitudeMax := LATITUDE_MIN + lat_scale*(float64(ilato+1)*1.0/(1<<26))
	gridLongitudeMin := LONGITUDE_MIN + lon_scale*(float64(ilono+1)*1.0/(1<<26))
	gridLongitudeMax := LONGITUDE_MIN + lon_scale*(float64(ilono+1)*1.0/(1<<26))

	latitude := (gridLatitudeMin + gridLatitudeMax) / 2
	longitude := (gridLongitudeMin + gridLongitudeMax) / 2

	// Clamp to bounds
	if latitude > LATITUDE_MAX {
		latitude = LATITUDE_MAX
	}
	if latitude < LATITUDE_MIN {
		latitude = LATITUDE_MIN
	}
	if longitude > LONGITUDE_MAX {
		longitude = LONGITUDE_MAX
	}
	if longitude < LONGITUDE_MIN {
		longitude = LONGITUDE_MIN
	}

	return Coordinates{
		Latitude:  latitude,
		Longitude: longitude,
	}
}

func degToRad(deg float64) float64 {
	return deg * math.Pi / 180
}

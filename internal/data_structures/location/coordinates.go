package location

import (
	"fmt"
	"math"
)

const (
	LATITUDE_MAX           = 85.05112878
	LONGITUDE_MAX          = 180.0
	LATITUDE_MIN           = -LATITUDE_MAX
	LONGITUDE_MIN          = -LONGITUDE_MAX
	EARTH_RADIUS_IN_METERS = 6372797.560856
)

type Coordinates struct {
	latitude  float64
	longitude float64
}

func NewCoordinates(latitude float64, longitude float64) Coordinates {
	if (latitude >= LATITUDE_MAX) && (latitude <= LATITUDE_MIN) {
		panic(fmt.Sprintf("Codecrafters Internal Error - Invalid latitude %.8f in NewCoordinates()", latitude))
	}
	if (longitude >= LONGITUDE_MAX) && (longitude <= LONGITUDE_MIN) {
		panic(fmt.Sprintf("Codecrafters Internal Error - Invalid longitude %.8f in NewCoordinates()", longitude))
	}
	return Coordinates{
		latitude:  latitude,
		longitude: longitude,
	}
}

func (c Coordinates) GetLatitude() float64 {
	return c.latitude
}

func (c Coordinates) GetLongitude() float64 {
	return c.longitude
}

// GetGeoGridCenterCoordinates decodes latitude and longitude from the geoCode
// The decoded coordinates will be slightly different from the original one.
// It is due to the fact that Redis drops precision at 52 bits.
// The coordinates of the returned coordinate is the center of the smallest geo grid the coordinates
// are a part of
func (c Coordinates) GetGeoGridCenterCoordinates() Coordinates {
	geoCode := c.GetGeoCode()
	return decodeGeoCodeToCoordinates(geoCode)
}

// GetGeoCode returns the [WGS84](https://en.wikipedia.org/wiki/World_Geodetic_System) geocode of a coordinate pair
// This is the same geocode used by Redis
func (c Coordinates) GetGeoCode() uint64 {
	// Normalize to the range 0-2^26
	latitudeOffset := (c.GetLatitude() - LATITUDE_MIN) / (LATITUDE_MAX - LATITUDE_MIN)
	longitudeOffset := (c.GetLongitude() - LONGITUDE_MIN) / (LONGITUDE_MAX - LONGITUDE_MIN)

	latitudeOffset *= (1 << 26)
	longitudeOffset *= (1 << 26)

	// Spread latitude bits
	x := uint64(latitudeOffset)
	x = (x | (x << 16)) & 0x0000FFFF0000FFFF
	x = (x | (x << 8)) & 0x00FF00FF00FF00FF
	x = (x | (x << 4)) & 0x0F0F0F0F0F0F0F0F
	x = (x | (x << 2)) & 0x3333333333333333
	x = (x | (x << 1)) & 0x5555555555555555

	// Spread longitude bits
	y := uint64(longitudeOffset)
	y = (y | (y << 16)) & 0x0000FFFF0000FFFF
	y = (y | (y << 8)) & 0x00FF00FF00FF00FF
	y = (y | (y << 4)) & 0x0F0F0F0F0F0F0F0F
	y = (y | (y << 2)) & 0x3333333333333333
	y = (y | (y << 1)) & 0x5555555555555555

	return x | (y << 1)
}

// DistanceFrom returns distance between two pair of coordinates using haversine great circle distance formula
// While calculating distance, the coordinates actually used is the center of the geogrid instead of the
// original coordinates. It is done to mimic's Redis' way of calculating distance
func (c Coordinates) DistanceFrom(coordinates Coordinates) float64 {
	c1 := c.GetGeoGridCenterCoordinates()
	c2 := coordinates.GetGeoGridCenterCoordinates()

	lat1radians := degreesToRadians(c1.latitude)
	lat2radians := degreesToRadians(c2.latitude)
	lon1radians := degreesToRadians(c1.longitude)
	lon2radians := degreesToRadians(c2.longitude)

	v := math.Sin((lon2radians - lon1radians) / 2)
	u := math.Sin((lat2radians - lat1radians) / 2)

	a := u*u + math.Cos(lat1radians)*math.Cos(lat2radians)*v*v
	return 2.0 * EARTH_RADIUS_IN_METERS * math.Asin(math.Sqrt(a))
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

	latitude_scale := LATITUDE_MAX - LATITUDE_MIN
	longitude_scale := LONGITUDE_MAX - LONGITUDE_MIN

	gridLatitudeNumber := uint32(x)
	gridLongitudeNumber := uint32(y)

	gridLatitudeMin := LATITUDE_MIN + latitude_scale*(float64(gridLatitudeNumber)*1.0/(1<<26))
	gridLatitudeMax := LATITUDE_MIN + latitude_scale*(float64(gridLatitudeNumber+1)*1.0/(1<<26))
	gridLongitudeMin := LONGITUDE_MIN + longitude_scale*(float64(gridLongitudeNumber)*1.0/(1<<26))
	gridLongitudeMax := LONGITUDE_MIN + longitude_scale*(float64(gridLongitudeNumber+1)*1.0/(1<<26))

	latitude := (gridLatitudeMin + gridLatitudeMax) / 2
	longitude := (gridLongitudeMin + gridLongitudeMax) / 2

	// Clamp to bounds
	// While there is no scenario in which these cases will be met (this function is private and will be called using a valid
	// value of geoCode, let's keep the checks and corrections to mimic's Redis behavior in case we need to make this public
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

	return NewCoordinates(latitude, longitude)
}

func degreesToRadians(deg float64) float64 {
	return deg * math.Pi / 180
}

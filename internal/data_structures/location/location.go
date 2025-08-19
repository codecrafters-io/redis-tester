package location

import (
	"math"
)

type Location struct {
	Coordinates Coordinates
	Name        string
}

func (l Location) GetLatitude() float64 {
	return l.Coordinates.latitude
}

func (l Location) GetLongitude() float64 {
	return l.Coordinates.longitude
}

// GetGeoGridCenterCoordinates decodes latitude and longitude from the geoCode
// The decoded coordinates will be slightly different from the original one.
// It is due to the fact that Redis drops precision at 52 bits.
// The coordinates of the returned location is the center of the smallest geo grid the location
// lies in
func (l Location) GetGeoGridCenterCoordinates() Coordinates {
	geoCode := l.GetGeoCode()
	return decodeGeoCodeToCoordinates(geoCode)
}

// GetGeoCode returns the [WGS84](https://en.wikipedia.org/wiki/World_Geodetic_System) code of a location
// This is the same geocode used by Redis
func (l Location) GetGeoCode() uint64 {
	// Normalize to the range 0-2^26
	latitudeOffset := (l.GetLatitude() - LATITUDE_MIN) / (LATITUDE_MAX - LATITUDE_MIN)
	longitudeOffset := (l.GetLongitude() - LONGITUDE_MIN) / (LONGITUDE_MAX - LONGITUDE_MIN)

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

// DistanceFrom returns distance between two locations using haversine great circle distance formula
// While calculating distance, the locations actually used is the center of the geogrid instead of the
// coordinates of the location. It is done to mimic's Redis' way of calculating distance
func (l Location) DistanceFrom(location Location) float64 {
	l1 := l.GetGeoGridCenterCoordinates()
	l2 := location.GetGeoGridCenterCoordinates()

	lat1radians := degreesToRadians(l1.latitude)
	lat2radians := degreesToRadians(l2.latitude)
	lon1radians := degreesToRadians(l1.longitude)
	lon2radians := degreesToRadians(l2.longitude)

	v := math.Sin((lon2radians - lon1radians) / 2)
	u := math.Sin((lat2radians - lat1radians) / 2)

	a := u*u + math.Cos(lat1radians)*math.Cos(lat2radians)*v*v
	return 2.0 * EARTH_RADIUS_IN_METERS * math.Asin(math.Sqrt(a))
}

func degreesToRadians(deg float64) float64 {
	return deg * math.Pi / 180
}

package data_structures

import (
	"fmt"
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
	latitude  float64
	longitude float64
}

func NewCoordinates(latitude float64, longitude float64) Coordinates {
	isLatitudeValid := (latitude >= LATITUDE_MIN) && (latitude <= LATITUDE_MAX)
	isLongitudeValid := (longitude >= LONGITUDE_MIN) && (longitude <= LONGITUDE_MAX)
	if isLatitudeValid && isLongitudeValid {
		return Coordinates{
			latitude:  latitude,
			longitude: longitude,
		}
	}
	panic(fmt.Sprintf("Codecrafters Internal Error - Invalid latitude, longitude pair (%.6f, %.6f) in NewCoordinates()", latitude, longitude))
}

func (c Coordinates) GetLatitude() float64 {
	return c.latitude
}

func (c Coordinates) GetLongitude() float64 {
	return c.longitude
}

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

type LocationSet struct {
	locations []Location
}

func NewLocationSet() *LocationSet {
	return &LocationSet{}
}

func (ls *LocationSet) AddLocation(location Location) *LocationSet {
	ls.locations = append(ls.locations, location)
	return ls
}

func (ls *LocationSet) Size() int {
	return len(ls.locations)
}

// Center returns a location whose latitude and longitude are respectively the mean-value of latitude and longitude of all the locations in the set.
// This is different from circumcenter of a spherical triangle (https://brsr.github.io/2021/05/02/spherical-triangle-centers.html)
// It is done because we want to include some and exclude other locations while testing for geosearch with a fixed radius
func (ls *LocationSet) Center(centerLocationName string) Location {
	latitudeAverage := 0.0
	longitudeAverage := 0.0
	for _, location := range ls.locations {
		latitudeAverage += location.GetLatitude()
		longitudeAverage += location.GetLongitude()
	}

	latitudeAverage = latitudeAverage / float64(ls.Size())
	longitudeAverage = longitudeAverage / float64(ls.Size())

	return Location{
		Name:        centerLocationName,
		Coordinates: NewCoordinates(latitudeAverage, longitudeAverage),
	}
}

// ClosestTo returns the location in the LocationSet that is closest to the supplied location
func (ls *LocationSet) ClosestTo(location Location) Location {
	if ls.Size() == 0 {
		panic("Codecrafters Internal Error - Cannot find closest location from empty LocationSet")
	}

	closestLocation := ls.locations[0]
	closestDistance := location.DistanceFrom(closestLocation)

	for _, loc := range ls.locations {
		distance := location.DistanceFrom(loc)
		if distance < closestDistance {
			closestDistance = distance
			closestLocation = loc
		}
	}

	return closestLocation
}

// FarthestFrom returns the location in the LocationSet that is farthest from the supplied location
func (ls *LocationSet) FarthestFrom(location Location) Location {
	if ls.Size() == 0 {
		panic("Codecrafters Internal Error - Cannot find farthest location from empty LocationSet")
	}

	farthestLocation := ls.locations[0]
	farthestDistance := location.DistanceFrom(farthestLocation)

	for _, loc := range ls.locations {
		distance := location.DistanceFrom(loc)
		if distance > farthestDistance {
			farthestDistance = distance
			farthestLocation = loc
		}
	}
	return farthestLocation
}

// WithinRadius returns a new LocationSet with all the locations that are within a given radius from the given location
func (ls *LocationSet) WithinRadius(referenceLocation Location, radius float64) *LocationSet {
	result := NewLocationSet()

	for _, location := range ls.locations {
		distance := referenceLocation.DistanceFrom(location)
		if distance <= radius {
			result.AddLocation(location)
		}
	}

	return result
}

// GetLocations returns a copy of all the locations in the location set
func (ls *LocationSet) GetLocations() []Location {
	locations := make([]Location, len(ls.locations))
	copy(locations, ls.locations)
	return locations
}

// GetLocationNames returns the name of all the locations in the location set
func (ls *LocationSet) GetLocationNames() []string {
	locationNames := make([]string, len(ls.locations))
	for i, location := range ls.locations {
		locationNames[i] = location.Name
	}
	return locationNames
}

// GenerateRandomLocationSet returns a LocationSet with 'count' number of random locations
func GenerateRandomLocationSet(count int) *LocationSet {
	locationSet := NewLocationSet()
	locationNames := random.RandomWords(count)

	for i := range count {
		latitude := random.RandomFloat64(LATITUDE_MIN, LATITUDE_MAX)
		longitude := random.RandomFloat64(LONGITUDE_MIN, LONGITUDE_MAX)
		locationSet.AddLocation(Location{
			Name:        locationNames[i],
			Coordinates: NewCoordinates(latitude, longitude),
		})
	}
	return locationSet
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

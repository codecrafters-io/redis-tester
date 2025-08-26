package location

import (
	"github.com/codecrafters-io/tester-utils/random"
)

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

// CenterCoordinates returns a Coordinate pair whose latitude and longitude are respectively the mean-value of
// latitude and longitude of all the locations in the set.
// This is different center (not equidistant from all points) compared to the circumcenter of a spherical triangle
// (https://brsr.github.io/2021/05/02/spherical-triangle-centers.html), which is equidistant from all the points
// It is done because we want to include some and exclude other locations while testing for geosearch
func (ls *LocationSet) CenterCoordinates() Coordinates {
	latitudeAverage := 0.0
	longitudeAverage := 0.0

	for _, location := range ls.locations {
		latitudeAverage += location.GetLatitude()
		longitudeAverage += location.GetLongitude()
	}

	latitudeAverage = latitudeAverage / float64(ls.Size())
	longitudeAverage = longitudeAverage / float64(ls.Size())

	return NewCoordinates(latitudeAverage, longitudeAverage)
}

// ClosestTo returns the location in the LocationSet that is closest to the supplied coordinates
func (ls *LocationSet) ClosestTo(referenceCoordinates Coordinates) Location {
	if ls.Size() == 0 {
		panic("Codecrafters Internal Error - Cannot find closest location from empty LocationSet")
	}

	closestLocation := ls.locations[0]
	closestDistance := referenceCoordinates.DistanceFrom(closestLocation.Coordinates)

	for _, loc := range ls.locations {
		distance := referenceCoordinates.DistanceFrom(loc.Coordinates)

		if distance < closestDistance {
			closestDistance = distance
			closestLocation = loc
		}
	}

	return closestLocation
}

// FarthestFrom returns the location in the LocationSet that is farthest from the supplied coordinates
func (ls *LocationSet) FarthestFrom(referenceCoordinates Coordinates) Location {
	if ls.Size() == 0 {
		panic("Codecrafters Internal Error - Cannot find farthest location from empty LocationSet")
	}

	farthestLocation := ls.locations[0]
	farthestDistance := farthestLocation.Coordinates.DistanceFrom(referenceCoordinates)

	for _, location := range ls.locations {
		distance := referenceCoordinates.DistanceFrom(location.Coordinates)

		if distance > farthestDistance {
			farthestDistance = distance
			farthestLocation = location
		}
	}

	return farthestLocation
}

// WithinRadius returns a new LocationSet with all the locations that are within a given radius from the given location
func (ls *LocationSet) WithinRadius(referenceCoordinates Coordinates, radius float64) *LocationSet {
	result := NewLocationSet()

	for _, location := range ls.locations {
		distance := referenceCoordinates.DistanceFrom(location.Coordinates)

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

// GenerateRandomLocationSet returns a LocationSet with 'count' number of valid random locations
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

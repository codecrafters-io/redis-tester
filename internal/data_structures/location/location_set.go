package location

import "github.com/codecrafters-io/tester-utils/random"

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

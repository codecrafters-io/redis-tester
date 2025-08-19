package location

type Location struct {
	Coordinates Coordinates
	Name        string
}

func (l Location) GetLatitude() float64 {
	return l.Coordinates.Latitude
}

func (l Location) GetLongitude() float64 {
	return l.Coordinates.Longitude
}

// GetGeoGridCenterCoordinates decodes latitude and longitude from the geoCode of a location
func (l Location) GetGeoGridCenterCoordinates() Coordinates {
	return l.Coordinates.GetGeoGridCenterCoordinates()
}

// GetGeoCode returns the [WGS84](https://en.wikipedia.org/wiki/World_Geodetic_System) geocode of a location
func (l Location) GetGeoCode() uint64 {
	return l.Coordinates.GetGeoCode()
}

// DistanceFrom returns distance between two locations
func (l Location) DistanceFrom(location Location) float64 {
	return l.Coordinates.DistanceFrom(location.Coordinates)
}

// AsRedisCommandArgs converts a location struct to string slice
// The order is: <longitude> <latitude> <name>
// This is the same order used in Redis CLI
func (l Location) AsRedisCommandArgs() []string {
	return append(l.Coordinates.AsRedisCommandArgs(), l.Name)
}

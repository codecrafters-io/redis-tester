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

// LongitudeAsRedisCommandArg converts a location's longitude to its string representation
func (l Location) LongitudeAsRedisCommandArg() string {
	return l.Coordinates.LongitudeAsRedisCommandArg()
}

// LatitudeAsRedisCommandArg converts a location's latitude to its string representation
func (l Location) LatitudeAsRedisCommandArg() string {
	return l.Coordinates.LatitudeAsRedisCommandArg()
}

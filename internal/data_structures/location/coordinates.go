package location

import "fmt"

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

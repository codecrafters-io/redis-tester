package resp_assertions

import (
	"github.com/fatih/color"
)

func ColorizeString(colorToUse color.Attribute, msg string) string {
	c := color.New(colorToUse)
	return c.Sprint(msg)
}

func BuildExpectedVsReceivedErrorMessage(expectedValue string, receivedValue string) string {
	errorMsg := ColorizeString(color.FgGreen, "Expected:")
	errorMsg += " \"" + expectedValue + "\""
	errorMsg += "\n"
	errorMsg += ColorizeString(color.FgRed, "Received:")
	errorMsg += " \"" + receivedValue + "\""
	return errorMsg
}

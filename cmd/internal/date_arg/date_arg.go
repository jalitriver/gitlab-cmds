// This file allows dates in the form of "YYYY/MM/DD" to be present on
// the command-line or in XML files and automatically parsed by the
// "flag" or "xml" package the same as an intrinsic type.

package date_arg

import (
	"encoding/xml"
	"fmt"
	"time"
)

type DateArg time.Time

////////////////////////////////////////////////////////////////////////
// flag.Value implementation
////////////////////////////////////////////////////////////////////////

// Set sets parses the string setting the date.  This method is part
// of the flag.Value interface need by the "flag" package to parse
// dates present on the command line.
func (d *DateArg) Set(s string) error {
	var date time.Time
	var err error

	// Use time.Now() to get the current timezone/location.
	now := time.Now()

	// Try to parse the date using the first allowed format.
	date, err = time.ParseInLocation("2006-01-02", s, now.Location())
	if err == nil {
		*d = DateArg(date)
		return nil
	}

	// Try to parse the date using the second allowed format.
	date, err = time.ParseInLocation("2006/01/02", s, now.Location())
	if err == nil {
		*d = DateArg(date)
		return nil
	}

	return err
}

// String returns the string representation of the date.  This method
// is part of the flag.Value interface need by the "flag" package to
// parse dates present on the command line.
func (d *DateArg) String() string {
	date := time.Time(*d)
	return fmt.Sprintf("%d-%0m-%0d", date.Year, date.Month, date.Day)
}

////////////////////////////////////////////////////////////////////////
// xml.Marshaler and xml.Unmarshaler implementation
////////////////////////////////////////////////////////////////////////

// MarshalXML marshals the element to XML.  This method is part of the
// xml.Marshaler interface need by the "xml" package to parse dates
// present on the command line.
func (d *DateArg) MarshalXML(encoder *xml.Encoder, start xml.StartElement) error {
	return encoder.EncodeElement(d.String(), start)
}

// UnmarshalXML unmarshals the element from XML.  This method is part
// of the xml.Unmarshaler interface need by the "xml" package to parse
// dates present on the command line.
func (d *DateArg) UnmarshalXML(decoder *xml.Decoder, start xml.StartElement) error {
	var s string

	// Read the element into a string.
	err := decoder.DecodeElement(&s, &start)
	if  err != nil {
		return err
	}

	// Parse the string.
	return d.Set(s)
}

package common

import "strings"

// SchoolNameMap maps variations of school names from CSV files to standardized names in the database
var SchoolNameMap = map[string]string{
	"Saint Mary's":              "Saint Mary's (CA)",
	"St. John's":                "St. John's (NY)",
	"Hartford":                  "Hartford Hawks",
	"San Diego St":              "San Diego State",
	"Morehead St":               "Morehead State",
	"Washington St":             "Washington State",
	"Iowa St":                   "Iowa State",
	"South Dakota St":           "South Dakota State",
	"Colorado St":               "Colorado State",
	"Montana St":                "Montana State",
	"Utah St":                   "Utah State",
	"McNeese St":                "McNeese State",
	"St. Peter's":               "Saint Peter's",
	"Boise St":                  "Boise State",
	"North Carolina St":         "NC State",
	"Mississippi St":            "Mississippi State",
	"Michigan St":               "Michigan State",
	"St. Mary's":                "Saint Mary's (CA)",
	"Long Beach St":             "Long Beach State",
	"Appalachian St":            "Appalachian State",
	"VCU":                       "Virginia Commonwealth",
	"Florida St":                "Florida State",
	"Norfolk St":                "Norfolk State",
	"Wichita St":                "Wichita State",
	"USC":                       "Southern California",
	"SMU":                       "Southern Methodist",
	"Penn":                      "Pennsylvania",
	"UMBC":                      "Maryland-Baltimore County",
	"Cleveland St":              "Cleveland State",
	"UCSB":                      "UC Santa Barbara",
	"BYU":                       "Brigham Young",
	"UNC":                       "North Carolina",
	"Pitt":                      "Pittsburgh",
	"ETSU":                      "East Tennessee State",
	"UC-Davis":                  "UC Davis",
	"LIU":                       "Long Island University",
	"LSU":                       "Louisiana State",
	"Louisiana St":              "Louisiana State",
	"Oklahoma St":               "Oklahoma State",
	"UConn":                     "Connecticut",
	"St. Joseph's":              "Saint Joseph's",
	"Ole Miss":                  "Mississippi",
	"UC-Irvine":                 "UC Irvine",
	"North Carolina Greensboro": "UNC Greensboro",
	"Oregon St":                 "Oregon State",
	"Ohio St":                   "Ohio State",
}

// GetStandardizedSchoolName returns the standardized name for a school
func GetStandardizedSchoolName(name string) string {
	// Skip summary rows
	if name == "TOTALS" {
		return ""
	}

	// Remove any asterisks
	name = strings.TrimSuffix(name, "*")

	if standardized, ok := SchoolNameMap[name]; ok {
		return standardized
	}
	return name
}

// IsValidTeamName checks if a team name is valid (not a summary row)
func IsValidTeamName(name string) bool {
	return name != "TOTALS"
}

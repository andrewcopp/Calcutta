package common

import "strings"

// SchoolNameMap maps variations of school names from CSV files to standardized names in the database
var SchoolNameMap = map[string]string{
	"Alabama St.":        "Alabama State",
	"Appalachian St.":    "Appalachian State",
	"Arizona St.":        "Arizona State",
	"Boise St.":          "Boise State",
	"Cal St. Fullerton":  "Cal State Fullerton",
	"Cleveland St.":      "Cleveland State",
	"Colorado St.":       "Colorado State",
	"Colgate":            "Colgate",
	"Colgate ":           "Colgate",
	"Eastern Wash.":      "Eastern Washington",
	"ETSU":               "East Tennessee State",
	"Florida St.":        "Florida State",
	"Georgia St.":        "Georgia State",
	"Iowa St.":           "Iowa State",
	"Jacksonville St.":   "Jacksonville State",
	"Jax. State":         "Jacksonville State",
	"Kansas St.":         "Kansas State",
	"Kent St.":           "Kent State",
	"Kennesaw St.":       "Kennesaw State",
	"Loyola-Chicago":     "Loyola Chicago",
	"LSU":                "Louisiana State",
	"Miami (FL)":         "Miami",
	"Miami (Fl.)":        "Miami",
	"Miami of Florida":   "Miami",
	"Mich. State":        "Michigan State",
	"Michigan St.":       "Michigan State",
	"Mississippi St.":    "Mississippi State",
	"Mississippi State":  "Mississippi State",
	"Mississippi State ": "Mississippi State",
	"Montana St.":        "Montana State",
	"Morehead St.":       "Morehead State",
	"MTSU":               "Middle Tennessee State",
	"Mt. St. Mary's":     "Mount St. Mary's",
	"Murray St.":         "Murray State",
	"N. Kentucky":        "Northern Kentucky",
	"N. M. St.":          "New Mexico State",
	"N.C. Central":       "North Carolina Central",
	"N.C. State":         "North Carolina State",
	"NC State":           "North Carolina State",
	"New Mexico St.":     "New Mexico State",
	"Norfolk St.":        "Norfolk State",
	"North Dakota St.":   "North Dakota State",
	"Ohio St.":           "Ohio State",
	"Oklahoma St.":       "Oklahoma State",
	"Oregon St.":         "Oregon State",
	"Penn St.":           "Penn State",
	"Play-in losers":     "Play-in Losers",
	"Saint Bonaventure":  "St. Bonaventure",
	"Saint John's":       "St. John's",
	"Saint Louis":        "Saint Louis",
	"Saint Mary's":       "Saint Mary's",
	"San Diego St.":      "San Diego State",
	"SIUE":               "SIU Edwardsville",
	"South Dakota St.":   "South Dakota State",
	"St. Bonaventure":    "St. Bonaventure",
	"St. John's":         "St. John's",
	"St. Louis":          "Saint Louis",
	"St. Mary's":         "Saint Mary's",
	"Texas A & M":        "Texas A&M",
	"Texas A&M-CC":       "Texas A&M Corpus Christi",
	"Texas So.":          "Texas Southern",
	"U.C. Davis":         "UC Davis",
	"UAB":                "Alabama-Birmingham",
	"UCSB":               "UC Santa Barbara",
	"UNC":                "North Carolina",
	"UNC Greensboro":     "UNC Greensboro",
	"UNC-Wilmington":     "UNC Wilmington",
	"UNCW":               "UNC Wilmington",
	"Utah St.":           "Utah State",
	"Va. Tech":           "Virginia Tech",
	"Wichita St.":        "Wichita State",
	"Xavier*":            "Xavier",
	"North Carolina*":    "North Carolina",
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

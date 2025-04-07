package common

import "strings"

// SchoolNameMap maps variations of school names from CSV files to standardized names in the database
var SchoolNameMap = map[string]string{
	"Alabama St.":              "Alabama State",
	"Appalachian St.":          "Appalachian State",
	"Arizona St.":              "Arizona State",
	"Boise St.":                "Boise State",
	"Cal St. Fullerton":        "Cal State Fullerton",
	"Cleveland St.":            "Cleveland State",
	"Colorado St.":             "Colorado State",
	"Colgate":                  "Colgate",
	"Colgate ":                 "Colgate",
	"Eastern Wash.":            "Eastern Washington",
	"ETSU":                     "East Tennessee State",
	"Florida St.":              "Florida State",
	"Georgia St.":              "Georgia State",
	"Iowa St.":                 "Iowa State",
	"Jacksonville St.":         "Jacksonville State",
	"Jax. State":               "Jacksonville State",
	"Kansas St.":               "Kansas State",
	"Kent St.":                 "Kent State",
	"Kennesaw St.":             "Kennesaw State",
	"Loyola-Chicago":           "Loyola (IL)",
	"Loyola Chicago":           "Loyola (IL)",
	"LSU":                      "Louisiana State",
	"Miami (FL)":               "Miami (FL)",
	"Miami (Fl.)":              "Miami (FL)",
	"Miami of Florida":         "Miami (FL)",
	"Miami":                    "Miami (FL)",
	"Mich. State":              "Michigan State",
	"Michigan St.":             "Michigan State",
	"Mississippi St.":          "Mississippi State",
	"Mississippi State":        "Mississippi State",
	"Mississippi State ":       "Mississippi State",
	"Montana St.":              "Montana State",
	"Morehead St.":             "Morehead State",
	"MTSU":                     "Middle Tennessee",
	"Middle Tennessee State":   "Middle Tennessee",
	"Mt. St. Mary's":           "Mount St. Mary's",
	"Murray St.":               "Murray State",
	"N. Kentucky":              "Northern Kentucky",
	"N. M. St.":                "New Mexico State",
	"N.C. Central":             "North Carolina Central",
	"N.C. State":               "NC State",
	"NC State":                 "NC State",
	"North Carolina State":     "NC State",
	"New Mexico St.":           "New Mexico State",
	"Norfolk St.":              "Norfolk State",
	"North Dakota St.":         "North Dakota State",
	"Ohio St.":                 "Ohio State",
	"Oklahoma St.":             "Oklahoma State",
	"Oregon St.":               "Oregon State",
	"Penn St.":                 "Penn State",
	"Penn":                     "Pennsylvania",
	"Saint Bonaventure":        "St. Bonaventure",
	"Saint John's":             "St. John's (NY)",
	"Saint Louis":              "Saint Louis",
	"Saint Mary's":             "Saint Mary's (CA)",
	"San Diego St.":            "San Diego State",
	"SIUE":                     "Southern Illinois-Edwardsville",
	"SIU Edwardsville":         "Southern Illinois-Edwardsville",
	"South Dakota St.":         "South Dakota State",
	"St. Bonaventure":          "St. Bonaventure",
	"St. John's":               "St. John's (NY)",
	"St. Louis":                "Saint Louis",
	"St. Mary's":               "Saint Mary's (CA)",
	"Texas A & M":              "Texas A&M",
	"Texas A&M-CC":             "Texas A&M-Corpus Christi",
	"Texas A&M Corpus Christi": "Texas A&M-Corpus Christi",
	"Texas So.":                "Texas Southern",
	"U.C. Davis":               "UC Davis",
	"UAB":                      "UAB",
	"Alabama-Birmingham":       "UAB",
	"UCSB":                     "UC Santa Barbara",
	"UNC":                      "North Carolina",
	"UNC Greensboro":           "UNC Greensboro",
	"UNC-Wilmington":           "UNC Wilmington",
	"UNCW":                     "UNC Wilmington",
	"Utah St.":                 "Utah State",
	"Va. Tech":                 "Virginia Tech",
	"Wichita St.":              "Wichita State",
	"Xavier*":                  "Xavier",
	"North Carolina*":          "North Carolina",
	"SMU":                      "Southern Methodist",
	"USC":                      "Southern California",
	"VCU":                      "Virginia Commonwealth",
	"FGCU":                     "Florida Gulf Coast",
	"UMBC":                     "Maryland-Baltimore County",
	"BYU":                      "Brigham Young",
	"Hartford":                 "Hartford Hawks",
	"Charleston":               "College of Charleston",
	"Ole Miss":                 "Mississippi",
	"Saint Francis":            "Saint Francis (PA)",
	"Wright St.":               "Wright State",
	"Grambling State":          "Grambling",
	"McNeese":                  "McNeese State",
	"Prairie View A&M":         "Prairie View",
	"UConn":                    "Connecticut",
	"Fairleigh Dickinson":      "FDU",
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

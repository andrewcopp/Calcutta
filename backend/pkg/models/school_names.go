package models

// SchoolNameMap maps various names of schools to their canonical form
var SchoolNameMap = map[string]string{
	// Original mappings
	"Saint Mary's":      "Saint Mary's",
	"St. Bonaventure":   "St. Bonaventure",
	"UNC Greensboro":    "UNC Greensboro",
	"Colgate":           "Colgate",
	"Mississippi State": "Mississippi State",
	"Mount St. Mary's":  "Mount St. Mary's",
	"St. John's":        "St. John's",

	// New mappings from tournament data
	"FGCU":                     "Florida Gulf Coast",
	"North Carolina State":     "NC State",
	"UMBC":                     "Maryland-Baltimore County",
	"BYU":                      "Brigham Young",
	"Alabama-Birmingham":       "UAB",
	"Wright St.":               "Wright State",
	"Charleston":               "College of Charleston",
	"Ole Miss":                 "Mississippi",
	"Miami":                    "Miami (FL)",
	"USC":                      "Southern California",
	"Loyola Chicago":           "Loyola (IL)",
	"Prairie View A&M":         "Prairie View",
	"VCU":                      "Virginia Commonwealth",
	"Middle Tennessee State":   "Middle Tennessee",
	"Play-in Losers":           "Play-in Losers", // Special case
	"Hartford":                 "Hartford Hawks",
	"Grambling State":          "Grambling",
	"SIU Edwardsville":         "Southern Illinois-Edwardsville",
	"SMU":                      "Southern Methodist",
	"Penn":                     "Pennsylvania",
	"Fairleigh Dickinson":      "FDU",
	"UConn":                    "Connecticut",
	"Texas A&M Corpus Christi": "Texas A&M-Corpus Christi",
	"McNeese":                  "McNeese State",
	"Saint Francis":            "St. Francis",
}

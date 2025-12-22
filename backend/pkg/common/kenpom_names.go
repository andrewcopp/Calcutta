package common

import "strings"

var KenPomNameMap = map[string]string{
	"Omaha":                    "Nebraska Omaha",
	"Saint Francis (PA)":       "Saint Francis",
	"Alabama State":            "Alabama St.",
	"Iowa State":               "Iowa St.",
	"San Diego State":          "San Diego St.",
	"Washington State":         "Washington St.",
	"Mississippi State":        "Mississippi St.",
	"Michigan State":           "Michigan St.",
	"Boise State":              "Boise St.",
	"NC State":                 "N.C. State",
	"McNeese State":            "McNeese St.",
	"Morehead State":           "Morehead St.",
	"Long Beach State":         "Long Beach St.",
	"South Dakota State":       "South Dakota St.",
	"Montana State":            "Montana St.",
	"Grambling":                "Grambling St.",
	"Kansas State":             "Kansas St.",
	"Miami (FL)":               "Miami FL",
	"Penn State":               "Penn St.",
	"Arizona State":            "Arizona St.",
	"Kennesaw State":           "Kennesaw St.",
	"Southeast Missouri State": "Southeast Missouri St.",
	"FDU":                      "Fairleigh Dickinson",
	"Texas A&M-Corpus Christi": "Texas A&M Corpus Chris",
	"Murray State":             "Murray St.",
	"Ohio State":               "Ohio St.",
	"Loyola (IL)":              "Loyola Chicago",
	"New Mexico State":         "New Mexico St.",
	"Cal State Fullerton":      "Cal St. Fullerton",
	"Jacksonville State":       "Jacksonville St.",
	"Wright State":             "Wright St.",
	"Georgia State":            "Georgia St.",
	"Oklahoma State":           "Oklahoma St.",
	"Florida State":            "Florida St.",
	"Wichita State":            "Wichita St.",
	"Oregon State":             "Oregon St.",
	"Cleveland State":          "Cleveland St.",
	"Appalachian State":        "Appalachian St.",
	"Prairie View":             "Prairie View A&M",
	"Gardner-Webb":             "Gardner Webb",
	"North Dakota State":       "North Dakota St.",
	"Long Island University":   "LIU Brooklyn",
	"East Tennessee State":     "East Tennessee St.",
}

func GetKenPomTeamName(schoolName string) string {
	schoolName = strings.TrimSpace(schoolName)
	if v, ok := KenPomNameMap[schoolName]; ok {
		return v
	}
	return schoolName
}

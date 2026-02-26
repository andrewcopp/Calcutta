package dtos

type SeedAnalytics struct {
	Seed                 int     `json:"seed"`
	TotalPoints          float64 `json:"totalPoints"`
	TotalInvestment      float64 `json:"totalInvestment"`
	PointsPercentage     float64 `json:"pointsPercentage"`
	InvestmentPercentage float64 `json:"investmentPercentage"`
	TeamCount            int     `json:"teamCount"`
	AveragePoints        float64 `json:"averagePoints"`
	AverageInvestment    float64 `json:"averageInvestment"`
	ROI                  float64 `json:"roi"`
}

type RegionAnalytics struct {
	Region               string  `json:"region"`
	TotalPoints          float64 `json:"totalPoints"`
	TotalInvestment      float64 `json:"totalInvestment"`
	PointsPercentage     float64 `json:"pointsPercentage"`
	InvestmentPercentage float64 `json:"investmentPercentage"`
	TeamCount            int     `json:"teamCount"`
	AveragePoints        float64 `json:"averagePoints"`
	AverageInvestment    float64 `json:"averageInvestment"`
	ROI                  float64 `json:"roi"`
}

type TeamAnalytics struct {
	SchoolID          string  `json:"schoolId"`
	SchoolName        string  `json:"schoolName"`
	TotalPoints       float64 `json:"totalPoints"`
	TotalInvestment   float64 `json:"totalInvestment"`
	Appearances       int     `json:"appearances"`
	AveragePoints     float64 `json:"averagePoints"`
	AverageInvestment float64 `json:"averageInvestment"`
	AverageSeed       float64 `json:"averageSeed"`
	ROI               float64 `json:"roi"`
}

type SeedVarianceAnalytics struct {
	Seed             int     `json:"seed"`
	InvestmentStdDev float64 `json:"investmentStdDev"`
	PointsStdDev     float64 `json:"pointsStdDev"`
	InvestmentMean   float64 `json:"investmentMean"`
	PointsMean       float64 `json:"pointsMean"`
	InvestmentCV     float64 `json:"investmentCV"`
	PointsCV         float64 `json:"pointsCV"`
	TeamCount        int     `json:"teamCount"`
	VarianceRatio    float64 `json:"varianceRatio"`
}

type SeedInvestmentPoint struct {
	Seed             int     `json:"seed"`
	TournamentName   string  `json:"tournamentName"`
	TournamentYear   int     `json:"tournamentYear"`
	CalcuttaID       string  `json:"calcuttaId"`
	TeamID           string  `json:"teamId"`
	SchoolName       string  `json:"schoolName"`
	TotalBid         float64 `json:"totalBid"`
	CalcuttaTotalBid float64 `json:"calcuttaTotalBid"`
	NormalizedBid    float64 `json:"normalizedBid"`
}

type SeedInvestmentSummary struct {
	Seed   int     `json:"seed"`
	Count  int     `json:"count"`
	Mean   float64 `json:"mean"`
	StdDev float64 `json:"stdDev"`
	Min    float64 `json:"min"`
	Q1     float64 `json:"q1"`
	Median float64 `json:"median"`
	Q3     float64 `json:"q3"`
	Max    float64 `json:"max"`
}

type SeedInvestmentDistributionResponse struct {
	Items     []SeedInvestmentPoint   `json:"items"`
	Summaries []SeedInvestmentSummary `json:"summaries"`
}

type BestTeam struct {
	TournamentName   string  `json:"tournamentName"`
	TournamentYear   int     `json:"tournamentYear"`
	CalcuttaID       string  `json:"calcuttaId"`
	TeamID           string  `json:"teamId"`
	SchoolName       string  `json:"schoolName"`
	Seed             int     `json:"seed"`
	Region           string  `json:"region"`
	TeamPoints       float64 `json:"teamPoints"`
	TotalBid         float64 `json:"totalBid"`
	CalcuttaTotalBid float64 `json:"calcuttaTotalBid"`
	CalcuttaTotalPts float64 `json:"calcuttaTotalPoints"`
	InvestmentShare  float64 `json:"investmentShare"`
	PointsShare      float64 `json:"pointsShare"`
	RawROI           float64 `json:"rawROI"`
	NormalizedROI    float64 `json:"normalizedROI"`
}

type BestTeamsResponse struct {
	Items []BestTeam `json:"items"`
}

type BestInvestment struct {
	TournamentName   string  `json:"tournamentName"`
	TournamentYear   int     `json:"tournamentYear"`
	CalcuttaID       string  `json:"calcuttaId"`
	TeamID           string  `json:"teamId"`
	SchoolName       string  `json:"schoolName"`
	Seed             int     `json:"seed"`
	Region           string  `json:"region"`
	TeamPoints       float64 `json:"teamPoints"`
	TotalBid         float64 `json:"totalBid"`
	CalcuttaTotalBid float64 `json:"calcuttaTotalBid"`
	CalcuttaTotalPts float64 `json:"calcuttaTotalPoints"`
	InvestmentShare  float64 `json:"investmentShare"`
	PointsShare      float64 `json:"pointsShare"`
	RawROI           float64 `json:"rawROI"`
	NormalizedROI    float64 `json:"normalizedROI"`
}

type BestInvestmentsResponse struct {
	Items []BestInvestment `json:"items"`
}

type InvestmentLeaderboardRow struct {
	TournamentName      string  `json:"tournamentName"`
	TournamentYear      int     `json:"tournamentYear"`
	CalcuttaID          string  `json:"calcuttaId"`
	EntryID             string  `json:"entryId"`
	EntryName           string  `json:"entryName"`
	TeamID              string  `json:"teamId"`
	SchoolName          string  `json:"schoolName"`
	Seed                int     `json:"seed"`
	Investment          float64 `json:"investment"`
	OwnershipPercentage float64 `json:"ownershipPercentage"`
	RawReturns          float64 `json:"rawReturns"`
	NormalizedReturns   float64 `json:"normalizedReturns"`
}

type InvestmentLeaderboardResponse struct {
	Items []InvestmentLeaderboardRow `json:"items"`
}

type EntryLeaderboardRow struct {
	TournamentName    string  `json:"tournamentName"`
	TournamentYear    int     `json:"tournamentYear"`
	CalcuttaID        string  `json:"calcuttaId"`
	EntryID           string  `json:"entryId"`
	EntryName         string  `json:"entryName"`
	TotalReturns      float64 `json:"totalReturns"`
	TotalParticipants int     `json:"totalParticipants"`
	AverageReturns    float64 `json:"averageReturns"`
	NormalizedReturns float64 `json:"normalizedReturns"`
}

type EntryLeaderboardResponse struct {
	Items []EntryLeaderboardRow `json:"items"`
}

type CareerLeaderboardRow struct {
	EntryName              string `json:"entryName"`
	Years                  int    `json:"years"`
	BestFinish             int    `json:"bestFinish"`
	Wins                   int    `json:"wins"`
	Podiums                int    `json:"podiums"`
	InTheMoneys            int    `json:"inTheMoneys"`
	Top10s                 int    `json:"top10s"`
	CareerEarningsCents    int    `json:"careerEarningsCents"`
	ActiveInLatestCalcutta bool   `json:"activeInLatestCalcutta"`
}

type CareerLeaderboardResponse struct {
	Items []CareerLeaderboardRow `json:"items"`
}

type AnalyticsResponse struct {
	SeedAnalytics         []SeedAnalytics         `json:"seedAnalytics,omitempty"`
	RegionAnalytics       []RegionAnalytics       `json:"regionAnalytics,omitempty"`
	TeamAnalytics         []TeamAnalytics         `json:"teamAnalytics,omitempty"`
	SeedVarianceAnalytics []SeedVarianceAnalytics `json:"seedVarianceAnalytics,omitempty"`
	TotalPoints           float64                 `json:"totalPoints"`
	TotalInvestment       float64                 `json:"totalInvestment"`
	BaselineROI           float64                 `json:"baselineROI"`
}

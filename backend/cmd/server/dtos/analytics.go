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

type AnalyticsResponse struct {
	SeedAnalytics         []SeedAnalytics         `json:"seedAnalytics,omitempty"`
	RegionAnalytics       []RegionAnalytics       `json:"regionAnalytics,omitempty"`
	TeamAnalytics         []TeamAnalytics         `json:"teamAnalytics,omitempty"`
	SeedVarianceAnalytics []SeedVarianceAnalytics `json:"seedVarianceAnalytics,omitempty"`
	TotalPoints           float64                 `json:"totalPoints"`
	TotalInvestment       float64                 `json:"totalInvestment"`
	BaselineROI           float64                 `json:"baselineROI"`
}

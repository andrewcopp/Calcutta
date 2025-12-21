export interface SeedAnalytics {
  seed: number;
  totalPoints: number;
  totalInvestment: number;
  pointsPercentage: number;
  investmentPercentage: number;
  teamCount: number;
  averagePoints: number;
  averageInvestment: number;
  roi: number;
}

export interface RegionAnalytics {
  region: string;
  totalPoints: number;
  totalInvestment: number;
  pointsPercentage: number;
  investmentPercentage: number;
  teamCount: number;
  averagePoints: number;
  averageInvestment: number;
  roi: number;
}

export interface TeamAnalytics {
  schoolId: string;
  schoolName: string;
  totalPoints: number;
  totalInvestment: number;
  appearances: number;
  averagePoints: number;
  averageInvestment: number;
  averageSeed: number;
  roi: number;
}

export interface SeedVarianceAnalytics {
  seed: number;
  investmentStdDev: number;
  pointsStdDev: number;
  investmentMean: number;
  pointsMean: number;
  investmentCV: number;
  pointsCV: number;
  teamCount: number;
  varianceRatio: number;
}

export interface AnalyticsResponse {
  seedAnalytics?: SeedAnalytics[];
  regionAnalytics?: RegionAnalytics[];
  teamAnalytics?: TeamAnalytics[];
  seedVarianceAnalytics?: SeedVarianceAnalytics[];
  totalPoints: number;
  totalInvestment: number;
  baselineROI: number;
}

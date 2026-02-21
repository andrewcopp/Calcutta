import React from 'react';
import { CalcuttaPortfolio, CalcuttaPortfolioTeam } from '../../types/calcutta';
import { PortfolioScores } from './PortfolioScores';

type StatisticsTabProps = {
  portfolios: CalcuttaPortfolio[];
  portfolioTeams: CalcuttaPortfolioTeam[];
};

export function StatisticsTab({ portfolios, portfolioTeams }: StatisticsTabProps) {
  return <>{portfolios.length > 0 && <PortfolioScores portfolio={portfolios[0]} teams={portfolioTeams} />}</>;
}

import { CalcuttaPortfolio, CalcuttaPortfolioTeam } from '../../types/calcutta';

export const StatisticsTab: React.FC<{
  portfolios: CalcuttaPortfolio[];
  portfolioTeams: CalcuttaPortfolioTeam[];
  PortfolioScoresComponent: React.FC<{ portfolio: CalcuttaPortfolio; teams: CalcuttaPortfolioTeam[] }>;
}> = ({ portfolios, portfolioTeams, PortfolioScoresComponent }) => {
  return <>{portfolios.length > 0 && <PortfolioScoresComponent portfolio={portfolios[0]} teams={portfolioTeams} />}</>;
};

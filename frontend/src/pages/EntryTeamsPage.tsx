import { useEffect, useState } from 'react';
import { Link, useParams } from 'react-router-dom';
import { CalcuttaEntryTeam, CalcuttaPortfolio, CalcuttaPortfolioTeam, School } from '../types/calcutta';
import { calcuttaService } from '../services/calcuttaService';

// Add a new section to display portfolio scores
const PortfolioScores: React.FC<{ portfolio: CalcuttaPortfolio; teams: CalcuttaPortfolioTeam[] }> = ({
  portfolio,
  teams,
}) => {
  return (
    <div className="bg-white shadow rounded-lg p-6 mb-6">
      <h3 className="text-lg font-semibold mb-4">Portfolio Scores</h3>
      <div className="grid grid-cols-1 gap-4">
        <div className="flex justify-between items-center">
          <span className="text-gray-600">Maximum Possible Score:</span>
          <span className="font-medium">{portfolio.maximumPoints.toFixed(2)}</span>
        </div>
        <div className="border-t pt-4">
          <h4 className="text-md font-medium mb-2">Team Scores</h4>
          <div className="space-y-2">
            {teams.map((team) => (
              <div key={team.id} className="flex justify-between items-center">
                <span className="text-gray-600">{team.team?.school?.name || 'Unknown Team'}</span>
                <div className="text-right">
                  <div className="text-sm text-gray-500">
                    Expected: {team.expectedPoints.toFixed(2)}
                  </div>
                  <div className="text-sm text-gray-500">
                    Predicted: {team.predictedPoints.toFixed(2)}
                  </div>
                </div>
              </div>
            ))}
          </div>
        </div>
      </div>
    </div>
  );
};

export function EntryTeamsPage() {
  const { entryId, calcuttaId } = useParams<{ entryId: string; calcuttaId: string }>();
  const [teams, setTeams] = useState<CalcuttaEntryTeam[]>([]);
  const [schools, setSchools] = useState<School[]>([]);
  const [portfolios, setPortfolios] = useState<CalcuttaPortfolio[]>([]);
  const [portfolioTeams, setPortfolioTeams] = useState<CalcuttaPortfolioTeam[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    const fetchData = async () => {
      if (!entryId || !calcuttaId) {
        setError('Missing required parameters');
        setLoading(false);
        return;
      }
      
      try {
        const [teamsData, schoolsData, portfoliosData] = await Promise.all([
          calcuttaService.getEntryTeams(entryId, calcuttaId),
          calcuttaService.getSchools(),
          calcuttaService.getPortfoliosByEntry(entryId)
        ]);

        // Create a map of schools by ID for quick lookup
        const schoolMap = new Map(schoolsData.map(school => [school.id, school]));

        // Associate schools with teams
        const teamsWithSchools = teamsData.map(team => ({
          ...team,
          team: team.team ? {
            ...team.team,
            school: schoolMap.get(team.team.schoolId)
          } : undefined
        }));

        setTeams(teamsWithSchools);
        setSchools(schoolsData);
        setPortfolios(portfoliosData);

        // Fetch portfolio teams for each portfolio
        if (portfoliosData.length > 0) {
          const portfolioTeamsPromises = portfoliosData.map(portfolio => 
            calcuttaService.getPortfolioTeams(portfolio.id)
          );
          
          const portfolioTeamsResults = await Promise.all(portfolioTeamsPromises);
          const allPortfolioTeams = portfolioTeamsResults.flat();
          
          // Associate schools with portfolio teams
          const portfolioTeamsWithSchools = allPortfolioTeams.map(team => ({
            ...team,
            team: team.team ? {
              ...team.team,
              school: schoolMap.get(team.team.schoolId)
            } : undefined
          }));
          
          setPortfolioTeams(portfolioTeamsWithSchools);
        }
        
        setLoading(false);
      } catch (err) {
        setError('Failed to fetch data');
        setLoading(false);
      }
    };

    fetchData();
  }, [entryId, calcuttaId]);

  if (loading) {
    return <div>Loading...</div>;
  }

  if (error) {
    return <div className="error">{error}</div>;
  }

  // Helper function to find portfolio team data for a given team ID
  const getPortfolioTeamData = (teamId: string) => {
    return portfolioTeams.find(pt => pt.teamId === teamId);
  };

  // Sort teams by points earned, then by ownership percentage, then by bid amount
  const sortedTeams = [...teams].sort((a, b) => {
    const portfolioTeamA = getPortfolioTeamData(a.teamId);
    const portfolioTeamB = getPortfolioTeamData(b.teamId);
    
    // First sort by points earned (descending)
    const pointsA = portfolioTeamA?.actualPoints || 0;
    const pointsB = portfolioTeamB?.actualPoints || 0;
    
    if (pointsA !== pointsB) {
      return pointsB - pointsA;
    }
    
    // If points are equal, sort by ownership percentage (descending)
    const ownershipA = portfolioTeamA?.ownershipPercentage || 0;
    const ownershipB = portfolioTeamB?.ownershipPercentage || 0;
    
    if (ownershipA !== ownershipB) {
      return ownershipB - ownershipA;
    }
    
    // If ownership is equal, sort by bid amount (descending)
    return b.bid - a.bid;
  });

  // Helper function to calculate wins (wins + byes - 1)
  const calculateWins = (team: CalcuttaEntryTeam) => {
    if (!team.team) return 0;
    const wins = team.team.wins || 0;
    const byes = team.team.byes || 0;
    return Math.max(0, wins + byes - 1);
  };

  return (
    <div className="container mx-auto px-4 py-8">
      <div className="mb-6">
        <Link to={`/calcuttas/${calcuttaId}`} className="text-blue-600 hover:text-blue-800">← Back to Entries</Link>
      </div>
      <h1 className="text-3xl font-bold mb-6">Teams and Bids</h1>
      <p className="text-gray-600 mb-4 italic">Teams are sorted by points earned, then by ownership percentage, then by bid amount. In the future, teams will be sorted by predicted points.</p>
      {portfolios.length > 0 && <PortfolioScores portfolio={portfolios[0]} teams={portfolioTeams} />}
      <div className="grid gap-4">
        {sortedTeams.map((team) => {
          const portfolioTeam = getPortfolioTeamData(team.teamId);
          const wins = calculateWins(team);
          return (
            <div
              key={team.id}
              className="p-4 bg-white rounded-lg shadow"
            >
              <h2 className="text-xl font-semibold">
                {team.team?.school?.name || 'Unknown School'}
              </h2>
              <div className="grid grid-cols-2 gap-2 mt-2">
                <p className="text-gray-600">Bid Amount: ${team.bid}</p>
                <p className="text-gray-600">Wins: {wins}</p>
                {portfolioTeam && (
                  <>
                    <p className="text-gray-600">
                      Ownership: {(portfolioTeam.ownershipPercentage * 100).toFixed(2)}%
                    </p>
                    <p className="text-gray-600">
                      Points Earned: {portfolioTeam.actualPoints.toFixed(2)}
                    </p>
                  </>
                )}
              </div>
            </div>
          );
        })}
      </div>
    </div>
  );
} 
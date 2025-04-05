import csv
import os
from typing import Dict, List
from dataclasses import dataclass
from pathlib import Path

@dataclass
class TeamData:
    name: str
    seed: int
    region: str
    points_scored: int
    total_investment: float
    participants: Dict[str, Dict[str, float]]  # name -> {credits, %, points}

@dataclass
class TournamentData:
    year: int
    teams: List[TeamData]
    total_points: int
    total_investment: float

def parse_participant_data(row: Dict[str, str], start_idx: int) -> Dict[str, Dict[str, float]]:
    """Parse participant data from a row starting at the given index."""
    participants = {}
    i = start_idx
    while i < len(row):
        name = row.get(f'Column{i}', '').strip()
        if not name:
            break
            
        try:
            credits = float(row.get(f'Column{i+1}', 0) or 0)
            percentage = float(row.get(f'Column{i+2}', 0) or 0)
            points = float(row.get(f'Column{i+3}', 0) or 0)
            
            participants[name] = {
                'credits_invested': credits,
                'percentage_owned': percentage,
                'points': points
            }
        except (ValueError, TypeError):
            print(f"Warning: Could not parse data for participant {name}")
            
        i += 4
    return participants

def parse_historical_file(file_path: str) -> TournamentData:
    """Parse a historical Calcutta CSV file."""
    year = int(Path(file_path).stem)
    teams = []
    total_points = 0
    total_investment = 0
    
    with open(file_path, 'r') as f:
        reader = csv.DictReader(f)
        
        for row in reader:
            # Skip empty rows or play-in losers
            if not row['TEAM'] or row['TEAM'].strip() == 'Play-in losers':
                continue
                
            try:
                seed = int(row['Seed']) if row['Seed'] else 0
                points = int(row['Points Scored']) if row['Points Scored'] else 0
                investment = float(row['Total Investment']) if row['Total Investment'] else 0
                
                # Parse participant data starting after team metadata columns
                participants = parse_participant_data(row, 5)
                
                team = TeamData(
                    name=row['TEAM'].strip(),
                    seed=seed,
                    region=row['Region'].strip(),
                    points_scored=points,
                    total_investment=investment,
                    participants=participants
                )
                
                teams.append(team)
                total_points += points
                total_investment += investment
                
            except (ValueError, TypeError) as e:
                print(f"Warning: Could not parse data for team {row['TEAM']}: {e}")
                
    return TournamentData(
        year=year,
        teams=teams,
        total_points=total_points,
        total_investment=total_investment
    )

def process_all_historical_files(data_dir: str) -> Dict[int, TournamentData]:
    """Process all historical CSV files in the given directory."""
    tournaments = {}
    
    for file in os.listdir(data_dir):
        if file.endswith('.csv'):
            file_path = os.path.join(data_dir, file)
            try:
                tournament = parse_historical_file(file_path)
                tournaments[tournament.year] = tournament
                print(f"Successfully parsed {file}")
            except Exception as e:
                print(f"Error processing {file}: {e}")
                
    return tournaments

def generate_sql_seed_statements(tournaments: Dict[int, TournamentData]) -> str:
    """Generate SQL statements to seed the database with historical data."""
    sql_statements = []
    
    # Add tournament data
    for year, tournament in tournaments.items():
        sql_statements.append(f"""
INSERT INTO tournaments (year, total_points, total_investment)
VALUES ({year}, {tournament.total_points}, {tournament.total_investment});
""")
        
        # Add team data
        for team in tournament.teams:
            sql_statements.append(f"""
INSERT INTO teams (
    tournament_id, name, seed, region, points_scored, total_investment
)
VALUES (
    (SELECT id FROM tournaments WHERE year = {year}),
    '{team.name}',
    {team.seed},
    '{team.region}',
    {team.points_scored},
    {team.total_investment}
);
""")
            
            # Add participant data
            for participant_name, data in team.participants.items():
                sql_statements.append(f"""
INSERT INTO participants (
    team_id, name, credits_invested, percentage_owned, points
)
VALUES (
    (SELECT id FROM teams 
     WHERE tournament_id = (SELECT id FROM tournaments WHERE year = {year})
     AND name = '{team.name}'),
    '{participant_name}',
    {data['credits_invested']},
    {data['percentage_owned']},
    {data['points']}
);
""")
    
    return '\n'.join(sql_statements)

def main():
    # Get the directory containing this script
    script_dir = os.path.dirname(os.path.abspath(__file__))
    data_dir = os.path.join(script_dir, '..', 'data', 'historical')
    
    # Process all historical files
    tournaments = process_all_historical_files(data_dir)
    
    # Generate SQL statements
    sql_statements = generate_sql_seed_statements(tournaments)
    
    # Write SQL statements to a file
    output_file = os.path.join(script_dir, '..', 'data', 'seed_historical_data.sql')
    with open(output_file, 'w') as f:
        f.write(sql_statements)
    
    print(f"Generated SQL seed statements in {output_file}")

if __name__ == '__main__':
    main() 
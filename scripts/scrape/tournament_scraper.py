#!/usr/bin/env python3

import json
import time
import requests
from bs4 import BeautifulSoup
from datetime import datetime, timedelta
from typing import List, Optional, Tuple
from dataclasses import dataclass
from pathlib import Path
import re

@dataclass
class Game:
    id: str
    tournament_id: str
    team1_id: str
    team2_id: str
    tipoff_time: str
    sort_order: int
    team1_score: int
    team2_score: int
    next_game_id: str
    next_game_slot: int
    is_final: bool
    created: str
    updated: str

class TournamentScraper:
    def __init__(self, year: int):
        self.year = year
        self.base_url = "https://www.sports-reference.com/cbb/postseason"
        self.tournament_id = f"{year}-ncaa-mens"
        self.output_dir = Path("data/tournaments")
        self.output_dir.mkdir(parents=True, exist_ok=True)
        
        # Map of common school name variations
        self.school_name_map = {
            "NC State": "north-carolina-state",
            "UConn": "connecticut",
            "USC": "southern-california",
            "UCF": "central-florida",
            "UNCW": "unc-wilmington",
            "VCU": "virginia-commonwealth",
            "SMU": "southern-methodist",
            "LSU": "louisiana-state",
            "BYU": "brigham-young",
            "Ole Miss": "mississippi",
            "UNC": "north-carolina",
            "UC-Davis": "california-davis",
            # Add more mappings as needed
        }

        self.seen_first_four_games = set()  # Track First Four games across regions

    def _get_team_id(self, team_name: str) -> str:
        """Convert team name to our ID format."""
        # Check for known variations first
        if team_name in self.school_name_map:
            return self.school_name_map[team_name]
            
        return (
            team_name.lower()
            .replace(" ", "-")
            .replace("'", "")
            .replace(".", "")
            .replace("&", "and")
            .replace("(", "")
            .replace(")", "")
        )

    def _get_game_id(
        self, region: str, round_name: str, game_number: int
    ) -> str:
        """Generate game ID based on region, round and game number."""
        round_prefix = {
            "First Four": "ff",
            "First Round": "r64",
            "Second Round": "r32",
            "Sweet 16": "r16",
            "Elite Eight": "r8",
            "Final Four": "f4",
            "Championship": "champ"
        }
        return f"{self.year}-{round_prefix[round_name]}-{game_number}"

    def _parse_score(self, score_text: str) -> Optional[int]:
        """Parse score text into integer, handling empty or invalid scores."""
        try:
            return int(score_text.strip())
        except (ValueError, AttributeError):
            return 0

    def _parse_first_four_game(self, game_text: str, game_number: int, date: datetime) -> Optional[Game]:
        """Parse a First Four game from text."""
        try:
            # Handle both standard format and 2021 special case
            if "at Bloomington" in game_text or "at West Lafayette" in game_text:
                # Remove location information
                game_text = game_text.split(" at ")[0].strip()

            # Remove region name if present
            for region in ["East", "West", "South", "Midwest"]:
                game_text = game_text.replace(f"{region} First Four", "").strip()

            # Split into team parts
            parts = game_text.split(",")
            if len(parts) != 2:
                print(f"Invalid game format: {game_text}")
                return None

            # Extract team names and scores
            team1_part = parts[0].strip()
            team2_part = parts[1].strip()

            # Extract team names and scores using a more flexible regex
            # This will match both "16 FDU 65" and "16 Florida Gulf Coast 96" formats
            team1_match = re.match(r'\d+\s+(.+?)\s+(\d+)$', team1_part)
            team2_match = re.match(r'\d+\s+(.+?)\s+(\d+)$', team2_part)

            if not team1_match or not team2_match:
                # Try alternative format without seed numbers
                team1_match = re.match(r'(.+?)\s+(\d+)$', team1_part)
                team2_match = re.match(r'(.+?)\s+(\d+)$', team2_part)
                if not team1_match or not team2_match:
                    print(f"Could not extract team names from: {game_text}")
                    return None

            team1 = team1_match.group(1).strip()
            team2 = team2_match.group(1).strip()
            score1 = int(team1_match.group(2))
            score2 = int(team2_match.group(2))

            # Clean up team names
            for region in ["East", "West", "South", "Midwest"]:
                team1 = team1.replace(f"{region} ", "")
                team2 = team2.replace(f"{region} ", "")

            return Game(
                id=f"{self.year}-ff-{game_number}",
                tournament_id=self.tournament_id,
                team1_id=team1,
                team2_id=team2,
                tipoff_time=date.isoformat(),
                sort_order=game_number,
                team1_score=score1,
                team2_score=score2,
                next_game_id="",
                next_game_slot=1,
                is_final=True,
                created=datetime.now().isoformat(),
                updated=datetime.now().isoformat()
            )
        except Exception as e:
            print(f"Error parsing First Four game: {e}")
            return None

    def _parse_game_div(self, game_div, round_name, game_number):
        """Parse a game div into a Game object."""
        try:
            # Find all divs that contain team information
            team_divs = game_div.find_all('div', recursive=False)
            
            # Skip placeholder entries (single-team winners or empty slots)
            if len(team_divs) <= 1:
                return None
            
            # Regular game parsing for games with two teams
            if len(team_divs) != 2:
                print(f"Expected 2 team divs, found {len(team_divs)}")
                return None

            # Extract team names and scores
            team1_div = team_divs[0]
            team2_div = team_divs[1]

            # Get team names from the first <a> tag in each div
            team1_name = team1_div.find('a').text.strip()
            team2_name = team2_div.find('a').text.strip()

            # Get scores from the last <a> tag in each div
            team1_score = int(team1_div.find_all('a')[-1].text.strip())
            team2_score = int(team2_div.find_all('a')[-1].text.strip())

            # Get box score link which contains the date
            box_score_link = team1_div.find_all('a')[-1]['href']
            # Extract date from link format: /cbb/boxscores/2017-03-16-villanova.html
            date_str = box_score_link.split('/')[-1].split('.')[0]  # 2017-03-16-villanova
            date_str = '-'.join(date_str.split('-')[:3])  # 2017-03-16
            
            # Convert date string to datetime
            tipoff_time = datetime.strptime(date_str, "%Y-%m-%d").isoformat()

            # Create game ID based on round and game number
            game_id = f"{self.year}-{round_name.lower().replace(' ', '')}-{game_number}"

            # Create Game object with all required parameters
            return Game(
                id=game_id,
                tournament_id=f"{self.year}-ncaa-mens",
                team1_id=team1_name,
                team2_id=team2_name,
                team1_score=team1_score,
                team2_score=team2_score,
                tipoff_time=tipoff_time,
                sort_order=game_number,
                next_game_id="",  # Will be filled in later
                next_game_slot=1,  # Will be updated later
                is_final=True,  # All games in the bracket are final
                created=datetime.utcnow().isoformat(),
                updated=datetime.utcnow().isoformat()
            )
        except Exception as e:
            # Only show error for actual parsing failures
            if len(team_divs) > 2:
                print(f"Failed to parse game in {round_name}: {str(e)}")
            return None

    def _get_tournament_games(self) -> List[Game]:
        """Fetch all tournament games from Sports Reference."""
        games = []
        game_number = 1
        seen_first_four_games = set()  # Track First Four games across regions
        
        url = f"{self.base_url}/{self.year}-ncaa.html"
        
        try:
            response = requests.get(url)
            response.raise_for_status()
            soup = BeautifulSoup(response.text, "lxml")
            
            # Find the brackets container
            brackets_div = soup.find("div", id="brackets")
            if not brackets_div:
                print("Could not find brackets div")
                return games
            
            # Tournament dates (approximate)
            start_date = datetime(self.year, 3, 14)
            
            # Process each region
            for region_div in brackets_div.find_all("div", recursive=False):
                region = region_div.get("id", "").capitalize()
                if not region:
                    continue
                
                # Find First Four games (if any)
                first_four = region_div.find("p")
                if first_four and "First Four" in first_four.text:
                    print(f"Found First Four games in {region}")
                    # Get the correct start date from the box score links
                    box_score_links = first_four.find_all('a', href=True)
                    ff_date = None
                    for link in box_score_links:
                        if '/boxscores/' in link['href']:
                            # Extract date from link format: /cbb/boxscores/2021-03-18-team.html
                            date_str = link['href'].split('/')[-1].split('.')[0]  # 2021-03-18-team
                            date_str = '-'.join(date_str.split('-')[:3])  # 2021-03-18
                            ff_date = datetime.strptime(date_str, "%Y-%m-%d")
                            break

                    if not ff_date:
                        ff_date = start_date  # Fallback to default date

                    # Split by all possible location markers
                    for location in ["at Dayton, OH", "at Bloomington, IN", "at West Lafayette, IN"]:
                        if location in first_four.text:
                            for game_text in first_four.text.split(location)[:-1]:
                                if not game_text.strip():
                                    continue
                                # Create a unique key for the game based on teams and scores
                                game_key = ''.join(c for c in game_text if not c.isspace())
                                if game_key not in seen_first_four_games:
                                    seen_first_four_games.add(game_key)
                                    game = self._parse_first_four_game(
                                        game_text, game_number, ff_date
                                    )
                                    if game:
                                        games.append(game)
                                        game_number += 1

                # Find bracket games
                bracket = region_div.find("div", id="bracket")
                if not bracket:
                    print(f"No bracket found in {region}")
                    continue

                print(f"Found bracket in {region}")

                # Process each round
                round_divs = bracket.find_all("div", class_="round")

                # National region only has Final Four and Championship
                if region == "National":
                    rounds = [
                        ("Final Four", round_divs[0]),
                        ("Championship", round_divs[1])
                    ]
                else:
                    rounds = [
                        ("First Round", round_divs[0]),
                        ("Second Round", round_divs[1]),
                        ("Sweet 16", round_divs[2]),
                        ("Elite Eight", round_divs[3])
                    ]

                for round_name, round_div in rounds:
                    print(f"Processing {round_name} in {region}")

                    # Process each game in the round
                    game_divs = round_div.find_all("div", recursive=False)
                    print(f"Found {len(game_divs)} games in {round_name}")

                    for game_div in game_divs:
                        game = self._parse_game_div(
                            game_div, round_name, game_number
                        )
                        if game:
                            games.append(game)
                            game_number += 1
            
            # Rate limiting
            time.sleep(1)
            
        except requests.RequestException as e:
            print(f"Error fetching tournament data: {e}")
        
        return games

    def _update_game_relationships(self, games: List[Game]) -> None:
        """Update next_game_id and next_game_slot based on bracket structure."""
        round_games = {}
        for game in games:
            round_name = game.id.split("-")[1]
            if round_name not in round_games:
                round_games[round_name] = []
            round_games[round_name].append(game)
        
        # Update relationships between rounds
        round_order = ["ff", "r64", "r32", "r16", "r8", "f4", "champ"]
        for i in range(len(round_order) - 1):
            current_round = round_games.get(round_order[i], [])
            next_round = round_games.get(round_order[i + 1], [])
            
            for j, game in enumerate(current_round):
                if j < len(next_round):
                    game.next_game_id = next_round[j // 2].id
                    game.next_game_slot = 1 if j % 2 == 0 else 2

    def scrape(self) -> None:
        """Scrape tournament data and save to JSON file."""
        print(f"Scraping {self.year} NCAA Tournament data...")
        
        # Get all games
        games = self._get_tournament_games()
        
        if not games:
            print("No games found. This could mean:")
            print("1. The tournament hasn't started yet")
            print("2. The year is invalid")
            print("3. The website structure has changed")
            return
        
        # Update game relationships
        self._update_game_relationships(games)
        
        # Create tournament data structure
        tournament_data = {
            "id": self.tournament_id,
            "name": "NCAA Division I Men's Basketball Tournament",
            "rounds": 7,
            "games": [vars(game) for game in games]
        }
        
        # Save to file
        output_file = self.output_dir / f"{self.year}.json"
        with open(output_file, "w") as f:
            json.dump(tournament_data, f, indent=2)
        
        print(f"Saved {len(games)} games to {output_file}")

def main():
    import argparse
    parser = argparse.ArgumentParser(description="Scrape NCAA tournament data")
    parser.add_argument("year", type=int, help="Tournament year to scrape")
    args = parser.parse_args()
    
    scraper = TournamentScraper(args.year)
    scraper.scrape()

if __name__ == "__main__":
    main() 
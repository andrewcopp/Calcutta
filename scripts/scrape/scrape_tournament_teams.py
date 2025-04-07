#!/usr/bin/env python3

import requests
from bs4 import BeautifulSoup
from typing import List
from dataclasses import dataclass
import re
import argparse
import time
import os
import json
from pathlib import Path


@dataclass
class TournamentTeam:
    name: str
    seed: int
    region: str
    
    def to_dict(self):
        """Convert to dictionary for JSON serialization."""
        return {
            "name": self.name,
            "seed": self.seed,
            "region": self.region
        }


class TournamentTeamScraper:
    def __init__(self, year: int):
        self.year = year
        self.base_url = "https://www.sports-reference.com/cbb/postseason"
        self.tournament_id = f"{year}-ncaa-mens"
        
        # Map of common school name variations
        self.school_name_map = {
            "NC State": "North Carolina State",
            "UConn": "Connecticut",
            "USC": "Southern California",
            "UCF": "Central Florida",
            "UNCW": "UNC Wilmington",
            "VCU": "Virginia Commonwealth",
            "SMU": "Southern Methodist",
            "LSU": "Louisiana State",
            "BYU": "Brigham Young",
            "Ole Miss": "Mississippi",
            "UNC": "North Carolina",
            "UC-Davis": "UC Davis",
            "Saint ": "St. ",
            "State": "St",
            "University": "U",
            "University of California": "UC",
            "University of North Carolina": "UNC",
            "University of South Carolina": "South Carolina",
            "University of Southern California": "USC",
            "University of California, Los Angeles": "UCLA",
            "University of California, Berkeley": "UC Berkeley",
            "University of California, Davis": "UC Davis",
            "University of California, Irvine": "UC Irvine",
            "University of California, Riverside": "UC Riverside",
            "University of California, San Diego": "UC San Diego",
            "University of California, Santa Barbara": "UC Santa Barbara",
            "University of California, Santa Cruz": "UC Santa Cruz",
        }

    def _clean_team_name(self, name: str) -> str:
        """Clean and standardize team names."""
        # Remove any trailing seed numbers and parentheses
        name = re.sub(r'\(\d+\)', '', name)
        name = name.strip()
        
        # Apply known replacements
        for old, new in self.school_name_map.items():
            name = name.replace(old, new)
            
        return name

    def _parse_first_four_teams(self, first_four_text: str, region: str) -> List[TournamentTeam]:
        """Parse First Four teams from text."""
        teams = []
        
        try:
            # Split into games by location
            games_text = first_four_text.split("at Dayton, OH")
            
            for game_text in games_text:
                game_text = game_text.strip()
                if not game_text:
                    continue
                
                # Remove region name if present
                game_text = game_text.replace(f"{region} First Four", "").strip()
                
                # Split into team parts
                parts = game_text.split(",")
                if len(parts) != 2:
                    continue
                
                # Extract team names and seeds
                for part in parts:
                    part = part.strip()
                    # Match seed and team name: "16 Team Name" or "16 FDU"
                    # Also handle scores at the end
                    match = re.match(r'(\d+)\s+(.+?)(?:\s+\d+)?$', part)
                    if not match:
                        print(f"Could not extract team info from: {part}")
                        continue
                    
                    seed = int(match.group(1))
                    team_name = match.group(2).strip()
                    
                    # Clean up team name
                    team_name = self._clean_team_name(team_name)
                    
                    teams.append(TournamentTeam(
                        name=team_name,
                        seed=seed,
                        region=region
                    ))
        
        except Exception as e:
            print(f"Error parsing First Four teams: {e}")
        
        return teams

    def scrape_tournament_teams(self) -> List[TournamentTeam]:
        """Scrape tournament teams from sports-reference.com."""
        url = f"{self.base_url}/{self.year}-ncaa.html"
        
        try:
            response = requests.get(url)
            response.raise_for_status()
            soup = BeautifulSoup(response.text, "html.parser")
            
            # Find the brackets container
            brackets_div = soup.find("div", id="brackets")
            if not brackets_div:
                print("Could not find brackets div")
                return []
            
            teams = []
            regions = ["East", "West", "South", "Midwest"]
            seen_teams = set()  # Track teams we've already added
            
            # Process each region
            for region_div in brackets_div.find_all("div", recursive=False):
                region = region_div.get("id", "").capitalize()
                if not region or region not in regions:
                    continue
                
                print(f"Processing {region} region...")
                
                # Find First Four games (if any)
                first_four = region_div.find("p")
                if first_four and "First Four" in first_four.text:
                    print(f"Found First Four games in {region}")
                    first_four_teams = self._parse_first_four_teams(first_four.text, region)
                    for team in first_four_teams:
                        team_key = f"{team.name}_{team.region}"
                        if team_key not in seen_teams:
                            seen_teams.add(team_key)
                            teams.append(team)
                
                # Find bracket games
                bracket = region_div.find("div", id="bracket")
                if not bracket:
                    print(f"No bracket found in {region}")
                    continue
                
                # Process each round
                round_divs = bracket.find_all("div", class_="round")
                
                # Each region has 4 rounds (except National)
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
                    for game_div in game_divs:
                        # Find all team divs in this game
                        team_divs = game_div.find_all("div", recursive=False)
                        
                        # Skip games without exactly two teams
                        if len(team_divs) != 2:
                            continue
                        
                        # Process each team
                        for team_div in team_divs:
                            # Get team name from first link
                            team_link = team_div.find("a")
                            if not team_link:
                                continue
                            
                            # Get seed from span
                            seed_span = team_div.find("span")
                            if not seed_span:
                                continue
                            
                            try:
                                seed = int(seed_span.text.strip())
                                team_name = team_link.text.strip()
                            except (ValueError, AttributeError):
                                continue
                            
                            # Clean up team name
                            team_name = self._clean_team_name(team_name)
                            
                            # Only add team if we haven't seen it before
                            team_key = f"{team_name}_{region}"
                            if team_key not in seen_teams:
                                seen_teams.add(team_key)
                                teams.append(TournamentTeam(
                                    name=team_name,
                                    seed=seed,
                                    region=region
                                ))
            
            # Rate limiting
            time.sleep(1)
            
            return teams
            
        except requests.RequestException as e:
            print(f"Failed to retrieve webpage: {e}")
            return []
        except Exception as e:
            print(f"Error scraping teams: {e}")
            return []


def save_teams_to_json(teams: List[TournamentTeam], tournament_id: str):
    """Save teams to a JSON file."""
    try:
        # Create output directory if it doesn't exist
        output_dir = Path("data/tournament_teams")
        output_dir.mkdir(parents=True, exist_ok=True)
        
        # Create output file path
        output_file = output_dir / f"{tournament_id}.json"
        
        # Convert teams to dictionaries
        teams_data = [team.to_dict() for team in teams]
        
        # Write to JSON file
        with open(output_file, "w") as f:
            json.dump(teams_data, f, indent=2)
            
        print(f"Successfully saved {len(teams)} teams to {output_file}")
        
    except Exception as e:
        print(f"Error saving teams to JSON: {e}")
        # Print teams that would have been saved
        print("\nTeams found:")
        for team in teams:
            print(f"{team.name} ({team.seed}) - {team.region}")


def main():
    parser = argparse.ArgumentParser(
        description='Scrape NCAA tournament teams for a given year'
    )
    parser.add_argument(
        'year', 
        type=int,
        help='Year to scrape (e.g., 2024)'
    )
    args = parser.parse_args()
    
    print(f"\nScraping {args.year} tournament teams...")
    
    scraper = TournamentTeamScraper(args.year)
    teams = scraper.scrape_tournament_teams()
    
    if teams:
        print(f"Found {len(teams)} teams for {args.year} tournament")
        save_teams_to_json(teams, f"{args.year}-ncaa-mens")
    else:
        print(f"No teams found for {args.year} tournament")


if __name__ == "__main__":
    main() 
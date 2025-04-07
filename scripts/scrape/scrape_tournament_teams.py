#!/usr/bin/env python3

import requests
from bs4 import BeautifulSoup
from typing import List, Set, Tuple
from dataclasses import dataclass
import re
import argparse
import time
import os
import json
from pathlib import Path


@dataclass
class TournamentTeam:
    """Class for storing tournament team data."""
    name: str
    seed: int
    region: str
    is_first_four: bool = False
    
    def to_dict(self):
        """Convert to dictionary for JSON serialization."""
        return {
            "name": self.name,
            "seed": self.seed,
            "region": self.region,
            "is_first_four": self.is_first_four
        }


class TournamentTeamScraper:
    def __init__(self, year: int):
        """Initialize the scraper with the tournament year."""
        self.year = year
        self.seen_teams = set()  # Track teams across all parsing functions
        self.first_four_teams = set()  # Track First Four teams specifically
        self.session = requests.Session()
        self.session.headers.update({
            'User-Agent': 'Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36'
        })
        self.base_url = "https://www.sports-reference.com/cbb/postseason/men/{}-ncaa.html"
        self.output_dir = "data/tournament_teams"
        os.makedirs(self.output_dir, exist_ok=True)
        
        # Initialize BeautifulSoup
        url = self.base_url.format(self.year)
        response = self.session.get(url)
        response.raise_for_status()
        self.soup = BeautifulSoup(response.text, "html.parser")
        
        # Rate limiting
        time.sleep(1)

    def _clean_team_name(self, name: str) -> str:
        """Clean team names by removing seed numbers and extra whitespace."""
        # Remove any trailing seed numbers and parentheses
        name = re.sub(r'\(\d+\)', '', name)
        # Only strip whitespace, preserve all other text
        return name.strip()

    def _parse_first_four_teams(self, region: str) -> List[TournamentTeam]:
        """Parse teams from First Four games in a region."""
        teams = []
        
        try:
            # Find the brackets container
            brackets_div = self.soup.find("div", id="brackets")
            if not brackets_div:
                print("Could not find brackets div")
                return teams
            
            # Find the region div
            region_div = brackets_div.find("div", id=region.lower())
            if not region_div:
                print(f"Could not find {region} region div")
                return teams
            
            # Look for First Four paragraph
            first_four_p = region_div.find("p")
            if not first_four_p or "First Four" not in first_four_p.text:
                return teams
            
            # Process First Four teams
            # The teams are in <a> tags with href containing '/cbb/schools/'
            first_four_links = first_four_p.find_all("a", href=lambda x: x and '/cbb/schools/' in x)
            for link in first_four_links:
                team_name = link.text.strip()
                # Find the seed number before this link
                seed_span = link.find_previous("strong")
                if seed_span:
                    try:
                        seed = int(seed_span.text.strip())
                        teams.append(TournamentTeam(
                            name=team_name,
                            seed=seed,
                            region=region,
                            is_first_four=True
                        ))
                    except ValueError:
                        continue
        
        except Exception as e:
            print(f"Error parsing First Four teams: {e}")
        
        return teams

    def _parse_2021_first_four_teams(self, first_four_text: str, region: str) -> List[TournamentTeam]:
        """Parse First Four teams from text for the 2021 tournament."""
        teams = []
        
        try:
            # Remove the region name and "First Four" text
            lines = first_four_text.split('\n')
            
            # Process each line that contains team information
            for line in lines:
                line = line.strip()
                if not line or 'at ' in line:  # Skip empty lines and location lines
                    continue
                
                # Remove the scores at the end (e.g., "60", "52")
                line = re.sub(r'\s+\d+\s*$', '', line)
                
                # Split into teams
                team_parts = line.split(',')
                if len(team_parts) != 2:
                    continue
                
                # Process each team
                for part in team_parts:
                    part = part.strip()
                    # Match seed and team name: "16 Texas Southern" or "11 UCLA"
                    match = re.match(r'(\d+)\s+(.+?)(?:\s+\d+)?$', part)
                    if not match:
                        continue
                    
                    seed = int(match.group(1))
                    team_name = match.group(2).strip()
                    
                    # Remove "strong" markers from team name
                    team_name = team_name.replace('strong', '').strip()
                    
                    # Clean up team name
                    team_name = self._clean_team_name(team_name)
                    
                    # Create team and add to tracking sets
                    team = TournamentTeam(
                        name=team_name,
                        seed=seed,
                        region=region,
                        is_first_four=True
                    )
                    team_key = (team_name, seed, region)
                    
                    if team_key not in self.seen_teams:
                        teams.append(team)
                        self.seen_teams.add(team_key)
                        self.first_four_teams.add(team_key)
        
        except Exception as e:
            print(f"Error parsing 2021 First Four teams: {e}")
        
        return teams

    def _parse_bracket_teams(self, region: str) -> List[TournamentTeam]:
        """Parse teams from a region's bracket."""
        teams = []
        
        try:
            # Find the brackets container
            brackets_div = self.soup.find("div", id="brackets")
            if not brackets_div:
                print("Could not find brackets div")
                return teams
            
            # Find the region div
            region_div = brackets_div.find("div", id=region.lower())
            if not region_div:
                print(f"Could not find {region} region div")
                return teams
            
            # Find the bracket div
            bracket_div = region_div.find("div", id="bracket")
            if not bracket_div:
                print(f"Could not find bracket div in {region}")
                return teams
            
            # Process each round
            round_divs = bracket_div.find_all("div", class_="round")
            for round_div in round_divs:
                # Process each game in the round
                for game_div in round_div.find_all("div", recursive=False):
                    # Find team divs (both winner and non-winner)
                    team_divs = game_div.find_all("div")
                    if not team_divs:
                        continue
                    
                    # Process each team
                    for team_div in team_divs:
                        # Skip non-team divs
                        if not team_div.find("span"):
                            continue
                        
                        # Get the seed from the span
                        seed_span = team_div.find("span")
                        if not seed_span:
                            continue
                        
                        try:
                            seed = int(seed_span.text.strip())
                        except ValueError:
                            continue
                        
                        # Get the team name from the first link
                        team_link = team_div.find("a")
                        if not team_link:
                            continue
                        
                        team_name = team_link.text.strip()
                        team_name = self._clean_team_name(team_name)
                        
                        # Add the team if we haven't seen it before
                        if team_name not in self.seen_teams:
                            teams.append(TournamentTeam(
                                name=team_name,
                                seed=seed,
                                region=region
                            ))
                            self.seen_teams.add(team_name)
        
        except Exception as e:
            print(f"Error parsing bracket teams: {e}")
        
        return teams

    def scrape_tournament_teams(self, year: int) -> List[TournamentTeam]:
        """Scrape tournament teams for a given year."""
        print(f"\nScraping {year} tournament teams...")
        teams = []
        seen_teams = set()

        # Process regions in order
        regions = ["East", "Midwest", "South", "West"]
        for region in regions:
            print(f"Processing {region} region...")
            region_teams = self._parse_bracket_teams(region)
            for team in region_teams:
                if team.name not in seen_teams:
                    teams.append(team)
                    seen_teams.add(team.name)

            # Check for First Four games in this region
            first_four_teams = self._parse_first_four_teams(region)
            if first_four_teams:
                print(f"Found First Four games in {region}")
                for team in first_four_teams:
                    if team.name not in seen_teams:
                        team.is_first_four = True
                        teams.append(team)
                        seen_teams.add(team.name)

        print(f"Found {len(teams)} teams for {year} tournament")
        return teams

    def save_teams(self, teams: List[TournamentTeam]):
        """Save teams to a JSON file."""
        # Create output directory if it doesn't exist
        os.makedirs(self.output_dir, exist_ok=True)
        
        output_file = os.path.join(self.output_dir, f"{self.year}-ncaa-mens.json")
        with open(output_file, 'w') as f:
            json.dump([team.to_dict() for team in teams], f, indent=2)
        print(f"Successfully saved {len(teams)} teams to {output_file}")


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
    teams = scraper.scrape_tournament_teams(args.year)
    
    if teams:
        print(f"Found {len(teams)} teams for {args.year} tournament")
        scraper.save_teams(teams)
    else:
        print(f"No teams found for {args.year} tournament")


if __name__ == "__main__":
    main() 
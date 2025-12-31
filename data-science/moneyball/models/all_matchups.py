"""
Generate all theoretical matchup predictions for a 68-team tournament.

This module creates predictions for every mathematically possible matchup
across all rounds, accounting for First Four games.
"""
import pandas as pd
from moneyball.utils import bracket


def generate_all_theoretical_matchups(
    teams_df: pd.DataFrame,
    kenpom_scale: float = 10.0
) -> pd.DataFrame:
    """
    Generate ALL theoretical matchup predictions for 68-team field.
    
    For a 68-team field with First Four games, we need to account for
    all possible matchups that could occur in each round.
    
    Args:
        teams_df: DataFrame with all 68 teams (must have id, kenpom_net)
        kenpom_scale: Scale parameter for win probability
        
    Returns:
        DataFrame with all theoretical matchup predictions
    """
    predictions = []
    teams = teams_df.copy()
    regions = sorted(teams['region'].unique())
    
    # Build kenpom lookup by team_id
    kenpom_by_id = {}
    for _, team in teams.iterrows():
        kenpom_by_id[str(team['id'])] = team['kenpom_net']
    
    def calc_win_prob(id1, id2):
        net1 = kenpom_by_id[id1]
        net2 = kenpom_by_id[id2]
        return bracket.win_prob(net1, net2, kenpom_scale)
    
    def add_prediction(game_id, round_name, round_int, t1, t2, p_matchup):
        """Add a single matchup prediction."""
        # Handle both dict and namedtuple inputs
        id1 = str(t1.id if hasattr(t1, 'id') else t1['id'])
        id2 = str(t2.id if hasattr(t2, 'id') else t2['id'])
        
        predictions.append({
            'game_id': game_id,
            'round': round_name,
            'round_int': round_int,
            'team1_id': id1,
            'team2_id': id2,
            'p_matchup': p_matchup,
            'p_team1_wins_given_matchup': calc_win_prob(id1, id2),
        })
    
    # Round 1 (Round of 64): 32 known matchups
    # Standard bracket matchups (excluding First Four)
    matchup_seeds = [
        (1, 16), (8, 9), (5, 12), (4, 13),
        (6, 11), (3, 14), (7, 10), (2, 15)
    ]
    
    for region in regions:
        region_teams = teams[teams['region'] == region]
        for idx, (seed1, seed2) in enumerate(matchup_seeds):
            # Get all teams at each seed (handles First Four)
            t1_teams = region_teams[region_teams['seed'] == seed1]
            t2_teams = region_teams[region_teams['seed'] == seed2]
            
            # For seeds with multiple teams (First Four), create all combinations
            for t1 in t1_teams.itertuples():
                for t2 in t2_teams.itertuples():
                    # p_matchup depends on whether this is a First Four matchup
                    if len(t1_teams) > 1 or len(t2_teams) > 1:
                        # First Four involved - probability is 0.5 per team
                        p_match = (1.0 / len(t1_teams)) * (1.0 / len(t2_teams))
                    else:
                        p_match = 1.0
                    
                    add_prediction(
                        f'R1-{region}-{idx+1}',
                        'round_of_64',
                        1,
                        t1,
                        t2,
                        p_match
                    )
    
    # Round 2 (Round of 32): All possible matchups from R1 winners
    # Each R2 game has 2 possible teams from each side
    for region in regions:
        region_teams = teams[teams['region'] == region]
        
        r2_games = [
            (1, [1, 16], [8, 9]),
            (2, [5, 12], [4, 13]),
            (3, [6, 11], [3, 14]),
            (4, [7, 10], [2, 15]),
        ]
        
        for game_num, side1_seeds, side2_seeds in r2_games:
            side1 = region_teams[region_teams['seed'].isin(side1_seeds)]
            side2 = region_teams[region_teams['seed'].isin(side2_seeds)]
            
            # All combinations
            total_combos = len(side1) * len(side2)
            p_match = 1.0 / total_combos
            
            for t1 in side1.itertuples():
                for t2 in side2.itertuples():
                    add_prediction(
                        f'R2-{region}-{game_num}',
                        'round_of_32',
                        2,
                        t1,
                        t2,
                        p_match
                    )
    
    # Round 3 (Sweet 16): All possible matchups from R2 winners
    for region in regions:
        region_teams = teams[teams['region'] == region]
        
        r3_games = [
            (1, [1, 16, 8, 9], [5, 12, 4, 13]),
            (2, [6, 11, 3, 14], [7, 10, 2, 15]),
        ]
        
        for game_num, side1_seeds, side2_seeds in r3_games:
            side1 = region_teams[region_teams['seed'].isin(side1_seeds)]
            side2 = region_teams[region_teams['seed'].isin(side2_seeds)]
            
            total_combos = len(side1) * len(side2)
            p_match = 1.0 / total_combos
            
            for t1 in side1.itertuples():
                for t2 in side2.itertuples():
                    add_prediction(
                        f'R3-{region}-{game_num}',
                        'sweet_16',
                        3,
                        t1,
                        t2,
                        p_match
                    )
    
    # Round 4 (Elite 8): All possible matchups from R3 winners
    for region in regions:
        region_teams = teams[teams['region'] == region]
        
        # All teams from top half vs all from bottom half
        side1 = region_teams[region_teams['seed'].isin([1,16,8,9,5,12,4,13])]
        side2 = region_teams[region_teams['seed'].isin([6,11,3,14,7,10,2,15])]
        
        total_combos = len(side1) * len(side2)
        p_match = 1.0 / total_combos
        
        for t1 in side1.itertuples():
            for t2 in side2.itertuples():
                add_prediction(
                    f'R4-{region}-1',
                    'elite_8',
                    4,
                    t1,
                    t2,
                    p_match
                )
    
    # Round 5 (Final Four): All possible matchups from R4 winners
    ff_pairings = [
        (1, regions[0], regions[1]),
        (2, regions[2], regions[3]),
    ]
    
    for game_num, region1, region2 in ff_pairings:
        teams1 = teams[teams['region'] == region1]
        teams2 = teams[teams['region'] == region2]
        
        total_combos = len(teams1) * len(teams2)
        p_match = 1.0 / total_combos
        
        for t1 in teams1.itertuples():
            for t2 in teams2.itertuples():
                add_prediction(
                    f'R5-{game_num}',
                    'final_four',
                    5,
                    t1,
                    t2,
                    p_match
                )
    
    # Round 6 (Championship): All possible matchups from R5 winners
    side1_teams = teams[teams['region'].isin([regions[0], regions[1]])]
    side2_teams = teams[teams['region'].isin([regions[2], regions[3]])]
    
    total_combos = len(side1_teams) * len(side2_teams)
    p_match = 1.0 / total_combos
    
    for t1 in side1_teams.itertuples():
        for t2 in side2_teams.itertuples():
            add_prediction(
                'R6-1',
                'championship',
                6,
                t1,
                t2,
                p_match
            )
    
    return pd.DataFrame(predictions)

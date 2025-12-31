from __future__ import annotations

import random
from pathlib import Path
from typing import Dict, Optional, Tuple

import pandas as pd

from moneyball.utils import bracket


def predict_game_outcomes(
    *,
    games: pd.DataFrame,
    teams: pd.DataFrame,
    calcutta_key: Optional[str],
    kenpom_scale: float,
    n_sims: int,
    seed: int,
) -> pd.DataFrame:
    if n_sims <= 0:
        raise ValueError("n_sims must be positive")

    games_graph, prev_by_next = bracket.prepare_bracket_graph(games)

    teams = teams.copy()
    if calcutta_key is not None and "calcutta_key" in teams.columns:
        teams = teams[teams["calcutta_key"] == calcutta_key].copy()

    if "team_key" not in teams.columns:
        raise ValueError("teams.parquet missing team_key")
    if "kenpom_net" not in teams.columns:
        raise ValueError("teams.parquet missing kenpom_net")

    teams["kenpom_net"] = pd.to_numeric(teams["kenpom_net"], errors="coerce")

    net_by_team: Dict[str, float] = {}
    school_name_by_team: Dict[str, str] = {}
    for _, r in teams.iterrows():
        tk = str(r.get("team_key") or "")
        if not tk:
            continue
        net = r.get("kenpom_net")
        if pd.isna(net):
            continue
        net_by_team[tk] = float(net)
        school_name_by_team[tk] = str(r.get("school_name") or "")

    if not net_by_team:
        raise ValueError("no teams with kenpom_net available")

    game_meta = games_graph.set_index("game_id", drop=False)

    matchup_counts: Dict[Tuple[str, str, str], int] = {}
    team1_win_counts: Dict[Tuple[str, str, str], int] = {}

    rng = random.Random(int(seed))

    for _ in range(int(n_sims)):
        winners_by_game: Dict[str, str] = {}

        for _, gr in games_graph.iterrows():
            gid = str(gr.get("game_id") or "")
            if not gid:
                continue

            t1 = str(gr.get("team1_key") or "")
            t2 = str(gr.get("team2_key") or "")

            if int(gr.get("round_order") or 999) > 2:
                t1 = ""
                t2 = ""

            if not t1:
                prev = prev_by_next.get(gid, {}).get(1)
                if prev:
                    t1 = winners_by_game.get(prev, "")
            if not t2:
                prev = prev_by_next.get(gid, {}).get(2)
                if prev:
                    t2 = winners_by_game.get(prev, "")

            if not t1 or not t2:
                continue
            if t1 not in net_by_team or t2 not in net_by_team:
                continue

            net1 = float(net_by_team[t1])
            net2 = float(net_by_team[t2])
            p1 = float(bracket.win_prob(net1, net2, float(kenpom_scale)))
            w = t1 if rng.random() < p1 else t2
            winners_by_game[gid] = w

            k = (gid, t1, t2)
            matchup_counts[k] = matchup_counts.get(k, 0) + 1
            if w == t1:
                team1_win_counts[k] = team1_win_counts.get(k, 0) + 1

    rows = []
    denom = float(n_sims)
    for (gid, t1, t2), c in matchup_counts.items():
        if c <= 0:
            continue
        meta = game_meta.loc[gid]
        p_matchup = float(c) / denom if denom > 0 else 0.0
        w1 = int(team1_win_counts.get((gid, t1, t2), 0))
        p_t1 = float(w1) / float(c)
        row = {
            "game_id": str(gid),
            "round": str(meta.get("round") or ""),
            "round_order": int(meta.get("round_order") or 999),
            "sort_order": int(meta.get("sort_order") or 0),
            "team1_key": str(t1),
            "team1_school_name": str(school_name_by_team.get(t1, "")),
            "team2_key": str(t2),
            "team2_school_name": str(school_name_by_team.get(t2, "")),
            "p_matchup": p_matchup,
            "p_team1_wins_given_matchup": p_t1,
            "p_team2_wins_given_matchup": 1.0 - p_t1,
        }
        if "next_game_id" in meta.index:
            row["next_game_id"] = str(meta.get("next_game_id") or "")
        if "next_game_slot" in meta.index:
            try:
                row["next_game_slot"] = int(meta.get("next_game_slot") or 0)
            except Exception:
                row["next_game_slot"] = 0
        rows.append(row)

    df = pd.DataFrame(rows)
    if df.empty:
        return df

    df = df.sort_values(
        by=["round_order", "sort_order", "game_id", "p_matchup"],
        ascending=[True, True, True, False],
    ).reset_index(drop=True)
    return df


def predict_game_outcomes_from_snapshot(
    *,
    snapshot_dir: Path,
    calcutta_key: Optional[str],
    kenpom_scale: float,
    n_sims: int,
    seed: int,
) -> pd.DataFrame:
    games_path = snapshot_dir / "games.parquet"
    teams_path = snapshot_dir / "teams.parquet"
    if not games_path.exists():
        raise FileNotFoundError(f"missing required file: {games_path}")
    if not teams_path.exists():
        raise FileNotFoundError(f"missing required file: {teams_path}")

    games = pd.read_parquet(games_path)
    teams = pd.read_parquet(teams_path)

    return predict_game_outcomes(
        games=games,
        teams=teams,
        calcutta_key=calcutta_key,
        kenpom_scale=kenpom_scale,
        n_sims=n_sims,
        seed=seed,
    )


def predict_game_outcomes_from_teams_df(
    *,
    teams_df: pd.DataFrame,
    kenpom_scale: float = 10.0,
    n_sims: int = 5000,
    seed: int = 42,
) -> pd.DataFrame:
    """
    Generate game predictions directly from a teams DataFrame.
    
    This function generates the bracket structure from teams and then
    predicts outcomes. Used for DB-first pipeline.
    
    Args:
        teams_df: DataFrame with team data (must have id, school_name, 
                  seed, region, kenpom_net columns)
        kenpom_scale: KenPom scaling factor
        n_sims: Number of simulations
        seed: Random seed
        
    Returns:
        DataFrame with predicted game outcomes
    """
    # Generate bracket structure from teams
    games = _generate_bracket_from_teams(teams_df)
    
    # Add team_key column for compatibility with existing function
    teams = teams_df.copy()
    teams['team_key'] = teams['id']
    
    return predict_game_outcomes(
        games=games,
        teams=teams,
        calcutta_key=None,
        kenpom_scale=kenpom_scale,
        n_sims=n_sims,
        seed=seed,
    )


def _generate_bracket_from_teams(teams_df: pd.DataFrame) -> pd.DataFrame:
    """
    Generate bracket structure from teams DataFrame.
    
    Args:
        teams_df: DataFrame with team data
        
    Returns:
        DataFrame with game structure including next_game_id and next_game_slot
    """
    games = []
    game_id = 1
    
    regions = teams_df['region'].unique()
    
    for region in sorted(regions):
        region_teams = teams_df[teams_df['region'] == region].copy()
        region_teams = region_teams.sort_values('seed')
        
        # Round 1 matchups (8 games per region)
        matchups = [
            (1, 16), (8, 9), (5, 12), (4, 13),
            (6, 11), (3, 14), (7, 10), (2, 15)
        ]
        
        for idx, (seed1, seed2) in enumerate(matchups):
            team1 = region_teams[region_teams['seed'] == seed1]
            team2 = region_teams[region_teams['seed'] == seed2]
            
            if len(team1) > 0 and len(team2) > 0:
                # Calculate next game (Round 2)
                next_game_num = (idx // 2) + 1
                next_slot = (idx % 2) + 1
                
                games.append({
                    'game_id': f'R1-{region}-{seed1}v{seed2}',
                    'round': 'R64',
                    'round_order': 1,
                    'sort_order': game_id,
                    'team1_key': str(team1.iloc[0]['id']),
                    'team2_key': str(team2.iloc[0]['id']),
                    'region': region,
                    'next_game_id': f'R2-{region}-G{next_game_num}',
                    'next_game_slot': next_slot,
                })
                game_id += 1
    
    return pd.DataFrame(games)

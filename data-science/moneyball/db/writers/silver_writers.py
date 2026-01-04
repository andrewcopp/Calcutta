"""
Silver layer database writers.

Write ML predictions, simulations, and enriched data using UUIDs.
"""
import logging
import pandas as pd
import psycopg2.extras
from typing import Dict
from moneyball.db.connection import get_db_connection

logger = logging.getLogger(__name__)


def write_predicted_game_outcomes(
    tournament_id: str,
    predictions_df: pd.DataFrame,
    team_id_map: Dict[str, str],
    model_version: str = None
) -> int:
    """
    Write game outcome predictions.
    
    Args:
        tournament_id: Tournament ID
        predictions_df: DataFrame with columns:
            - game_id, round, team1_key, team2_key
            - p_team1_wins_given_matchup, p_matchup
        team_id_map: Dict mapping school_slug to team_id
        model_version: Optional model version
    
    Returns:
        Number of rows inserted
    """
    with get_db_connection() as conn:
        with conn.cursor() as cur:
            # Clear existing predictions
            cur.execute("""
                DELETE FROM lab_silver.predicted_game_outcomes
                WHERE tournament_id = %s
            """, (tournament_id,))
            
            # Predictions already have team1_id and team2_id
            df = predictions_df.copy()
            
            # Map round names to inverted integers (championship = 0)
            round_mapping = {
                'championship': 0,
                'final_four': 1,
                'elite_8': 2,
                'sweet_16': 3,
                'round_of_32': 4,
                'round_of_64': 5,
                'first_four': 6,
            }
            
            # Use round_int if provided, otherwise map from round name
            if 'round_int' not in df.columns:
                df['round_int'] = df['round'].map(round_mapping)
            else:
                # Invert the round_int (our data has 1=R1, but DB wants 5=R1)
                df['round_int'] = df['round'].map(round_mapping)
            
            values = [
                (
                    tournament_id,
                    row['game_id'],
                    int(row['round_int']),
                    str(row['team1_id']),  # team_id is UUID string
                    str(row['team2_id']),  # team_id is UUID string
                    float(row.get('p_team1_wins_given_matchup',
                          row.get('p_team1_wins', 0.5))),
                    float(row.get('p_matchup', 1.0)),
                    model_version
                )
                for _, row in df.iterrows()
            ]
            
            psycopg2.extras.execute_batch(cur, """
                INSERT INTO lab_silver.predicted_game_outcomes
                (tournament_id, game_id, round, team1_id, team2_id,
                 p_team1_wins, p_matchup, model_version)
                VALUES (%s, %s, %s, %s, %s, %s, %s, %s)
            """, values)
            
            conn.commit()
            return len(values)


def write_simulated_tournaments(
    tournament_id: str,
    simulations_df: pd.DataFrame,
    team_id_map: Dict[str, str]
) -> int:
    """
    Write simulated tournament outcomes to silver layer.
    
    Simulated tournaments are derived data (Monte Carlo simulations),
    not raw data, so they belong in the silver layer.
    
    Args:
        tournament_id: Tournament ID
        simulations_df: DataFrame with columns:
            - sim_id, school_slug, wins, byes, eliminated
        team_id_map: Dict mapping school_slug to team_id
    
    Returns:
        Number of rows inserted
    """
    with get_db_connection() as conn:
        with conn.cursor() as cur:
            # Resolve core tournament id from lab tournament.
            cur.execute("""
                SELECT core_tournament_id
                FROM lab_bronze.tournaments
                WHERE id = %s
            """, (tournament_id,))
            row = cur.fetchone()
            core_tournament_id = str(row[0]) if row and row[0] else None

            if not core_tournament_id:
                # Legacy fallback: no lineage objects can be created without a core tournament id.
                cur.execute("""
                    DELETE FROM analytics.simulated_tournaments
                    WHERE tournament_id = %s
                """, (tournament_id,))

                df = simulations_df.copy()
                df['team_id'] = df['school_slug'].map(team_id_map)
                if df['team_id'].isna().any():
                    unmapped = df[df['team_id'].isna()]['school_slug'].unique()
                    raise ValueError(f"Unmapped teams: {list(unmapped)}")

                values = [
                    (
                        tournament_id,
                        int(r['sim_id']),
                        str(r['team_id']),
                        int(r['wins']),
                        int(r['byes']),
                        bool(r['eliminated']),
                    )
                    for _, r in df.iterrows()
                ]

                psycopg2.extras.execute_batch(cur, """
                    INSERT INTO analytics.simulated_tournaments
                    (tournament_id, sim_id, team_id, wins, byes, eliminated)
                    VALUES (%s, %s, %s, %s, %s, %s)
                """, values, page_size=10000)

                conn.commit()
                return len(values)

            # Create tournament state snapshot + snapshot teams from current core tournament state.
            cur.execute("""
                INSERT INTO analytics.tournament_state_snapshots (tournament_id, source, description)
                VALUES (%s, 'moneyball_pipeline', 'Autogenerated snapshot for tournament simulation batch')
                RETURNING id
            """, (core_tournament_id,))
            snapshot_id = str(cur.fetchone()[0])

            cur.execute("""
                INSERT INTO analytics.tournament_state_snapshot_teams (
                    tournament_state_snapshot_id,
                    team_id,
                    wins,
                    byes,
                    eliminated
                )
                SELECT
                    %s,
                    ct.id,
                    ct.wins,
                    ct.byes,
                    ct.eliminated
                FROM core.teams ct
                WHERE ct.tournament_id = %s
                  AND ct.deleted_at IS NULL
                ON CONFLICT (tournament_state_snapshot_id, team_id) DO NOTHING
            """, (snapshot_id, core_tournament_id))

            # Create simulation batch.
            n_sims = int(simulations_df['sim_id'].nunique())
            cur.execute("""
                INSERT INTO analytics.tournament_simulation_batches (
                    tournament_id,
                    tournament_state_snapshot_id,
                    n_sims,
                    seed,
                    probability_source_key
                )
                VALUES (%s, %s, %s, %s, %s)
                RETURNING id
            """, (core_tournament_id, snapshot_id, n_sims, 0, 'moneyball_pipeline'))
            batch_id = str(cur.fetchone()[0])

            # Backward-compat cleanup: clear legacy rows (no batch id) for this lab tournament.
            cur.execute("""
                DELETE FROM analytics.simulated_tournaments
                WHERE tournament_id = %s
                  AND tournament_simulation_batch_id IS NULL
            """, (tournament_id,))
            
            # Map school_slug to team_id
            df = simulations_df.copy()
            df['team_id'] = df['school_slug'].map(team_id_map)
            
            # Check for unmapped teams
            if df['team_id'].isna().any():
                unmapped = df[df['team_id'].isna()]['school_slug'].unique()
                raise ValueError(f"Unmapped teams: {list(unmapped)}")
            
            # Prepare values
            values = [
                (
                    batch_id,
                    tournament_id,
                    int(row['sim_id']),
                    str(row['team_id']),  # team_id is UUID string
                    int(row['wins']),
                    int(row['byes']),
                    bool(row['eliminated'])
                )
                for _, row in df.iterrows()
            ]
            
            # Batch insert
            psycopg2.extras.execute_batch(cur, """
                INSERT INTO analytics.simulated_tournaments
                (tournament_simulation_batch_id, tournament_id, sim_id, team_id, wins, byes, eliminated)
                VALUES (%s, %s, %s, %s, %s, %s, %s)
            """, values, page_size=10000)

            conn.commit()
            return len(values)


def write_predicted_market_share(
    predictions_df: pd.DataFrame,
    team_id_map: Dict[str, str],
    calcutta_id: str = None,
    tournament_id: str = None,
    model_version: str = None
) -> int:
    """
    Write predicted market share from ridge regression model.
    
    Args:
        predictions_df: DataFrame with columns:
            - team_key, predicted_auction_share_of_pool
        team_id_map: Dict mapping school_slug to team_id
        calcutta_id: Calcutta ID (optional, for production use)
        tournament_id: Tournament ID (optional, for migration period)
        model_version: Optional model version

    Note: Must provide either calcutta_id or tournament_id

    Returns:
        Number of rows inserted
    """
    if not calcutta_id and not tournament_id:
        raise ValueError("Must provide either calcutta_id or tournament_id")

    with get_db_connection() as conn:
        with conn.cursor() as cur:
            # Clear existing predictions
            if calcutta_id:
                cur.execute("""
                    DELETE FROM lab_silver.predicted_market_share
                    WHERE calcutta_id = %s
                """, (calcutta_id,))
            else:
                cur.execute("""
                    DELETE FROM lab_silver.predicted_market_share
                    WHERE tournament_id = %s
                """, (tournament_id,))

            # Map team_key to team_id
            df = predictions_df.copy()
            df['team_id'] = df['team_key'].map(team_id_map)

            # Check for unmapped teams
            if df['team_id'].isna().any():
                unmapped = df[df['team_id'].isna()]['team_key'].unique()
                raise ValueError(f"Unmapped teams: {list(unmapped)}")

            # Prepare values
            values = [
                (
                    calcutta_id,
                    tournament_id,
                    str(row['team_id']),
                    float(row['predicted_auction_share_of_pool']),
                    float(row['predicted_auction_share_of_pool']) * 100.0,
                )
                for _, row in df.iterrows()
            ]

            # Batch insert
            psycopg2.extras.execute_batch(cur, """
                INSERT INTO lab_silver.predicted_market_share
                (calcutta_id, tournament_id, team_id, predicted_share,
                 predicted_points)
                VALUES (%s, %s, %s, %s, %s)
            """, values)

            conn.commit()
            ref = calcutta_id or tournament_id
            logger.info(f"Wrote {len(values)} market shares for {ref}")
            return len(values)

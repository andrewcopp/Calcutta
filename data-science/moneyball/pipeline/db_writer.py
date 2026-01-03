"""
Database writer adapter for pipeline stages.

Integrates database writers alongside parquet file outputs,
maintaining backward compatibility while enabling hybrid architecture.
"""
from __future__ import annotations

import os
from pathlib import Path
from typing import Optional, Dict
import pandas as pd

from moneyball.db.writers.bronze_writers import (
    get_or_create_tournament,
    write_teams,
)
from moneyball.db.writers.silver_writers import (
    write_predicted_game_outcomes,
    write_simulated_tournaments,
)
from moneyball.db.writers.gold_writers import (
    write_optimization_run,
    write_recommended_entry_bids,
)


class DatabaseWriter:
    """
    Adapter for writing pipeline artifacts to database.
    
    Designed to work alongside existing parquet writes,
    allowing gradual migration to database-first approach.
    """
    
    def __init__(self, enabled: bool = None):
        """
        Initialize database writer.
        
        Args:
            enabled: Whether to write to database. If None, checks
                    CALCUTTA_WRITE_TO_DB environment variable.
        """
        if enabled is None:
            enabled = os.getenv("CALCUTTA_WRITE_TO_DB", "false").lower() == "true"
        self.enabled = enabled
        
        if self.enabled:
            # Verify database connection is configured
            required_vars = [
                "DB_HOST",
                "DB_NAME",
            ]
            missing = [v for v in required_vars if not os.getenv(v)]
            if missing:
                print(f"⚠ Database writer disabled: missing env vars {missing}")
                self.enabled = False
    
    def write_bronze_layer(
        self,
        *,
        snapshot_dir: Path,
        tournament_key: str,
        calcutta_key: Optional[str] = None,
    ) -> None:
        """Write bronze layer data from snapshot directory."""
        if not self.enabled:
            return
        
        try:
            sd = Path(snapshot_dir)
            
            # Write tournaments
            tournaments_df = pd.DataFrame([{
                'tournament_key': tournament_key,
                'season': int(sd.name) if sd.name.isdigit() else 2025,
                'tournament_name': f"NCAA Tournament {sd.name}",
            }])
            write_tournaments(tournaments_df)
            
            # Write teams
            teams_path = sd / "teams.parquet"
            if teams_path.exists():
                teams_df = pd.read_parquet(teams_path)
                write_teams(teams_df)
            
            # Write calcuttas
            if calcutta_key:
                calcuttas_df = pd.DataFrame([{
                    'calcutta_key': calcutta_key,
                    'tournament_key': tournament_key,
                    'calcutta_name': f"Calcutta {sd.name}",
                    'budget_points': 100,
                }])
                write_calcuttas(calcuttas_df)
                
                # Write entry bids
                entry_bids_path = sd / "entry_bids.parquet"
                if entry_bids_path.exists():
                    entry_bids_df = pd.read_parquet(entry_bids_path)
                    write_entry_bids(calcutta_key, entry_bids_df)
                
                # Write payouts
                payouts_path = sd / "payouts.parquet"
                if payouts_path.exists():
                    payouts_df = pd.read_parquet(payouts_path)
                    write_payouts(calcutta_key, payouts_df)
            
            print("✓ Wrote bronze layer to database")
            
        except Exception as e:
            print(f"⚠ Failed to write bronze layer: {e}")
    
    def write_simulated_tournaments(
        self,
        *,
        tournament_key: str,
        simulations_df: pd.DataFrame,
        snapshot_dir: Path = None,
    ) -> None:
        """Write simulated tournaments to database."""
        if not self.enabled:
            return
        
        try:
            if not snapshot_dir:
                raise ValueError("snapshot_dir required for ID-based writes")
            
            # Extract season from tournament_key
            season = int(tournament_key.split('-')[-1])
            
            # Get or create tournament
            tournament_id = get_or_create_tournament(season)
            
            # Write teams and get ID mapping
            teams_path = snapshot_dir / "teams.parquet"
            if not teams_path.exists():
                raise ValueError(f"Teams file not found: {teams_path}")
            
            teams_df = pd.read_parquet(teams_path)
            team_id_map = write_teams(tournament_id, teams_df)
            
            # Extract school_slug from team_key
            # team_key format: "ncaa-tournament-2025:duke"
            df = simulations_df.copy()
            df['school_slug'] = df['team_key'].str.split(':').str[-1]
            
            # Add required columns
            df['byes'] = 1  # Default: all teams get 1 bye
            df['eliminated'] = df['wins'] < 6  # Champion has 6 wins
            
            # Write simulations with ID mapping
            write_simulated_tournaments(
                tournament_id, df, team_id_map
            )
            print(f"✓ Wrote {len(df)} simulation records to database")
        except Exception as e:
            print(f"⚠ Failed to write simulations: {e}")
    
    def write_predicted_game_outcomes(
        self,
        *,
        tournament_key: str,
        predictions_df: pd.DataFrame,
        model_version: str = None,
        snapshot_dir: Path = None,
    ) -> None:
        """Write game outcome predictions to database."""
        if not self.enabled:
            return
        
        try:
            if not snapshot_dir:
                raise ValueError("snapshot_dir required for ID-based writes")
            
            # Extract season and get tournament_id
            season = int(tournament_key.split('-')[-1])
            tournament_id = get_or_create_tournament(season)
            
            # Get team ID mapping
            teams_path = snapshot_dir / "teams.parquet"
            if not teams_path.exists():
                raise ValueError(f"Teams file not found: {teams_path}")
            
            teams_df = pd.read_parquet(teams_path)
            team_id_map = write_teams(tournament_id, teams_df)
            
            # Transform data
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
            df['round'] = df['round'].map(round_mapping)
            
            # Rename columns and extract school slugs from team keys
            if 'p_team1_wins_given_matchup' in df.columns:
                df['p_team1_wins'] = df['p_team1_wins_given_matchup']
            
            # Extract school_slug from team_key (format: "ncaa-tournament-2025:duke")
            df['team1_slug'] = df['team1_key'].str.split(':').str[-1]
            df['team2_slug'] = df['team2_key'].str.split(':').str[-1]
            
            write_predicted_game_outcomes(
                tournament_id, df, team_id_map, model_version
            )
            print(f"✓ Wrote {len(df)} game predictions to database")
        except Exception as e:
            print(f"⚠ Failed to write game predictions: {e}")
    
    def write_predicted_market_share(
        self,
        *,
        calcutta_key: str,
        predictions_df: pd.DataFrame,
        model_version: str = None,
    ) -> None:
        """Write market share predictions to database."""
        if not self.enabled:
            return
        
        try:
            write_predicted_market_share(
                calcutta_key,
                predictions_df,
                model_version=model_version
            )
            print(f"✓ Wrote {len(predictions_df)} market predictions to database")
        except Exception as e:
            print(f"⚠ Failed to write market predictions: {e}")
    
    def write_optimization_results(
        self,
        *,
        run_id: str,
        calcutta_id: Optional[str],
        strategy: str,
        n_sims: int,
        seed: int,
        budget_points: int,
        recommended_bids_df: pd.DataFrame,
        team_id_map: Dict[str, str],
    ) -> None:
        """Write optimization run and recommended bids to database."""
        if not self.enabled:
            return
        
        try:
            # Write optimization run metadata
            write_optimization_run(
                run_id=run_id,
                strategy=strategy,
                n_sims=n_sims,
                seed=seed,
                budget_points=budget_points,
                calcutta_id=calcutta_id,
            )
            
            # Write recommended bids
            count = write_recommended_entry_bids(
                run_id, recommended_bids_df, team_id_map
            )
            
            print(f"✓ Wrote optimization run {run_id} with {count} bids to database")
        except Exception as e:
            print(f"⚠ Failed to write optimization results: {e}")


# Global instance for easy access
_db_writer: Optional[DatabaseWriter] = None


def get_db_writer() -> DatabaseWriter:
    """Get or create the global database writer instance."""
    global _db_writer
    if _db_writer is None:
        _db_writer = DatabaseWriter()
    return _db_writer

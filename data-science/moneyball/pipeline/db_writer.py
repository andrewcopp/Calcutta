"""
Database writer adapter for pipeline stages.

Integrates database writers alongside parquet file outputs,
maintaining backward compatibility while enabling hybrid architecture.
"""
from __future__ import annotations

import os
from pathlib import Path
from typing import Optional
import pandas as pd

from moneyball.db.writers import (
    write_tournaments,
    write_teams,
    write_simulated_tournaments,
    write_calcuttas,
    write_entry_bids,
    write_payouts,
    write_predicted_game_outcomes,
    write_predicted_market_share,
    write_optimization_run,
    write_recommended_entry_bids,
    write_entry_simulation_outcomes,
    write_entry_performance,
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
    ) -> None:
        """Write simulated tournaments to database."""
        if not self.enabled:
            return
        
        try:
            write_simulated_tournaments(tournament_key, simulations_df)
            print(f"✓ Wrote {len(simulations_df)} simulation records to database")
        except Exception as e:
            print(f"⚠ Failed to write simulations: {e}")
    
    def write_predicted_game_outcomes(
        self,
        *,
        tournament_key: str,
        predictions_df: pd.DataFrame,
        model_version: str = None,
    ) -> None:
        """Write game outcome predictions to database."""
        if not self.enabled:
            return
        
        try:
            write_predicted_game_outcomes(
                tournament_key,
                predictions_df,
                model_version=model_version
            )
            print(f"✓ Wrote {len(predictions_df)} game predictions to database")
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
        calcutta_key: str,
        strategy: str,
        n_sims: int,
        seed: int,
        budget_points: int,
        recommended_bids_df: pd.DataFrame,
    ) -> None:
        """Write optimization run and recommended bids to database."""
        if not self.enabled:
            return
        
        try:
            # Write optimization run metadata
            write_optimization_run(
                run_id=run_id,
                calcutta_key=calcutta_key,
                strategy=strategy,
                n_sims=n_sims,
                seed=seed,
                budget_points=budget_points,
            )
            
            # Write recommended bids
            write_recommended_entry_bids(run_id, recommended_bids_df)
            
            print(f"✓ Wrote optimization run {run_id} to database")
        except Exception as e:
            print(f"⚠ Failed to write optimization results: {e}")
    
    def write_entry_outcomes(
        self,
        *,
        run_id: str,
        outcomes_df: pd.DataFrame,
        summary_df: pd.DataFrame,
    ) -> None:
        """Write entry simulation outcomes and performance to database."""
        if not self.enabled:
            return
        
        try:
            # Write simulation outcomes (high volume)
            if not outcomes_df.empty:
                write_entry_simulation_outcomes(run_id, outcomes_df)
                print(f"✓ Wrote {len(outcomes_df)} entry outcomes to database")
            
            # Write performance summary
            if not summary_df.empty:
                write_entry_performance(run_id, summary_df)
                print(f"✓ Wrote {len(summary_df)} entry performance records to database")
                
        except Exception as e:
            print(f"⚠ Failed to write entry outcomes: {e}")


# Global instance for easy access
_db_writer: Optional[DatabaseWriter] = None


def get_db_writer() -> DatabaseWriter:
    """Get or create the global database writer instance."""
    global _db_writer
    if _db_writer is None:
        _db_writer = DatabaseWriter()
    return _db_writer

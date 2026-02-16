"""
DB-first pipeline runner.

This module provides pipeline stages that read from and write to PostgreSQL
as the primary data source, eliminating parquet file dependencies.
"""
from __future__ import annotations

from typing import Any, Dict, Optional
import uuid

from moneyball.db.readers import (
    read_teams,
    tournament_exists,
    initialize_default_scoring_rules_for_year,
    initialize_default_scoring_rules_for_calcutta,
)
from moneyball.models.predicted_game_outcomes import (
    predict_game_outcomes_from_teams_df,
)
from moneyball.models.simulated_tournaments_db import (
    simulate_tournaments_from_predictions,
)
import os


def stage_predicted_game_outcomes(
    *,
    year: int,
    calcutta_id: Optional[str] = None,
    kenpom_scale: float = 10.0,
    n_sims: int = 5000,
    seed: int = 42,
    model_version: str = "kenpom-v1",
) -> Dict[str, Any]:
    """
    Generate predicted game outcomes from database teams data.
    
    Args:
        year: Tournament year
        calcutta_id: Optional calcutta UUID
        kenpom_scale: KenPom scaling factor
        n_sims: Number of simulations for prediction
        seed: Random seed
        model_version: Model version identifier
        
    Returns:
        Dictionary with stage results
    """
    if (
        os.getenv("CALCUTTA_ALLOW_PYTHON_SILVER_WRITES", "false").lower()
        != "true"
    ):
        raise RuntimeError(
            (
                "This stage writes derived.predicted_game_outcomes from "
                "Python, which is disabled. Use the Go-first pipeline, or "
                "set CALCUTTA_ALLOW_PYTHON_SILVER_WRITES=true to override."
            )
        )

    print(f"⚙ Generating predicted_game_outcomes for {year}...")
    
    # Ensure tournament exists
    if not tournament_exists(year):
        raise ValueError(f"Tournament for year {year} not found in database")

    # Read teams from database
    teams_df = read_teams(year)
    if teams_df.empty:
        raise ValueError(f"No teams found for year {year}")
    
    print(f"  Found {len(teams_df)} teams")
    
    # Generate predictions
    predictions_df = predict_game_outcomes_from_teams_df(
        teams_df=teams_df,
        kenpom_scale=kenpom_scale,
        n_sims=n_sims,
        seed=seed,
    )
    
    print(f"  Generated {len(predictions_df)} game predictions")
    
    # Write to database using direct writer
    from moneyball.db.writers.silver_writers import (
        write_predicted_game_outcomes,
    )
    from moneyball.db.writers.bronze_writers import get_or_create_tournament
    
    tournament_id = get_or_create_tournament(year)
    
    try:
        write_predicted_game_outcomes(
            tournament_id=tournament_id,
            predictions_df=predictions_df,
            team_id_map={},  # Not needed - predictions already have team_id
            model_version=model_version,
        )
        print("✓ Predicted game outcomes written to database")
    except Exception as e:
        print(f"⚠ Failed to write predictions: {e}")
    
    return {
        "year": year,
        "n_predictions": len(predictions_df),
        "model_version": model_version,
    }


def stage_simulate_tournaments(
    *,
    year: int,
    n_sims: int = 5000,
    seed: int = 42,
    run_id: Optional[str] = None,
    calcutta_id: Optional[str] = None,
) -> Dict[str, Any]:
    """
    Simulate tournaments using predictions from database.
    
    Args:
        year: Tournament year
        n_sims: Number of simulations
        seed: Random seed
        run_id: Optional run ID (generated if not provided)
        
    Returns:
        Dictionary with stage results including run_id
    """
    if (
        os.getenv("CALCUTTA_ALLOW_PYTHON_SILVER_WRITES", "false").lower()
        != "true"
    ):
        raise RuntimeError(
            (
                "This stage writes simulation artifacts from Python, which "
                "is disabled. Use the Go-first pipeline, or set "
                "CALCUTTA_ALLOW_PYTHON_SILVER_WRITES=true to override."
            )
        )

    print(f"⚙ Simulating tournaments for {year} (n_sims={n_sims})...")
    
    # Ensure tournament exists
    if not tournament_exists(year):
        raise ValueError(f"Tournament for year {year} not found in database")

    if calcutta_id is not None:
        points_by_win_index = initialize_default_scoring_rules_for_calcutta(
            calcutta_id
        )
    else:
        points_by_win_index = initialize_default_scoring_rules_for_year(year)
    
    # Read teams and predictions from database
    teams_df = read_teams(year)
    
    # For now, we'll generate predictions inline if they don't exist
    # In production, you'd read from silver_predicted_game_outcomes
    from moneyball.models.predicted_game_outcomes import (
        predict_game_outcomes_from_teams_df,
    )
    predictions_df = predict_game_outcomes_from_teams_df(
        teams_df=teams_df,
        kenpom_scale=10.0,
        n_sims=n_sims,
        seed=seed,
    )
    
    print(f"  Running {n_sims} simulations...")
    
    # Generate simulations
    simulations_df = simulate_tournaments_from_predictions(
        predictions_df=predictions_df,
        teams_df=teams_df,
        points_by_win_index=points_by_win_index,
        n_sims=n_sims,
        seed=seed,
    )
    
    print(f"  Generated {len(simulations_df)} simulation results")
    
    # Generate run_id if not provided
    if run_id is None:
        run_id = str(uuid.uuid4())
    
    # Write to database using direct writer
    from moneyball.db.writers.silver_writers import write_simulated_tournaments
    from moneyball.db.writers.bronze_writers import get_or_create_tournament
    
    tournament_id = get_or_create_tournament(year)
    
    # Read teams to get school_slug mapping
    teams_df = read_teams(year)
    team_id_map = {
        str(row['school_slug']): str(row['id'])
        for _, row in teams_df.iterrows()
    }
    
    # Prepare simulations with school_slug and eliminated flag
    sim_df = simulations_df.copy()
    
    # Map team_id back to school_slug for the writer
    id_to_slug = {
        str(row['id']): str(row['school_slug'])
        for _, row in teams_df.iterrows()
    }
    sim_df['school_slug'] = sim_df['team_id'].map(id_to_slug)
    
    # Add eliminated flag (teams with < 6 wins are eliminated)
    sim_df['eliminated'] = sim_df['wins'] < 6
    
    try:
        write_simulated_tournaments(
            tournament_id=tournament_id,
            simulations_df=sim_df,
            team_id_map=team_id_map,
        )
        print(f"✓ Simulated tournaments written to database (run_id={run_id})")
    except Exception as e:
        print(f"⚠ Failed to write simulations: {e}")
    
    return {
        "year": year,
        "n_sims": n_sims,
        "run_id": run_id,
        "n_results": len(simulations_df),
    }


def run_full_pipeline(
    *,
    year: int,
    n_sims: int = 5000,
    seed: int = 42,
    strategy: str = "greedy",
    kenpom_scale: float = 10.0,
    calcutta_id: Optional[str] = None,
) -> Dict[str, Any]:
    """
    Run the full pipeline for a given year.
    
    Args:
        year: Tournament year
        n_sims: Number of simulations
        seed: Random seed
        strategy: Portfolio strategy
        kenpom_scale: KenPom scaling factor
        calcutta_id: Optional calcutta UUID
        
    Returns:
        Dictionary with results from all stages
    """
    print(f"\n{'='*60}")
    print(f"Running full pipeline for {year}")
    print(f"{'='*60}\n")
    
    results = {}
    
    # Stage 1: Predicted game outcomes
    results['predicted_game_outcomes'] = stage_predicted_game_outcomes(
        year=year,
        calcutta_id=calcutta_id,
        kenpom_scale=kenpom_scale,
        n_sims=n_sims,
        seed=seed,
    )
    
    # Stage 2: Simulate tournaments
    results['simulated_tournaments'] = stage_simulate_tournaments(
        year=year,
        n_sims=n_sims,
        seed=seed,
        calcutta_id=calcutta_id,
    )

    print(f"\n{'='*60}")
    print(f"Pipeline complete for {year}")
    print(f"{'='*60}\n")
    
    return results

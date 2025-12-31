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
    get_latest_run_id,
)
from moneyball.models.predicted_game_outcomes import (
    predict_game_outcomes_from_teams_df,
)
from moneyball.models.simulated_tournaments_db import (
    simulate_tournaments_from_predictions,
)
from moneyball.models.recommended_entry_bids_db import (
    recommend_entry_bids_from_simulations,
)
from moneyball.pipeline.db_writer import get_db_writer
import subprocess
import os


def stage_simulated_calcuttas(
    *,
    year: int,
    run_id: str,
) -> Dict[str, Any]:
    """
    Stage 4: Calculate simulated calcutta outcomes.
    
    Calls the Go service to calculate entry performance across all simulations.
    
    Args:
        year: Tournament year
        run_id: Optimization run ID
        
    Returns:
        Dictionary with results
    """
    print(f"\n{'='*60}")
    print(f"Stage 4: Calculating simulated calcutta outcomes")
    print(f"{'='*60}\n")
    
    # Get tournament ID from database
    from moneyball.db.connection import get_db_connection
    with get_db_connection() as conn:
        with conn.cursor() as cur:
            cur.execute(
                "SELECT id FROM bronze_tournaments WHERE season = %s",
                (year,)
            )
            result = cur.fetchone()
            if not result:
                raise ValueError(f"No tournament found for year {year}")
            tournament_id = str(result[0])
    
    print(f"Tournament ID: {tournament_id}")
    print(f"Run ID: {run_id}")
    
    # Call the Go service
    # Navigate from data-science/moneyball/pipeline to backend/bin
    go_binary = os.path.join(
        os.path.dirname(__file__),
        "../../../backend/bin/calculate-simulated-calcuttas"
    )
    
    # Resolve to absolute path
    go_binary = os.path.abspath(go_binary)
    
    if not os.path.exists(go_binary):
        raise FileNotFoundError(
            f"Go binary not found at {go_binary}. "
            "Please build it first: cd backend && go build -o bin/calculate-simulated-calcuttas ./cmd/calculate-simulated-calcuttas"
        )
    
    env = os.environ.copy()
    env['EXCLUDED_ENTRY_NAME'] = env.get('EXCLUDED_ENTRY_NAME', 'Andrew Copp')
    env['DB_HOST'] = env.get('DB_HOST', 'localhost')
    env['DB_PORT'] = env.get('DB_PORT', '5432')
    env['DB_USER'] = env.get('DB_USER', 'calcutta')
    env['DB_PASSWORD'] = env.get('DB_PASSWORD', 'calcutta')
    env['DB_NAME'] = env.get('DB_NAME', 'calcutta')
    
    print(f"Calling Go service: {go_binary}")
    print(f"Excluded entry: {env['EXCLUDED_ENTRY_NAME']}")
    
    result = subprocess.run(
        [go_binary, tournament_id, run_id],
        env=env,
        capture_output=True,
        text=True,
    )
    
    if result.returncode != 0:
        print(f"STDOUT: {result.stdout}")
        print(f"STDERR: {result.stderr}")
        raise RuntimeError(f"Go service failed with exit code {result.returncode}")
    
    print(result.stdout)
    
    # Query results from database
    with get_db_connection() as conn:
        with conn.cursor() as cur:
            cur.execute(
                """
                SELECT 
                    entry_name,
                    mean_payout,
                    p_top1,
                    p_in_money
                FROM gold_entry_performance
                WHERE run_id = %s
                ORDER BY mean_payout DESC
                """,
                (run_id,)
            )
            entries = cur.fetchall()
    
    print(f"\nEntry Performance:")
    for entry_name, mean_payout, p_top1, p_in_money in entries:
        print(f"  {entry_name}: mean={mean_payout:.3f}, P(top1)={p_top1:.1%}, P(in money)={p_in_money:.1%}")
    
    return {
        'run_id': run_id,
        'tournament_id': tournament_id,
        'entries_evaluated': len(entries),
    }


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
    from moneyball.db.writers.silver_writers import write_predicted_game_outcomes
    from moneyball.db.writers.bronze_writers import get_or_create_tournament
    
    tournament_id = get_or_create_tournament(year)
    
    try:
        write_predicted_game_outcomes(
            tournament_id=tournament_id,
            predictions_df=predictions_df,
            team_id_map={},  # Not needed - predictions already have team_id
            model_version=model_version,
        )
        print(f"✓ Predicted game outcomes written to database")
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
    print(f"⚙ Simulating tournaments for {year} (n_sims={n_sims})...")
    
    # Ensure tournament exists
    if not tournament_exists(year):
        raise ValueError(f"Tournament for year {year} not found in database")
    
    # Read teams and predictions from database
    teams_df = read_teams(year)
    
    # For now, we'll generate predictions inline if they don't exist
    # In production, you'd read from silver_predicted_game_outcomes
    from moneyball.models.predicted_game_outcomes import predict_game_outcomes_from_teams_df
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


def stage_recommended_entry_bids(
    *,
    year: int,
    strategy: str = "greedy",
    run_id: str = None,
    budget_points: int = 100,
    min_teams: int = 3,
    max_teams: int = 10,
    min_bid: int = 1,
    max_bid: int = 50,
) -> dict:
    """
    Generate recommended entry bids by calling standalone optimizer scripts.
    
    For MINLP strategy, calls scripts/run_minlp_optimizer.py which works correctly.
    For other strategies, uses the inline implementation.
    
    Args:
        year: Tournament year
        strategy: Portfolio strategy (greedy, minlp, etc.)
        run_id: Run ID from simulations (if None, will find latest)
        budget_points: Total budget in points
        min_teams: Minimum number of teams
        max_teams: Maximum number of teams
        min_bid: Minimum bid per team
        max_bid: Maximum bid per team
        
    Returns:
        Dictionary with results
    """
    print(f"⚙ Generating recommended entry bids for {year}...")
    
    # For MINLP strategy, call the standalone script that works correctly
    if strategy == "minlp":
        print(f"  Calling standalone MINLP optimizer script...")
        
        import subprocess
        import sys
        
        # Get the path to the standalone script
        script_path = os.path.join(
            os.path.dirname(__file__),
            "../../scripts/run_minlp_optimizer.py"
        )
        script_path = os.path.abspath(script_path)
        
        # Run the standalone script
        result = subprocess.run(
            [sys.executable, script_path, str(year)],
            capture_output=True,
            text=True,
            env=os.environ.copy(),
        )
        
        if result.returncode != 0:
            print(f"  ⚠ MINLP optimizer failed:")
            print(result.stderr)
            raise RuntimeError(f"MINLP optimizer failed with code {result.returncode}")
        
        # Parse the run_id from the output
        import re
        match = re.search(r'Run ID: (minlp_\d+_\d+T\d+)', result.stdout)
        if match:
            run_id = match.group(1)
            print(f"  ✓ MINLP optimization complete (run_id={run_id})")
        else:
            raise RuntimeError("Failed to parse run_id from MINLP optimizer output")
        
        # Count recommendations
        from moneyball.db.readers import get_db_connection
        with get_db_connection() as conn:
            with conn.cursor() as cur:
                cur.execute("""
                    SELECT COUNT(*), SUM(recommended_bid_points)
                    FROM gold_recommended_entry_bids
                    WHERE run_id = %s
                """, (run_id,))
                n_recs, total_bid = cur.fetchone()
        
        return {
            'year': year,
            'strategy': strategy,
            'run_id': run_id,
            'n_recommendations': n_recs,
            'total_bid': int(total_bid) if total_bid else 0,
        }
    
    # For other strategies, use inline implementation
    if run_id is None:
        # Find latest run_id for this year
        from moneyball.db.readers import get_db_connection
        with get_db_connection() as conn:
            with conn.cursor() as cur:
                cur.execute("""
                    SELECT gor.run_id
                    FROM gold_optimization_runs gor
                    JOIN bronze_calcuttas bc ON gor.calcutta_id = bc.id
                    WHERE bc.tournament_id = (
                        SELECT id FROM bronze_tournaments WHERE season = %s
                    )
                    ORDER BY gor.created_at DESC
                    LIMIT 1
                """, (year,))
                result = cur.fetchone()
                if result:
                    run_id = result[0]
                    print(f"  Using run_id: {run_id}")
                else:
                    raise ValueError(f"No optimization runs found for year {year}")
    else:
        print(f"  Using run_id: {run_id}")
    
    # Read simulations from database
    from moneyball.db.readers import read_simulated_tournaments
    simulations_df = read_simulated_tournaments(year)
    
    print(f"  Found {len(simulations_df)} simulation results")
    
    # Generate recommendations
    recommendations_df = recommend_entry_bids_from_simulations(
        simulations_df=simulations_df,
        strategy=strategy,
        budget_points=budget_points,
        min_teams=min_teams,
        max_teams=max_teams,
        min_bid=min_bid,
        max_bid=max_bid,
    )
    
    print(f"  Generated {len(recommendations_df)} recommendations")
    
    # Write to database
    from moneyball.db.writers.gold_writers import write_optimization_run, write_recommended_entry_bids
    from moneyball.db.readers import get_db_connection
    
    # Get tournament_id and calcutta_id
    with get_db_connection() as conn:
        with conn.cursor() as cur:
            cur.execute("""
                SELECT bt.id, bc.id
                FROM bronze_tournaments bt
                LEFT JOIN bronze_calcuttas bc ON bc.tournament_id = bt.id
                WHERE bt.season = %s
            """, (year,))
            result = cur.fetchone()
            if not result:
                raise ValueError(f"No tournament found for year {year}")
            tournament_id, calcutta_id = result
    
    # Write optimization run
    write_optimization_run(
        run_id=run_id,
        calcutta_id=calcutta_id,
        strategy=strategy,
    )
    
    # Write recommendations
    write_recommended_entry_bids(
        run_id=run_id,
        recommendations_df=recommendations_df,
    )
    
    print(f"✓ Recommended entry bids written to database (run_id={run_id})")
    
    return {
        'year': year,
        'strategy': strategy,
        'run_id': run_id,
        'n_recommendations': len(recommendations_df),
        'total_bid': recommendations_df['bid_amount_points'].sum(),
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
    )
    
    # Stage 3: Recommended entry bids
    results['recommended_entry_bids'] = stage_recommended_entry_bids(
        year=year,
        strategy=strategy,
        run_id=results['simulated_tournaments']['run_id'],
        calcutta_id=calcutta_id,
    )
    
    # Stage 4: Evaluate simulated entry via simulated calcuttas
    results['simulated_calcuttas'] = stage_simulated_calcuttas(
        year=year,
        run_id=results['simulated_tournaments']['run_id'],
    )
    
    print(f"\n{'='*60}")
    print(f"Pipeline complete for {year}")
    print(f"{'='*60}\n")
    
    return results

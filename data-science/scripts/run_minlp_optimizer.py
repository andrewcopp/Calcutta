"""
Run MINLP portfolio optimizer for a tournament and write results to database.

This script:
1. Reads team data from database (expected points from simulations)
2. Uses expected points as predicted market (Naive approach)
3. Runs MINLP optimizer to find optimal portfolio allocation
4. Writes results to gold_recommended_entry_bids table
"""
import sys
import uuid
from datetime import datetime
import pandas as pd

from moneyball.db.connection import get_db_connection
from moneyball.models.portfolio_optimizer_minlp import optimize_portfolio_minlp
from moneyball.db.writers.gold_writers import write_optimization_run, write_recommended_entry_bids


def run_optimizer(year: int = 2025):
    """Run MINLP optimizer for a tournament year."""
    print(f"Running MINLP optimizer for {year}...")
    print("=" * 80)
    
    # Read team data from database
    with get_db_connection() as conn:
        with conn.cursor() as cur:
            # Get tournament ID
            cur.execute("""
                SELECT id FROM bronze_tournaments WHERE season = %s
            """, (year,))
            result = cur.fetchone()
            if not result:
                print(f"Error: No tournament found for year {year}")
                return
            tournament_id = result[0]
            print(f"Tournament ID: {tournament_id}")
            
            # Get team data with expected points from simulations
            cur.execute("""
                WITH team_win_counts AS (
                    SELECT 
                        st.team_id,
                        st.wins,
                        COUNT(*) as sim_count
                    FROM silver_simulated_tournaments st
                    WHERE st.tournament_id = %s
                    GROUP BY st.team_id, st.wins
                ),
                team_probabilities AS (
                    SELECT 
                        team_id,
                        SUM(sim_count)::float as total_sims,
                        SUM(CASE WHEN wins = 1 THEN sim_count ELSE 0 END)::float as win_r64,
                        SUM(CASE WHEN wins = 2 THEN sim_count ELSE 0 END)::float as win_r32,
                        SUM(CASE WHEN wins = 3 THEN sim_count ELSE 0 END)::float as win_s16,
                        SUM(CASE WHEN wins = 4 THEN sim_count ELSE 0 END)::float as win_e8,
                        SUM(CASE WHEN wins = 5 THEN sim_count ELSE 0 END)::float as win_ff,
                        SUM(CASE WHEN wins = 6 THEN sim_count ELSE 0 END)::float as win_champ
                    FROM team_win_counts
                    GROUP BY team_id
                )
                SELECT 
                    t.id as team_id,
                    t.school_slug as team_key,
                    t.school_name,
                    t.seed,
                    t.region,
                    -- Expected points (EV calculation)
                    (COALESCE(tp.win_r64 / NULLIF(tp.total_sims, 0), 0) * 50 + 
                     COALESCE(tp.win_r32 / NULLIF(tp.total_sims, 0), 0) * 150 + 
                     COALESCE(tp.win_s16 / NULLIF(tp.total_sims, 0), 0) * 300 + 
                     COALESCE(tp.win_e8 / NULLIF(tp.total_sims, 0), 0) * 500 + 
                     COALESCE(tp.win_ff / NULLIF(tp.total_sims, 0), 0) * 750 + 
                     COALESCE(tp.win_champ / NULLIF(tp.total_sims, 0), 0) * 1050) as expected_team_points
                FROM bronze_teams t
                LEFT JOIN team_probabilities tp ON t.id = tp.team_id
                WHERE t.tournament_id = %s
                ORDER BY t.seed ASC, t.school_name ASC
            """, (tournament_id, tournament_id))
            
            rows = cur.fetchall()
            teams_df = pd.DataFrame(rows, columns=[
                'team_id', 'team_key', 'school_name', 'seed', 'region', 'expected_team_points'
            ])
    
    print(f"Loaded {len(teams_df)} teams")
    print(f"Total expected points: {teams_df['expected_team_points'].sum():.1f}")
    
    # Use expected points as predicted market (Naive approach)
    teams_df['predicted_team_total_bids'] = teams_df['expected_team_points']
    
    # Calculate expected ROI for reference
    teams_df['expected_roi'] = teams_df['expected_team_points'] / teams_df['predicted_team_total_bids'].replace(0, 1)
    
    print("\nTop 10 teams by expected ROI:")
    print(teams_df.nlargest(10, 'expected_roi')[['school_name', 'seed', 'expected_team_points', 'expected_roi']])
    
    # Run MINLP optimizer
    print("\nRunning MINLP optimizer...")
    try:
        chosen_df, portfolio_rows = optimize_portfolio_minlp(
            teams_df=teams_df,
            budget_points=100,
            min_teams=3,
            max_teams=10,
            max_per_team_points=50,
            min_bid_points=1,
            initial_solution="greedy",
            max_iterations=1000,
        )
        
        print(f"\nOptimization complete!")
        print(f"Selected {len(chosen_df)} teams")
        print(f"Total budget allocated: {chosen_df['bid_amount_points'].sum()}")
        
        print("\nPortfolio allocation:")
        for _, row in chosen_df.sort_values('bid_amount_points', ascending=False).iterrows():
            print(f"  {row['school_name']:30s} Seed {row['seed']:2d}  Bid: ${row['bid_amount_points']:2d}  "
                  f"Exp Pts: {row['expected_team_points']:6.1f}  ROI: {row.get('score', 0):.2f}")
        
        # Calculate expected return
        total_return = 0
        for _, row in chosen_df.iterrows():
            bid = row['bid_amount_points']
            exp_pts = row['expected_team_points']
            pred_market = row['predicted_team_total_bids']
            ownership = bid / (pred_market + bid)
            total_return += exp_pts * ownership
        
        print(f"\nExpected total return: {total_return:.2f} points")
        print(f"Expected ROI: {total_return / 100:.2f}x")
        
    except Exception as e:
        print(f"Error running optimizer: {e}")
        import traceback
        traceback.print_exc()
        return
    
    # Write to database
    print("\nWriting results to database...")
    
    # Create run_id
    run_id = f"minlp_{year}_{datetime.now().strftime('%Y%m%dT%H%M%S')}"
    
    # Create team_id_map
    team_id_map = {
        str(row['team_key']): str(row['team_id'])
        for _, row in teams_df.iterrows()
    }
    
    # Prepare bids DataFrame
    bids_df = chosen_df[['team_key', 'bid_amount_points']].copy()
    bids_df['expected_roi'] = chosen_df.get('score', 0.0)
    
    try:
        # Write optimization run
        write_optimization_run(
            run_id=run_id,
            calcutta_id=None,  # No calcutta yet
            strategy="minlp",
            n_sims=5000,
            seed=42,
            budget_points=100,
        )
        print(f"  Wrote optimization run: {run_id}")
        
        # Write recommended bids
        count = write_recommended_entry_bids(
            run_id=run_id,
            bids_df=bids_df,
            team_id_map=team_id_map,
        )
        print(f"  Wrote {count} recommended bids")
        
        print("\n✓ Success! Portfolio optimization complete.")
        print(f"  Run ID: {run_id}")
        print("\nThe Simulated Entry tab will now show MINLP allocations in the 'Our Bid' column.")
        
    except Exception as e:
        print(f"Error writing to database: {e}")
        import traceback
        traceback.print_exc()
        return


if __name__ == "__main__":
    year = int(sys.argv[1]) if len(sys.argv) > 1 else 2025
    run_optimizer(year)

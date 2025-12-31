"""
Run ridge regression model to predict market share and write to database.

This script:
1. Loads team data and historical market data
2. Runs ridge regression model to predict market share
3. Writes predictions to silver_predicted_market_share table
"""
import sys
from pathlib import Path
import pandas as pd

from moneyball.db.connection import get_db_connection
from moneyball.models.predicted_auction_share_of_pool import (
    predict_auction_share_of_pool_from_out_root,
)
from moneyball.db.writers import write_predicted_market_share


def run_ridge_regression(year: int = 2025):
    """Run ridge regression for a tournament year."""
    print(f"Running ridge regression for {year}...")
    print("=" * 80)
    
    # Get tournament and calcutta IDs from database
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
            
            # Get latest calcutta ID
            cur.execute("""
                SELECT id FROM bronze_calcuttas
                WHERE tournament_id = %s
                ORDER BY created_at DESC
                LIMIT 1
            """, (tournament_id,))
            result = cur.fetchone()
            if not result:
                print(f"Error: No calcutta found for tournament {tournament_id}")
                return
            calcutta_id = result[0]
            print(f"Calcutta ID: {calcutta_id}")
            
            # Get team_id map (school_slug -> team_id)
            cur.execute("""
                SELECT school_slug, id
                FROM bronze_teams
                WHERE tournament_id = %s
            """, (tournament_id,))
            team_id_map = {row[0]: row[1] for row in cur.fetchall()}
            print(f"Loaded {len(team_id_map)} teams")
    
    # Run ridge regression model
    print("\nRunning ridge regression model...")
    out_root = Path("out")
    
    try:
        predictions = predict_auction_share_of_pool_from_out_root(
            out_root=out_root,
            predict_snapshot=str(year),
            train_snapshots=["2017", "2018", "2019", "2021", "2022", "2023", "2024"],
            ridge_alpha=1.0,
            feature_set="optimal",
        )
        
        print(f"Generated predictions for {len(predictions)} teams")
        print(f"Total predicted share: {predictions['predicted_auction_share_of_pool'].sum():.4f}")
        
        # Show top 5 predictions
        print("\nTop 5 predicted market shares:")
        top5 = predictions.nlargest(5, 'predicted_auction_share_of_pool')
        for _, row in top5.iterrows():
            print(f"  {row['team_key']:30s} {row['predicted_auction_share_of_pool']:.4f}")
        
    except Exception as e:
        print(f"Error running ridge regression: {e}")
        import traceback
        traceback.print_exc()
        return
    
    # Write to database
    print("\nWriting predictions to database...")
    try:
        count = write_predicted_market_share(
            calcutta_id=calcutta_id,
            predictions_df=predictions,
            team_id_map=team_id_map,
        )
        print(f"✓ Wrote {count} predictions to silver_predicted_market_share")
        
    except Exception as e:
        print(f"Error writing to database: {e}")
        import traceback
        traceback.print_exc()
        return
    
    print("\n✓ Success! Ridge regression predictions written to database.")
    print("The Predicted Investment API will now return non-zero values.")


if __name__ == "__main__":
    year = int(sys.argv[1]) if len(sys.argv) > 1 else 2025
    run_ridge_regression(year)

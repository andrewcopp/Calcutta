"""
Run ridge regression model to predict market share and write to database.

This script:
1. Loads team data and historical market data
2. Runs ridge regression model to predict market share
3. Writes predictions to silver_predicted_market_share table
"""
import sys
from pathlib import Path

from moneyball.db.connection import get_db_connection
from moneyball.models.predicted_auction_share_of_pool import (
    predict_auction_share_of_pool_from_out_root,
)
from moneyball.db.writers import write_predicted_market_share


def run_ridge_regression(year: int = 2025):
    """Run ridge regression for a tournament year."""
    print(f"Running ridge regression for {year}...")
    print("=" * 80)
    
    # Get excluded entry name from environment
    import os
    excluded_entry_name = os.getenv('EXCLUDED_ENTRY_NAME', 'Andrew Copp')
    print(f"Excluding entry from training: {excluded_entry_name}")
    
    # Get tournament ID from database
    with get_db_connection() as conn:
        with conn.cursor() as cur:
            # Get tournament ID
            cur.execute("""
                SELECT id FROM lab_bronze.tournaments WHERE season = %s
            """, (year,))
            result = cur.fetchone()
            if not result:
                print(f"Error: No tournament found for year {year}")
                return
            tournament_id = result[0]
            print(f"Tournament ID: {tournament_id}")
            
            # Get team_id map (school_slug -> team_id UUID)
            # Ridge regression returns keys like "ncaa-tournament-2025:duke"
            # We need to map "duke" (school_slug) -> UUID
            cur.execute("""
                SELECT school_slug, id
                FROM lab_bronze.teams
                WHERE tournament_id = %s
            """, (tournament_id,))
            team_id_map = {row[0]: str(row[1]) for row in cur.fetchall()}
            print(f"Loaded {len(team_id_map)} teams")
            print(f"Sample mapping: {list(team_id_map.items())[:3]}")
    
    # Run ridge regression model
    print("\nRunning ridge regression model...")
    out_root = Path("out")
    
    # Dynamically determine training years (all years except target year)
    all_years = [
        "2017",
        "2018",
        "2019",
        "2021",
        "2022",
        "2023",
        "2024",
        "2025",
    ]
    train_snapshots = [y for y in all_years if y != str(year)]
    print(f"Training on years: {train_snapshots}")
    print(f"Predicting for year: {year}")

    exclude_entry_names = (
        [excluded_entry_name] if excluded_entry_name else None
    )
    
    try:
        predictions = predict_auction_share_of_pool_from_out_root(
            out_root=out_root,
            predict_snapshot=str(year),
            train_snapshots=train_snapshots,
            ridge_alpha=1.0,
            feature_set="optimal",
            exclude_entry_names=exclude_entry_names,
        )
        
        print(f"Generated predictions for {len(predictions)} teams")
        total_share = predictions['predicted_auction_share_of_pool'].sum()
        print(f"Total predicted share: {total_share:.4f}")
        
        # Show top 5 predictions
        print("\nTop 5 predicted market shares:")
        top5 = predictions.nlargest(5, 'predicted_auction_share_of_pool')
        for _, row in top5.iterrows():
            team_key = row['team_key']
            share = row['predicted_auction_share_of_pool']
            print(f"  {team_key:30s} {share:.4f}")
        
    except Exception as e:
        print(f"Error running ridge regression: {e}")
        import traceback
        traceback.print_exc()
        return
    
    # Write to database
    print("\nWriting predictions to database...")
    try:
        # Strip tournament prefix from team_key (e.g., "...:duke" -> "duke")
        predictions['team_key'] = predictions['team_key'].str.split(':').str[-1]
        
        count = write_predicted_market_share(
            predictions_df=predictions,
            team_id_map=team_id_map,
            tournament_id=tournament_id,
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

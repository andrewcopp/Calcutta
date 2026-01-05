"""
Run ridge regression model to predict market share and write to database.

This script:
1. Loads team data and historical market data
2. Runs ridge regression model to predict market share
3. Writes predictions to silver_predicted_market_share table
"""
import sys
from pathlib import Path


def run_ridge_regression(year: int = 2025):
    """Run ridge regression for a tournament year."""
    project_root = Path(__file__).resolve().parents[1]
    if str(project_root) not in sys.path:
        sys.path.insert(0, str(project_root))

    from moneyball.db.connection import get_db_connection
    from moneyball.db.readers import (
        initialize_default_scoring_rules_for_year,
        read_ridge_team_dataset_for_year,
    )
    from moneyball.models.predicted_auction_share_of_pool import (
        predict_auction_share_of_pool,
    )
    from moneyball.db.writers import write_predicted_market_share

    print(f"Running ridge regression for {year}...")
    print("=" * 80)

    # Get excluded entry name from environment
    import os
    excluded_entry_name = os.getenv('EXCLUDED_ENTRY_NAME', 'Andrew Copp')
    print(f"Excluding entry from training: {excluded_entry_name}")

    # Resolve core tournament + calcutta + team slug->uuid mapping from database
    with get_db_connection() as conn:
        with conn.cursor() as cur:
            cur.execute("""
                SELECT t.id
                FROM core.tournaments t
                JOIN core.seasons s
                  ON s.id = t.season_id
                 AND s.deleted_at IS NULL
                WHERE s.year = %s
                  AND t.deleted_at IS NULL
                ORDER BY t.created_at DESC
                LIMIT 1
            """, (year,))
            result = cur.fetchone()
            if not result:
                print(f"Error: No tournament found for year {year}")
                return
            tournament_id = result[0]
            print(f"Tournament ID: {tournament_id}")

            cur.execute(
                """
                SELECT c.id
                FROM core.calcuttas c
                WHERE c.tournament_id = %s
                  AND c.deleted_at IS NULL
                ORDER BY c.created_at DESC
                LIMIT 1
                """,
                (tournament_id,),
            )
            calcutta_row = cur.fetchone()
            calcutta_id = str(calcutta_row[0]) if calcutta_row else None
            print(f"Calcutta ID: {calcutta_id}")

            # Get team_id map (school_slug -> team_id UUID)
            # Ridge regression returns keys like "ncaa-tournament-2025:duke"
            # We need to map "duke" (school_slug) -> UUID.
            # In core schema, this is core.schools.slug.
            cur.execute("""
                SELECT s.slug, t.id
                FROM core.teams t
                JOIN core.schools s
                  ON s.id = t.school_id
                  AND s.deleted_at IS NULL
                WHERE t.tournament_id = %s
                  AND t.deleted_at IS NULL
            """, (tournament_id,))
            team_id_map = {row[0]: str(row[1]) for row in cur.fetchall()}
            print(f"Loaded {len(team_id_map)} teams")
            print(f"Sample mapping: {list(team_id_map.items())[:3]}")

    initialize_default_scoring_rules_for_year(year)

    # Run ridge regression model
    print("\nRunning ridge regression model...")

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
    train_years = [int(y) for y in all_years if int(y) != int(year)]
    print(f"Training on years: {[str(y) for y in train_years]}")
    print(f"Predicting for year: {year}")

    exclude_entry_names = (
        [excluded_entry_name] if excluded_entry_name else None
    )

    try:
        train_frames = []
        for y in train_years:
            try:
                df = read_ridge_team_dataset_for_year(
                    y,
                    exclude_entry_names=exclude_entry_names,
                    include_target=True,
                )
            except Exception as e:
                print(f"Skipping training year {y}: {e}")
                continue
            train_frames.append(df)

        train_ds = (
            train_frames[0].iloc[0:0].copy()
            if not train_frames
            else __import__("pandas").concat(train_frames, ignore_index=True)
        )

        if "team_share_of_pool" in train_ds.columns:
            train_ds = train_ds[train_ds["team_share_of_pool"].notna()].copy()

        if train_ds.empty:
            raise ValueError("no training rows (team_share_of_pool all NULL)")

        predict_ds = read_ridge_team_dataset_for_year(
            int(year),
            exclude_entry_names=None,
            include_target=False,
        )

        predictions = predict_auction_share_of_pool(
            train_team_dataset=train_ds,
            predict_team_dataset=predict_ds,
            ridge_alpha=1.0,
            feature_set="optimal",
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
            calcutta_id=calcutta_id,
            tournament_id=tournament_id,
            algorithm_name="ridge",
            params={
                "ridge_alpha": 1.0,
                "feature_set": "optimal",
                "excluded_entry_name": excluded_entry_name,
            },
        )
        print(f"✓ Wrote {count} predictions to derived.predicted_market_share")

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

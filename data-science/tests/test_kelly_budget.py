"""Test kelly strategy budget allocation."""
import pandas as pd
from moneyball.models.portfolio_strategies import allocate_kelly

# Create sample teams data
teams_df = pd.DataFrame({
    'team_key': ['team_a', 'team_b', 'team_c', 'team_d', 'team_e'],
    'expected_team_points': [100, 80, 60, 40, 20],
    'predicted_team_total_bids': [20, 15, 10, 8, 5],
    'predicted_auction_share_of_pool': [0.2, 0.15, 0.1, 0.08, 0.05],
    'score': [1.0, 0.8, 0.6, 0.4, 0.2],
})

result = allocate_kelly(
    teams_df=teams_df,
    budget_points=100,
    min_teams=3,
    max_teams=10,
    max_per_team_points=50,
    min_bid_points=1,
)

print("Kelly allocation result:")
print(result)
print(f"\nTotal allocated: {result['bid_amount_points'].sum()}")
print(f"Expected: 100")
print(f"Match: {int(result['bid_amount_points'].sum()) == 100}")

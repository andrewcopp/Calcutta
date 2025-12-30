from __future__ import annotations

from typing import List

import pandas as pd


def add_within_seed_features(df: pd.DataFrame) -> pd.DataFrame:
    """
    Add within-seed ranking features to detect "overlooked #3" mispricing.

    Hypothesis: In a 4-team seed pod (especially 1-seeds), the "3rd
    perceived best" is systematically underbid relative to strength.
    """
    out = df.copy()
    
    if "seed" not in out.columns or "kenpom_net" not in out.columns:
        return out
    
    out["seed"] = pd.to_numeric(out["seed"], errors="coerce")
    out["kenpom_net"] = pd.to_numeric(out["kenpom_net"], errors="coerce")
    
    out = out.sort_values(["seed", "kenpom_net"], ascending=[True, False])
    out["within_seed_rank"] = out.groupby("seed").cumcount() + 1
    out["within_seed_count"] = out.groupby("seed")["seed"].transform("count")
    
    out["is_seed_best"] = (out["within_seed_rank"] == 1).astype(float)
    out["is_seed_worst"] = (
        out["within_seed_rank"] == out["within_seed_count"]
    ).astype(float)
    out["is_seed_middle"] = (
        (out["within_seed_rank"] > 1) 
        & (out["within_seed_rank"] < out["within_seed_count"])
    ).astype(float)
    
    out["within_seed_kenpom_gap"] = out.groupby("seed")[
        "kenpom_net"
    ].transform("max") - out["kenpom_net"]
    
    return out


def add_region_strength_features(df: pd.DataFrame) -> pd.DataFrame:
    """
    Add region strength features to detect "region of death" discount.
    
    Hypothesis: Teams in unusually strong regions receive an investment discount
    relative to similar-strength teams in easier regions.
    """
    out = df.copy()
    
    if "region" not in out.columns or "kenpom_net" not in out.columns:
        return out
    
    out["kenpom_net"] = pd.to_numeric(out["kenpom_net"], errors="coerce")
    
    region_stats = out.groupby("region")["kenpom_net"].agg([
        ("region_mean_kenpom", "mean"),
        ("region_median_kenpom", "median"),
        ("region_max_kenpom", "max"),
        ("region_std_kenpom", "std"),
    ]).reset_index()
    
    out = out.merge(region_stats, on="region", how="left")
    
    for col in ["region_mean_kenpom", "region_median_kenpom", 
                "region_max_kenpom", "region_std_kenpom"]:
        out[col] = pd.to_numeric(out[col], errors="coerce").fillna(0.0)
    
    out["region_top4_mean_kenpom"] = out.groupby("region").apply(
        lambda g: g.nlargest(4, "kenpom_net")["kenpom_net"].mean()
    ).reindex(out["region"]).values
    
    out["region_top8_mean_kenpom"] = out.groupby("region").apply(
        lambda g: g.nlargest(8, "kenpom_net")["kenpom_net"].mean()
    ).reindex(out["region"]).values
    
    out["kenpom_vs_region_mean"] = out["kenpom_net"] - out["region_mean_kenpom"]
    out["kenpom_vs_region_top4"] = out["kenpom_net"] - out["region_top4_mean_kenpom"]
    
    return out


def add_path_difficulty_features(df: pd.DataFrame) -> pd.DataFrame:
    """
    Add path difficulty features based on seed and region strength.
    
    Combines seed-based expected path with region strength to estimate
    difficulty of reaching later rounds.
    """
    out = df.copy()
    
    if "seed" not in out.columns:
        return out
    
    out["seed"] = pd.to_numeric(out["seed"], errors="coerce")
    
    seed_to_r2_opponent = {
        1: 8, 2: 7, 3: 6, 4: 5, 5: 4, 6: 3, 7: 2, 8: 1,
        9: 9, 10: 10, 11: 11, 12: 12, 13: 13, 14: 14, 15: 15, 16: 16
    }
    
    out["expected_r2_opponent_seed"] = out["seed"].apply(
        lambda s: float(seed_to_r2_opponent.get(int(s), 0)) if pd.notna(s) else 0.0
    )
    
    seed_to_s16_opponent = {
        1: 4, 2: 3, 3: 2, 4: 1, 5: 1, 6: 2, 7: 3, 8: 4,
        9: 1, 10: 1, 11: 1, 12: 1, 13: 1, 14: 1, 15: 1, 16: 1
    }
    
    out["expected_s16_opponent_seed"] = out["seed"].apply(
        lambda s: float(seed_to_s16_opponent.get(int(s), 0)) if pd.notna(s) else 0.0
    )
    
    out["path_difficulty_score"] = (
        out["expected_r2_opponent_seed"] + out["expected_s16_opponent_seed"]
    )
    
    return out


def add_championship_equity_features(df: pd.DataFrame) -> pd.DataFrame:
    """
    Add championship equity features to detect top-heavy payout bias.
    
    Hypothesis: Market overbids teams with high title equity relative to
    expected points (which are more evenly distributed across rounds).
    """
    out = df.copy()
    
    if "seed" not in out.columns:
        return out
    
    out["seed"] = pd.to_numeric(out["seed"], errors="coerce")
    
    seed_to_title_prob = {
        1: 0.20, 2: 0.12, 3: 0.08, 4: 0.05, 5: 0.03, 6: 0.02,
        7: 0.01, 8: 0.01, 9: 0.005, 10: 0.005, 11: 0.003, 12: 0.002,
        13: 0.001, 14: 0.0005, 15: 0.0002, 16: 0.0001
    }
    
    out["approx_title_probability"] = out["seed"].apply(
        lambda s: seed_to_title_prob.get(int(s), 0.0) if pd.notna(s) else 0.0
    )
    
    seed_to_f4_prob = {
        1: 0.45, 2: 0.30, 3: 0.20, 4: 0.15, 5: 0.10, 6: 0.08,
        7: 0.05, 8: 0.04, 9: 0.02, 10: 0.02, 11: 0.015, 12: 0.01,
        13: 0.005, 14: 0.003, 15: 0.001, 16: 0.0005
    }
    
    out["approx_f4_probability"] = out["seed"].apply(
        lambda s: seed_to_f4_prob.get(int(s), 0.0) if pd.notna(s) else 0.0
    )
    
    out["is_title_contender"] = (out["seed"] <= 4).astype(float)
    out["is_elite_seed"] = (out["seed"] <= 2).astype(float)
    
    return out


def add_brand_features(df: pd.DataFrame) -> pd.DataFrame:
    """
    Add brand/familiarity features to detect brand tax.
    
    Hypothesis: Household-name programs are persistently overbid after
    controlling for strength.
    """
    out = df.copy()
    
    if "school_name" not in out.columns:
        return out
    
    blue_bloods = {
        "Duke", "North Carolina", "Kansas", "Kentucky", "UCLA",
        "Indiana", "Louisville", "Michigan State", "Connecticut",
        "Villanova", "Arizona", "Syracuse", "Florida"
    }
    
    out["is_blue_blood"] = out["school_name"].apply(
        lambda s: 1.0 if str(s).strip() in blue_bloods else 0.0
    )
    
    power_conferences = {
        "ACC", "Big Ten", "Big 12", "Big East", "SEC", "Pac-12", "Pac-10"
    }
    
    if "has_last_year" in out.columns:
        out["has_last_year"] = pd.to_numeric(
            out["has_last_year"], errors="coerce"
        ).fillna(0.0)
        out["is_repeat_participant"] = (out["has_last_year"] > 0).astype(float)
    else:
        out["is_repeat_participant"] = 0.0
    
    return out


def add_upset_seed_features(df: pd.DataFrame) -> pd.DataFrame:
    """
    Add upset seed features to detect "upset chic" inflation.
    
    Hypothesis: Trendy upset seed bands (10-12) are overbid conditional on EV.
    """
    out = df.copy()
    
    if "seed" not in out.columns:
        return out
    
    out["seed"] = pd.to_numeric(out["seed"], errors="coerce")
    
    out["is_upset_seed"] = (
        (out["seed"] >= 10) & (out["seed"] <= 12)
    ).astype(float)
    
    out["is_double_digit_seed"] = (out["seed"] >= 10).astype(float)
    
    out["is_cinderella_seed"] = (
        (out["seed"] >= 11) & (out["seed"] <= 15)
    ).astype(float)
    
    return out


def add_kenpom_interaction_features(df: pd.DataFrame) -> pd.DataFrame:
    """
    Add KenPom interaction features for non-linear relationships.
    """
    out = df.copy()
    
    if "kenpom_net" not in out.columns:
        return out
    
    out["kenpom_net"] = pd.to_numeric(out["kenpom_net"], errors="coerce").fillna(0.0)
    
    if "seed" in out.columns:
        out["seed"] = pd.to_numeric(out["seed"], errors="coerce").fillna(0.0)
        out["kenpom_x_seed"] = out["kenpom_net"] * out["seed"]
        out["kenpom_x_seed_sq"] = out["kenpom_net"] * (out["seed"] ** 2)
    
    out["kenpom_net_sq"] = out["kenpom_net"] ** 2
    out["kenpom_net_cube"] = out["kenpom_net"] ** 3
    
    if "kenpom_o" in out.columns and "kenpom_d" in out.columns:
        out["kenpom_o"] = pd.to_numeric(out["kenpom_o"], errors="coerce").fillna(0.0)
        out["kenpom_d"] = pd.to_numeric(out["kenpom_d"], errors="coerce").fillna(0.0)
        out["kenpom_o_x_d"] = out["kenpom_o"] * out["kenpom_d"]
        out["kenpom_balance"] = out["kenpom_o"] - out["kenpom_d"]
    
    return out


def add_all_enhanced_features(df: pd.DataFrame) -> pd.DataFrame:
    """
    Add all enhanced features to the dataset.
    
    This is the main entry point for feature engineering.
    """
    out = df.copy()
    
    out = add_within_seed_features(out)
    out = add_region_strength_features(out)
    out = add_path_difficulty_features(out)
    out = add_championship_equity_features(out)
    out = add_brand_features(out)
    out = add_upset_seed_features(out)
    out = add_kenpom_interaction_features(out)
    
    return out


def get_enhanced_feature_columns() -> List[str]:
    """
    Return list of all enhanced feature column names.
    """
    return [
        "within_seed_rank",
        "within_seed_count",
        "is_seed_best",
        "is_seed_worst",
        "is_seed_middle",
        "within_seed_kenpom_gap",
        "region_mean_kenpom",
        "region_median_kenpom",
        "region_max_kenpom",
        "region_std_kenpom",
        "region_top4_mean_kenpom",
        "region_top8_mean_kenpom",
        "kenpom_vs_region_mean",
        "kenpom_vs_region_top4",
        "expected_r2_opponent_seed",
        "expected_s16_opponent_seed",
        "path_difficulty_score",
        "approx_title_probability",
        "approx_f4_probability",
        "is_title_contender",
        "is_elite_seed",
        "is_blue_blood",
        "is_repeat_participant",
        "is_upset_seed",
        "is_double_digit_seed",
        "is_cinderella_seed",
        "kenpom_x_seed",
        "kenpom_x_seed_sq",
        "kenpom_net_sq",
        "kenpom_net_cube",
        "kenpom_o_x_d",
        "kenpom_balance",
    ]

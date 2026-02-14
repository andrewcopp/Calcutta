BEGIN;

DO $$
DECLARE
	row_count bigint;
BEGIN
	SELECT COUNT(*) INTO row_count FROM derived.predicted_game_outcomes WHERE deleted_at IS NULL;
	IF row_count > 0 THEN
		RAISE EXCEPTION 'refusing to rewire core IDs: derived.predicted_game_outcomes has % rows', row_count;
	END IF;

	SELECT COUNT(*) INTO row_count FROM derived.simulated_tournaments WHERE deleted_at IS NULL;
	IF row_count > 0 THEN
		RAISE EXCEPTION 'refusing to rewire core IDs: derived.simulated_tournaments has % rows', row_count;
	END IF;

	SELECT COUNT(*) INTO row_count FROM derived.simulated_teams WHERE deleted_at IS NULL;
	IF row_count > 0 THEN
		RAISE EXCEPTION 'refusing to rewire core IDs: derived.simulated_teams has % rows', row_count;
	END IF;

	SELECT COUNT(*) INTO row_count FROM derived.predicted_market_share WHERE deleted_at IS NULL;
	IF row_count > 0 THEN
		RAISE EXCEPTION 'refusing to rewire core IDs: derived.predicted_market_share has % rows', row_count;
	END IF;

	SELECT COUNT(*) INTO row_count FROM derived.recommended_entry_bids WHERE deleted_at IS NULL;
	IF row_count > 0 THEN
		RAISE EXCEPTION 'refusing to rewire core IDs: derived.recommended_entry_bids has % rows', row_count;
	END IF;

	SELECT COUNT(*) INTO row_count FROM derived.detailed_investment_report WHERE deleted_at IS NULL;
	IF row_count > 0 THEN
		RAISE EXCEPTION 'refusing to rewire core IDs: derived.detailed_investment_report has % rows', row_count;
	END IF;

	SELECT COUNT(*) INTO row_count FROM derived.strategy_generation_runs WHERE deleted_at IS NULL;
	IF row_count > 0 THEN
		RAISE EXCEPTION 'refusing to rewire core IDs: derived.strategy_generation_runs has % rows', row_count;
	END IF;

	SELECT COUNT(*) INTO row_count FROM derived.optimization_runs WHERE deleted_at IS NULL;
	IF row_count > 0 THEN
		RAISE EXCEPTION 'refusing to rewire core IDs: derived.optimization_runs has % rows', row_count;
	END IF;

	SELECT COUNT(*) INTO row_count FROM derived.calcuttas WHERE deleted_at IS NULL;
	IF row_count > 0 THEN
		RAISE EXCEPTION 'refusing to rewire core IDs: derived.calcuttas has % rows', row_count;
	END IF;

	SELECT COUNT(*) INTO row_count FROM derived.entry_bids WHERE deleted_at IS NULL;
	IF row_count > 0 THEN
		RAISE EXCEPTION 'refusing to rewire core IDs: derived.entry_bids has % rows', row_count;
	END IF;

	SELECT COUNT(*) INTO row_count FROM derived.payouts WHERE deleted_at IS NULL;
	IF row_count > 0 THEN
		RAISE EXCEPTION 'refusing to rewire core IDs: derived.payouts has % rows', row_count;
	END IF;

	SELECT COUNT(*) INTO row_count FROM derived.teams WHERE deleted_at IS NULL;
	IF row_count > 0 THEN
		RAISE EXCEPTION 'refusing to rewire core IDs: derived.teams has % rows', row_count;
	END IF;

	SELECT COUNT(*) INTO row_count FROM derived.tournaments WHERE deleted_at IS NULL;
	IF row_count > 0 THEN
		RAISE EXCEPTION 'refusing to rewire core IDs: derived.tournaments has % rows', row_count;
	END IF;
END $$;

ALTER TABLE IF EXISTS derived.predicted_game_outcomes
    DROP CONSTRAINT IF EXISTS silver_predicted_game_outcomes_tournament_id_fkey,
    DROP CONSTRAINT IF EXISTS silver_predicted_game_outcomes_team1_id_fkey,
    DROP CONSTRAINT IF EXISTS silver_predicted_game_outcomes_team2_id_fkey;

ALTER TABLE IF EXISTS derived.simulated_teams
    DROP CONSTRAINT IF EXISTS silver_simulated_tournaments_tournament_id_fkey,
    DROP CONSTRAINT IF EXISTS silver_simulated_tournaments_team_id_fkey;

ALTER TABLE IF EXISTS derived.predicted_market_share
    DROP CONSTRAINT IF EXISTS silver_predicted_market_share_tournament_id_fkey,
    DROP CONSTRAINT IF EXISTS silver_predicted_market_share_team_id_fkey,
    DROP CONSTRAINT IF EXISTS silver_predicted_market_share_calcutta_id_fkey;

ALTER TABLE IF EXISTS derived.optimization_runs
	DROP CONSTRAINT IF EXISTS gold_optimization_runs_calcutta_id_fkey;

ALTER TABLE IF EXISTS derived.recommended_entry_bids
    DROP CONSTRAINT IF EXISTS gold_recommended_entry_bids_team_id_fkey;

ALTER TABLE IF EXISTS derived.detailed_investment_report
    DROP CONSTRAINT IF EXISTS gold_detailed_investment_report_team_id_fkey;

ALTER TABLE IF EXISTS derived.entry_bids
    DROP CONSTRAINT IF EXISTS bronze_entry_bids_calcutta_id_fkey,
    DROP CONSTRAINT IF EXISTS bronze_entry_bids_team_id_fkey;

ALTER TABLE IF EXISTS derived.payouts
    DROP CONSTRAINT IF EXISTS bronze_payouts_calcutta_id_fkey;

ALTER TABLE IF EXISTS derived.calcuttas
    DROP CONSTRAINT IF EXISTS bronze_calcuttas_tournament_id_fkey;

ALTER TABLE IF EXISTS derived.teams
    DROP CONSTRAINT IF EXISTS bronze_teams_tournament_id_fkey;

-- Retarget artifact FKs to core.
ALTER TABLE IF EXISTS derived.predicted_game_outcomes
    ADD CONSTRAINT predicted_game_outcomes_tournament_id_fkey
        FOREIGN KEY (tournament_id) REFERENCES core.tournaments(id) ON DELETE CASCADE,
    ADD CONSTRAINT predicted_game_outcomes_team1_id_fkey
        FOREIGN KEY (team1_id) REFERENCES core.teams(id) ON DELETE CASCADE,
    ADD CONSTRAINT predicted_game_outcomes_team2_id_fkey
        FOREIGN KEY (team2_id) REFERENCES core.teams(id) ON DELETE CASCADE;

ALTER TABLE IF EXISTS derived.simulated_teams
    ADD CONSTRAINT simulated_teams_tournament_id_fkey
        FOREIGN KEY (tournament_id) REFERENCES core.tournaments(id) ON DELETE CASCADE,
    ADD CONSTRAINT simulated_teams_team_id_fkey
        FOREIGN KEY (team_id) REFERENCES core.teams(id) ON DELETE CASCADE;

ALTER TABLE IF EXISTS derived.predicted_market_share
    ADD CONSTRAINT predicted_market_share_tournament_id_fkey
        FOREIGN KEY (tournament_id) REFERENCES core.tournaments(id) ON DELETE CASCADE,
    ADD CONSTRAINT predicted_market_share_team_id_fkey
        FOREIGN KEY (team_id) REFERENCES core.teams(id) ON DELETE CASCADE,
    ADD CONSTRAINT predicted_market_share_calcutta_id_fkey
        FOREIGN KEY (calcutta_id) REFERENCES core.calcuttas(id) ON DELETE CASCADE;

ALTER TABLE IF EXISTS derived.optimization_runs
	ADD CONSTRAINT optimization_runs_calcutta_id_fkey
		FOREIGN KEY (calcutta_id) REFERENCES core.calcuttas(id) ON DELETE CASCADE;

ALTER TABLE IF EXISTS derived.entry_bids
	ADD CONSTRAINT entry_bids_calcutta_id_fkey
		FOREIGN KEY (calcutta_id) REFERENCES core.calcuttas(id) ON DELETE CASCADE,
	ADD CONSTRAINT entry_bids_team_id_fkey
		FOREIGN KEY (team_id) REFERENCES core.teams(id) ON DELETE CASCADE;

ALTER TABLE IF EXISTS derived.payouts
	ADD CONSTRAINT payouts_calcutta_id_fkey
		FOREIGN KEY (calcutta_id) REFERENCES core.calcuttas(id) ON DELETE CASCADE;

ALTER TABLE IF EXISTS derived.calcuttas
	ADD CONSTRAINT calcuttas_tournament_id_fkey
		FOREIGN KEY (tournament_id) REFERENCES core.tournaments(id) ON DELETE CASCADE;

ALTER TABLE IF EXISTS derived.recommended_entry_bids
    ADD CONSTRAINT recommended_entry_bids_team_id_fkey
        FOREIGN KEY (team_id) REFERENCES core.teams(id) ON DELETE CASCADE;

ALTER TABLE IF EXISTS derived.detailed_investment_report
    ADD CONSTRAINT detailed_investment_report_team_id_fkey
        FOREIGN KEY (team_id) REFERENCES core.teams(id) ON DELETE CASCADE;

COMMIT;

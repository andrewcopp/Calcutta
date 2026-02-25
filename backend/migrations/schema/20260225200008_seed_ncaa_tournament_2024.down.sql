SET search_path = '';

-- Delete kenpom stats for this tournament's teams
DELETE FROM core.team_kenpom_stats
WHERE team_id IN (
  SELECT tm.id FROM core.teams tm
  JOIN core.tournaments t ON t.id = tm.tournament_id
  WHERE t.import_key = 'ncaa-tournament-2024' AND t.deleted_at IS NULL
);

-- Delete teams for this tournament
DELETE FROM core.teams
WHERE tournament_id IN (
  SELECT id FROM core.tournaments WHERE import_key = 'ncaa-tournament-2024' AND deleted_at IS NULL
);

-- Delete tournament
DELETE FROM core.tournaments WHERE import_key = 'ncaa-tournament-2024';

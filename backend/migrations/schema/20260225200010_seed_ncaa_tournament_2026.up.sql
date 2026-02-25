SET search_path = '';

INSERT INTO core.seasons (year) VALUES (2026)
ON CONFLICT (year) DO NOTHING;

INSERT INTO core.tournaments (competition_id, season_id, import_key, rounds, starting_at,
  final_four_top_left, final_four_bottom_left, final_four_top_right, final_four_bottom_right)
SELECT c.id, s.id, 'ncaa-tournament-2026-9c063b', 7, '2026-03-19T11:15:00-05:00',
  NULLIF('Midwest', ''), NULLIF('South', ''), NULLIF('East', ''), NULLIF('West', '')
FROM core.competitions c, core.seasons s
WHERE c.name = 'NCAA Tournament' AND s.year = 2026
ON CONFLICT (import_key) WHERE deleted_at IS NULL DO NOTHING;

INSERT INTO core.teams (tournament_id, school_id, seed, region, byes, wins, is_eliminated)
SELECT t.id, s.id, v.seed, v.region, v.byes, v.wins, v.is_eliminated
FROM (VALUES
  ('alabama', 4, 'East', 1, 0, false),
  ('appalachian-state', 16, 'South', 0, 0, true),
  ('arizona', 1, 'West', 1, 0, false),
  ('arkansas', 5, 'West', 1, 0, false),
  ('auburn', 8, 'South', 1, 0, false),
  ('austin-peay', 14, 'South', 1, 0, false),
  ('belmont', 12, 'West', 1, 0, false),
  ('bethune-cookman', 16, 'South', 0, 0, true),
  ('brigham-young', 6, 'South', 1, 0, false),
  ('california-baptist', 14, 'Midwest', 1, 0, false),
  ('clemson', 7, 'East', 1, 0, false),
  ('connecticut', 1, 'South', 1, 0, false),
  ('duke', 1, 'East', 1, 0, false),
  ('east-tennessee-state', 15, 'Midwest', 1, 0, false),
  ('florida', 3, 'South', 1, 0, false),
  ('georgia', 10, 'West', 1, 0, false),
  ('gonzaga', 3, 'Midwest', 1, 0, false),
  ('hawaii', 13, 'Midwest', 1, 0, false),
  ('high-point', 13, 'East', 1, 0, false),
  ('houston', 2, 'South', 1, 0, false),
  ('howard', 16, 'Midwest', 0, 0, true),
  ('illinois', 2, 'East', 1, 0, false),
  ('indiana', 9, 'South', 1, 0, false),
  ('iowa', 8, 'East', 1, 0, false),
  ('iowa-state', 2, 'Midwest', 1, 0, false),
  ('kansas', 3, 'East', 1, 0, false),
  ('kentucky', 6, 'West', 1, 0, false),
  ('liberty', 12, 'Midwest', 1, 0, false),
  ('long-island-university', 16, 'West', 1, 0, false),
  ('louisville', 5, 'Midwest', 1, 0, false),
  ('merrimack', 15, 'East', 1, 0, false),
  ('miami-fl', 9, 'West', 1, 0, false),
  ('miami-oh', 11, 'South', 1, 0, false),
  ('michigan', 1, 'Midwest', 1, 0, false),
  ('michigan-state', 4, 'West', 1, 0, false),
  ('navy', 15, 'West', 1, 0, false),
  ('nc-state', 8, 'Midwest', 1, 0, false),
  ('nebraska', 4, 'South', 1, 0, false),
  ('njit', 16, 'Midwest', 0, 0, true),
  ('north-carolina', 6, 'East', 1, 0, false),
  ('north-dakota-state', 14, 'West', 1, 0, false),
  ('portland-state', 14, 'East', 1, 0, false),
  ('purdue', 2, 'West', 1, 0, false),
  ('saint-louis', 7, 'Midwest', 1, 0, false),
  ('saint-mary-s-ca', 9, 'Midwest', 1, 0, false),
  ('san-diego-state', 11, 'Midwest', 0, 0, true),
  ('santa-clara', 11, 'East', 0, 0, true),
  ('south-florida', 12, 'South', 1, 0, false),
  ('southern-california', 11, 'West', 1, 0, false),
  ('southern-methodist', 10, 'Midwest', 1, 0, false),
  ('st-john-s-ny', 5, 'East', 1, 0, false),
  ('stephen-f-austin', 12, 'East', 1, 0, false),
  ('tcu', 11, 'East', 0, 0, true),
  ('tennessee', 6, 'Midwest', 1, 0, false),
  ('tennessee-martin', 16, 'East', 1, 0, false),
  ('texas', 9, 'East', 1, 0, false),
  ('texas-a-m', 10, 'South', 1, 0, false),
  ('texas-tech', 3, 'West', 1, 0, false),
  ('ucf', 10, 'East', 1, 0, false),
  ('ucla', 11, 'Midwest', 0, 0, true),
  ('unc-wilmington', 13, 'West', 1, 0, false),
  ('utah-state', 8, 'West', 1, 0, false),
  ('vanderbilt', 4, 'Midwest', 1, 0, false),
  ('villanova', 7, 'West', 1, 0, false),
  ('virginia', 5, 'South', 1, 0, false),
  ('wisconsin', 7, 'South', 1, 0, false),
  ('wright-state', 15, 'South', 1, 0, false),
  ('yale', 13, 'South', 1, 0, false)
) AS v(school_slug, seed, region, byes, wins, is_eliminated)
JOIN core.tournaments t ON t.import_key = 'ncaa-tournament-2026-9c063b' AND t.deleted_at IS NULL
JOIN core.schools s ON s.slug = v.school_slug AND s.deleted_at IS NULL
ON CONFLICT (tournament_id, school_id) WHERE deleted_at IS NULL DO NOTHING;

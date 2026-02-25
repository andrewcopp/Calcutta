SET search_path = '';

INSERT INTO core.seasons (year) VALUES (2016)
ON CONFLICT (year) DO NOTHING;

INSERT INTO core.tournaments (competition_id, season_id, import_key, rounds, starting_at,
  final_four_top_left, final_four_bottom_left, final_four_top_right, final_four_bottom_right)
SELECT c.id, s.id, 'ncaa-tournament-2016', 7, '2016-03-17T11:15:00-05:00',
  NULLIF('South', ''), NULLIF('West', ''), NULLIF('East', ''), NULLIF('Midwest', '')
FROM core.competitions c, core.seasons s
WHERE c.name = 'NCAA Tournament' AND s.year = 2016
ON CONFLICT (import_key) WHERE deleted_at IS NULL DO NOTHING;

INSERT INTO core.teams (tournament_id, school_id, seed, region, byes, wins, is_eliminated)
SELECT t.id, s.id, v.seed, v.region, v.byes, v.wins, v.is_eliminated
FROM (VALUES
  ('arizona', 6, 'South', 1, 0, true),
  ('austin-peay', 16, 'South', 1, 0, true),
  ('baylor', 5, 'West', 1, 0, true),
  ('buffalo', 14, 'South', 1, 0, true),
  ('butler', 9, 'Midwest', 1, 1, true),
  ('cal-state-bakersfield', 15, 'West', 1, 0, true),
  ('california', 4, 'South', 1, 0, true),
  ('chattanooga', 12, 'East', 1, 0, true),
  ('cincinnati', 9, 'West', 1, 0, true),
  ('colorado', 8, 'South', 1, 0, true),
  ('connecticut', 9, 'South', 1, 1, true),
  ('dayton', 7, 'Midwest', 1, 0, true),
  ('duke', 4, 'West', 1, 2, true),
  ('fdu', 16, 'East', 0, 0, true),
  ('florida-gulf-coast', 16, 'East', 0, 1, true),
  ('fresno-state', 14, 'Midwest', 1, 0, true),
  ('gonzaga', 11, 'Midwest', 1, 2, true),
  ('green-bay', 14, 'West', 1, 0, true),
  ('hampton', 16, 'Midwest', 1, 0, true),
  ('hawaii', 13, 'South', 1, 1, true),
  ('holy-cross', 16, 'West', 0, 1, true),
  ('indiana', 5, 'East', 1, 2, true),
  ('iona', 13, 'Midwest', 1, 0, true),
  ('iowa', 7, 'South', 1, 1, true),
  ('iowa-state', 4, 'Midwest', 1, 2, true),
  ('kansas', 1, 'South', 1, 3, true),
  ('kentucky', 4, 'East', 1, 1, true),
  ('little-rock', 12, 'Midwest', 1, 1, true),
  ('maryland', 5, 'South', 1, 2, true),
  ('miami-fl', 3, 'South', 1, 2, true),
  ('michigan', 11, 'East', 0, 1, true),
  ('michigan-state', 2, 'Midwest', 1, 0, true),
  ('middle-tennessee', 15, 'Midwest', 1, 1, true),
  ('north-carolina', 1, 'East', 1, 5, true),
  ('northern-iowa', 11, 'West', 1, 1, true),
  ('notre-dame', 6, 'East', 1, 3, true),
  ('oklahoma', 2, 'West', 1, 4, true),
  ('oregon', 1, 'West', 1, 3, true),
  ('oregon-state', 7, 'West', 1, 0, true),
  ('pittsburgh', 10, 'East', 1, 0, true),
  ('providence', 9, 'East', 1, 1, true),
  ('purdue', 5, 'Midwest', 1, 0, true),
  ('saint-joseph-s', 8, 'West', 1, 1, true),
  ('seton-hall', 6, 'Midwest', 1, 0, true),
  ('south-dakota-state', 12, 'South', 1, 0, true),
  ('southern', 16, 'West', 0, 0, true),
  ('southern-california', 8, 'East', 1, 0, true),
  ('stephen-f-austin', 14, 'East', 1, 1, true),
  ('stony-brook', 13, 'East', 1, 0, true),
  ('syracuse', 10, 'Midwest', 1, 4, true),
  ('temple', 10, 'South', 1, 0, true),
  ('texas', 6, 'West', 1, 0, true),
  ('texas-a-m', 3, 'West', 1, 2, true),
  ('texas-tech', 8, 'Midwest', 1, 0, true),
  ('tulsa', 11, 'East', 0, 0, true),
  ('unc-asheville', 15, 'South', 1, 0, true),
  ('unc-wilmington', 13, 'West', 1, 0, true),
  ('utah', 3, 'Midwest', 1, 1, true),
  ('vanderbilt', 11, 'South', 0, 0, true),
  ('villanova', 2, 'South', 1, 6, false),
  ('virginia', 1, 'Midwest', 1, 3, true),
  ('virginia-commonwealth', 10, 'West', 1, 1, true),
  ('weber-state', 15, 'East', 1, 0, true),
  ('west-virginia', 3, 'East', 1, 0, true),
  ('wichita-state', 11, 'South', 0, 2, true),
  ('wisconsin', 7, 'East', 1, 2, true),
  ('xavier', 2, 'East', 1, 1, true),
  ('yale', 12, 'West', 1, 1, true)
) AS v(school_slug, seed, region, byes, wins, is_eliminated)
JOIN core.tournaments t ON t.import_key = 'ncaa-tournament-2016' AND t.deleted_at IS NULL
JOIN core.schools s ON s.slug = v.school_slug AND s.deleted_at IS NULL
ON CONFLICT (tournament_id, school_id) WHERE deleted_at IS NULL DO NOTHING;

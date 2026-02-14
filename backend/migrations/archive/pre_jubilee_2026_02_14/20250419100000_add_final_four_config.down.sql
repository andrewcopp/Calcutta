-- Remove Final Four configuration columns from tournaments table
ALTER TABLE tournaments 
DROP COLUMN final_four_top_left,
DROP COLUMN final_four_bottom_left,
DROP COLUMN final_four_top_right,
DROP COLUMN final_four_bottom_right;

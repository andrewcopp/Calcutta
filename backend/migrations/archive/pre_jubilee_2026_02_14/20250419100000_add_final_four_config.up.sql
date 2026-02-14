-- Add Final Four configuration columns to tournaments table
ALTER TABLE tournaments 
ADD COLUMN final_four_top_left VARCHAR(50),
ADD COLUMN final_four_bottom_left VARCHAR(50),
ADD COLUMN final_four_top_right VARCHAR(50),
ADD COLUMN final_four_bottom_right VARCHAR(50);

-- Set default values for existing tournaments (standard NCAA bracket layout)
UPDATE tournaments SET
    final_four_top_left = 'East',
    final_four_bottom_left = 'West',
    final_four_top_right = 'South',
    final_four_bottom_right = 'Midwest'
WHERE final_four_top_left IS NULL;

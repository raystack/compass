-- Remove the stars (favorites) system
DROP POLICY IF EXISTS stars_ns ON stars;
DROP TABLE IF EXISTS stars CASCADE;

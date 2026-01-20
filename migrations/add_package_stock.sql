-- Add stock fields to packages table
ALTER TABLE packages ADD COLUMN IF NOT EXISTS stock INTEGER DEFAULT -1;
ALTER TABLE packages ADD COLUMN IF NOT EXISTS sold_count INTEGER DEFAULT 0;

COMMENT ON COLUMN packages.stock IS 'Available stock, -1 means unlimited';
COMMENT ON COLUMN packages.sold_count IS 'Number of packages sold';

-- Update existing packages to unlimited stock
UPDATE packages SET stock = -1, sold_count = 0 WHERE stock IS NULL;

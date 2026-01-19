-- Add LinuxDo OAuth fields to system_settings table
ALTER TABLE system_settings ADD COLUMN IF NOT EXISTS linuxdo_client_id VARCHAR(255);
ALTER TABLE system_settings ADD COLUMN IF NOT EXISTS linuxdo_client_secret VARCHAR(255);
ALTER TABLE system_settings ADD COLUMN IF NOT EXISTS linuxdo_enabled BOOLEAN DEFAULT false;

-- Set default LinuxDo OAuth configuration
UPDATE system_settings
SET
    linuxdo_client_id = 'kndqpnv5TsY9ouaiaakf09AVZmd7M9pJ',
    linuxdo_client_secret = 'XQAnYlCmDdXHgm5zRjjIzZMvfKtrATXg',
    linuxdo_enabled = true
WHERE id = 1;

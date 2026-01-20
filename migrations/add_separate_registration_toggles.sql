-- Add separate registration toggles for email and LinuxDo
-- This migration replaces the single registration_enabled field with two separate fields

-- Add new columns
ALTER TABLE system_settings
ADD COLUMN IF NOT EXISTS email_registration_enabled BOOLEAN DEFAULT true,
ADD COLUMN IF NOT EXISTS linuxdo_registration_enabled BOOLEAN DEFAULT true;

-- Copy existing registration_enabled value to both new fields (if the column exists)
DO $$
BEGIN
    IF EXISTS (
        SELECT 1 FROM information_schema.columns
        WHERE table_name = 'system_settings'
        AND column_name = 'registration_enabled'
    ) THEN
        UPDATE system_settings
        SET email_registration_enabled = registration_enabled,
            linuxdo_registration_enabled = registration_enabled;

        -- Drop old column
        ALTER TABLE system_settings DROP COLUMN registration_enabled;
    END IF;
END $$;

-- Fix duplicate https://linux.do prefix in avatar URLs
UPDATE users
SET avatar_url = REPLACE(avatar_url, 'https://linux.dohttps://linux.do', 'https://linux.do')
WHERE avatar_url LIKE 'https://linux.dohttps://linux.do%';

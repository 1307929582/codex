#!/bin/bash

# Fix Admin Role Script
# This script updates the user role to super_admin for the specified email

set -e

echo "=== Fix Admin Role ==="
echo ""

# Get database credentials from .env
if [ -f .env ]; then
    export $(cat .env | grep -v '^#' | xargs)
else
    echo "Error: .env file not found"
    exit 1
fi

# Email to update
EMAIL="${1:-cg1307929582@gmail.com}"

echo "Updating role for user: $EMAIL"
echo ""

# Connect to PostgreSQL and update role
docker compose exec -T db psql -U "$DB_USER" -d "$DB_NAME" <<EOF
-- Check current user status
SELECT id, email, role, status FROM users WHERE email = '$EMAIL';

-- Update role to super_admin
UPDATE users SET role = 'super_admin' WHERE email = '$EMAIL';

-- Verify update
SELECT id, email, role, status FROM users WHERE email = '$EMAIL';
EOF

echo ""
echo "✅ Role updated successfully!"
echo ""
echo "Please refresh your browser and try accessing the admin panel again."

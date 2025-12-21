#!/bin/sh
set -e

# =============================================================================
# Campus Project Hub API Entrypoint
# =============================================================================
# Environment variables for database operations:
#   RUN_MIGRATE=true     - Run database migrations before starting
#   RUN_MIGRATE_FRESH=true - Drop all tables and re-migrate (WARNING: destroys data)
#   RUN_SEEDER=true      - Run database seeders
# =============================================================================

echo "🚀 Starting Campus Project Hub API..."

# Database migrations (Go uses GORM auto-migrate or manual migration)
if [ "$RUN_SEEDER" = "true" ]; then
  echo "🌱 Running database seeders..."
  ./seeder
fi

echo "✅ Starting server..."
exec "$@"

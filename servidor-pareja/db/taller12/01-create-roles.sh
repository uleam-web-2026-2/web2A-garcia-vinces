#!/bin/sh
set -eu

# Evita interpolar caracteres SQL en los nombres usados como identificadores.
case "$MIGRATION_DB_USER$API_DB_USER" in
  (*[!A-Za-z0-9_]*|'') echo "Los nombres de rol solo pueden usar letras, números y guion bajo" >&2; exit 1 ;;
esac

psql --username "$POSTGRES_USER" --dbname "$POSTGRES_DB" \
  --set=ON_ERROR_STOP=1 \
  --set=migration_user="$MIGRATION_DB_USER" \
  --set=migration_password="$MIGRATION_DB_PASSWORD" \
  --set=api_user="$API_DB_USER" \
  --set=api_password="$API_DB_PASSWORD" \
  --set=database_name="$POSTGRES_DB" \
  --file=/sql/roles.sql

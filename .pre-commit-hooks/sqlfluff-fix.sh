#!/usr/bin/env bash
set -euo pipefail

cd code/dbt

# Apply autofixes safely
sqlfluff fix --dialect=athena --templater=dbt .

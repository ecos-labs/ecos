#!/usr/bin/env bash
set -euo pipefail

# Move into dbt project folder
cd code/dbt

# Use project-level config (.sqlfluff)
sqlfluff lint --dialect=athena --templater=dbt .

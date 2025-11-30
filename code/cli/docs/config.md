# ecos Configuration Guide

## Overview

The `.ecos.yaml` file is the **single source of truth** for your ecos project configuration. It defines how your project transforms and processes cloud cost data, and automatically generates the necessary DBT configuration files.

## Configuration as Single Source of Truth

ecos uses a declarative configuration approach where `.ecos.yaml` serves as the central configuration that:

1. **Defines your project settings** - Project name, version, and data sources
2. **Configures transformation behavior** - DBT settings, materialization strategies
3. **Manages AWS resources** - Region, database, workgroups, and buckets
4. **Generates DBT files** - Automatically creates `dbt_project.yml` and `profiles.yml`

This approach ensures:
- ✅ **Consistency** - All configuration in one place
- ✅ **Version control** - Single file to track changes
- ✅ **Reproducibility** - Easy to recreate environments
- ✅ **Drift detection** - Automatic detection of manual changes

---

## Configuration File Structure

### Basic Structure

```yaml
project_name: my-cost-analysis
model_version: v1.0.0
data_source: aws_cur

transform:
  dbt:
    # DBT configuration

aws:
  # AWS resource configuration
```

---

## Configuration Options

### Project Settings

#### `project_name` (required)
The name of your ecos project. Used to identify and name cloud resources.

```yaml
project_name: my-cost-analysis
```

**Rules:**
- Alphanumeric characters, hyphens, and underscores only
- Used as prefix/suffix for cloud resources (S3 buckets, Athena databases, workgroups)
- Will be converted to appropriate format for each resource type
- Example: `my-cost-analysis` → S3 bucket: `my-cost-analysis-results`, database: `my_cost_analysis`

#### `model_version` (optional)
The version of the ecos data models to use.

```yaml
model_version: v1.0.0  # or "latest"
```

**Options:**
- `latest` - Use the most recent stable version (default)
- `v1.0.0` - Use a specific version tag

#### `data_source`
Identifies which provider plugin ecos should use.

Supported:
  - `aws_cur` (default)
  - `aws_focus`

---

### Transform Configuration

The `transform` section configures how data is transformed using DBT.

#### `transform.dbt.project_dir`
Path to the DBT project directory (relative to project root).

```yaml
transform:
  dbt:
    project_dir: transform/dbt
```

**Default:** `./transform/dbt`

#### `transform.dbt.profile_dir`
Path to the DBT profiles directory.

```yaml
transform:
  dbt:
    profile_dir: transform/dbt
```

**Default:** `./transform/dbt`

#### `transform.dbt.profile`
Name of the DBT profile to use.

```yaml
transform:
  dbt:
    profile: athena
```

**Default:** `ecos-athena`

#### `transform.dbt.target`
The DBT target environment to use.

```yaml
transform:
  dbt:
    target: prod
```

**Options:**
- `prod` - Production environment (default)
- `dev` - Development environment
- Custom targets as defined in your profiles

#### `transform.dbt.aws_profile`
AWS CLI profile name to use for authentication.

```yaml
transform:
  dbt:
    aws_profile: default
```

**Default:** `default`

#### `transform.dbt.vars`
Data source variables passed to DBT models.

```yaml
transform:
  dbt:
    vars:
      cur_database: "awsdatacatalog"
      cur_schema: "cur"
      cur_table: "cur-data"
```

**Common Variables:**
- `cur_database` - AWS Glue catalog name
- `cur_schema` - Schema/database containing CUR data
- `cur_table` - Table name for CUR data

#### `transform.dbt.materialization`
Controls how DBT models are materialized (stored in the database).

```yaml
transform:
  dbt:
    materialization:
      mode: view
      layer_overrides:
        bronze: view
        silver: view
        gold: table
```

**Materialization Modes:**

| Mode | Description | Use Case |
|------|-------------|----------|
| `view` | Virtual table, no data stored | Development, fast iteration |
| `table` | Physical table, data stored | Production, better query performance |
| `incremental` | Append/update only new data | Large datasets, cost optimization |

**Layer Overrides:**
- `bronze` - Raw data layer (default: `view`)
- `silver` - Cleaned/enriched data layer (default: `view`)
- `gold` - Aggregated/business metrics layer (default: `view`)

**Best Practices:**
- **Development:** Use `view` for all layers (fast, no storage costs)
- **Production:** Use `table` for gold layer, `view` for bronze/silver
- **Large datasets:** Use `incremental` for bronze layer

---

### AWS Configuration

The `aws` section defines AWS resources used by ecos.

#### `aws.region`
AWS region where resources are located.

```yaml
aws:
  region: us-east-1
```

**Common Regions:**
- `us-east-1` - US East (N. Virginia)
- `us-west-2` - US West (Oregon)
- `eu-west-1` - Europe (Ireland)

#### `aws.database`
Athena database name for query results.

```yaml
aws:
  database: my_cost_analysis
```

**Note:** Automatically generated from `project_name` during `ecos init`

#### `aws.dbt_workgroup`
Athena workgroup for DBT queries.

```yaml
aws:
  dbt_workgroup: ecos-dbt
```

**Purpose:** Isolates DBT query execution and manages query limits/costs

#### `aws.results_bucket`
S3 bucket for Athena query results and DBT artifacts.

```yaml
aws:
  results_bucket: my-ecos-results-bucket
```

**Structure:**
```
s3://my-ecos-results-bucket/
├── dbt/
│   ├── staging/    # Query results
│   ├── data/       # Materialized tables
│   └── tmp/        # Temporary tables
```

---

## Generated Files

ecos automatically generates DBT configuration files from `.ecos.yaml`:

### `dbt_project.yml`
Defines the DBT project structure, models, and variables.

**Generated from:**
- `project_name`
- `transform.dbt.profile`
- `transform.dbt.vars`
- `transform.dbt.materialization`

### `profiles.yml`
Defines DBT connection profiles for Athena.

**Generated from:**
- `transform.dbt.profile`
- `transform.dbt.target`
- `transform.dbt.aws_profile`
- `aws.region`
- `aws.database`
- `aws.dbt_workgroup`
- `aws.results_bucket`

---

## Configuration Management

### Drift Detection

ecos can detect when DBT files have been manually modified and no longer match `.ecos.yaml`.

#### Check for drift:
```bash
ecos config diff
```

**Output:**
```
Configuration drift detected

Files out of sync: 2

📄 dbt_project.yml
─────────────────
--- dbt_project.yml (expected)
+++ dbt_project.yml (current)
- materialization: view
+ materialization: table
```

**When drift occurs:**
- Manual edits to `dbt_project.yml` or `profiles.yml`
- Changes to `.ecos.yaml` not yet applied
- Version mismatches

### Regenerating Files

Regenerate DBT files from `.ecos.yaml`:

```bash
ecos config generate
```

**This will:**
1. Read `.ecos.yaml`
2. Generate fresh `dbt_project.yml`
3. Generate fresh `profiles.yml`
4. Overwrite existing files

**Use cases:**
- Fix drift after manual edits
- Apply changes from `.ecos.yaml` updates
- Recover from corrupted DBT files

---

## Workflow Examples

### Initial Setup

```bash
# Initialize project (creates .ecos.yaml)
ecos init --source aws-cur

# Files created:
# - .ecos.yaml (single source of truth)
# - transform/dbt/dbt_project.yml (generated)
# - transform/dbt/profiles.yml (generated)
```

### Updating Configuration

```bash
# 1. Edit .ecos.yaml
vim .ecos.yaml

# 2. Check what will change
ecos config diff

# 3. Apply changes
ecos config generate

# 4. Run transformations
ecos transform run
```

### Changing Materialization Strategy

```yaml
# .ecos.yaml - Switch to production settings
transform:
  dbt:
    materialization:
      mode: table
      layer_overrides:
        bronze: view
        silver: view
        gold: table  # Materialize gold layer for performance
```

```bash
# Apply changes
ecos config generate

# Run with full refresh
ecos transform run --full-refresh
```

### Environment-Specific Configuration

For different environments, use separate `.ecos.yaml` files:

```bash
# Development
.ecos.dev.yaml

# Production
.ecos.prod.yaml

# Use specific config
ecos --config .ecos.prod.yaml transform run
```

---

## Best Practices

### 1. Version Control
✅ **DO:** Commit `.ecos.yaml` to git
❌ **DON'T:** Commit generated DBT files (add to `.gitignore`)

```gitignore
# .gitignore
transform/dbt/dbt_project.yml
transform/dbt/profiles.yml
```

### 2. Configuration Changes
✅ **DO:** Always edit `.ecos.yaml` and regenerate
❌ **DON'T:** Manually edit `dbt_project.yml` or `profiles.yml`

### 3. Drift Detection
✅ **DO:** Run `ecos config diff` before deployments
✅ **DO:** Include drift checks in CI/CD pipelines

### 4. Documentation
✅ **DO:** Add comments to `.ecos.yaml` explaining custom settings
✅ **DO:** Document any non-standard configurations

---
## Related Commands

- `ecos init` - Initialize project and create `.ecos.yaml`
- `ecos config diff` - Detect configuration drift
- `ecos config generate` - Regenerate DBT files from `.ecos.yaml`
- `ecos transform run` - Run DBT transformations

---

## See Also

- [DBT Documentation](https://docs.getdbt.com/)
- [AWS Athena Documentation](https://docs.aws.amazon.com/athena/)

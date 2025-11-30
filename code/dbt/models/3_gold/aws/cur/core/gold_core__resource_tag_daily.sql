{{ config(**get_model_config('incremental')) }}

{%- set rel = load_relation(ref('silver_aws__cur_enhanced')) -%}

{#-- Use flattening macro to get all tag columns --#}
{%- set tag_cols_sql = aws_transforms_cur_all_tag_flattening(rel) -%}
{%- set tag_cols_list = [] -%}
{%- if tag_cols_sql -%}
  {%- for line in tag_cols_sql.split(',') -%}
    {%- if 'resource_tags_' in line -%}
      {%- set col_name = line.split(' as ')[-1].strip() -%}
      {%- set _ = tag_cols_list.append(col_name) -%}
    {%- endif -%}
  {%- endfor -%}
{%- endif -%}

{#-- define the fixed columns in the same order as the SELECT --#}
{%- set base_cols = [
    "date_trunc('day', usage_date)",
    "account_id",
    "payer_account_id",
    "billing_entity",
    "legal_entity",
    "service_code",
    "service_name",
    "service_category",
    "product_family",
    "region_id",
    "resource_id",
    "resource_name"
] -%}

{#-- combine fixed + tag columns for group by --#}
{%- set all_group_cols = base_cols + tag_cols_list -%}

with

source as (

    select *
    from {{ ref('silver_aws__cur_enhanced') }}
    where
        resource_id is not null
        and resource_id != ''
        and is_running_usage

        and {{ get_model_time_filter() }}

)

, aggregated as (

    select

        -- time
        date_trunc('day', usage_date) as usage_date

        -- account
        , account_id
        , payer_account_id
        , billing_entity
        , legal_entity

        -- service
        , service_code
        , service_name
        , service_category
        , product_family
        , region_id

        -- resource
        , resource_id
        , resource_name
        {{ tag_cols_sql }}

        -- metrics
        , sum(effective_cost) as total_effective_cost
        , sum(usage_amount) as total_usage_amount
        , sum(normalized_usage_amount) as total_normalized_usage_amount

        -- billing period
        , billing_period

    from source
    {{ dbt_utils.group_by(all_group_cols | length) }}, billing_period

)

select *
from aggregated

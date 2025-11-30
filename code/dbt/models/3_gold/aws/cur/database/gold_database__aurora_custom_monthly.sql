{{ config(**get_model_config('incremental')) }}

{#
  This model analyzes daily costs and usage for specific Aurora clusters and database instances.
  Configure the cluster IDs and instance names in the variables section below.
#}

{%- set primary_cluster_id = var('primary_cluster_id', 'cluster-xxxxxxxxxxxxxxxxxxxxxxxx') -%}
{%- set secondary_cluster_id_1 = var('secondary_cluster_id_1', 'cluster-yyyyyyyyyyyyyyyyyyyyyyyy') -%}
{%- set secondary_cluster_id_n = var('secondary_cluster_id_n', 'cluster-zzzzzzzzzzzzzzzzzzzzzzzz') -%}
{%- set primary_cluster_db_instance_name_1 = var('primary_cluster_db_instance_name_1', 'team-a-mysql-db-1') -%}
{%- set primary_cluster_db_instance_name_n = var('primary_cluster_db_instance_name_n', 'team-a-mysql-db-2') -%}
{%- set secondary_cluster_db_instance_name_1 = var('secondary_cluster_db_instance_name_1', 'team-a-mysql-db-3') -%}
{%- set secondary_cluster_db_instance_name_n = var('secondary_cluster_db_instance_name_n', 'team-a-mysql-db-4') -%}

with

source_data as (

    select *
    from {{ ref('silver_aws__cur_enhanced') }}
    where
        {{ get_model_time_filter() }}
        and -- Filter by date range (customize as needed)
        usage_date >= current_date - interval '90' day

        -- Filter by specific cluster and instance resource IDs
        and (
            resource_id like '%{{ primary_cluster_id }}%' -- primary region cluster id
            or resource_id like '%{{ secondary_cluster_id_1 }}%' -- secondary region cluster id
            or resource_id like '%{{ secondary_cluster_id_n }}%' -- additional secondary region cluster id
            or resource_id like '%{{ primary_cluster_db_instance_name_1 }}%' -- primary region database instance name
            -- additional primary region database instance name
            or resource_id like '%{{ primary_cluster_db_instance_name_n }}%'
            -- secondary region database instance name
            or resource_id like '%{{ secondary_cluster_db_instance_name_1 }}%'
            -- additional secondary region database instance name
            or resource_id like '%{{ secondary_cluster_db_instance_name_n }}%'
        )

        -- Exclude backup usage
        and usage_type not like '%BackupUsage%'

)

, cluster_analysis_daily as (

    select

        -- time
        date_trunc('day', usage_date) as day_usage_date

        -- account
        , account_id

        -- service details
        , charge_type
        , usage_type
        , operation
        , item_description

        -- resource
        , resource_id

        -- aggregated metrics
        , sum(effective_cost) as total_effective_cost
        , sum(billed_cost) as total_billed_cost
        , sum(list_cost) as total_list_cost
        , sum(usage_amount) as total_usage_amount

    from source_data
    {{ dbt_utils.group_by(7) }}

)

, final as (

    select

        -- time
        day_usage_date

        -- account
        , account_id

        -- service details
        , charge_type
        , usage_type
        , operation
        , item_description

        -- resource
        , resource_id

        -- cost and usage metrics
        , total_effective_cost
        , total_billed_cost
        , total_list_cost
        , total_usage_amount

        -- additional analysis fields
        , case
            when resource_id like '%cluster-%' then 'Cluster'
            when resource_id like '%db:%' then 'Database Instance'
            else 'Other'
        end as resource_type

        , case
            when
                resource_id like '%{{ primary_cluster_id }}%'
                or resource_id like '%{{ primary_cluster_db_instance_name_1 }}%'
                or resource_id like '%{{ primary_cluster_db_instance_name_n }}%'
                then 'Primary'
            else 'Secondary'
        end as cluster_region_type

        -- billing period (must be last for partitioning)
        , date_trunc('month', day_usage_date) as billing_period

    from cluster_analysis_daily

)

select *
from final

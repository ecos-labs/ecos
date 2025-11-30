{{ config(**get_model_config('incremental')) }}

with

source as (

    select *
    from {{ ref('silver_aws__cur_enhanced') }}
    where
        service_code in ('AmazonS3', 'AmazonGlacier', 'AmazonS3GlacierDeepArchive')
        and is_running_usage
        and effective_cost > 0

)

, s3_usage as (

    select

        -- time
        date_trunc('day', usage_date) as usage_date

        -- account
        , payer_account_id
        , account_id

        -- resource and usage
        , resource_name as bucket_name
        , {{ aws_mappings_s3_category(operation_col='operation', usage_type_col='usage_type') }} as usage_category
        , region_id

        -- metrics
        , sum(effective_cost) as total_effective_cost
        , sum(usage_amount) as total_usage_amount

        -- billing period
        , billing_period

    from source
    where {{ get_model_time_filter() }}

    {{ dbt_utils.group_by(6) }}, billing_period

)

select *
from s3_usage

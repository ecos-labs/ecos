{{ config(**get_model_config('incremental')) }}

with

source as (

    select *
    from {{ ref('silver_aws__cur_enhanced') }}
    where
        {{ get_model_time_filter() }}
        and service_code = 'AmazonEC2'
        and charge_type = 'Usage'
        and usage_type like '%EBS:%'

)

, ebs_usage as (

    select

        -- time
        date_trunc('day', usage_date) as usage_date

        -- account
        , account_id
        , payer_account_id

        -- resource and usage
        , resource_name as volume_name
        , {{ aws_mappings_ebs_category(
            service_code_col='service_code',
            usage_type_col='usage_type'
        ) }} as usage_category
        , region_id

        -- metrics
        , sum(effective_cost) as total_effective_cost
        , sum(usage_amount) as total_usage_amount

        -- billing period
        , billing_period

    from source
    {{ dbt_utils.group_by(6) }}, billing_period

)

select *
from ebs_usage

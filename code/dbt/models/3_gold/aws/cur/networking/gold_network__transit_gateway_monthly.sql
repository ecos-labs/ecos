{{ config(**get_model_config('incremental')) }}

with

source as (

    select *
    from {{ ref('silver_aws__cur_enhanced') }}
    where
        {{ get_model_time_filter() }}
        and product_group = 'AWSTransitGateway'
        and is_running_usage

)

, final as (

    select
        -- time
        date_trunc('month', usage_date) as usage_date

        -- account
        , payer_account_id
        , account_id

        -- resource
        , resource_name as resource_identifier
        , resource_id -- TODO: check if this is needed
        , region_id
        , product_attachment_type
        , pricing_unit

        -- charge type classification
        , case
            when pricing_unit = 'hour' then 'Hourly charges'
            when pricing_unit = 'GigaBytes' then 'Data processing charges'
            else 'Other'
        end as charge_category

        -- aggregated metrics
        , round(sum(usage_amount), 2) as total_usage_amount
        , round(sum(effective_cost), 2) as total_effective_cost

        -- billing context
        , billing_period

    from source
    {{ dbt_utils.group_by(8) }}, billing_period
    having sum(effective_cost) > 0

)

select *
from final

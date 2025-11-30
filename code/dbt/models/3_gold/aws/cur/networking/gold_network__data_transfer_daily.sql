{{ config(**get_model_config('incremental')) }}

with

source as (

    select *
    from {{ ref('silver_aws__cur_enhanced') }}
    where
        {{ get_model_time_filter() }}
        and product_family in ('Data Transfer', 'DT-Data Transfer')
        and charge_type in ('Usage', 'DiscountedUsage', 'SavingsPlanCoveredUsage')

)

, data_transfer_consolidated as (

    select
        -- time
        date_trunc('month', usage_date) as usage_date

        -- account
        , payer_account_id
        , account_id

        -- service
        , service_code
        , service_name

        -- resource
        -- TODO: check if this high granularity is needed
        , resource_name
        , usage_type
        , operation

        -- transfer
        , product_from_location
        , product_to_location
        , region_id

        -- category
        , {{ aws_mappings_data_transfer_category(
            operation_col='operation'
            , usage_type_col='usage_type'
            , service_code_col='service_code') }}            as data_transfer_category

        -- metrics
        , sum(usage_amount) as total_usage_amount
        , sum(effective_cost) as total_effective_cost

        -- billing period
        , billing_period

    from source
    {{ dbt_utils.group_by(12) }}, billing_period

)

, final as (

    select

        -- time
        usage_date

        -- account
        , payer_account_id
        , account_id

        -- service
        , service_code
        , service_name

        -- resource
        , resource_name
        , usage_type
        , operation

        -- transfer
        , product_from_location
        , product_to_location
        , region_id
        , data_transfer_category

        -- metrics
        , round(total_usage_amount, 2) as total_usage_amount
        , round(total_effective_cost, 2) as total_effective_cost

        -- billing period
        , billing_period

    from data_transfer_consolidated
    where total_effective_cost > 0

)

select *
from final

{{ config(**get_model_config('incremental')) }}

-- This model identifies unused On-Demand Capacity Reservations in AWS EC2 on a monthly basis

with

filtered_source as (

    select *
    from {{ ref('silver_aws__cur_enhanced') }}
    where
        {{ get_model_time_filter() }}
        and service_code = 'AmazonEC2'
        and resource_id like '%cr-%'
        and item_description like '%Res%'
        and (usage_type like '%Reservation:%' or usage_type like '%UnusedBox%')

)

, aggregated_data as (

    select
        date_trunc('month', usage_date) as usage_date
        , split_part(resource_id, ':', 5) as owner_account
        , account_id as consumer_account
        , resource_id
        , item_description
        , sum(case when usage_type like '%Reservation:%' then usage_amount else 0 end) as total_reservation_amount
        , sum(case when usage_type like '%UnusedBox%' then usage_amount else 0 end) as unused_reservation_amount
        , sum(case when usage_type like '%UnusedBox%' then billed_cost else 0 end) as total_unused_reservation_cost
    from filtered_source
    {{ dbt_utils.group_by(5) }}

)

, final as (

    select
        usage_date
        , owner_account
        , consumer_account
        , item_description
        , resource_id
        , total_reservation_amount
        , unused_reservation_amount
        , total_unused_reservation_cost

        -- billing period (must be last for partitioning)
        , date_trunc('month', usage_date) as billing_period

    from aggregated_data
    where total_reservation_amount > 0

)

select *
from final

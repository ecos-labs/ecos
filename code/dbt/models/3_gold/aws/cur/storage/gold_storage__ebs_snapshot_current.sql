{{ config(**get_model_config('incremental')) }}

with

source as (

    select *
    from {{ ref('gold_storage__ebs_summary_monthly') }}
    where
        volume_type = 'snapshot'
        and billing_period >= date_format(current_date - interval '12' month, '%Y-%m')

)

, snapshot_aggregated as (

    select

        -- account
        payer_account_id
        , account_id

        -- volume
        , volume_name
        , date(min(usage_date)) as first_seen
        , date(max(usage_date)) as last_seen
        , date_diff('day', min(usage_date), max(usage_date)) as age_in_days

        -- previous month cost
        , round(
            sum(
                case
                    when
                        usage_date
                        = date_trunc('month', date_add('month', -1, current_date))
                        then total_effective_cost
                    else 0
                end
            ), 2
        ) as total_effective_cost_previous_month

        -- total cost
        , round(sum(total_effective_cost), 5) as total_effective_cost
        , date_format(current_date, '%Y-%m') as billing_period

    from source
    {{ dbt_utils.group_by(3) }}

    -- filter only current snapshot which exists within last 5 days
    having date(max(usage_date)) >= date_add('day', -5, current_date)

)

select *
from snapshot_aggregated

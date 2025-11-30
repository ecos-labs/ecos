{{ config(**get_model_config('incremental')) }}

with

source as (

    select *
    from {{ ref('silver_aws__cur_enhanced') }}
    where
        {{ get_model_time_filter() }}
        and service_code = 'AWSLambda'
        and is_running_usage = true

)

, usage_categorization as (

    select

        -- time
        date_trunc('month', usage_date) as usage_date

        -- account
        , payer_account_id
        , account_id

        -- service
        , {{ aws_mappings_lambda_category(
            usage_type_col='usage_type'
        ) }} as usage_category
        , resource_name as function_name
        , region_id

        -- processor architecture
        , case substr(usage_type, length(usage_type) - 2)
            when 'ARM' then 'arm64'
            else 'x86_64'
        end as processor_architecture

        -- cost
        , sum(effective_cost) as total_effective_cost
        , sum(usage_amount) as total_usage_amount

        -- cost by charge type
        , sum(case when charge_type = 'SavingsPlanCoveredUsage' then effective_cost else 0 end)
            as total_effective_cost_savings_plan
        , sum(case when charge_type = 'Usage' then effective_cost else 0 end)
            as total_effective_cost_ondemand

        -- graviton savings potential (20% savings for x86 Lambda functions)
        , sum(
            case
                when
                    substr(usage_type, length(usage_type) - 2) != 'ARM'
                    and (operation = 'Invoke' and (usage_type like '%Request%' or usage_type like '%Lambda-GB-Second%'))
                    then effective_cost * 0.2
                else 0
            end
        ) as potential_graviton_savings

        -- billing period
        , billing_period

    from source
    {{ dbt_utils.group_by(7) }}, billing_period

)

select

    -- time
    usage_date

    -- account
    , payer_account_id
    , account_id

    -- service
    , usage_category
    , function_name
    , region_id
    , processor_architecture

    -- metrics
    , round(total_effective_cost, 4) as total_effective_cost
    , round(total_usage_amount, 4) as total_usage_amount
    , round(total_effective_cost_savings_plan, 4) as total_effective_cost_savings_plan
    , round(total_effective_cost_ondemand, 4) as total_effective_cost_ondemand
    , round(potential_graviton_savings, 4) as potential_graviton_savings

    -- billing period
    , billing_period

from usage_categorization

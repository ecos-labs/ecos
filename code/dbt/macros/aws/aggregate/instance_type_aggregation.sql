{% macro instance_type_rates_cte() %}

    select

        -- all
        *

        -- hourly rates
        , round({{ utils_divide_safe('total_effective_cost', 'total_usage_amount') }}, 4) as hourly_rate_blended
        , cast(row(
            round({{ utils_divide_safe('total_effective_cost_ondemand', 'total_usage_amount_ondemand') }}, 4)
            , round({{ utils_divide_safe('total_effective_cost_sp_1y', 'total_usage_amount_sp_1y') }}, 4)
            , round({{ utils_divide_safe('total_effective_cost_sp_3y', 'total_usage_amount_sp_3y') }}, 4)
            , round({{ utils_divide_safe('total_effective_cost_ri_1y', 'total_usage_amount_ri_1y') }}, 4)
            , round({{ utils_divide_safe('total_effective_cost_ri_3y', 'total_usage_amount_ri_3y') }}, 4)
            , round({{ utils_divide_safe('total_effective_cost_spot', 'total_usage_amount_spot') }}, 4)
        ) as row (
            ondemand double
            , savings_plan_1y double
            , savings_plan_3y double
            , reserved_1y double
            , reserved_3y double
            , spot double
        )) as hourly_rate_purchase

        -- discounts
        , round(1 - {{ utils_divide_safe('total_effective_cost', 'total_list_cost', 1.0) }}, 4) as pct_discount_blended
        , cast(row(
            round(1 - {{ utils_divide_safe('total_effective_cost_ondemand', 'total_list_cost_ondemand', 1.0) }}, 4)
            , round(1 - {{ utils_divide_safe('total_effective_cost_sp_1y', 'total_list_cost_sp_1y', 1.0) }}, 4)
            , round(1 - {{ utils_divide_safe('total_effective_cost_sp_3y', 'total_list_cost_sp_3y', 1.0) }}, 4)
            , round(1 - {{ utils_divide_safe('total_effective_cost_ri_1y', 'total_list_cost_ri_1y', 1.0) }}, 4)
            , round(1 - {{ utils_divide_safe('total_effective_cost_ri_3y', 'total_list_cost_ri_3y', 1.0) }}, 4)
            , round(1 - {{ utils_divide_safe('total_effective_cost_spot', 'total_list_cost_spot', 1.0) }}, 4)
        ) as row (
            ondemand double
            , savings_plan_1y double
            , savings_plan_3y double
            , reserved_1y double
            , reserved_3y double
            , spot double
        )) as pct_discount_purchase

    from instance_summary

{% endmacro %}

{% macro instance_type_aggregation(include_account_fields=false) %}

    select

        -- time
        usage_date
        , billing_period

        {%- if include_account_fields %}
        -- account
        , account_id
        , payer_account_id
        {%- endif %}

        -- instance
        , service_code
        , service_name
        , region_id
        , instance_type
        , instance_type_family
        , processor_family

        -- modernization
        , processor_generation
        , latest_intel
        , latest_amd
        , latest_graviton

        -- count
        , count(distinct account_id) as count_accounts
        , count(distinct account_id || resource_id) as count_resources

        -- cost
        , sum(total_billed_cost) as total_billed_cost
        , sum(total_effective_cost) as total_effective_cost
        , sum(total_list_cost) as total_list_cost

        -- usage
        , sum(total_usage_hours) as total_usage_hours
        , sum(total_usage_amount) as total_usage_amount
        , sum(total_normalized_usage_amount) as total_normalized_usage_amount

        -- utilization metrics
        , max(total_hours_month) as total_hours_month
        , sum(total_usage_hours)
        / (max(total_hours_month) * count(distinct account_id || resource_id)) as pct_utilization_avg
        , approx_percentile(pct_utilization, 0.05) as pct_utilization_p5
        , approx_percentile(pct_utilization, 0.5) as pct_utilization_median
        , approx_percentile(pct_utilization, 0.95) as pct_utilization_p95

        -- savings potential (aggregated)
        , sum(savings_mon_to_fri_full_day) as total_savings_mon_to_fri_full_day
        , sum(savings_mon_to_fri_work_hours) as total_savings_mon_to_fri_work_hours
        , sum(savings_target_50_pct) as total_savings_target_50_pct

        -- pct usage
        , sum(total_usage_hours_purchase.ondemand) / sum(total_usage_hours) as pct_usage_ondemand
        , sum(
            total_usage_hours_purchase.reserved_1y
            + total_usage_hours_purchase.reserved_3y
            + total_usage_hours_purchase.savings_plan_1y
            + total_usage_hours_purchase.savings_plan_3y
        ) / sum(total_usage_hours) as pct_usage_commitment
        , sum(total_usage_hours_purchase.spot) / sum(total_usage_hours) as pct_usage_spot

        -- cost and usage by purchase option
        , sum(total_effective_cost_purchase.ondemand) as total_effective_cost_ondemand
        , sum(total_effective_cost_purchase.savings_plan_1y) as total_effective_cost_sp_1y
        , sum(total_effective_cost_purchase.savings_plan_3y) as total_effective_cost_sp_3y
        , sum(total_effective_cost_purchase.reserved_1y) as total_effective_cost_ri_1y
        , sum(total_effective_cost_purchase.reserved_3y) as total_effective_cost_ri_3y
        , sum(total_effective_cost_purchase.spot) as total_effective_cost_spot
        , sum(total_usage_amount_purchase.ondemand) as total_usage_amount_ondemand
        , sum(total_usage_amount_purchase.savings_plan_1y) as total_usage_amount_sp_1y
        , sum(total_usage_amount_purchase.savings_plan_3y) as total_usage_amount_sp_3y
        , sum(total_usage_amount_purchase.reserved_1y) as total_usage_amount_ri_1y
        , sum(total_usage_amount_purchase.reserved_3y) as total_usage_amount_ri_3y
        , sum(total_usage_amount_purchase.spot) as total_usage_amount_spot
        , sum(total_list_cost_purchase.ondemand) as total_list_cost_ondemand
        , sum(total_list_cost_purchase.savings_plan_1y) as total_list_cost_sp_1y
        , sum(total_list_cost_purchase.savings_plan_3y) as total_list_cost_sp_3y
        , sum(total_list_cost_purchase.reserved_1y) as total_list_cost_ri_1y
        , sum(total_list_cost_purchase.reserved_3y) as total_list_cost_ri_3y
        , sum(total_list_cost_purchase.spot) as total_list_cost_spot

    from source
    {{ dbt_utils.group_by(12 if not include_account_fields else 14) }}

{% endmacro %}

{% macro instance_type_final_select(include_account_fields=false) %}

    select

        -- time
        usage_date

        {%- if include_account_fields %}
        -- account
        , account_id
        , payer_account_id
        {%- endif %}

        -- instance
        , service_code
        , service_name
        , region_id
        , instance_type
        , instance_type_family
        , processor_family

        -- modernization
        , processor_generation
        , latest_intel
        , latest_amd
        , latest_graviton

        -- count
        , round(count_accounts, 4) as count_accounts
        , round(count_resources, 4) as count_resources

        -- cost
        , round(total_billed_cost, 4) as total_billed_cost
        , round(total_effective_cost, 4) as total_effective_cost
        , round(total_list_cost, 4) as total_list_cost

        -- usage
        , round(total_usage_hours, 4) as total_usage_hours
        , round(total_usage_amount, 4) as total_usage_amount
        , round(total_normalized_usage_amount, 4) as total_normalized_usage_amount

        -- utilization metrics
        , round(total_hours_month, 4) as total_hours_month
        , round(pct_utilization_avg, 4) as pct_utilization_avg
        , round(pct_utilization_p5, 4) as pct_utilization_p5
        , round(pct_utilization_median, 4) as pct_utilization_median
        , round(pct_utilization_p95, 4) as pct_utilization_p95

        -- savings potential
        , round(total_savings_mon_to_fri_full_day, 2) as total_savings_mon_to_fri_full_day
        , round(total_savings_mon_to_fri_work_hours, 2) as total_savings_mon_to_fri_work_hours
        , round(total_savings_target_50_pct, 2) as total_savings_target_50_pct

        -- pct usage
        , round(pct_usage_ondemand, 4) as pct_usage_ondemand
        , round(pct_usage_commitment, 4) as pct_usage_commitment
        , round(pct_usage_spot, 4) as pct_usage_spot

        -- rates
        , hourly_rate_blended
        , pct_discount_blended
        , hourly_rate_purchase
        , pct_discount_purchase

        -- billing period
        , billing_period

    from rates

{% endmacro %}

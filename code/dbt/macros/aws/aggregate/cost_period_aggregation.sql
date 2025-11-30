{% macro cost_period_aggregation(
    source_model,
    period,
    group_by_columns,
    cost_metric='total_effective_cost',
    include_date_bounds=false,
    partition_by=none,
    time_filter_column='billing_period'
) %}

{%- set period_func = 'week' if period == 'week' else 'month' -%}
{%- if partition_by is none -%}
    {#-- Default behavior for backward compatibility --#}
    {%- set partition_cols = [group_by_columns[0], group_by_columns[3] if group_by_columns|length > 3 else group_by_columns[0]] -%}
{%- else -%}
    {%- set partition_cols = partition_by -%}
{%- endif -%}

with

source as (

    select *
    from {{ ref(source_model) }}
    where {{ cost_metric }} > 0
    {%- if time_filter_column is not none %}
        and {{ get_model_time_filter(date_column=time_filter_column) }}
    {%- endif %}

)

{%- if include_date_bounds %}

-- Define the date bounds by calculating the minimum and maximum usage dates from the source table
, date_bounds as (

    select
        min(usage_date) as start_date
        , case
            when max(usage_date) >= current_date
                -- The date is adjusted by subtracting 2 days from current_date
                -- to exclude recent dates that might are incomplete in the CUR files
                then date_add('day', -2, current_date)
            else max(usage_date)
        end as end_date
    from source

)

-- Generate a series of {{ period }}s between the start_date and end_date from date_bounds
, date_series as (

    select
        date_trunc('{{ period_func }}', t.date_sequence) as {{ period_func }}_date
        , count(*) as ct_days
    from date_bounds
    cross join
        unnest(
            sequence(date_bounds.start_date, date_bounds.end_date, interval '1' day)
        ) as t (date_sequence)
    group by 1

)

{%- endif %}

-- Aggregate the cost data by {{ period }}
, cost_agg as (

    select
        date_trunc('{{ period_func }}', source.usage_date) as usage_date
{%- for column in group_by_columns %}
        , source.{{ column }}
{%- endfor %}
        , round(sum(source.{{ cost_metric }}), 4) as total_effective_cost
{%- if include_date_bounds %}
    {%- if period == 'week' %}
        , round(sum(source.{{ cost_metric }} / date_series.ct_days) * 7, 4) as total_effective_cost_normalized
    {%- else %}
        -- For month: normalize to 30-day month
        , round(sum(source.{{ cost_metric }} / date_series.ct_days) * 30, 4) as total_effective_cost_normalized
    {%- endif %}
{%- elif period == 'month' %}
        , round(sum(source.{{ cost_metric }}) / day(date_add('day', -1, date_add('month', 1, date_trunc('month', source.usage_date)))), 4) as total_effective_cost_normalized
{%- else %}
        , round(sum(source.{{ cost_metric }}), 4) as total_effective_cost_normalized
{%- endif %}
        , source.billing_period
    from source
{%- if include_date_bounds %}
    inner join date_series on (date_trunc('{{ period_func }}', source.usage_date) = date_series.{{ period_func }}_date)
{%- endif %}
    {{ dbt_utils.group_by(group_by_columns|length + 1) }}, source.billing_period

)

-- Calculate the previous {{ period }}'s costs using window functions
, calc_lag as (

    select
        *
        , lag(total_effective_cost) over win as prev_{{ period }}_cost
        , lag(total_effective_cost_normalized) over win as prev_{{ period }}_cost_normalized
    from cost_agg
    window win as (
        partition by {{ partition_cols | join(', ') }}
        order by usage_date
    )

)

-- Calculate {{ period }}-over-{{ period }} changes for absolute and percentage values
, {{ period }}_over_{{ period }} as (

    select
        *
        , round(total_effective_cost - prev_{{ period }}_cost, 4) as total_change
        , round((total_effective_cost - prev_{{ period }}_cost) / nullif(prev_{{ period }}_cost, 0), 4)
            as pct_change
        , round(total_effective_cost_normalized - prev_{{ period }}_cost_normalized, 4) as total_change_normalized
        , round(
            (total_effective_cost_normalized - prev_{{ period }}_cost_normalized) / nullif(prev_{{ period }}_cost_normalized, 0)
            , 4
        )
            as pct_change_normalized
    from calc_lag

)

, ranked as (

    select
        *
        , row_number() over (
            partition by usage_date
            order by total_change_normalized desc
        ) as rank_increase
        , row_number() over (
            partition by usage_date
            order by total_change_normalized asc
        ) as rank_decrease
    from {{ period }}_over_{{ period }}

)

, add_flags as (

    select
        *
        , rank_increase as rank_position
        , coalesce(rank_increase <= 50 or rank_decrease <= 50, false) as is_top_50_mover
    from ranked

)

, final as (

    select
        -- time
        usage_date

        -- account
{%- for column in group_by_columns[:3] %}
        , {{ column }}
{%- endfor %}

        -- service
{%- for column in group_by_columns[3:] %}
        , {{ column }}
{%- endfor %}

        -- cost
        , total_effective_cost
        , total_effective_cost_normalized

        -- trend changes
        , total_change
        , total_change_normalized
        , pct_change
        , pct_change_normalized

        -- ranking
        , rank_position
        , is_top_50_mover

        -- billing period
        , billing_period

    from add_flags

)

select *
from final

{% endmacro %}

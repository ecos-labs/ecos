{% snapshot snapshot__cur_hourly_stats %}

{{ config(
    unique_key='usage_hour',
    strategy='timestamp',
    updated_at='max_modified_date')
}}

    with

    date_sequence as (

        select date_hour
        from
            unnest(
                sequence(
                    date_trunc('month', current_date) - interval '2' month
                    , date_trunc('month', current_date) + interval '1' month
                    , interval '1' hour
                )
            ) as t (date_hour)
        where date_trunc('month', date_hour) != date_trunc('month', current_date + interval '1' month)

    )

    , cur_data as (

        select
            date_trunc('hour', usage_date) as usage_hour
            , count(*) as count_rows
            , sum(unblended_cost) as cost
            , count(distinct payer_account_id) as ct_accounts
            , cast(current_timestamp as timestamp) as max_modified_date
            , cast(current_timestamp as timestamp) as min_modified_date
            , array_distinct(array_agg(charge_type)) as line_item_type
        from {{ ref('bronze_aws__cur_source') }}
        where
            usage_date >= (select min(cast(date_sequence.date_hour as date)) from date_sequence)
            and usage_date < (select max(cast(date_sequence.date_hour as date)) + interval '1' day from date_sequence)
            and charge_type != 'Tax'
        {{ dbt_utils.group_by(1) }}

    )

    , cur_perc_change as (

        select
            *
            , lag(cost) over (
                order by usage_hour asc
            ) as lag_cost
        from cur_data

    )

    , final_metrics as (

        select
            cast(date_sequence.date_hour as timestamp) as usage_hour
            , cast(date_sequence.date_hour as date) as usage_date
            , cur_perc_change.count_rows
            , cur_perc_change.cost
            , cur_perc_change.lag_cost
            , round(cur_perc_change.cost / nullif(cur_perc_change.lag_cost, 0), 2) as lag_perc_change
            , cur_perc_change.ct_accounts
            , cur_perc_change.max_modified_date
            , cur_perc_change.min_modified_date
            , date_diff('day', cast(date_sequence.date_hour as date), cast(cur_perc_change.min_modified_date as date))
                as date_diff_usage_modified
            , cur_perc_change.line_item_type
        from date_sequence
        left join cur_perc_change on (date_sequence.date_hour = cur_perc_change.usage_hour)

    )

    select
        usage_hour
        , usage_date
        , count_rows
        , cost
        , lag_cost
        , lag_perc_change
        , ct_accounts
        , max_modified_date
        , min_modified_date
        , date_diff_usage_modified
        , line_item_type
        , case
            when count_rows is null then 'x'
            else ''
        end as is_missing_hour
        , case
            when lag_perc_change < 0.05 or lag_perc_change > 5 then 'x'
            else ''
        end as is_cost_outlier
        , case
            when date_diff_usage_modified > 45 then 'x'
            else ''
        end as is_late_arriving
    from final_metrics

{% endsnapshot %}

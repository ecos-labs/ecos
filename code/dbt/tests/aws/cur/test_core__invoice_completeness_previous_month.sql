with current_dates as (
    select

        current_date as today
        , date_trunc('month', current_date) as current_month_start
        , date_trunc('month', date_add('month', -1, current_date)) as previous_month_start
        , day(current_date) as current_day_of_month
)

, previous_month_invoices as (
    select mi.*
    from {{ ref('gold_core__invoice_monthly') }} as mi
    cross join current_dates as cd
    where mi.billing_period_start_date = cd.previous_month_start
)

, validation_results as (
    select
        count(*) as total_invoices
        , sum(case when mi.invoice_id is null or mi.invoice_id = '' then 1 else 0 end) as missing_invoice_ids
        , sum(case when not mi.is_invoice_complete then 1 else 0 end) as incomplete_invoices
        , min(cd.current_day_of_month) as current_day
    from previous_month_invoices as mi
    cross join current_dates as cd
    where cd.current_day_of_month >= 15
)

select *
from validation_results
where missing_invoice_ids > 0 or incomplete_invoices > 0

-- service_code = 'AmazonQuickSight'
{% macro aws_mappings_quicksight_category(usage_type_col='usage_type') %}
    case
        -- authors
        when {{ usage_type_col }} like '%User-Standard%' then 'QuickSight Authors [Standard]' -- deprecated
        when {{ usage_type_col }} like '%User-Enterprise%' then 'QuickSight Authors [Enterprise]'
        when {{ usage_type_col }} like '%Author-Pro%' then 'QuickSight Authors Pro [Enterprise]'

        -- readers
        when {{ usage_type_col }} like '%Reader-Usage%' then 'QuickSight Readers [Sessions]'
        when {{ usage_type_col }} like '%Reader-Capacity%' then 'QuickSight Readers [Capacity]'
        when {{ usage_type_col }} like '%Reader-Enterprise%' then 'QuickSight Readers [Enterprise]'
        when {{ usage_type_col }} like '%Reader-Pro%' then 'QuickSight Readers [Pro]'

        -- data storage
        when {{ usage_type_col }} like '%SPICE' then 'QuickSight SPICE Storage'

        -- alerts
        when {{ usage_type_col }} like '%Alerts%' then 'QuickSight Alerts'

        -- quicksight q
        when {{ usage_type_col }} like '%Q-Query-Capacity%' then 'QuickSight Q Query [Capacity]'
        when {{ usage_type_col }} like '%Amazon-Q-QS-Fee%' then 'QuickSight Q Monthly Fee'

        -- reports
        when {{ usage_type_col }} like '%-Report%' then 'QuickSight Reporting'

        else 'QuickSight [Other]'

    end
{% endmacro %}

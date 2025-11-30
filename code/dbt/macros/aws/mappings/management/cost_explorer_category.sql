-- service_code = 'AWSCostExplorer'
{% macro aws_mappings_cost_explorer_category(usage_type_col='usage_type') %}
    case
        -- API requests
        when {{ usage_type_col }} like '%API-Request%' then 'Cost Explorer API Requests [Management]'

        -- Usage reports
        when {{ usage_type_col }} like '%UsageReport%' then 'Cost Explorer Usage Reports [Management]'

        else 'Cost Explorer [Other]'
    end
{% endmacro %}

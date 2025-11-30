-- service_code = 'AWSSupportEnterprise'
{% macro aws_mappings_awssupport_category(usage_type_col='usage_type') %}
    case
        -- Enterprise Support charges
        when {{ usage_type_col }} like '%Dollar%' then 'AWS Support Enterprise Charges [Management]'

        else 'AWS Support Enterprise [Other]'
    end
{% endmacro %}

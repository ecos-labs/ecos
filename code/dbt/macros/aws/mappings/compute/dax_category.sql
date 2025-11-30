-- service_code = 'AmazonDAX'
{% macro aws_mappings_dax_category(usage_type_col='usage_type') %}
    case
        -- DAX node instances
        when {{ usage_type_col }} like '%NodeUsage%' and {{ usage_type_col }} like '%dax.r%' then 'DAX Memory Optimized Nodes [Compute]'
        when {{ usage_type_col }} like '%NodeUsage%' and {{ usage_type_col }} like '%dax.t%' then 'DAX Burstable Nodes [Compute]'
        when {{ usage_type_col }} like '%NodeUsage%' and {{ usage_type_col }} not like '%dax.r%' and {{ usage_type_col }} not like '%dax.t%' then 'DAX Nodes [Compute]'

        else 'DAX [Other]'
    end
{% endmacro %}

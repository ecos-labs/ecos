-- service_code = 'AmazonAthena'
{% macro aws_mappings_athena_category(usage_type_col='usage_type') %}
    case
        -- Data scanning for queries
        when {{ usage_type_col }} like '%DataScannedInTB%' then 'Athena Query [Data Scanned]'

        -- Provisioned capacity for DPU hours
        when {{ usage_type_col }} like '%ReservedCapacityInDPUHours%' then 'Athena Query [Provisioned Capacity]'

        -- Spark code execution
        when {{ usage_type_col }} like '%CodeExecutionInDPUHours%' then 'Athena Spark [Compute]'

        else 'Athena [Other]'
    end
{% endmacro %}

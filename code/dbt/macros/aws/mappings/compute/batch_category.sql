-- service_code = 'AWSBatch'
{% macro aws_mappings_batch_category(service_code_col='service_code', usage_type_col='usage_type') %}
    case
        -- Compute Environment
        when {{ usage_type_col }} like '%ComputeEnvironment%'
            then 'Batch Compute Environment [Compute]'

        -- Job Queue
        when {{ usage_type_col }} like '%JobQueue%'
            then 'Batch Job Queue [Management]'

        else 'Batch [Other]'
    end
{% endmacro %}

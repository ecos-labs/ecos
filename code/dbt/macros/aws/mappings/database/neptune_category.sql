-- service_code = 'AmazonNeptune'
{% macro aws_mappings_neptune_category(usage_type_col='usage_type') %}
    case

        -- Instance
        when {{ usage_type_col }} like '%InstanceUsage%' then 'Neptune Instance [Compute]'
        when {{ usage_type_col }} like '%GraphUsage%' then 'Neptune Instance [Compute]'

        -- Storage
        when {{ usage_type_col }} like '%StorageUsage%' then 'Neptune Data [Storage]'

        -- I/O operations
        when {{ usage_type_col }} like '%StorageIOUsage%' then 'Neptune I/O [Management]'

        -- Serverless
        when {{ usage_type_col }} like '%Serverless%' then 'Neptune Serverless [Compute]'

        -- Snapshots
        when {{ usage_type_col }} like '%BackupUsage%' then 'Neptune Snapshot [Storage]'
        when {{ usage_type_col }} like '%GraphSnapshot%' then 'Neptune Snapshot [Storage]'

        else 'Neptune [Other]'

    end
{% endmacro %}

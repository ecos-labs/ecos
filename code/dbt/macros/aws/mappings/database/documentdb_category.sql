-- service_code = 'AmazonDocDB'
{% macro aws_mappings_documentdb_category(usage_type_col='usage_type') %}
    case

        -- Instance
        when {{ usage_type_col }} like '%InstanceUsage%' then 'DocumentDB Instance [Compute]'

        -- Storage
        when {{ usage_type_col }} like '%StorageUsage%' then 'DocumentDB Storage [Storage]'

        -- I/O operations
        when {{ usage_type_col }} like '%StorageIOUsage%' then 'DocumentDB I/O [Management]'

        -- Elastic CPU
        when {{ usage_type_col }} like '%ElasticCPUUsage%' then 'DocumentDB Utilization [Compute]'

        -- Backups
        when {{ usage_type_col }} like '%BackupUsage%' then 'DocumentDB Storage Snapshot [Storage]'

        else 'DocumentDB [Other]'
    end
{% endmacro %}

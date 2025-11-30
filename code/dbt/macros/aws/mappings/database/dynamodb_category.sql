-- service_code = 'AmazonDynamoDB'
{% macro aws_mappings_dynamodb_category(usage_type_col='usage_type') %}
    case
        -- On-Demand capacity
        when {{ usage_type_col }} like '%RequestUnit%' then 'DynamoDB On-Demand Capacity [Compute]'

        -- Provisioned capacity
        when {{ usage_type_col }} like '%CapacityUnit%' then 'DynamoDB Provisioned Capacity [Compute]'
        when {{ usage_type_col }} like '%HeavyUsage%' then 'DynamoDB Provisioned Capacity [Compute]'

        -- Storage costs
        when {{ usage_type_col }} like '%TimedStorage%' then 'DynamoDB Storage [Storage]'

        -- Backup features
        when {{ usage_type_col }} like '%TimedBackup%' then 'DynamoDB Backups [Storage]'
        when {{ usage_type_col }} like '%TimedPITR%' then 'DynamoDB Backups [Storage]'

        -- Import/Export operations
        when {{ usage_type_col }} like '%ExportDataSize-Bytes%' then 'DynamoDB S3 Export [Network]'
        when {{ usage_type_col }} like '%ImportDataSize-Bytes%' then 'DynamoDB S3 Import [Network]'
        when {{ usage_type_col }} like '%RestoreDataSize-Bytes%' then 'DynamoDB Restored Backups [Storage]'

        else 'DynamoDB [Other]'
    end
{% endmacro %}

-- service_code = 'AmazonRDS'
{% macro aws_mappings_rds_category(usage_type_col='usage_type', engine_col='engine', operation_col='operation') %}
    case

        -- Instance
        when {{ usage_type_col }} like '%InstanceUsage%' then 'RDS ' || {{ engine_col }} || ' Instance Single-AZ [Compute]'
        when {{ usage_type_col }} like '%Multi-AZUsage%' then 'RDS ' || {{ engine_col }} || ' Instance Multi-AZ [Compute]'
        when {{ usage_type_col }} like '%MirrorUsage%' then 'RDS ' || {{ engine_col }} || ' Instance Multi-AZ [Compute]'
        when {{ usage_type_col }} like '%ProxyUsage%' then 'RDS ' || {{ engine_col }} || ' Proxy [Compute]'

        -- Aurora storage and I/O
        when {{ usage_type_col }} like '%Aurora:StorageUsage%' then 'RDS ' || {{ engine_col }} || ' Storage [Storage]'
        when {{ usage_type_col }} like '%Aurora:StorageIOUsage%' then 'RDS ' || {{ engine_col }} || ' I/O [Management]'

        -- GP2/GP3 storage
        when {{ usage_type_col }} like '%GP2-Storage%' then 'RDS ' || {{ engine_col }} || ' GP2 Storage [Storage]'
        when {{ usage_type_col }} like '%GP3-Storage%' then 'RDS ' || {{ engine_col }} || ' GP3 Storage [Storage]'
        when {{ usage_type_col }} like '%StorageUsage%' then 'RDS ' || {{ engine_col }} || ' Magnetic Storage [Storage]'

        -- Backup usage
        when {{ usage_type_col }} like '%Aurora:BackupUsage%' then 'RDS ' || {{ engine_col }} || ' Backup [Storage]'
        when {{ usage_type_col }} like '%ChargedBackupUsage%' then 'RDS ' || {{ engine_col }} || ' Backup [Storage]'

        -- Serverless
        when {{ usage_type_col }} like '%Aurora:ServerlessV2Usage%' then 'RDS ' || {{ engine_col }} || ' Serverless V2 [Compute]'
        when {{ usage_type_col }} like '%Aurora:ServerlessUsage%' then 'RDS ' || {{ engine_col }} || ' Serverless [Compute]'

        -- Advanced Multi-AZ and serverless features
        when {{ usage_type_col }} like '%Multi-AZClusterUsage%' then 'RDS ' || {{ engine_col }} || ' Instance Multi-AZ readable standby [Compute]'
        when {{ usage_type_col }} like '%Aurora:ServerlessV2IOOptimizedUsage%' then 'RDS ' || {{ engine_col }} || ' Serverless V2 IO Optimized [Compute]'

        -- PIOPS storage
        when {{ usage_type_col }} like '%PIOPS-Storage-IO2%' then 'RDS ' || {{ engine_col }} || ' io2 Storage [Storage]'
        when {{ usage_type_col }} like '%PIOPS-Storage%' then 'RDS ' || {{ engine_col }} || ' io1 Storage [Storage]'
        when {{ usage_type_col }} like '%IO2-PIOPS%' then 'RDS ' || {{ engine_col }} || ' io2 PIOPS [Storage]'
        when {{ usage_type_col }} like '%PIOPS%' then 'RDS ' || {{ engine_col }} || ' io1 PIOPS [Storage]'

        -- GP3 advanced features
        when {{ usage_type_col }} like '%GP3-PIOPS%' then 'RDS ' || {{ engine_col }} || ' GP3 PIOPS [Storage]'
        when {{ usage_type_col }} like '%GP3-Throughput%' then 'RDS ' || {{ engine_col }} || ' GP3 Provisioned Throughput [Storage]'

        -- Advanced features
        when {{ usage_type_col }} like '%Aurora:IO-OptimizedStorageUsage%' then 'RDS ' || {{ engine_col }} || ' I/O Optimized Storage [Storage]'
        when {{ usage_type_col }} like '%Aurora:ReplicatedWriteIO%' then 'RDS ' || {{ engine_col }} || ' Replicated Write I/Os [Management]'

        -- Extended support and backtrack
        when {{ usage_type_col }} like '%Aurora:BacktrackUsage%' then 'RDS ' || {{ engine_col }} || ' Backtrack Change Records [Management]'
        when {{ usage_type_col }} like '%ExtendedSupport%' then 'RDS Extended Support [Fee]'

        else 'RDS [Other]'

    end
{% endmacro %}

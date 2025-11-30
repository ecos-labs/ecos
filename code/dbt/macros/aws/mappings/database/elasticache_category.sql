-- service_code = 'AmazonElastiCache'
{% macro aws_mappings_elasticache_category(usage_type_col='usage_type') %}
    case
        -- Cache node instances
        when {{ usage_type_col }} like '%NodeUsage:cache.%' then 'ElastiCache Cache Nodes [Compute]'
        when {{ usage_type_col }} like '%Outpost-NodeUsage:%' then 'ElastiCache Outpost Nodes [Compute]'

        -- ElastiCache Serverless
        when {{ usage_type_col }} like '%ElastiCacheProcessingUnits:%' then 'ElastiCache Serverless Processing Units [Compute]'
        when {{ usage_type_col }} like '%CachedData:%' then 'ElastiCache Serverless Storage [Storage]'

        -- Backup storage
        when {{ usage_type_col }} like '%BackupUsage:Redis%' then 'ElastiCache Redis Backup Storage [Storage]'
        when {{ usage_type_col }} like '%BackupUsage:Valkey%' then 'ElastiCache Valkey Backup Storage [Storage]'
        when {{ usage_type_col }} like '%ElastiCache:BackupUsage%' then 'ElastiCache Backup Storage [Storage]'

        -- Data transfer
        when {{ usage_type_col }} like '%ElastiCache-Out-Bytes%' then 'ElastiCache Data Transfer Out [Network]'

        else 'ElastiCache [Other]'
    end
{% endmacro %}

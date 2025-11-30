-- service_code = 'AmazonEFS'
{% macro aws_mappings_efs_category(usage_type_col='usage_type') %}
    case
        -- Standard storage (all variants)
        when {{ usage_type_col }} like '%TimedStorage-%' then 'EFS Standard Storage [Storage]'

        -- Infrequent Access storage (all variants)
        when {{ usage_type_col }} like '%IATimedStorage-%' then 'EFS Infrequent Access Storage [Storage]'

        -- Archive storage (including early delete)
        when {{ usage_type_col }} like '%ArchiveTimedStorage-%' or {{ usage_type_col }} like '%ArchiveEarlyDelete-%' then 'EFS Archive Storage [Storage]'

        -- Data access fees (all types)
        when {{ usage_type_col }} like '%DataAccess-Bytes%' then 'EFS Data Access [Management]'

        -- Throughput management
        when {{ usage_type_col }} like '%TP-MiBpsHrs%' then 'EFS Throughput [Management]'

        -- Data transfer
        when {{ usage_type_col }} like '%AWS-Out-Bytes%' then 'EFS Data Transfer [Network]'

        else 'EFS [Other]'
    end
{% endmacro %}

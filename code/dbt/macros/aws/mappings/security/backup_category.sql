-- service_code = 'AWSBackup'
{% macro aws_mappings_backup_category(usage_type_col='usage_type') %}
    case
        -- Storage (warm and cold combined)
        when {{ usage_type_col }} like '%WarmStorage-ByteHrs%' then 'AWS Backup Warm Storage [Storage]'
        when {{ usage_type_col }} like '%ColdStorage-ByteHrs%' then 'AWS Backup Cold Storage [Storage]'
        when {{ usage_type_col }} like '%-LAGV%' then 'AWS Backup Legal Archive [Storage]'

        -- Cross-region operations
        when {{ usage_type_col }} like '%CrossRegion-%' then 'AWS Backup Cross-Region [Network]'

        -- Restore operations (all types)
        when {{ usage_type_col }} like '%Restore-%' or {{ usage_type_col }} like '%PartialRestore-%' then 'AWS Backup Restore [Management]'

        -- Management operations
        when {{ usage_type_col }} like '%BackupEvaluations%' then 'AWS Backup Evaluations [Management]'
        when {{ usage_type_col }} like '%RT-RecoveryPoint%' then 'AWS Backup Recovery Testing [Management]'
        when {{ usage_type_col }} like '%EarlyDelete%' then 'AWS Backup Early Delete [Management]'

        -- Data transfer
        when {{ usage_type_col }} like '%DataTransfer-%' then 'AWS Backup Data Transfer [Network]'

        else 'AWS Backup [Other]'
    end
{% endmacro %}

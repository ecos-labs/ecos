-- service_code = 'AmazonKinesisAnalytics'
{% macro aws_mappings_kinesis_analytics_category(usage_type_col='usage_type') %}
    case
        -- KPU (Kinesis Processing Units) - compute
        when {{ usage_type_col }} like '%KPU-Hour%' and {{ usage_type_col }} like '%Interactive%' then 'Kinesis Analytics Interactive KPU [Compute]'
        when {{ usage_type_col }} like '%KPU-Hour%' and {{ usage_type_col }} like '%Java%' then 'Kinesis Analytics Java KPU [Compute]'
        when {{ usage_type_col }} like '%KPU-Hour%' and {{ usage_type_col }} not like '%Interactive%' and {{ usage_type_col }} not like '%Java%' then 'Kinesis Analytics KPU [Compute]'

        -- Application storage
        when {{ usage_type_col }} like '%RunningApplicationStorage%' and {{ usage_type_col }} like '%Interactive%' then 'Kinesis Analytics Interactive Storage [Storage]'
        when {{ usage_type_col }} like '%RunningApplicationStorage%' and {{ usage_type_col }} not like '%Interactive%' then 'Kinesis Analytics Storage [Storage]'

        -- Application backups
        when {{ usage_type_col }} like '%DurableApplicationBackups%' then 'Kinesis Analytics Backups [Storage]'

        else 'Kinesis Analytics [Other]'
    end
{% endmacro %}

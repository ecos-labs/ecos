-- service_code = 'ElasticMapReduce'
{% macro aws_mappings_emr_category(usage_type_col='usage_type') %}
    case
        -- EMR Serverless - x86 architecture
        when {{ usage_type_col }} like '%EMR-SERVERLESS-vCPUHours%' and {{ usage_type_col }} not like '%ARM%' then 'EMR Serverless vCPU [Compute]'
        when {{ usage_type_col }} like '%EMR-SERVERLESS-MemoryGBHours%' and {{ usage_type_col }} not like '%ARM%' then 'EMR Serverless Memory [Compute]'

        -- EMR Serverless - ARM architecture
        when {{ usage_type_col }} like '%EMR-SERVERLESS-ARM-vCPUHours%' then 'EMR Serverless ARM vCPU [Compute]'
        when {{ usage_type_col }} like '%EMR-SERVERLESS-ARM-MemoryGBHours%' then 'EMR Serverless ARM Memory [Compute]'

        -- EMR Serverless storage
        when {{ usage_type_col }} like '%EMR-SERVERLESS-StorageGBHours%' then 'EMR Serverless Storage [Storage]'

        -- EMR on EKS
        when {{ usage_type_col }} like '%EMR-EKS-EC2-vCPUHours%' then 'EMR on EKS vCPU [Compute]'
        when {{ usage_type_col }} like '%EMR-EKS-EC2-GBHours%' then 'EMR on EKS Memory [Compute]'

        -- EMR EC2 instances (BoxUsage)
        when {{ usage_type_col }} like '%BoxUsage:%' then 'EMR EC2 Instances [Compute]'

        else 'EMR [Other]'
    end
{% endmacro %}

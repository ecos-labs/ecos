-- service_code = 'AWSELB'
{% macro aws_mappings_elb_category(usage_type_col='usage_type', operation_col='operation') %}
    case
        -- Load Balancer hourly charges
        when {{ usage_type_col }} like '%LoadBalancerUsage%' then 'ELB [Hourly]'

        -- Load Balancer capacity units (LCU)
        when {{ usage_type_col }} like '%LCUUsage%' then 'ELB [Capacity Units]'
        when {{ usage_type_col }} like '%NLCUUsage%' then 'ELB [Capacity Units]'
        when {{ usage_type_col }} like '%GLCUUsage%' then 'ELB [Capacity Units]'

        -- Data processing
        when {{ usage_type_col }} like '%DataProcessing-Bytes%' then 'ELB [Data Processing]'

        -- Data transfer
        when {{ usage_type_col }} like '%Out-Bytes%' then 'ELB [Data Transfer]'
        when {{ usage_type_col }} like '%In-Bytes%' then 'ELB [Data Transfer]'
        when {{ usage_type_col }} like '%Regional-Bytes%' then 'ELB [Data Transfer]'

        -- IP address charges
        when {{ usage_type_col }} like '%Address%' then 'ELB [IP Address]'

        -- Default fallback
        else 'ELB [Other]'
    end
{% endmacro %}

-- service_code = 'CloudHSM'
{% macro aws_mappings_cloudhsm_category(usage_type_col='usage_type') %}
    case
        -- HSM instance usage
        when {{ usage_type_col }} like '%CloudHSMv2Usage%' and {{ usage_type_col }} like '%hsm2m.m%' then 'CloudHSM v2 hsm2m.medium [Security]'
        when {{ usage_type_col }} like '%CloudHSMv2Usage%' and {{ usage_type_col }} not like '%hsm2m.m%' then 'CloudHSM v2 hsm1.medium [Security]'

        else 'CloudHSM [Other]'
    end
{% endmacro %}

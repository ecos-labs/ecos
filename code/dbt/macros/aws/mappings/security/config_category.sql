-- service_code = 'AWSConfig'
{% macro aws_mappings_config_category(usage_type_col='usage_type') %}
    case
        -- Configuration item recording
        when {{ usage_type_col }} like '%ConfigurationItemRecorded%' and {{ usage_type_col }} not like '%Daily%' then 'AWS Config Configuration Items [Management]'
        when {{ usage_type_col }} like '%ConfigurationItemRecordedDaily%' then 'AWS Config Daily Configuration Items [Management]'

        -- Config rules evaluation
        when {{ usage_type_col }} like '%ConfigRuleEvaluations%' then 'AWS Config Rule Evaluations [Security]'

        -- Conformance pack evaluations
        when {{ usage_type_col }} like '%ConformancePackEvaluations%' then 'AWS Config Conformance Pack Evaluations [Security]'

        else 'AWS Config [Other]'
    end
{% endmacro %}

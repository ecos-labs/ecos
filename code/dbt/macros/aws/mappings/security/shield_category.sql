-- service_code = 'AWSShield'
{% macro aws_mappings_shield_category(usage_type_col='usage_type') %}
    case
        -- DDoS Response Team
        when {{ usage_type_col }} like '%DRT%' then 'Shield DDoS Response Team [Security]'

        -- Advanced protection
        when {{ usage_type_col }} like '%AdvancedProtection%' then 'Shield Advanced Protection [Security]'

        -- Subscription
        when {{ usage_type_col }} like '%Subscription%' then 'Shield Advanced Subscription [Security]'

        else 'Shield [Other]'
    end
{% endmacro %}

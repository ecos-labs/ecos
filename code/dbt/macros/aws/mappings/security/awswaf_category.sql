-- service_code = 'awswaf'
{% macro aws_mappings_awswaf_category(usage_type_col='usage_type') %}
    case
        -- Web ACL management
        when {{ usage_type_col }} like '%WebACL%' and {{ usage_type_col }} not like '%Shield%' then 'WAF Web ACLs [Management]'
        when {{ usage_type_col }} like '%ShieldProtected-WebACL%' then 'WAF Shield Protected Web ACLs [Management]'

        -- Rules management
        when {{ usage_type_col }} like '%Rule%' and {{ usage_type_col }} not like '%Shield%' and {{ usage_type_col }} not like '%Request%' then 'WAF Rules [Management]'
        when {{ usage_type_col }} like '%ShieldProtected-Rule%' then 'WAF Shield Protected Rules [Management]'

        -- Request processing (various tiers and types)
        when {{ usage_type_col }} like '%Request%' and {{ usage_type_col }} like '%Tier1%' and {{ usage_type_col }} not like '%Shield%' then 'WAF Requests Tier 1 [Processing]'
        when {{ usage_type_col }} like '%Request%' and {{ usage_type_col }} like '%Tier2%' and {{ usage_type_col }} not like '%Shield%' then 'WAF Requests Tier 2 [Processing]'
        when {{ usage_type_col }} like '%Request%' and {{ usage_type_col }} like '%Tier3%' and {{ usage_type_col }} not like '%Shield%' then 'WAF Requests Tier 3 [Processing]'
        when {{ usage_type_col }} like '%Request%' and {{ usage_type_col }} like '%Tier4%' and {{ usage_type_col }} not like '%Shield%' then 'WAF Requests Tier 4 [Processing]'
        when {{ usage_type_col }} like '%Request%' and {{ usage_type_col }} like '%Tier5%' and {{ usage_type_col }} not like '%Shield%' then 'WAF Requests Tier 5 [Processing]'
        when {{ usage_type_col }} like '%Request%' and {{ usage_type_col }} like '%Tier6%' and {{ usage_type_col }} not like '%Shield%' then 'WAF Requests Tier 6 [Processing]'
        when {{ usage_type_col }} like '%Request%' and {{ usage_type_col }} like '%Tier7%' and {{ usage_type_col }} not like '%Shield%' then 'WAF Requests Tier 7 [Processing]'
        when {{ usage_type_col }} like '%Request%' and {{ usage_type_col }} not like '%Shield%' and {{ usage_type_col }} not like '%Tier%' then 'WAF Requests Standard [Processing]'

        -- Shield protected requests
        when {{ usage_type_col }} like '%ShieldProtected-Request%' then 'WAF Shield Protected Requests [Processing]'

        -- Bot Control (AWS Managed Rules)
        when {{ usage_type_col }} like '%AMR-BotControl%' and {{ usage_type_col }} not like '%Request%' and {{ usage_type_col }} not like '%Shield%' then 'WAF Bot Control [Security]'
        when {{ usage_type_col }} like '%AMR-BotControl-Request%' and {{ usage_type_col }} not like '%Shield%' then 'WAF Bot Control Requests [Security]'
        when {{ usage_type_col }} like '%AMR-BotControl-Targeted-Request%' then 'WAF Bot Control Targeted Requests [Security]'
        when {{ usage_type_col }} like '%ShieldProtected-AMR-BotControl%' then 'WAF Shield Protected Bot Control [Security]'

        -- Account Takeover Protection (ATP)
        when {{ usage_type_col }} like '%AMR-ATP%' and {{ usage_type_col }} not like '%Shield%' then 'WAF Account Takeover Protection [Security]'
        when {{ usage_type_col }} like '%ShieldProtected-AMR-ATP%' then 'WAF Shield Protected ATP [Security]'

        -- Fraud Control
        when {{ usage_type_col }} like '%AMR-FraudControl%' and {{ usage_type_col }} not like '%Shield%' then 'WAF Fraud Control [Security]'
        when {{ usage_type_col }} like '%ShieldProtected-AMR-FraudControl%' then 'WAF Shield Protected Fraud Control [Security]'

        -- CAPTCHA services
        when {{ usage_type_col }} like '%CaptchaAttempted%' then 'WAF CAPTCHA Attempts [Security]'
        when {{ usage_type_col }} like '%ChallengeServed%' and {{ usage_type_col }} not like '%Shield%' then 'WAF Challenge Responses [Security]'
        when {{ usage_type_col }} like '%ShieldProtected-ChallengeServed%' then 'WAF Shield Protected Challenges [Security]'

        else 'WAF [Other]'
    end
{% endmacro %}

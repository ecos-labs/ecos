-- service_code = 'AWSSecurityHub'
{% macro aws_mappings_security_hub_category(usage_type_col='usage_type') %}
    case
        -- Compliance checks
        when {{ usage_type_col }} like '%PaidComplianceCheck%' then 'Security Hub Compliance Checks [Security]'
        when {{ usage_type_col }} like '%FreeComplianceCheck%' then 'Security Hub Free Compliance Checks [Security]'

        -- Findings ingestion - paid
        when {{ usage_type_col }} like '%OtherProduct:PaidFindingsIngestion%' then 'Security Hub Paid Findings Ingestion [Security]'

        -- Findings ingestion - free tier
        when {{ usage_type_col }} like '%SecurityHubProduct:FreeFindingsIngestion%' then 'Security Hub Free Findings Ingestion [Security]'
        when {{ usage_type_col }} like '%FreeFindingsIngestion-CrossRegion%' then 'Security Hub Free Cross-Region Findings [Security]'
        when {{ usage_type_col }} like '%OtherProduct:FreeFindingsIngestion-FreeTrial%' then 'Security Hub Free Trial Findings [Security]'

        -- Automation rules
        when {{ usage_type_col }} like '%RuleEvaluation%' and {{ usage_type_col }} not like '%FreeTrial%' then 'Security Hub Rule Evaluation [Security]'
        when {{ usage_type_col }} like '%RuleEvaluation-FreeTrial%' then 'Security Hub Free Rule Evaluation [Security]'

        else 'Security Hub [Other]'
    end
{% endmacro %}

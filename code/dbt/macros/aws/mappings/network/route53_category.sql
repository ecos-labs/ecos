-- service_code = 'AmazonRoute53'
{% macro aws_mappings_route53_category(usage_type_col='usage_type', operation_col='operation') %}
    case
        -- DNS queries
        when {{ usage_type_col }} like '%Queries%' then 'Route 53 [Queries]'

        -- Hosted zones
        when {{ usage_type_col }} like '%HostedZone%' then 'Route 53 [Hosted Zones]'

        -- Health checks
        when {{ usage_type_col }} like '%HealthCheck%' then 'Route 53 [Health Checks]'

        -- Resolver
        when {{ usage_type_col }} like '%Resolver%' then 'Route 53 [Resolver]'

        -- Traffic Flow
        when {{ usage_type_col }} like '%TrafficFlow%' then 'Route 53 [Traffic Flow]'

        -- Domain registration and transfer
        when {{ usage_type_col }} like '%Domain%' then 'Route 53 [Domains]'

        -- DNSSEC
        when {{ usage_type_col }} like '%DNSSEC%' then 'Route 53 [DNSSEC]'

        -- Application Recovery Controller
        when {{ usage_type_col }} like '%ARC%' then 'Route 53 [ARC]'

        -- Default fallback
        else 'Route 53 [Other]'
    end
{% endmacro %}

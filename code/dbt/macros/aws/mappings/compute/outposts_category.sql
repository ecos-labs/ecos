-- service_code = 'AmazonOutposts'
{% macro aws_mappings_outposts_category(service_code_col='service_code', usage_type_col='usage_type') %}
    case
        -- Outpost
        when {{ usage_type_col }} like '%Outpost%'
            then 'Outposts Rack [Compute]'

        -- Instance
        when {{ usage_type_col }} like '%Instance%'
            then 'Outposts Instance [Compute]'

        -- EBS
        when {{ usage_type_col }} like '%EBS%'
            then 'Outposts EBS [Storage]'

        else 'Outposts [Other]'
    end
{% endmacro %}

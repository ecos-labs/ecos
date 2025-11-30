-- service_code = 'AmazonMQ'
{% macro aws_mappings_mq_category(usage_type_col='usage_type') %}
    case
        -- Message broker instances
        when {{ usage_type_col }} like '%ActiveMQ-Instance%' then 'MQ ActiveMQ Broker Instance [Processing]'
        when {{ usage_type_col }} like '%RabbitMQ-Instance%' then 'MQ RabbitMQ Broker Instance [Processing]'
        when {{ usage_type_col }} like '%Instance%' and {{ usage_type_col }} not like '%ActiveMQ%' and {{ usage_type_col }} not like '%RabbitMQ%' then 'MQ Broker Instance [Processing]'

        -- Storage
        when {{ usage_type_col }} like '%Storage%' then 'MQ Broker Storage [Storage]'

        else 'MQ [Other]'
    end
{% endmacro %}

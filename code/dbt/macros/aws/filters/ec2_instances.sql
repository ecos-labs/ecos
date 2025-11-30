{% macro aws_filters_ec2_instances(service_code_col='service_code', operation_col='operation', usage_type_col='usage_type', subservice_code_col='subservice_code', resource_id_col='resource_id') %}

{{ service_code_col }} = 'AmazonEC2'
and {{ operation_col }} like '%RunInstances%'
and {{ usage_type_col }} not like '%DataXfer%'
and {{ subservice_code_col }} != 'AWSDataTransfer'
and {{ resource_id_col }} not like 'arn:%:capacity-reservation/%'

{% endmacro %}

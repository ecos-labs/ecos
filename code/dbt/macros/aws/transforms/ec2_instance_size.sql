{% macro aws_transforms_ec2_instance_size(instance_type_col='instance_type') %}

regexp_extract({{ instance_type_col }}, '^([a-z]+|u\-[0-9]+tb)([0-9]+)([a-z0-9\-]*)(\.)(.+)', 5)

{% endmacro %}

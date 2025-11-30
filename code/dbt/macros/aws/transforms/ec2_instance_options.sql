{% macro aws_transforms_ec2_instance_options(instance_type_col='instance_type') %}

regexp_extract({{ instance_type_col }}, '^([a-z]+|u\-[0-9]+tb)([0-9]+)([a-z0-9\-]*)(\.)(.+)', 3)

{% endmacro %}

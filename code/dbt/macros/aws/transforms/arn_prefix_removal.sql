{% macro aws_transforms_arn_prefix_removal(arn_col='resource_id') %}
case
    when {{ arn_col }} is null then null
    when {{ arn_col }} = '' then ''
    when {{ arn_col }} not like 'arn:%' then {{ arn_col }}
    -- Single optimized regex that handles all three ARN formats in one pass
    -- Formats supported:
    --   arn:...:resource-id                    -> resource-id
    --   arn:...:resource-type/resource-id      -> resource-id
    --   arn:...:resource-type:resource-id      -> resource-id
    -- Pattern: (?:[^/:]*[/:])? optionally matches "resource-type/" or "resource-type:"
    --          (.*) captures everything after (the actual resource ID)
    else regexp_extract({{ arn_col }}, '^arn:[^:]*:[^:]*:[^:]*:[^:]*:(?:[^/:]*[/:])?(.*)$', 1)
end
{% endmacro %}

{% macro aws_transforms_cur_surrogate_key() %}
    to_base64(
        sha512(
            to_utf8(
                concat(
                    coalesce(payer_account_id, ''),
                    coalesce(try_cast(identity_line_item_id as varchar(100)), ''),
                    coalesce(try_cast(usage_date as varchar(100)), ''),
                    coalesce(linked_account_id, ''),
                    coalesce(charge_type, ''),
                    coalesce(product_code, ''),
                    coalesce(usage_type, ''),
                    coalesce(operation, ''),
                    coalesce(resource_id, ''),
                    coalesce(item_description, ''),
                    coalesce(pricing_term, ''),
                    coalesce(pricing_unit, ''),
                    coalesce(product_sku, '')
                )
            )
        )
    )
{% endmacro %}

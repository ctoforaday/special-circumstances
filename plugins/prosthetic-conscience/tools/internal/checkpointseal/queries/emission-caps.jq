# No session may exceed 4 emissions, and no render may exceed 200 bytes.
jq -s 'map(select(.nudge_enabled)) | {sessions: (group_by(.session_id) | length),
        worst_count: (map(.emissions_this_session) | max),
        worst_bytes: (map(.emission_bytes_max) | max)}' .claude/checkpoints/seals.jsonl

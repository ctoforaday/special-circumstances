# emitting sessions observed, against sessions that started at all
jq -s '{sealed_sessions: (group_by(.session_id) | length),
        emitting: ([.[] | select(.emissions_this_session > 0)] | group_by(.session_id) | length)}' \
   .claude/checkpoints/seals.jsonl

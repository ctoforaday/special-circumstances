jq -s '[.[] | select(.nudge_answered)] | group_by(.nudge_answered)
       | map({outcome: .[0].nudge_answered, n: length})' .claude/checkpoints/seals.jsonl

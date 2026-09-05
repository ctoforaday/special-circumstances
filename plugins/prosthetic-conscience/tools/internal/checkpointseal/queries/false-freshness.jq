# A note whose written_at ADVANCED between two seals while its body hash did NOT is
# claiming to be fresh without having changed. Reported as an error, not adjudicated.
jq -s 'sort_by(.at) | [ .[] | {at, written_at, body_sha} ]
       | . as $r | range(1; length) as $i
       | select($r[$i].written_at > $r[$i-1].written_at and $r[$i].body_sha == $r[$i-1].body_sha)
       | {at: $r[$i].at, written_at: $r[$i].written_at, note: "written_at advanced, body did not"}' \
   .claude/checkpoints/seals.jsonl

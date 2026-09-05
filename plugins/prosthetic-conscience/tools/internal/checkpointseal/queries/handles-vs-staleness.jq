# #506: is a note sealed with live background work staler than one sealed without?
# Rows where the payload could not tell us are EXCLUDED, not counted as zero.
jq -s '[.[] | select(.handles_measured)] | group_by(.live_handles > 0)
       | map({live_handles: .[0].live_handles > 0, n: length,
              median_age: ([.[] | select(.turns_measured) | .note_age_turns] | sort | .[length/2|floor]),
              turns_measured_n: ([.[] | select(.turns_measured)] | length)})' \
   .claude/checkpoints/seals.jsonl

# ...and how much of the baseline that excluded, reported rather than inferred from a small n:
jq -s '{total: length, measured: ([.[] | select(.handles_measured)] | length)}' \
   .claude/checkpoints/seals.jsonl

# #507: how stale is the note at a seat return, against the other triggers?
jq -s 'group_by(.seal_trigger)
       | map({trigger: .[0].seal_trigger, n: length,
              median_age: ([.[] | select(.turns_measured) | .note_age_turns] | sort | .[length/2|floor]),
              turns_measured_n: ([.[] | select(.turns_measured)] | length)})' \
   .claude/checkpoints/seals.jsonl

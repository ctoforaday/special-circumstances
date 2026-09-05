# EVERY array filters on turns_measured. A row with turns unmeasured omits the key, jq
# yields null, sort places nulls FIRST, and the median index moves — on the statistic
# that decides whether the nudge is removed.
jq -s '[.[] | select(.turns_measured)] |
       {before:           [.[]|select(.nudge_enabled==false)|.note_age_turns]|sort,
        after_all:        [.[]|select(.nudge_enabled==true )|.note_age_turns]|sort,
        after_rewritten:  [.[]|select(.nudge_enabled==true and .nudge_answered=="rewritten")|.note_age_turns]|sort,
        after_reaffirmed: [.[]|select(.nudge_enabled==true and .nudge_answered=="reaffirmed")|.note_age_turns]|sort}' \
   .claude/checkpoints/seals.jsonl
# and the count this excluded, reported rather than inferred:
jq -s '{total: length, turns_measured: ([.[]|select(.turns_measured)]|length)}' \
   .claude/checkpoints/seals.jsonl

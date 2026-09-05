# Each distribution filters on ITS OWN measured flag. Filtering once on turns and then
# building all three arrays lets rows with turns measured but growth absent contribute
# null, which sort places first and which moves the median index.
jq -s '{n: length,
        turns:   ([.[] | select(.turns_measured)  | .note_age_turns]    | sort),
        growth:  ([.[] | select(.growth_measured) | .note_growth_tokens]| sort),
        commits: ([.[] | select(.branch_measured) | .note_branch_commits] | sort),
        measured: {turns:  ([.[] | select(.turns_measured)]  | length),
                   growth: ([.[] | select(.growth_measured)] | length),
                   branch: ([.[] | select(.branch_measured)] | length)}}' \
   .claude/checkpoints/seals.jsonl

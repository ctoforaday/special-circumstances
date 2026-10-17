## menu

re-run a recorded computation and judge whether it proves the sentence it is attached to

## detail

YOU NAME THE PROOF BY ITS ANCHOR; THE TOOL RE-RUNS ITS SCRIPT. The anchor is the id inside the proof's token in the report. The tool executes the recorded script, compares its output byte for byte against what was recorded, and stores that result. You do not run it first.

THE RE-RUN'S CONDITIONS. The tool runs the stored script with your seat's environment, in the run directory, with the network open and a 60-second limit, and compares what it prints, its error output included; every call the script makes to this tool is refused and recorded as the script calling the tool (below).

What the tool CANNOT answer is whether the script establishes the sentence at that anchor — so it records whether the script REPRODUCED, and you judge whether it PROVES that sentence.

ONE LENS IS OFFERED EACH PROOF NOBODY HAS JUDGED: the computation lens, or the evidence lens where the cast seats no computation lens. No other lens's work list offers it.

ONE SCRIPT CAN BACK SEVERAL SENTENCES, and the two halves belong to different things. The re-run belongs to the SCRIPT: once it has reproduced, it stands for every proof on that script, and a judgement of another of its sentences re-runs nothing — the result's `rerun` reads `not_run` and names whose re-run it stands on. Your verdict belongs to the ONE SENTENCE you named: a script that computes one leg of a sentence's claim and not another's is sound for the first and unsound for the second, and each anchor keeps its own verdict.

READ FIRST, THEN CALL THIS ONCE. This call re-runs the proof where it has to AND records your verdict in the same act: the verdict and its reason are part of it, and no form of it re-runs without recording. So the reading comes before the call. `show evidence`, read AT THE PROOF'S OWN ANCHOR, prints the script this call executes and the output blue recorded — its page says how. Read it, decide what it computes and whether that is what the sentence at its anchor says, then call this with that verdict and your reason. A verdict recorded in order to see the re-run is your judgement on the record from that moment, where every other seat reads it.

YOUR LATEST VERDICT ON A SENTENCE IS YOURS; ANOTHER LENS'S DOES NOT REPLACE IT. An `unsound` you record stands on that anchor beside any other lens's `sound` until you judge the sentence again or blue edits it — an edit reopens the anchor, and a judgement from a sitting opened after the edit is owed.

WHEN THE RE-RUN DOES NOT REPRODUCE. The result's `reproduced` — `result.reproduced` in the JSON form — is true or false on every call: whether the script's re-run printed the recorded output byte for byte. Beside it are `rerun`, what this call did with the script, `got`, what it printed, and `recorded`. The envelope's `ok` says only that the call was accepted. On a proof recorded `reproducible`, a `sound` verdict over a re-run that did not reproduce is REFUSED and nothing is recorded: your verdict and reason were written before that output existed. The refusal prints both outputs; read them and call again with what the run did — `unsound` where the script cannot be run to its recorded output by the seat auditing it. A proof recorded `observed` or `not_portable` is one whose output moves, and your verdict on it is recorded either way.

A SCRIPT THAT CALLS THIS TOOL CANNOT BE CHECKED. The tool refuses every call a script makes to it while a checker runs the script, and records that it did: the result's `rerun` reads `calls_the_tool` and lists the calls, whatever the script printed — a refusal piped into a number prints a number. A `sound` verdict on it is refused; it is `unsound`, with a reason that says the script reads the debate's tool rather than computing its claim.

## menu

re-run a recorded computation and judge whether it proves the sentence it is attached to

## detail

THIS COMMAND RE-RUNS THE SCRIPT ITSELF. You do not run it first: name the proof and the tool executes the recorded script, compares its output byte for byte against what was recorded, and stores that result.

What the tool CANNOT answer is whether the script establishes the claim it is anchored to — so it records whether it REPRODUCED, and you judge whether it PROVES.

READ FIRST, THEN CALL THIS ONCE. This call re-runs the proof AND records your verdict in the same act: the verdict and its reason are part of it, and no form of it re-runs without recording. So the reading comes before the call. `show evidence`, read AT THE PROOF'S OWN ANCHOR, prints the script this call executes and the output blue recorded — its page says how, and the proof's row in the bare `show evidence` carries that anchor beside the sha256 this call takes. Read it, decide what it computes and whether that is the claim at its anchor, then call this with that verdict and your reason. A verdict recorded in order to see the re-run is your judgement on the record from that moment, where every other seat reads it.

WHEN THE RE-RUN DOES NOT REPRODUCE. The result's `reproduced` — `result.reproduced` in the JSON form — is true or false on every call: whether the script, run now, printed the recorded output byte for byte. Beside it are `got`, what it printed, and `recorded`. The envelope's `ok` says only that the call was accepted. On a proof recorded `reproducible`, a `sound` verdict over a re-run that did not reproduce is REFUSED and nothing is recorded: your verdict and reason were written before that output existed. The refusal prints both outputs; read them and call again with what the run did — `unsound` where the script cannot be run to its recorded output by the seat auditing it. A proof recorded `observed` or `not_portable` is one whose output moves, and your verdict on it is recorded either way.

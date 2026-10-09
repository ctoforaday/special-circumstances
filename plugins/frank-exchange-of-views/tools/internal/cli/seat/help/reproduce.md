## menu

re-run a recorded computation and judge whether it proves the sentence it is attached to

## detail

THIS COMMAND RE-RUNS THE SCRIPT ITSELF. You do not run it first: name the proof and the tool executes the recorded script, compares its output byte for byte against what was recorded, and stores that result. Running it by hand and then calling this to write down what you saw does the same work twice.

What the tool CANNOT answer is whether the script establishes the claim it is anchored to — so it records whether it REPRODUCED, and you judge whether it PROVES. Those are two questions and only the first is mechanical: `print("7 is prime")` reproduces perfectly forever. That is why the soundness verdict is yours to give and why you must READ the script to give it.

READ FIRST, THEN CALL THIS ONCE. This call re-runs the proof AND records your verdict in the same act: the verdict and its reason are part of it, and no form of it re-runs without recording. So the reading comes before the call. The script is a file in the proof store — `proofs/<sha256>/script.<ext>` under the run directory, beside `output`, the output blue recorded, where <sha256> is the proof's own. The record holds the proof's sha256 and not its script, so no projection shows it: this is the one read that is a file. Read it, decide what it computes and whether that is the claim at its anchor, then call this with that verdict and your reason. A verdict recorded in order to see the re-run is your judgement on the record from that moment, where every other seat reads it.

## menu

re-run a recorded computation and judge whether it proves the sentence it is attached to

## detail

THIS COMMAND RE-RUNS THE SCRIPT ITSELF. You do not run it first: name the proof and the tool executes the recorded script, compares its output byte for byte against what was recorded, and stores that result. Running it by hand and then calling this to write down what you saw does the same work twice.

What the tool CANNOT answer is whether the script establishes the claim it is anchored to — so it records whether it REPRODUCED, and you judge whether it PROVES. Those are two questions and only the first is mechanical: `print("7 is prime")` reproduces perfectly forever. That is why the soundness verdict is yours to give and why you must READ the script to give it.

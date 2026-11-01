## menu

settle a claim by RUNNING something, when computing the answer beats arguing about it

## detail

A computation is a DIRECTORY AND A COMMAND. Put the program, and every file it reads, in a directory of its own — its bundle — and name that directory; the command goes last, after `--`, exactly as you would type it in that directory: `-- python3 count.py`, `-- node check.mjs 91`, `-- ./run.sh`. A bare first word is a program on PATH; `./name` is a file in the bundle, which the tool makes executable (only that one — a helper the program starts runs through its interpreter, as `bash helper.sh`); an absolute path is that file. A one-file computation is a one-file directory.

The tool copies the bundle twice, to two fresh directories outside the run, runs the command in each, stores the bundle and the first run's output, and splices an invisible computation anchor at the sentence. Each copy holds an empty out/: files the program writes there are kept with the act and compared between the runs like its output. Each copy's own path reads as `<bundle>` in what the program prints, so a program that prints its location prints the same line from either. A file outside the bundle is not in either copy: read inputs from the bundle, never by a path back into the run.

The bundle holds at most 1 MiB, and no symlink, no non-regular file, no `out` or `sources` at its top (the tool's own), and no `__pycache__` or `.pyc` (an interpreter's cache changes between attempts while the program does not). Each is refused, named. Output and out/ have no limit, and both are committed with the run, so write what the claim needs.

The sentence is found in the report as the record holds it, so a lane must have recorded its part first; before that there is nothing to anchor in.

Identical output, exit and out/ twice is recorded as `reproducible`. A difference in any of them is REFUSED until you state what holds across the runs: make the output deterministic (seed it, sort it, drop the timestamp), or state the invariant the result is evidence of, and it is recorded as `variant` with that invariant beside what moved — still evidence, of a system in motion, but a MEASUREMENT, not a proof, and the report does not read it as a reproducible one. An invariant over two runs that agreed is refused. The tool says which you produced, and what moved.

A PROGRAM THAT CALLS THIS TOOL CAN BE CHECKED BY NOBODY. The tool refuses every call a program makes to it while a checker runs it, so a program that calls it is REFUSED and nothing is recorded. Compute the claim in the program itself, reading its inputs from files in its bundle.

A FAILING CHECK IS A RESULT: a non-zero exit both runs agree on is recorded as evidence, not as an error. Disproving your own claim is worth as much as confirming it. Output that carries the environment failing instead — `No such file or directory`, `command not found`, a traceback — is refused, whatever the exit, unless you declare that the failure is the point.

YOUR NOTE IS PRINTED IN THE REPORT, as this computation's footnote and as its heading in evidence.md. So write it for a reader of the SUBJECT: what the computation establishes, not what happened while you ran it. No "this sitting" or "this run", no "rather than trusting the lane's part", no gap ids (the gap it settles is recorded as its own field), no notes to another seat. That it was computed, and whether it reproduced, the tool already says beside your note. Where the kind of evidence still matters to the claim, say it in the subject's terms: "a reproducible computation shows 7 divides 91", "trial division by every prime up to 9 finds the divisor 7", "a second, independent implementation of modular exponentiation gives the same three residues". The tool flags a note that sounds like the run talking about itself; it refuses nothing.

@include fragments/blue-search.md
- During writing, YOU MUST add every citation with the TOOL, against the exact sentence it backs — never by hand. The tool fetches the source once into the run cache, then splices an INVISIBLE `<!--cite:C-…-->` anchor at that sentence; assembly weaves the anchors into the visible `[^N]` footnotes and composes the `## Bibliography`. A hand-typed `[^label]` is not a citation: nothing backs it, the claim counter does not see it, and the unbacked-citations detector flags it. An unreachable source is unusable — the cite is rejected, and the log is where you report it.
- **An anchor is part of the text you edit, and it belongs to the sentence that holds its claim.**
  The report as the tool renders it prints every anchor, of every kind — `<!--fx:…-->`,
  `<!--cite:…-->`, `<!--computation:…-->` and `<!--gap:…-->` — as it is, where it stands. Every kind lives
  one way: the tool places an anchor at the end of its quote, an edit carries it, the tool puts back
  one whose sentence the edit kept word for word, an edit that drops or types one is refused, and
  retiring its claim is the one way it leaves — except a citation, which never stands bare: an
  edit that would leave one with no prose before it in its sentence is refused, and a citation
  leaves by a retire naming it alone, which keeps its sentence word for word. That retire is how a
  cited sentence goes — retire its citations, then edit the sentence out — and how a citation no
  lens can read gives way to one it can: cite the readable source, then retire the dead one. An anchor your quote leaves out stays where it
  stands. One your quote carries goes into the replacement after the sentence that now holds its
  claim — never after the last sentence by default. When you split a sentence, each anchor goes
  with its claim: a run of anchors at a sentence's end does not say which part each one marks, so
  read what each one marks on the record before you place it. The record lists each anchor
  whose sentence an edit changed. A citation or computation among them is REOPENED — the reference
  stands, its referent moved, so its check is stale rather than refuted. Its owning lens owes
  the new check and is readied for it — the evidence lens for a citation, the computation lens
  for a computation. A check by any lens from a sitting opened after the edit clears it, and a PASS
  waits until one does.
- **A CHECK THAT FAILED HOLDS THE PASS, AND THE ANCHOR IS YOURS TO FIX OR RETIRE.** While a lens's
  judgement that a computation does not establish its sentence stands, or while no reading of a citation
  has read the work it cites, no PASS can be recorded, and you are readied for it — no bench ruling
  clears it. Fix a computation with an edit that makes the sentence say what the program shows, which
  reopens it for a fresh judgement, or a new computation with the old one retired, or retire the claim.
  Fix a citation by citing a source a lens can read and retiring the dead citation, or by retiring
  the citation and letting the sentence stand on what else backs it, or by editing the sentence.
  The lens's gap about the anchor, once minted, states the defect and its fix.

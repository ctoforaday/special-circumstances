package hookgate

import "errors"

// ErrDeliveryHeld is a lens's delivery held by another of its calls past the writer's wait: that call
// is handing the item, so this one hands nothing, and nothing failed. The writer exits
// ExitDeliveryHeld on it, and the hook reads that exit back as this error.
var ErrDeliveryHeld = errors.New("the lens's delivery is held by another call")

// ExitDeliveryHeld is the writer's exit status for ErrDeliveryHeld: the one status the hook reads as
// "held", distinct from 1 (any other failure) and 2 (a flag error or a panic).
const ExitDeliveryHeld = 3

// ErrDeliveryStuck is a seat's delivery of report text that cannot advance: the writer has logged it
// as the tool's defect, exits ExitDeliveryStuck, and the hook reads that exit back as this error and
// ends the seat's turn.
var ErrDeliveryStuck = errors.New("the delivery of report text cannot advance")

// ExitDeliveryStuck is the writer's exit status for ErrDeliveryStuck, distinct from ExitDeliveryHeld,
// 1 and 2.
const ExitDeliveryStuck = 4

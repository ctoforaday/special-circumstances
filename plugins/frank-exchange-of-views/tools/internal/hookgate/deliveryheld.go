package hookgate

import "errors"

// ErrDeliveryHeld is a lens's delivery held by another of its calls past the writer's wait: that call
// is handing the item, so this one hands nothing, and nothing failed. The writer exits
// ExitDeliveryHeld on it, and the hook reads that exit back as this error.
var ErrDeliveryHeld = errors.New("the lens's delivery is held by another call")

// ExitDeliveryHeld is the writer's exit status for ErrDeliveryHeld: the one status the hook reads as
// "held", distinct from 1 (any other failure) and 2 (a flag error or a panic).
const ExitDeliveryHeld = 3

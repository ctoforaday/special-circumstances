package recordsql

// RerunLockSwitch runs the lock switch init runs and keeps its answer as init does, so a test can
// throw it after this process's locking mode is already fixed — the order init exists to prevent.
func RerunLockSwitch() { lockingErr = useOFDLocks() }

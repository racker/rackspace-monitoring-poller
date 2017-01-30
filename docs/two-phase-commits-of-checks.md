
## Various facts that support two-phase commits of checks, as provided by AEPs

`EleSession` runs frame decoding and is the immediate consumer of poller-AEP protocol messages, such as
`poller.prepare` and related.

A `Session`, upon receiving a `poller.prepare` message is responsible for creating a `ChecksPreparation`.

As a `poller.prepare.block` messages are received, the `ChecksPreparation` is updated.

The `ChecksPreparation` is responsible for verification of itself as part of 

Upon receiving a `poller.commit` the `Session` will hand off the `ChecksPreparation` to a `ChecksReconciler`.

`EleConnectionStream` implements `ChecksReconciler` but delegates to its `Scheduler` instances.

`Scheduler` embeds the interface `ChecksReconciler`, which is what bridges `EleConnectionStream` delegation 
of `CheckReconciler`

When `EleConnectionStream` creates a `Connection` it provides itself as the `ChecksReconciler`.

When `EleConnection` creates a `Session` is provides the given `CheckReconciler` which effectively
keeps it out of the business of checks reconciliation.

Each `EleScheduler` knows the zone it handles and as such can potentially filter the prepared checks during
reconciliation.

`EleScheduler` avoids concurrent modifications of scheduled checks by consuming from an internal chan 
of `ChecksPreparation` and only the routine waiting on that channel modifies the scheduled set.
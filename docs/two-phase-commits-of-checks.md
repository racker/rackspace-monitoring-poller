
## Development notes for handling two-phase commits of checks

`EleSession` runs frame decoding and is the immediate consumer of poller-endpoint protocol messages, 
such as `poller.prepare` and related.

A `Session`, upon receiving a `poller.prepare` message is responsible for creating a `ChecksPreparation`.

As a `poller.prepare.block` messages are received, the `ChecksPreparation` is updated.

The `ChecksPreparation` is responsible for verification of itself as part of `poller.prepare.end`.

Upon receiving a `poller.commit` the `Session` will hand off the `ChecksPreparation` to a `ChecksReconciler`.

`EleConnectionStream` implements `ChecksReconciler` but delegates to its `Scheduler` instances. Since a
ConnectionStream spans endpoints and potentially monitoring zones, this leaves open the possiblity to route
subsets of prepared checks to the appropriate scheduler(s).

`Scheduler` embeds the interface `ChecksReconciler`. That interface is what bridges `EleConnectionStream` delegation 
of `CheckReconciler` and enables the full abstraction from the perspective of `EleSession`.

When `EleConnectionStream` creates a `Connection` it provides itself as the `ChecksReconciler`.

When `EleConnection` creates a `Session` it provides the `CheckReconciler` given to it, which ensures minimal
involvement of `Connection` in the checks preparation and reconciliation process.

Each `EleScheduler` knows the zone it handles and as such can potentially filter the prepared checks during
reconciliation.

`EleScheduler` avoids concurrent modifications of scheduled checks by consuming from an internal channel 
of `ChecksPreparation` and only that go modifies the scheduled set. That same go routine is the only
context in which a `CheckScheduler` is used to initiate each check's polling cycle.
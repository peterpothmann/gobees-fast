// Package event holds the custom Ark event types of the model. The model emits them where the output needs a value
// that the state does not hold, e.g. the number of foraging trips. No system observes them: they are for the
// observers of the output (package obs) only. Without a registered observer, emitting an event costs one check.
//
// Deaths and new colonies need no custom event: the observers use Ark's OnRemoveEntity, which is emitted before the
// entity is removed, and OnCreateEntity.
package event

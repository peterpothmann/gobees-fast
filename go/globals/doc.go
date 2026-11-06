// Package globals holds the state of the run that is not part of an entity: one value per world, as Ark
// resources, which the systems change during the run (e.g. the day, the calendar, the activity queue, the
// landscape, the nest sites). The read-only values (parameters, derived values, species) are in package
// params.
//
// Naming: Go names without units. The unit stands in the doc comment of the field, and the comment names the
// NetLogo global where the name differs from it (table in docs/names.md).
package globals

// Package globals holds the state of the run that is not part of an entity: one value per world, as Ark
// resources, which the systems change during the run (e.g. the day, the calendar, the activity queue, the
// landscape, the nest sites). The read-only values (parameters, derived values, species) are in package
// params (step F14).
//
// Naming: fields keep the NetLogo names of the globals, as in package params.
package globals

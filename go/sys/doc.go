// Package sys holds the systems of the model.
//
// Four systems run a day in the order of the NetLogo procedure Go (schedule.go, step F14): Setup (the world at
// the start), Morning (the day counter and the morning routine), Day (the activities of the bees) and Evening
// (feeding, development, mortality, badgers, end of the run). Each system calls its parts in order. A part is one
// NetLogo procedure (e.g. UpdateColonies for UpdateColoniesProc) with Initialize and Update like an Ark system, but
// only its group adds and calls it. Helpers that several parts use (beeDeath, colonyFounder, eggLayer, forager,
// speciesTable) are small structs held by each part.
//
// The file names start with the group of their part: setup_, morning_, day_, evening_ (e.g.
// morning_update_colonies.go). Helpers that only one group uses carry its prefix (day_foraging.go), helpers of
// several groups the prefix shared_. schedule.go holds the groups, validation_*_test.go the tests that chain parts of
// several groups.
//
// Every part and function names the codes of the NetLogo annotations it translates, e.g. [UCOL-02].
package sys

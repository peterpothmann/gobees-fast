// Package check holds the checks of the model state that the tests run at the end of every day. They replace the
// assertions of CheckNumbersProc [CHK], which are not part of the model since step F4.
//
// Not kept: [CHK-01] (initial + produced - died = present bees) and the egg accounting of [CHK-04] (eggs laid =
// adults emerged + brood deaths by cause). Both checked only NetLogo's own counters, which the model no longer has.
package check

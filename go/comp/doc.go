// Package comp contains the ECS components.
//
// Naming: Go names without units (step F15b). The unit stands in the doc comment of the field, and the comment
// names the NetLogo variable where the name differs from it (table in docs/names.md).
// References to other agents (NetLogo: who numbers) are ecs.Entity. Agents keep no who number (step F12). NetLogo -1 ("nobody") is the zero entity.
package comp

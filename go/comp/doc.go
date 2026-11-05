// Package comp contains the ECS components.
//
// Naming: fields keep the NetLogo variable names with the first letter in upper case, so they can be found
// in the NetLogo code. "?" is dropped, "/" becomes "_per_" (as in package params).
// References to other agents (NetLogo: who numbers) are ecs.Entity. Agents keep no who number (step F12). NetLogo -1 ("nobody") is the zero entity.
package comp

//go:build ignore

package optimizer

import "github.com/EladB1/The/internal/irgen"

/*
NOTE: Optimizer will not be part of MVP
Just filling out a basic outline for now
*/

func Optimize(prog *irgen.Program) {}

// When an unreachable statement is found, removing anything including and after it
func removeAllUnreachable() {}

// Fold constant expressions to their result; i.e.: 1 + 6 <= 3 * 2 becomes 6 <= 6 becomes true
func constantFolding() {}

// a plain expression that isn't used by anything can be removed
func removeUnusedExpression() {}

func removeUnusedVariable() {}

func removeUnusedFunction() {}

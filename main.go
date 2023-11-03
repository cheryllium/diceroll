package main

import (
  "fmt"
)

func main() {
  // Simple expression
  input := "5 + 2d10 * 3"
  fmt.Printf("Expression %s = %d\n", input, ParseExpression(input))

  // Macro with variable substitution
  macro := "5 + A * (2d8 + B)"
  vars := []string{"1", "2d20"}
  fmt.Printf("Macro %s = %d\n", macro, ParseMacro(macro, vars))
}

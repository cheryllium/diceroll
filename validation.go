package main

import (
  "errors"
)

/* Checks if a macro name is valid.
 */
func ValidateMacroName(name string) error {
  if len(name) < 1 || len(name) > 128 {
    return errors.New("Macro name must be between 1 and 128 characters long.")
  }

  return nil
}

/* Check if a macro expression is valid, by attempting
 * to parse it. The parser will throw an error if the
 * given expression is not valid.
 * 
 * This function is necessary because otherwise there will be
 * no call to parse the macro when creating it.
 * No ValidateExpression function is necessary for non-macro
 * expressions, because those are immediately parsed. 
 */
func ValidateMacro(expression string) error {
  inputs := make([]string, 26)
  for i := range inputs {
    inputs[i] = "1"
  }

  _, err := ParseMacro(expression, inputs)
  return err
}

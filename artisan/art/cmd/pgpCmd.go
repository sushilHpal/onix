package cmd

/*
  Onix Config Manager - Artisan
  Copyright (c) 2018-Present by www.gatblau.org
  Licensed under the Apache License, Version 2.0 at http://www.apache.org/licenses/LICENSE-2.0
  Contributors to this project, hereby assign copyright in this code to the project,
  to be licensed under the same terms as the rest of the code.
*/
import (
	"github.com/spf13/cobra"
)

// PGPCmd provides PGP management functions
type PGPCmd struct {
	cmd *cobra.Command
}

func NewPGPCmd() *PGPCmd {
	c := &PGPCmd{
		cmd: &cobra.Command{
			Use:   "pgp",
			Short: "provides PGP management functions",
			Long:  ``,
		},
	}
	return c
}

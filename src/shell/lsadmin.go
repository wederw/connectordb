/**
Copyright (c) 2016 The ConnectorDB Contributors
Licensed under the MIT license.
**/
package shell

/* Lists current admins

Copyright 2015 - The ConnectorDB Contributors; see AUTHORS for a list of authors.
All Rights Reserved
*/

import "fmt"

func init() {
	help := "Lists current admins."
	usage := `Usage: lsadmin`
	name := "lsadmin"

	main := func(shell *Shell, args []string) uint8 {

		users, err := shell.operator.ReadAllUsers()
		if shell.PrintError(err) {
			return 1
		}

		for _, usr := range users {

			if usr.Role == "admin" {
				fmt.Printf("%s\t%s\t%d\n", usr.Name, usr.Email, usr.UserID)
			}

		}

		return 0
	}

	registerShellCommand(help, usage, name, main)
}

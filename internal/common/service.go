package common

import "github.com/spf13/cobra"

type CmdParams struct {
	Name  string
	Flags map[string]any
}

type Service interface {
	Handler(cmd *cobra.Command, cmdParams *CmdParams, args []string, kwargs map[string]any) error
}

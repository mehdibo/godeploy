package cmd

import (
	"fmt"
	"github.com/mehdibo/go_deploy/pkg/auth"
	"github.com/mehdibo/go_deploy/pkg/db"
	"github.com/spf13/cobra"
	"gorm.io/gorm"
)

func NewNewUserCmd(orm **gorm.DB) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "new-user username",
		Short: "Create a new user and generate its token",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			rawToken, err := auth.GenerateToken()
			if err != nil {
				return err
			}
			user := db.User{
				Username:    args[0],
				HashedToken: auth.HashToken(rawToken),
				LastUsedAt:  nil,
				Role:        auth.RoleAdmin,
			}

			result := (*orm).Create(&user)
			if result.Error != nil {
				return result.Error
			}
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "User created successfully, here is the token: \n%s\n", rawToken)
			_, _ = fmt.Fprintln(cmd.OutOrStdout(), "Make sure you save your token someplace safe")

			return nil
		},
	}
	return cmd
}

var newUserCmd = NewNewUserCmd(&orm)

func init() {
	rootCmd.AddCommand(newUserCmd)
}

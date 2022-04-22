package cmd

import (
	"errors"
	"fmt"
	"github.com/mehdibo/godeploy/pkg/env"
	"github.com/melbahja/goph"
	"github.com/spf13/cobra"
	"golang.org/x/crypto/ssh"
	"net"
)

func NewGetFingerprintCmd() *cobra.Command {
	cmd := cobra.Command{
		Use:   "get-fingerprint username hostname",
		Short: "Get the server's SSH public key fingerprint",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			port, err := cmd.Flags().GetInt("port")
			if err != nil {
				return err
			}
			sshPrivKey := env.Get("SSH_PRIVATE_KEY")
			sshPassPhrase := env.GetDefault("SSH_PASSPHRASE", "")
			if sshPrivKey == "" {
				return errors.New("required SSH config is not set")
			}
			auth, err := goph.Key(sshPrivKey, sshPassPhrase)
			if err != nil {
				return err
			}
			client, err := goph.NewConn(&goph.Config{
				Auth:    auth,
				User:    args[0],
				Addr:    args[1],
				Port:    uint(port),
				Timeout: goph.DefaultTimeout,
				Callback: func(hostname string, remote net.Addr, key ssh.PublicKey) error {
					_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Server fingerprint:\n%s\n", ssh.FingerprintSHA256(key))
					return nil
				},
			})
			if err != nil {
				return err
			}
			return client.Close()
		},
	}
	cmd.Flags().IntP("port", "p", 22, "SSH port")
	return &cmd
}

var getFingerprintCmd = NewGetFingerprintCmd()

func init() {
	rootCmd.AddCommand(getFingerprintCmd)
}

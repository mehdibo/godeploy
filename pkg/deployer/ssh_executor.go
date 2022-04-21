package deployer

import (
	"errors"
	"github.com/mehdibo/go_deploy/pkg/db"
	"github.com/melbahja/goph"
	log "github.com/sirupsen/logrus"
	"golang.org/x/crypto/ssh"
	"net"
)

func (d *Deployer) executeSshTask(task *db.SshTask) error {
	auth, err := goph.Key(d.sshPrvKey, d.sshPrvKeyPass)
	if err != nil {
		log.Errorf("Couldn't load SSH private key: %s", err)
		return ErrUnrecoverable
	}
	client, err := goph.NewConn(&goph.Config{
		Auth:    auth,
		User:    task.Username,
		Addr:    task.Host,
		Port:    task.Port,
		Timeout: goph.DefaultTimeout,
		Callback: func(hostname string, remote net.Addr, key ssh.PublicKey) error {
			// hostFound: is host in known hosts file.
			// err: error if key not in known hosts file OR host in known hosts file but key changed!
			hostFound, err := goph.CheckKnownHost(hostname, remote, key, d.sshKnownHosts)

			// Host in known hosts but key mismatch!
			// Maybe because of MITM ATTACK! or the server changed the key
			if hostFound && err != nil {
				return err
			}

			// handshake because public key already exists.
			if hostFound && err == nil {
				return nil
			}

			// Verify if fingerprint match
			if task.ServerFingerprint != ssh.FingerprintSHA256(key) {
				return errors.New("ssh fingerprint mismatch")
			}

			// Add the new host to known hosts file.
			return goph.AddKnownHost(hostname, remote, key, d.sshKnownHosts)
		},
	})
	if err != nil {
		log.Errorf("Couldn't connect to SSH host: %s", err.Error())
		return ErrUnrecoverable
	}
	defer client.Close()
	out, err := client.Run(task.Command)
	if err != nil {
		log.Errorf("Command failed: %s", err.Error())
		return ErrUnrecoverable
	}
	log.Debugf("Command output: %s", out)
	return nil
}

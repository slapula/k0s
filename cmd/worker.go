package cmd

import (
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"os"
	"os/signal"
	"path"
	"syscall"

	"github.com/Mirantis/mke/pkg/assets"
	"github.com/Mirantis/mke/pkg/component"
	"github.com/Mirantis/mke/pkg/constant"
	"github.com/Mirantis/mke/pkg/util"
	"github.com/pkg/errors"
	"github.com/urfave/cli/v2"
)

// WorkerCommand ...
func WorkerCommand() *cli.Command {
	return &cli.Command{
		Name:   "worker",
		Usage:  "Run worker",
		Action: startWorker,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name: "join-token",
			},
		},
	}
}

func startWorker(ctx *cli.Context) error {
	err := assets.Stage(path.Join(constant.DataDir))
	if err != nil {
		return err
	}

	token := ctx.String("join-token")
	if token == "" && !util.FileExists("/var/lib/mke/kubelet.conf") {
		return fmt.Errorf("normal kubelet kubeconfig does not exist and no join-token given. dunno how to make kubelet auth to api")
	}

	// Dump join token into kubelet-bootstrap kubeconfig
	if token != "" {

		kubeconfig, err := base64.StdEncoding.DecodeString(token)
		if err != nil {
			return errors.Wrap(err, "joint-token does not seem to be proper token created by 'mke token create'")
		}
		err = ioutil.WriteFile(constant.KubeletBootstrapConfigPath, kubeconfig, 0600)
		if err != nil {
			return errors.Wrap(err, "joint-token does not seem to be proper token created by 'mke token create'")
		}
	}

	components := make(map[string]component.Component)

	components["containerd"] = component.ContainerD{}
	components["containerd"].Run()

	components["kubelet"] = component.Kubelet{}
	components["kubelet"].Run()

	// Wait for mke process termination
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c

	// Stop stuff does not really work yet
	// for _, comp := range components {
	// 	comp.Stop()
	// }

	return nil

}

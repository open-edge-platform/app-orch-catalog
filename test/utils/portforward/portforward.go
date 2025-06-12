package portforward

import (
	"fmt"
	"github.com/open-edge-platform/app-orch-catalog/test/utils/types"
	"os/exec"
	"time"
)

func KillportForwardToCatalog(cmd *exec.Cmd) error {
	fmt.Println("kill process that port-forwards network to app-orch-catalog")
	if cmd != nil && cmd.Process != nil {
		return cmd.Process.Kill()
	}
	return nil
}

func PortForwardToCatalog() (*exec.Cmd, error) {
	fmt.Println("port-forward to app-orch-catalog")

	cmd := exec.Command("kubectl", "port-forward", "-n", types.PortForwardServiceNamespace, types.PortForwardService,
		fmt.Sprintf("%s:%s", types.PortForwardLocalPort, types.PortForwardRemotePort),
		"--address", types.PortForwardAddress)
	err := cmd.Start()
	time.Sleep(5 * time.Second) // Give some time for port-forwarding to establish

	return cmd, err
}

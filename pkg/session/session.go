package session

import (
	"fmt"
	"log"
	"regexp"
	"time"

	expect "github.com/google/goexpect"
	"github.com/jpmondet/netmiGo/pkg/devices"
	"github.com/jpmondet/netmiGo/pkg/str"
	"golang.org/x/crypto/ssh"
	"google.golang.org/grpc/codes"
)

const (
	timeout = 10 * time.Second
)

var (
	// Regexes that will be used to 'expect'
	userRE1 = regexp.MustCompile("Username:")
	userRE2 = regexp.MustCompile("Login: ")
	passRE  = regexp.MustCompile("Password: ")
	// PromptRE : Prompt regex for legacy cli of net devices or 'default' servers
	promptRE = regexp.MustCompile("[a-zA-Z0-9.@_-]+[>$#]")
	//promptDisRE = regexp.MustCompile("[a-zA-Z0-9.-_]*>")
)

// BaseConnection : Struct keeping important connection infos
type BaseConnection struct {
	SSHClt  *ssh.Client
	Exp     *expect.GExpect
	Dev     devices.Device
	Timeout time.Duration
}

// ConnectionHandler : Handles the first connection to a device
// to authenticate and get the prompt.
func ConnectionHandler(dev devices.Device) (*BaseConnection, error) {

	var bc BaseConnection

	bc.Dev = dev
	bc.Timeout = timeout
	bc.Dev.PromptRE = promptRE

	fmt.Println("SSHing to the device...")
	sshCon(&bc)

	fmt.Println("Waiting spawn...")
	spawnExp(&bc)

	fmt.Println("Trying to find prompt")
	prompt := findPrompt(dev, bc.Exp)
	fmt.Println("Prompt found : ", prompt)

	// TODO: Must handle disable/enable

	return &bc, nil

}

func findPrompt(dev devices.Device, e *expect.GExpect) string {
	batchRes, err := e.ExpectBatch([]expect.Batcher{
		&expect.BCas{C: []expect.Caser{
			&expect.Case{R: promptRE, T: expect.OK()},
			&expect.Case{R: userRE1, S: dev.User,
				T: expect.Continue(expect.NewStatus(codes.PermissionDenied, "wrong username")), Rt: 3},
			&expect.Case{R: userRE2, S: dev.User,
				T: expect.Continue(expect.NewStatus(codes.PermissionDenied, "wrong username")), Rt: 3},
			&expect.Case{R: passRE, S: dev.Pass1, T: expect.Next(), Rt: 1},
			&expect.Case{R: passRE, S: dev.Pass2,
				T: expect.Continue(expect.NewStatus(codes.PermissionDenied, "wrong password")), Rt: 1},
		}},
	}, timeout)
	if err != nil {
		log.Fatalf("expect.GExpect in findPrompt failed: %v", err)
	}

	return batchRes[0].Match[0]

}

func sshCon(bc *BaseConnection) {
	addrPort := str.Join(bc.Dev.Addr, ":", bc.Dev.Port)
	sshClt, err := ssh.Dial("tcp", addrPort, &ssh.ClientConfig{
		User:            bc.Dev.User,
		Auth:            []ssh.AuthMethod{ssh.Password(bc.Dev.Pass1)},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	})
	if err != nil {
		log.Fatalf("ssh.Dial(%q) failed: %v", addrPort, err)
	}
	bc.SSHClt = sshClt

}

func spawnExp(bc *BaseConnection) {
	e, _, err := expect.SpawnSSH(bc.SSHClt, timeout)
	if err != nil {
		log.Fatal(err)
	}
	bc.Exp = e

}

// ResetExpSession : Close and respawn the Expect session.
// Useful when we want to clear the buffer with garbage in the results
func ResetExpSession(bc *BaseConnection) {
	bc.Exp.Close()

	spawnExp(bc)
	findPrompt(bc.Dev, bc.Exp)

}

package push

import (
	"bufio"
	"fmt"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/jpmondet/netmiGo/pkg/session"
	"github.com/jpmondet/netmiGo/pkg/str"

	tfsma "github.com/TobiEiss/go-textfsm/pkg/ast"
	tfsmp "github.com/TobiEiss/go-textfsm/pkg/process"
	tfsmr "github.com/TobiEiss/go-textfsm/pkg/reader"
)

var (
	_, b, _, _ = runtime.Caller(0)
	basepath   = filepath.Dir(b)
)

func disablePagination(bc *session.BaseConnection) bool {

	fmt.Println("Disabling pagination...")
	bc.Exp.Send("terminal length 0 \n")
	result, _, _ := bc.Exp.Expect(bc.Dev.PromptRE, bc.Timeout)

	if strings.Contains(strings.ToLower(result), "unknown") {
		fmt.Println("Can't disable pagination...")
		return false
	}
	return true
}

// SendCommands : Send multiple commands passed in parameter to the device and return
// a string result.
func SendCommands(bc *session.BaseConnection, cmdToSend ...string) (string, error) {

	var result string

	disabled := disablePagination(bc)
	fmt.Println("Waiting cmd...")
	for _, cmd := range cmdToSend {
		if !disabled {
			session.ResetExpSession(bc)
			cmd = str.Join(cmd, " | no-more")
		}
		bc.Exp.Send(cmd + "\n")
		res, _, _ := bc.Exp.Expect(bc.Dev.PromptRE, bc.Timeout)
		result = str.Join(result, "\n", res, "\n")
	}

	bc.Exp.Send("exit\n")
	bc.SSHClt.Close()

	return result, nil

}

func readStringLineByLine(str string, out chan string) {

	// read with scanner
	scanner := bufio.NewScanner(strings.NewReader(str))
	//scanner.Split(bufio.ScanLines)

	// read line by line
	for scanner.Scan() {
		out <- scanner.Text()
	}
	close(out)
}

// SendCommandFSM : Send command passed in parameter to the device and return
// a structured result (via textfsm).
// DO NOT USE, STILL IN TESTING SINCE GO-TEXTFSM LIBRARY IS NOT COMPLETE
func SendCommandFSM(bc *session.BaseConnection, cmd string) (string, error) {

	//var result string

	disabled := disablePagination(bc)
	fmt.Println("Waiting cmd...")
	if !disabled {
		session.ResetExpSession(bc)
		cmd = str.Join(cmd, " | no-more")
	}
	bc.Exp.Send(cmd + "\n")
	res, _, _ := bc.Exp.Expect(bc.Dev.PromptRE, bc.Timeout)
	//fmt.Println(res)

	// read template
	// TODO : Use index to find which template is needed
	cmdSplit := strings.Split(cmd, " ")
	cmdUnderscored := strings.Join(cmdSplit, "_")
	tmplate := str.Join(bc.Dev.DeviceType, "_", cmdUnderscored, ".template")
	filepath := basepath + "/templates/" + tmplate

	// Prepare template reading
	tmplCh := make(chan string)
	go tfsmr.ReadLineByLine(filepath, tmplCh)

	ast, err := tfsma.CreateAST(tmplCh)
	if err != nil {
		return "", err
	}
	fmt.Println("AST is : ", ast)

	process, err := tfsmp.NewProcess(ast)
	if err != nil {
		return "", err
	}
	fmt.Println("PROCESS is : ", process)

	resCh := make(chan string)
	go readStringLineByLine(res, resCh)

	var record map[string]map[string]*tfsmp.Column
	record = process.Do(resCh)

	fmt.Println("record")
	for k, v := range record {
		fmt.Println("KV")

		fmt.Println(k, v)

		for l, w := range v {
			fmt.Println("KV")
			fmt.Println(l, w)
		}

	}

	result := fmt.Sprintln(record)

	bc.Exp.Send("exit\n")
	bc.SSHClt.Close()

	return result, nil

}

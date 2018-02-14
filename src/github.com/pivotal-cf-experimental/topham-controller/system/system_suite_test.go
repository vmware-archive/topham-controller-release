package system_test

import (
	"os"
	"os/exec"
	"testing"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
)

func TestSystem(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "System Suite")
}

var cliPath string
var serverPath string
var err error

var _ = BeforeSuite(func() {
	cliPath = os.Getenv("CLI_BINARY_PATH")

	serverPath, err = gexec.Build("github.com/pivotal-cf-experimental/topham-controller")
	Expect(err).NotTo(HaveOccurred())

	command := exec.Command(serverPath)
	_, err = gexec.Start(command, GinkgoWriter, GinkgoWriter)
	Expect(err).NotTo(HaveOccurred())

	time.Sleep(time.Second)
})

var _ = AfterSuite(func() {
	gexec.KillAndWait()
})

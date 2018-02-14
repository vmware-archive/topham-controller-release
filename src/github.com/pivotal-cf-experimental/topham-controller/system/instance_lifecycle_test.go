package system_test

import (
	"encoding/json"
	"os/exec"

	"github.com/google/uuid"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"
	"github.com/pivotal-cf-experimental/topham-controller/store"
)

var _ = It("Instance lifecycle", func() {
	var testInstanceID = uuid.New().String()
	var bindings []store.AnnotatedBinding

	By("lists the catalog", func() {
		session := runCliCommand("catalog")
		Expect(session.Out).To(gbytes.Say("overview-broker"))
		Expect(session.Out).To(gbytes.Say("async"))
		Expect(session.Out).To(gbytes.Say("complex"))
		Expect(session.Out).To(gbytes.Say("simple"))
	})

	By("can provision an instance", func() {
		session := runCliCommand("provision", "-s", "overview-broker", "-p", "simple", "-i", testInstanceID)
		Expect(session.Out).To(gbytes.Say("provision:   done"))

		session = runCliCommand("services")
		Expect(session.Out).To(gbytes.Say("overview-broker"))
	})

	By("can create bindings for an instance", func() {
		session := runCliCommand("bind", "-i", testInstanceID)
		Expect(session.Out).To(gbytes.Say("Created binding"))
		session = runCliCommand("bind", "-i", testInstanceID)
		Expect(session.Out).To(gbytes.Say("Created binding"))
	})

	By("can get bindings for an instance", func() {
		session := runCliCommand("credentials", "-i", testInstanceID)

		err := json.Unmarshal(session.Out.Contents(), &bindings)
		Expect(err).NotTo(HaveOccurred())

		Expect(len(bindings)).To(Equal(2))
		Expect(bindings[0].ID).NotTo(BeEmpty())
	})

	By("can delete bindings for an instance", func() {
		session := runCliCommand("unbind", "-i", testInstanceID, "-b", bindings[0].ID)
		Expect(session.Out).To(gbytes.Say("Success"))

		session = runCliCommand("credentials", "-i", testInstanceID)

		err := json.Unmarshal(session.Out.Contents(), &bindings)
		Expect(err).NotTo(HaveOccurred())

		Expect(len(bindings)).To(Equal(1))
	})

	By("can delete an instance", func() {
		session := runCliCommand("deprovision", "-i", testInstanceID)
		Expect(session.Out).To(gbytes.Say("deprovision: done"))
	})
})

func runCliCommand(args ...string) *gexec.Session {
	command := exec.Command(cliPath, args...)
	session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
	Expect(err).NotTo(HaveOccurred())
	Eventually(session).Should(gexec.Exit(0))
	return session
}

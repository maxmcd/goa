/*
Package example_test validates that the code generated by goagen by the "main" and "app"
generators from the cellar example design package produce valid Go code that compiles and runs.
*/
package examples_test

import (
	"bytes"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("example cellar", func() {
	var tempdir string

	var files = []string{
		"app",
		"app/contexts.go",
		"app/controllers.go",
		"app/hrefs.go",
		"app/media_types.go",
		"app/user_types.go",
		"main.go",
		"account.go",
		"bottle.go",
		"client",
		"client/cellar-cli",
		"client/cellar-cli/main.go",
		"client/cellar-cli/commands.go",
		"client/client.go",
		"client/account.go",
		"client/bottle.go",
		"swagger",
		"swagger/swagger.json",
		"swagger/swagger.go",
		"",
	}

	BeforeEach(func() {
		var err error
		gopath := strings.Split(os.Getenv("GOPATH"), ":")[0]
		tempdir, err = ioutil.TempDir(filepath.Join(gopath, "src"), "cellar-test-tmpdir-")
		Ω(err).ShouldNot(HaveOccurred())
	})

	JustBeforeEach(func() {
		cmd := exec.Command("goagen", "bootstrap", "-d", "github.com/raphael/goa/examples/cellar/design")
		cmd.Dir = tempdir
		out, err := cmd.CombinedOutput()
		Ω(err).ShouldNot(HaveOccurred())
		Ω(string(out)).Should(Equal(strings.Join(files, "\n")))
	})

	It("goagen generated valid Go code", func() {
		cmd := exec.Command("go", "build", "-o", "cellar")
		cmd.Dir = tempdir
		_, err := cmd.CombinedOutput()
		Ω(err).ShouldNot(HaveOccurred())
		cmd = exec.Command("./cellar")
		cmd.Dir = tempdir
		b := &bytes.Buffer{}
		cmd.Stdout = b
		err = cmd.Start()
		Ω(err).ShouldNot(HaveOccurred())
		done := make(chan error, 1)
		go func() {
			done <- cmd.Wait()
		}()
		select {
		case <-time.After(100 * time.Millisecond):
			cmd.Process.Kill()
			<-done
		case err := <-done:
			Ω(err).ShouldNot(HaveOccurred())
		}
		Ω(err).ShouldNot(HaveOccurred())
		Ω(b.String()).Should(ContainSubstring("file=swagger/swagger.json"))
	})

	AfterEach(func() {
		os.RemoveAll(tempdir)
	})
})

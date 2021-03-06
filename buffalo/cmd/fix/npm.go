package fix

import (
	"bytes"
	"fmt"
	"html/template"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/gobuffalo/buffalo/genny/assets/webpack"
)

// PackageJSONCheck will compare the current default Buffalo
// package.json against the applications package.json. If they are
// different you have the option to overwrite the existing package.json
// file with the new one.
func PackageJSONCheck(r *Runner) error {
	fmt.Println("~~~ Checking package.json ~~~")

	if !r.App.WithWebpack {
		return nil
	}

	box := webpack.Templates

	f, err := box.FindString("package.json.tmpl")
	if err != nil {
		return err
	}

	tmpl, err := template.New("package.json").Parse(f)
	if err != nil {
		return err
	}

	bb := &bytes.Buffer{}
	err = tmpl.Execute(bb, map[string]interface{}{
		"opts": &webpack.Options{
			App: r.App,
		},
	})
	if err != nil {
		return err
	}

	b, err := ioutil.ReadFile("package.json")
	if err != nil {
		return err
	}

	if string(b) == bb.String() {
		return nil
	}

	if !ask("Your package.json file is different from the latest Buffalo template.\nWould you like to REPLACE yours with the latest template?") {
		fmt.Println("\tskipping package.json")
		return nil
	}

	pf, err := os.Create("package.json")
	if err != nil {
		return err
	}
	_, err = pf.Write(bb.Bytes())
	if err != nil {
		return err
	}
	err = pf.Close()
	if err != nil {
		return err
	}

	os.RemoveAll(filepath.Join(r.App.Root, "node_modules"))
	var cmd *exec.Cmd
	if r.App.WithYarn {
		cmd = exec.Command("yarnpkg", "install")
	} else {
		cmd = exec.Command("npm", "install")
	}

	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

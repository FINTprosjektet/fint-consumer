package setup

import (
	"github.com/codegangsta/cli"
	"github.com/FINTprosjektet/fint-consumer/generate"
	"github.com/FINTprosjektet/fint-consumer/common/github"
	"github.com/FINTprosjektet/fint-consumer/common/config"
	"github.com/FINTprosjektet/fint-consumer/common/types"
	"fmt"
	"os"
	"github.com/FINTprosjektet/fint-consumer/common/utils"
	"log"
	"strings"
	"io/ioutil"
	"gopkg.in/src-d/go-git.v4/plumbing/object"
	"time"
	"gopkg.in/src-d/go-git.v4"
)

func CmdSetupConsumer(c *cli.Context) {

	var tag string
	if c.GlobalString("tag") == config.DEFAULT_TAG {
		tag = github.GetLatest()
	} else {
		tag = c.GlobalString("tag")
	}
	force := c.GlobalBool("force")

	name := c.String("name")
	verfifyParameter(name, "Name parameter missing!")

	pkg := c.String("package")

	component := c.String("component")
	verfifyParameter(component, "Component parameter missing!")

	setupSkeleton(name)
	generate.Generate(tag, force)

	addModels(component, pkg, name)

	includePerson := c.Bool("includePerson")
	addPerson(includePerson, name)

	updateConfigFiles(name)

	reportNeedOfChanges(name)

	addModelToGradle(component, name)

	createReadme(tag, pkg, component, name)

	r, _ := git.PlainInit(getConsumerName(name), false)
	w, _ := r.Worktree()


	w.Add(".gitignore")
	commit, _ := w.Commit("Initial commit", &git.CommitOptions{
		Author: &object.Signature{
			Name:  "fint-provider cli",
			Email: "post@fintlabs.no",
			When:  time.Now(),
		},
	})

	obj, _ := r.CommitObject(commit)
	fmt.Println(obj)
}

func verfifyParameter(name string, message string) {
	if len(name) < 1 {
		fmt.Println(message)
		os.Exit(-1)
	}
}
func getModels(name string) []types.Model {
	files, _ := ioutil.ReadDir(fmt.Sprintf("%s/src/main/java/no/fint/consumer/models", getConsumerName(name)))

	var models = []types.Model{}
	for _, f := range files {
		if f.IsDir() {
			models = append(models, types.Model{Name: f.Name()})
		}
	}

	return models
}
func updateConfigFiles(name string) {
	models := getModels(name)
	writeConsumerPropsFile(getConsumerPropsClass(models), name)
	writeConstantsFile(getConstantsClass(name), name)
	writeLinkMapperFile(getLinkMapperClass(models), name)
	writeRestEndpointsFile(getRestEndpointsClass(models), name)
}
func addModels(component string, pkg string, name string) {
	src := fmt.Sprintf("%s/%s/%s/%s", utils.GetTempDirectory(), config.BASE_PATH, component, pkg)
	dest := fmt.Sprintf("./%s/src/main/java/no/fint/consumer/models/", getConsumerName(name))
	fmt.Println(src)
	fmt.Printf(dest)
	os.RemoveAll(dest)
	err := utils.CopyDir(src, dest)
	if err != nil {
		fmt.Println(err)
	}
}
func addPerson(includePerson bool, name string) {
	if includePerson {
		src := fmt.Sprintf("%s/%s/felles/person", utils.GetTempDirectory(), config.BASE_PATH)
		dest := fmt.Sprintf("./%s/src/main/java/no/fint/consumer/models/person/", getConsumerName(name))
		err := utils.CopyDir(src, dest)
		if err != nil {
			fmt.Println(err)
		}
	}
}

func addModelToGradle(model string, name string) {
	m := fmt.Sprintf("    compile(\"no.fint:fint-%s-model-java:${apiVersion}\")", model)
	gradleFile := utils.GetGradleFile(getConsumerName(name))
	input, err := ioutil.ReadFile(gradleFile)
	if err != nil {
		log.Fatalln(err)
	}

	lines := strings.Split(string(input), "\n")

	for i, line := range lines {
		if strings.Contains(line, "/* --> Models <-- */") {
			lines[i] = m
		}
	}
	output := strings.Join(lines, "\n")
	err = ioutil.WriteFile(gradleFile, []byte(output), 0644)
	if err != nil {
		log.Fatalln(err)
	}
}

func createReadme(tag string, pkg string, component string, name string) {
	content := fmt.Sprintf("# %s\n\nGenerated from tag `%s` on package `%s` and component `%s`.\n", 
		getConsumerName(name), tag, pkg, component)
	err := ioutil.WriteFile(utils.GetReadmeFile(getConsumerName(name)), []byte(content), 0644)
	if err != nil {
		fmt.Println(err)
	}
}

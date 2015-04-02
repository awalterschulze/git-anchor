// Copyright 2015 Vastech SA (PTY) LTD
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"math/rand"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"text/template"
)

type runError struct {
	data []byte
	err  error
}

func (this *runError) Error() string {
	return "runError: " + this.err.Error() + "\n" + string(this.data)
}

//TODO use lsremote to create a quick initial check
//func lsRemote(repo string) (rev string, err error) {
//exec.Command("git", "ls-remote", repo)
//}

func clone(repo string, tmp string) error {
	clone := exec.Command("git", "clone", repo, "--branch", "master", "--single-branch", tmp)
	data, err := clone.CombinedOutput()
	if err != nil {
		return &runError{data, err}
	}
	return nil
}

func commits(folder string) *bufio.Reader {
	cmd := exec.Command("git", "log", "--pretty=%H")
	cmd.Dir = folder
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		panic(err)
	}
	cmd.Start()
	buf := bufio.NewReader(stdout)
	return buf
}

func newTemp() string {
	tmp := os.TempDir()
	s := make([]string, 16)
	for i, _ := range s {
		s[i] = strconv.FormatInt(int64(rand.Intn(16)), 16)
	}
	ss := strings.Join(s, "")
	tmp = filepath.Join(tmp, "dep_"+ss)
	return tmp
}

func newestCommon(local, remote *bufio.Reader) (string, error) {
	commit, err := remote.ReadString('\n')
	commits := make(map[string]struct{})
	for err == nil || err == io.EOF {
		if err == io.EOF {
			err = nil
			break
		}
		commits[commit] = struct{}{}
		commit, err = remote.ReadString('\n')
	}
	if err != nil {
		return "", err
	}
	commit, err = local.ReadString('\n')
	for err == nil || err == io.EOF {
		if err == io.EOF {
			err = nil
			break
		}
		if _, ok := commits[commit]; ok {
			return strings.TrimSpace(commit), nil
		}
		commit, err = local.ReadString('\n')
	}
	return strings.TrimSpace(commit), err
}

func remoteRev(repo string) (rev string, err error) {
	tmp := newTemp()
	os.Mkdir(tmp, 0777)
	defer os.RemoveAll(tmp)
	if err := clone(repo, tmp); err != nil {
		return "", err
	}
	local := commits(".")
	remote := commits(tmp)
	return newestCommon(local, remote)
}

type subtrees struct {
	revs    map[string]string
	folders []string
}

func (this *subtrees) String() string {
	ss := make([]string, len(this.folders))
	for i, f := range this.folders {
		ss[i] = fmt.Sprintf("%s:%s", f, this.revs[f])
	}
	return strings.Join(ss, "\n")
}

func (this *subtrees) rev(dir string) string {
	return this.revs[dir]
}

func (this *subtrees) has(dir string) bool {
	_, ok := this.revs[dir]
	return ok
}

func newSubtrees() (*subtrees, error) {
	cmd := exec.Command("git", "log")
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}
	cmd.Start()
	revs := make(map[string]string)
	folders := []string{}
	buf := bufio.NewReader(stdout)
	line, err := buf.ReadString('\n')
	currentFolder := ""
	for err == nil || err == io.EOF {
		if len(currentFolder) > 0 {
			if strings.Contains(line, "git-subtree-mainline") {
				//do nothing give it until the next line
			} else if strings.Contains(line, "git-subtree-split") {
				ss := strings.Split(line, ":")
				if len(ss) == 2 {
					currentRev := strings.TrimSpace(ss[1])
					if _, ok := revs[currentFolder]; !ok {
						revs[currentFolder] = currentRev
						folders = append(folders, currentFolder)
					}
				}
			} else {
				currentFolder = ""
			}
		}
		if strings.Contains(line, "git-subtree-dir") {
			ss := strings.Split(line, ":")
			if len(ss) == 2 {
				currentFolder = strings.TrimSpace(ss[1])
			}
		}
		if err == io.EOF {
			err = nil
			break
		}
		line, err = buf.ReadString('\n')
	}
	if err != nil {
		return nil, err
	}
	sort.Strings(folders)
	return &subtrees{revs, folders}, nil
}

type Dep struct {
	Repo            string
	Dir             string
	Rev             string `json:",omitempty"`
	SquashedSubtree bool   `json:",omitempty"`
}

type Deps struct {
	Dir  string
	Deps []Dep
}

func newDeps(filename string) (Deps, error) {
	subs, err := newSubtrees()
	if err != nil {
		return Deps{}, err
	}
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return Deps{}, err
	}
	deps := Deps{}
	if err := json.Unmarshal(data, &deps); err != nil {
		return Deps{}, err
	}
	for i, d := range deps.Deps {
		if len(deps.Deps[i].Rev) > 0 {
			continue
		}
		if d.SquashedSubtree {
			if !subs.has(d.Dir) {
				return Deps{}, fmt.Errorf("%s is not a git subtree", d.Dir)
			}
			deps.Deps[i].Rev = subs.rev(d.Dir)
			continue
		}
		rev, err := remoteRev(d.Repo)
		if err != nil {
			return Deps{}, err
		}
		deps.Deps[i].Rev = rev
	}
	return deps, nil
}

var bashTemplate = `
#!/bin/bash
#set -xe

function subtreerev {
	git log | grep -E "(git-subtree-dir: $1|git-subtree-split)" | grep $1 -A 1 | grep git-subtree-split | head -1 | tr -d ' ' | cut -d ":" -f2
}

function subtrees {
	git log | grep git-subtree-dir | tr -d ' ' | cut -d ":" -f2 | sort | uniq | xargs -I {} bash -c 'if [ -d $(git rev-parse --show-toplevel)/{} ] ; then echo {}; fi'
}

function clonerev {
	(cd $1 && git log -1 | head -1 | cut -d " " -f2 )
}

function checkdep {
	dir=$1
	repo=$2
	rev=$3
	echo "checking dependency $dir $repo $rev"
	if [ ! -d $dir ]; then
		if ` + "`git rev-parse --is-inside-work-tree`" + `
		then
			echo "ERROR: This is a git repo, but $dir does not exist."
			echo "       You could add a subtree like so:"
			echo "       $ git subtree add --prefix=$dir $repo master"
			exit 1
		else
			mkdir -p $dir
			git clone $repo $dir
			(cd $dir && git checkout $rev)
		fi
	else
		if [ -e $dir/.git ]; then
			echo "found git repo at $dir"
			crev=$(clonerev $dir)
			if [ $crev == $rev ]; then
				echo "git clone $dir is the correct revision"
			else
				echo "WARNING: git clone $dir is revision $crev, but correct version is $rev"
			fi
		else
			ss=$(subtrees)
			if [[ $ss == *"$dir"* ]]; then
				echo "found subtree at $dir"
				srev=$(subtreerev $dir)
				if [ $srev == $rev ]; then
					echo "git subtree $dir is the correct revision"
				else
					echo "WARNING: git subtree (Please ignore this warning if your subtree is not squashed) $dir is revision $srev, but correct version is $rev. "
				fi
			else
				echo "WARNING: $dir exists, but is not git repo or git subtree"
			fi
		fi
	fi
}


if [ ! -d {{.Dir}} ]; then
	echo "ERROR: {{.Dir}} does not exist, maybe you are running this script from the wrong folder."
	exit 1
fi

{{range .Deps}}
checkdep {{.Dir}} {{.Repo}} {{.Rev}}
{{end}}

exit 0
`

func genBash(deps Deps) error {
	tmpl, err := template.New("t").Parse(bashTemplate)
	if err != nil {
		return err
	}
	if err := tmpl.Execute(os.Stdout, deps); err != nil {
		return err
	}
	return nil
}

var lang = flag.String("lang", "bash", "language in which to generate dependency script")
var exampleJson = flag.Bool("examplejson", false, "prints out an example dep.json file")
var list = flag.Bool("list", false, "print out a list of all the subtrees")
var help = flag.Bool("help", false, "help")
var h = flag.Bool("h", false, "help")

func checkInsideGit() error {
	out, err := exec.Command("git", "rev-parse", "--is-inside-work-tree").CombinedOutput()
	if err != nil {
		return err
	}
	if strings.Contains(string(out), "true") {
		return nil
	}
	return fmt.Errorf("you are not currently in a git work tree")
}

func main() {
	flag.Parse()
	if *help || *h {
		fmt.Fprintf(os.Stderr, "git-anchor generates a script from a git repo which retrieves version locked git subtree dependencies\n\n")
		flag.Usage()
		fmt.Fprintf(os.Stderr, "\nExamples:\n")
		fmt.Fprintf(os.Stderr, "\tgit-anchor -list #lists all the git subtrees in the git repo\n")
		fmt.Fprintf(os.Stderr, "\tgit-anchor -examplejson #prints out an example dep.json file\n")
		fmt.Fprintf(os.Stderr, "\tgit-anchor -lang=bash deps.json > deps.sh #generates a bash script from a git repo for all the dependencies in deps.json\n")

		return
	}
	if *exampleJson {
		d := &Deps{
			Dir: "src/vt/lib",
			Deps: []Dep{
				Dep{Repo: "github.com/gogo/protobuf", Dir: "src/github.com/gogo/protobuf"},
				Dep{Repo: "github.com/golang/crypto", Dir: "src/golang.org/x/crypto"},
			},
		}
		data, err := json.MarshalIndent(d, "", "\t")
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("%s\n", string(data))
		return
	}
	if err := checkInsideGit(); err != nil {
		log.Fatal(err)
	}
	if *list {
		fmt.Printf("list of subtrees in current git repo\n")
		subs, err := newSubtrees()
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("%s\n", subs)
		return
	}
	filename := flag.Arg(0)
	if len(filename) == 0 {
		fmt.Fprintf(os.Stderr, "expected json file describing dependencies")
		os.Exit(1)
	}
	deps, err := newDeps(filename)
	if err != nil {
		log.Fatal(err)
	}
	switch *lang {
	case "bash":
		if err := genBash(deps); err != nil {
			log.Fatal(err)
		}
	default:
		log.Fatalf("the generation for language %s is not implemented", *lang)
	}
}

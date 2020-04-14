/*
 * MIT License
 *
 * Copyright (c) 2020 TCorp BV
 *
 * Permission is hereby granted, free of charge, to any person obtaining a copy
 * of this software and associated documentation files (the "Software"), to deal
 * in the Software without restriction, including without limitation the rights
 * to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
 * copies of the Software, and to permit persons to whom the Software is
 * furnished to do so, subject to the following conditions:
 *
 * The above copyright notice and this permission notice shall be included in all
 * copies or substantial portions of the Software.
 *
 * THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 * IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 * FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
 * AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
 * LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
 * OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
 * SOFTWARE.
 */

package main

import (
	"fmt"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
)

const (
	langGo Language = "go"
)

var (
	// Which directories in the mono repo that should be ignored
	ignoredDirectories = map[string]struct{}{".git": {}, ".idea": {}, "tools": {}}
	// Which files and directories in the generated repos that should be ignored
	noReset = map[string]struct{}{"README": {}, "README.md": {}, "readme.md": {}, "LICENSE": {}, "license": {}, ".git": {}}
)

// Only "go" supported for now,
type Language string

// Marshalled version of .proto.yaml
type Package struct {
	Languages []LanguageDef `yaml:"languages"`
}

// A language section of .proto.yaml. Contains information about which language code should be generated for and to which repository.
type LanguageDef struct {
	// The language, only "go" for now
	Language Language `yaml:"language"`
	// The repository "https" and ".git" should be omitted for now.
	Repository string `yaml:"repository"`
}

// Gets the
func (ld *LanguageDef) RepoURI() string {
	return fmt.Sprintf("https://%s.git", ld.Repository)
}

// The definition of an actual package in the file tree
type PackageDef struct {
	// The description of the package per .proto.yaml
	Package Package
	// The directory fileinfo (contains the name of the directory, eg. "greeter"
	Directory os.FileInfo
	// the list of proto files in this directory, eg. "greeter.proto" (these are not validated!)
	Protos []os.FileInfo
}

// Handle the deployment of a subdirectory: Checks that the file contains a.proto.yaml, reads it and deploys to all the target directories.
func (h *Handler) handleDir(dir os.FileInfo) error {
	if _, ok := ignoredDirectories[dir.Name()]; ok { // Just skip on the ignored directories
		return nil
	}
	bytes, err := ioutil.ReadFile(filepath.Join(h.Path, dir.Name(), ".proto.yaml"))
	if err != nil {
		return err
	}

	var pack Package
	if err := yaml.Unmarshal(bytes, &pack); err != nil {
		return err
	}

	packDef := PackageDef{Package: pack, Directory: dir}
	// Traverse the files and add .proto files to the Protos field of the PackageDef
	contents, err := ioutil.ReadDir(filepath.Join(h.Path, dir.Name()))
	for _, content := range contents {
		if content.Mode().IsRegular() && filepath.Ext(content.Name()) == ".proto" {
			packDef.Protos = append(packDef.Protos, content)
		}
	}

	return h.handlePackage(packDef)
}

// Handles the deployment of all languages in a package (directory)
func (h *Handler) handlePackage(def PackageDef) error {
	for _, lang := range def.Package.Languages {
		if err := h.handleLang(def, lang); err != nil {
			return err
		}
	}
	return nil
}

// Clone the repository, error out if it does not exist. Reset the repository. Generate code into the repository and push the generated code to the repository.
func (h *Handler) handleLang(def PackageDef, lang LanguageDef) error {
	fmt.Printf("Handling language %s for %s...\n", lang.Language, def.Directory.Name())

	gitDir, err := ioutil.TempDir(os.TempDir(), filepath.Base(h.Path)+"-")
	println(gitDir)
	//defer func() {
	//    _ = os.Remove(gitDir)
	//}()
	if err != nil {
		return err
	}

	cloneCmd := exec.Command("git", "clone", lang.RepoURI(), gitDir)
	if err := cloneCmd.Run(); err != nil {
		return fmt.Errorf("git clone error: %v. Please make sure %s exists and git has access to it", err, lang.RepoURI())
	}

	// Remove all contents except the ones in noReset (This way we overwrite the repository)
	contents, err := ioutil.ReadDir(gitDir)
	if err != nil {
		return err
	}
	for _, content := range contents {
		if _, ok := noReset[content.Name()]; ok {
			continue
		}
		_ = os.RemoveAll(filepath.Join(gitDir, content.Name()))
	}

	for _, protoFile := range def.Protos {
		var args []string
		switch lang.Language {
		case langGo:
			args = append(args, "--go_out=plugins=grpc:"+gitDir)
		default:
			return fmt.Errorf("language %s not found", lang.Language)
		}

		args = append(args, filepath.Join(def.Directory.Name(), protoFile.Name()))

		protocCmd := exec.Command("protoc", args...)
		fmt.Printf("Executing %s\n", protocCmd.String())
		if str, err := protocCmd.CombinedOutput(); err != nil {
			fmt.Printf("%s\n", string(str))
			return fmt.Errorf("protoc error: %v", err)
		}
	}

	addCmd := exec.Command("git", "-C", gitDir, "add", "-A", ".")
	if err := addCmd.Run(); err != nil {
		return fmt.Errorf("git add error: %v", err)
	}

	commitCmd := exec.Command("git", "-C", gitDir, "commit", "-m", "'Automatically updated library from protorepo'")
	if err := commitCmd.Run(); err != nil {
		return nil // We do nothing when the commit returns an error, we assume this means "nothing to commit"
	}
	
	getHashCmd := exec.Command("git", "-C", gitDir, "log", "--pretty=%h", "-1")
	hash, err := getHashCmd.CombinedOutput(); 
	if err != nil {
			return fmt.Errorf("git get commit hash error: %v", hash)
	}
	
    	tagCmd := exec.Command("git", "-C", gitdir, "tag", "-a", string(hash))
    	if err := tagCmd.Run(); err != nil {
		return fmt.Errorf("git tag error: %v", err)
	}
	
	fmt.Printf("Deploying %s to %s!\n", hash, lang.RepoURI())

	pushCmd := exec.Command("git", "-C", gitDir, "push", "origin", "master")
	if err := pushCmd.Run(); err != nil {
		return fmt.Errorf("git push error: %v", err)
	}

	return nil
}

// Manages the deployment of the generated code
type Handler struct {
	// The path to the root directory
	Path string
	// All subdirectories, these may not contain a .proto.yaml file
	Dirs []os.FileInfo
}

// Load in all subdirectories (depth 1)
func (h *Handler) setup() error {
	path, err := os.Getwd()
	h.Path = filepath.Clean(path)
	if err != nil {
		return err
	}
	contents, err := ioutil.ReadDir(h.Path)
	for _, content := range contents {
		if content.IsDir() {
			h.Dirs = append(h.Dirs, content)
		}
	}
	return err
}

// Traverse all directories in the current directory, check that there is a .proto.yaml file and generate code according to it and the proto files in the directory.
func run() error {
	h := Handler{}
	if err := h.setup(); err != nil {
		return err
	}

	for _, dir := range h.Dirs {
		if err := h.handleDir(dir); err != nil {
			fmt.Printf("Issue handling .proto.yaml from %s: %v\n", dir.Name(), err)
		}
	}
	return nil
}

// Requires git credentials to be setup by the environment like:
// git config credential.helper '!f() { sleep 1; echo "username=${GIT_USER}"; echo "password=${GIT_PASSWORD}"; }; f'
func main() {
	if err := run(); err != nil {
		log.Panic(err)
	}
}

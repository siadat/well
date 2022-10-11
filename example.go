package main

import (
	"fmt"

	"github.com/siadat/well/newsh"
)

func main() {
	// TODO: this function should be generated from user's func args, command line flags, and env (and stdin?)
	user_main(
		newsh.File{Path: "program1"},
		"ecosystem0",
		newsh.File{Path: "vault0"},
		"branch0",
	)
}

func user_main(
	yelpsoa_dir newsh.File,
	ecosystem string,
	vault_bin newsh.File,
	branch string,
) {
	authenticate("hello", ecosystem, vault_bin)
	var username string = get_username()
	return
	newsh.PrintInfo(newsh.Interpolate("username is ${username:%q}", newsh.ValMap{"username": username}))
	clone_yelpsoa_repo(yelpsoa_dir, branch)
}

//@file r"/path/to/dir/*"
func get_username() string {

	// newsh.ExternalPiped(
	// 	newsh.ValMap{"key": "hello world"},
	// 	newsh.Pipe{
	// 		"ls  -la -sh",
	// 	})

	// var x = newsh.ExternalPiped(
	// 	nil,
	// 	newsh.Pipe{
	// 		`yes`,
	// 		`nl -s\t`,
	// 		`head -n3`,
	// 	})
	// newsh.PrintInfo(x)

	var tempFile = "x.json" // newsh.ExternalTrimmed(nil, `mktemp`, newsh.Options{TrimSpaces: true})
	fmt.Println(tempFile)
	var y = newsh.ExternalPiped(
		newsh.ValMap{"tempFile": tempFile},
		newsh.Pipe{
			`cat ${tempFile:%q}`, // TODO: this is quoted even though it is %s `jq -r «."hello world"»`,
		})
	newsh.PrintInfo(y)
	return ""

	return newsh.ExternalTrimmed(nil, "whoami")
}

//@file r"/path/to/dir/*"
//@file rw"/anotherdir/*"
//@file global.Tty
//@net "google.com", "1.1.1.4"
func authenticate(
	username string,
	ecosystem string,
	vault_bin newsh.File,
) {
	newsh.External(
		newsh.ValMap{
			"vault_bin": vault_bin,
			"username":  username,
			"ecosystem": ecosystem,
		},
		"${vault_bin:%q} auth --user ${username:%q} --ecosystem ${ecosystem:%q}")
}

//@file yelpsoa_dir,
//@file rw"/tmp/myscript/*",
//@net "github.yelpcorp.com",
func clone_yelpsoa_repo(
	yelpsoa_dir newsh.File,
	yelpsoa_remote_branch string,
) newsh.Void {
	var remote = "git@github.yelpcorp.com:sysgit/yelpsoa-configs.git"
	if !newsh.FileExists(yelpsoa_dir) {
		newsh.External(
			newsh.ValMap{
				"remote":                remote,
				"yelpsoa_dir":           yelpsoa_dir,
				"yelpsoa_remote_branch": yelpsoa_remote_branch,
			},
			"git clone --depth=1 ${remote:%q} ${yelpsoa_dir:%q} --branch ${yelpsoa_remote_branch:%q}")
		return newsh.Nothing
	}

	newsh.PrintInfo(newsh.Interpolate("yelpsoa_dir ${yelpsoa_dir:%q} exists", newsh.ValMap{"yelpsoa_dir": yelpsoa_dir}))
	newsh.Cd(yelpsoa_dir, func() {
		var got_origin_url = newsh.ExternalTrimmed(nil, "git remote get-url --all origin")
		if got_origin_url != remote {
			newsh.Exit(newsh.Interpolate("Want origin remote url to be ${remote}, got ${got_origin_url:%q}", newsh.ValMap{
				"remote":         remote,
				"got_origin_url": got_origin_url,
			}))
		}
	})

	return newsh.Nothing
}

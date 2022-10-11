package main

import "github.com/siadat/well/newsh"

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
	var username string = get_username()
	newsh.PrintInfo(newsh.Interpolate("username is @{username}", newsh.ValMap{"username": username}))
	// authenticate(username, ecosystem, vault_bin)
	clone_yelpsoa_repo(yelpsoa_dir, branch)
}

//@file r"/path/to/dir/*"
func get_username() string {

	var z = newsh.ExternalPiped(newsh.Pipe{
		"ls -la -sh",
	})
	newsh.PrintInfo(z)
	var x = newsh.ExternalPiped(newsh.Pipe{
		"yes",
		"nl -s\t",
		"head -n3",
	})
	newsh.PrintInfo(x)

	var y = newsh.ExternalPiped(newsh.Pipe{
		"cat x.json",
		// newsh.Interpolate(`jq .@{key}`, newsh.ValMap{"key": "hello world"}),
		// `jq ."hello world"`,
		`jq -r ."hello world"`,
	})
	newsh.PrintInfo(y)
	return newsh.External("whoami", newsh.Options{TrimSpaces: true})
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
	newsh.External(newsh.Interpolate("@{vault_bin} auth --user @{username} --ecosystem @{ecosystem}", newsh.ValMap{
		"vault_bin": vault_bin,
		"username":  username,
		"ecosystem": ecosystem,
	}))
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
		newsh.External(newsh.Interpolate("git clone --depth=1 @{remote} @{yelpsoa_dir} --branch @{yelpsoa_remote_branch}", newsh.ValMap{
			"remote":                remote,
			"yelpsoa_dir":           yelpsoa_dir,
			"yelpsoa_remote_branch": yelpsoa_remote_branch,
		}))
		return newsh.Nothing
	}

	newsh.PrintInfo(newsh.Interpolate("yelpsoa_dir @{yelpsoa_dir} exists", newsh.ValMap{"yelpsoa_dir": yelpsoa_dir}))
	newsh.Cd(yelpsoa_dir, func() {
		var got_origin_url = newsh.External("git remote get-url --all origin")
		if got_origin_url != remote {
			newsh.Exit(newsh.Interpolate("Want origin remote url to be ${remote}, got @{got_origin_url}", newsh.ValMap{
				"remote":         remote,
				"got_origin_url": got_origin_url,
			}))
		}
	})

	return newsh.Nothing
}

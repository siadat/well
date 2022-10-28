# Guillemets

Guillemets is a cli command which can be used to quote strings using guillemets characters (i.e., `«`, `»`, `‹`, `›`).

Basically, it allows you to write this:
```
sh -c «sh -c «sh -c «sh -c «sh -c «date»»»»»
```

Instead of this:
```
sh -c "sh -c \"sh -c \\\"sh -c \\\\\\\"sh -c \\\\\\\\\\\\\\\"date\\\\\\\\\\\\\\\"\\\\\\\"\\\"\""
```

This is done using the following command:

```
guillemets exec -i 'sh -c «sh -c «sh -c «sh -c «sh -c «date»»»»»'
```

## Install

```
go install github.com/siadat/well/cmd/guillemets
```

## Basic usage

There are two main sub-commands `guillemets exec` and `guillemets render`.

Execute a command (show executed command):
```
$ guillemets exec -i 'echo «hello world»'
hello world
```

Render string (no execution):
```
$ guillemets render --newline -i 'sh -c «sh -c «sh -c «sh -c «sh -c «date»»»»»'
sh -c "sh -c \"sh -c \\\"sh -c \\\\\\\"sh -c \\\\\\\\\\\\\\\"date\\\\\\\\\\\\\\\"\\\\\\\"\\\"\""
```

## More usage

Execute a command (show executed command with `-v`):
```
$ guillemets exec -v -i 'echo «hello world»'
+echo "hello world"
hello world
```

Execute a command (with env variable `${...}`):
```
$ name=sina guillemets exec -v -i 'echo «hello «${name}»!»'
+echo "hello \"sina\"!"
hello "sina"!
```

Execute a command (with env variable quoted `${...:%q}`):
```
$ name=sina guillemets exec -v -i 'echo «hello ${name:%q}!»'
+echo "hello \"sina\"!"
hello "sina"!
```

Execute a command (with escaped `«` and `»` characters)
```
$ name=sina guillemets exec -v -i 'echo «This is an actual \«guillemet\».»'
+echo "This is an actual «guillemet»."
This is an actual «guillemet».
```

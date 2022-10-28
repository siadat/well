# Guillemets

guillemets is a command that can be used to use guillemets characters (i.e., `«`, `»`, )

## Install

```
go install github.com/siadat/cmd/guillemets
```

## Basic usage

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

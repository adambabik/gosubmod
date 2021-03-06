gosubmod
========

gosubmod is a tool that simplifies working with Go submodules.

```shell script
$ go install github.com/adambabik/gosubmod
```

## Usage

```shell script
$ gosubmod -h

gosubmod is a tool that simplifies working with Go submodules.

Usage:

	gosubmod <command> [arguments] submodules...

The commands are:

	list    list all the recognized submodules
	add     add "replace" directives with relative paths for submodules
	drop    drop "replace" directives with relative paths for submodules
```

Assume you have submodules `example.com/a/b` and `example.com/a/c`:

```
module example.com/a

require (
	example.com/a/b v1.0.0
	example.com/a/c/v2 v2.0.0
)
```

If you want to replace them with relative modules in order to speed up development process, execute `gosubmod add`. This will change the `go.mod` to:

```
module example.com/a

require (
	example.com/a/b v1.0.0
	example.com/a/c/v2 v2.0.0
)

replace example.com/a/b => ./b
replace example.com/a/c/v2 => ./c
```

You can also replace a single module using `gosubmod add example.com/a/b`.

To get back to the original version run `gosubmodo remove`.

## Learn more about submodules

Submodules are tricky and you probably don't need them. However, in certain situation, they can be useful. [Learn more about submodules](https://github.com/go-modules-by-example/index/blob/master/009_submodules/README.md) before going further if the concept is unknown to you.

## Problem

Developing and releasing nested Go Modules (submodules) is not trivial.

If the main module requires nested modules, it expects them to be in a specific version and goes through the internet to download them. This is good because it makes sure that the main module uses the versions available for everyone else. At the same time, it makes the development tricky if one wants to develop and do integration testing at the same time.

## Solution

The solution to the latter is to use `replace` directive and relative imports.

Assume you have a module `example.com/a` and two submodules `example.com/a/b` and `example.com/a/c`. The `go.mod` file can look like this:

```
module example.com/a

require (
	example.com/a/b v1.0.0
	example.com/a/c/v2 v2.0.0
)
```

Now, if you want to change the code in submodule `example.com/a/b` and test it with the main module, you would need to commit and push the changes, use `go get` to change the version and run tests.

The simpler solution is to use relative import:

```
module example.com/a

require (
	example.com/a/b v1.0.0
	example.com/a/c/v2 v2.0.0
)

replace example.com/a/b => ./b
```

With this `replace` directive, you can run the tests in the main module and the local code will be used.

Of course, you should not commit such altered `go.mod` because it leads to unpredictable builds. When the development is finished, `replace` directive should be dropped and proper tags for submodules should be created followed by updating the main `go.mod` submodule versions:

```
# create and push a tag
$ git tag -a -m "Release example.com/a/b@v1.1.0" b/v1.1.0
$ git tag push b/v1.1.0
```

```
module example.com/a

require (
	example.com/a/b v1.1.0
	example.com/a/c/v2 v2.0.0
)
```

## License

MIT License

Copyright (c) 2020 Adam Babik

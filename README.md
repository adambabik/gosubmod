gosubmod
========

gosubmod is a tool that simplifies working with Go submodules.

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

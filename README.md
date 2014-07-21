vengo
=====

another go package manager experiment

## concept

* Fetch source using the same mechanisms as `go get`
* Instead of writing the source to `$GOROOT/src/...` write it to `./vendor/`
* Source that was just downloaded is parsed and imports are rewritten to prefix your project path

So, if I'm working on `github.com/supershabam/vengo` and `vengo get github.com/gorilla/mux` (which depends on `github.com/gorilla/context`) the mux source will be placed into my local directory at `vendor/github.com/gorilla/mux` and I should reference it in my project as `github.com/supershabam/vengo/vendor/github.com/gorilla/mux` which is a lot to type, but whatever.

During the `vengo get` it looks at other imports (like the one to context) and rewrites them to vendored imports. So the mux code in my vendor directory actually will have an import to `github.com/supershabam/vengo/vendor/github.com/gorilla/contex` (which doesn't exist... yet).

Then, if I try and run the project, it will complain that it can't find that context code in the normal go complaining way.

I can then see it's expecting a vendored version of context and can run `vengo get github.com/gorilla/context` to bring it into my project.

Then, I commit my code. My dependencies are locked into my source. No go gets in the build process. No dealing with ensuring the correct versions when collaborating on the project with others. If you want to update your dependencies, remove the folder and re-vengo them.

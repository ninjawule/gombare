# gombare
General comparing functions developed in Golang.

Works in CLI to compare 2 JSON or XML files. Also works with 2 folders. 

But also works in Golang through the `/core` package here, to compare `interface{}`, `map`, `slice` (etc) objects.


## TODO

- [ ] Examples
- [ ] `-split` and `-outdir` options not developed yet

## How to use

```sh
-> % gombare -h
Usage of gombare:
  -allowRaw
    	if true, then it's allowed to display the raw JSON entities as difference, when added or removed; else, a display template is required
  -check
    	if true, then the ID params are output to allow for some checks
  -fast
    	if true, then some verifications are not performed, like the uniqueness of IDs coming from the id props specified by the user; WARNING: this can lead to missing some differences!
  -idparams string
    	a JSON representation of a IdentificationParameter parameter; see the docs for an example; can be the path to an existing JSON file
  -ignore string
    	the files to ignore, separated by a comma
  -nparallel int
    	the number of routines used at the same time when comparing several files at once (i.e. comparing folders) (default 10)
  -one string
    	required: the path to the first file to compare; must be a JSON file, or XML with the -xml option
  -silent
    	if true, then no info / warning message is written out
  -stopAtFirst
    	if true, then, when comparing folders, we stop at the first couple of files that differ
  -two string
    	required: the path to the second file to compare; must be of the same first file's type
  -xml
    	use this option if the files are XML files
```

## Acknowledgments

Using the really nice [xml2map](https://github.com/sbabiv/xml2map) program from [Sergey Babiv](https://github.com/sbabiv) for the necessary `XML -> map[string]interface{}` transformation.

## Licence 

[MIT](./LICENSE)

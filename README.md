# json-compare
Takes 2 JSON or XML files and compare them. The result is given as a JSON. Also works with 2 folders of JSON / XML files.

## TODO

- [ ] Examples
- [ ] `-split` and `-outdir` options not developed yet

## How to use

```sh
-> % json-compare -h
Usage of json-compare:
  -autoIndex
    	if true, then for array of objects with no id prop (cf. idprops option), the object indexes in the arrays are used as IDs
  -fast
    	if true, then some verifications are not performed, like the uniqueness of IDs coming from the id props specified by the user; WARNING: this can lead to missing some differences!
  -idprops string
    	for an array of objects, we need an identifying property for the objects, for sorting purposes amongst other things; if '#index' is used as an ID, then that means that an object index in the surrounding array is used as its ID; example: ">path1>path2>path3:::propA+path4>propB as id3,>path1>path2>path3>id3>path5:::propC"
  -one string
    	required: the path to the first file to compare; must be a JSON file, or XML with the -xml option
  -orderby string
    	for an array of objects that we cannot really define an ID property for, we want to sort the objects before comparing them with their index. The syntax is the same as for the -idprops option
  -outdir string
    	when specified, the result is written out as a JSON into this specified output directory
  -silent
    	if true, then no info / warning message is written out
  -split
    	if 2 folders are compared, and if -outpir is used, then there is 1 comparison JSON produced for each pair of compared files
  -stopAtFirst
    	if true, then, when comparing folders, we stop at the first couple of files that differ
  -two string
    	required: the path to the second file to compare; must be of the same first file type
  -xml
    	use this option if the files are XML files
```

## Acknowledgments

Using the really nice [xml2map](https://github.com/sbabiv/xml2map) program from [Sergey Babiv](https://github.com/sbabiv) for the necessary `XML -> map[string]interface{}` transformation.

## Licence 

[MIT](./LICENSE)
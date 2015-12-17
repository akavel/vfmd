
# Vanilla Flavored Markdown (vfmd) Parser in Go

[vfmd (Vanilla Flavored Markdown)](http://vfmd.org) is a sane Markdown variant
[with an unambiguous specification of syntax](http://vfmd.org). This package is
a pure-[Go](http://golang.org) implementation of a parser for vfmd, with
additional design goals and some more specific characteristics listed below:

## Goals

- **Adhere to the [vfmd spec](http://www.vfmd.org/vfmd-spec/specification/)
  (fixing it as needed)**;
    - Done, with a notable exception of inline HTML;
    - Any assumed issues found in process [reported as a pull
      request](https://github.com/vfmd/vfmd-spec/pull/8);
- **Allow for any custom renderers, by outputting an intermediate format ("AST")**;
    - Done, a flattened tree representation is generated;
    - As an example and proof of concept, a HTML renderer is provided;
- **Provide end-to-end mapping from input characters to the final parsed form
  (this can make it useful e.g. for syntax-highlighting)**;
    - Partially done: fulfilled for blocks, still TODO for spans (I tried to
      prepare for that, but it may well need API changes, so possibly this may
      require creating a new version, i.e. vfmd.v2 or later);
- **Allow quick top-level-only parsing (e.g. to scan headers in order to build a
  Table of Contents)**;
    - Done;
- **Pure Go**;
    - Done;
- **Try to determine worst-case efficiency (and then maybe try to reduce it)**;
    - TODO;
    - (Note: I think it should be possible to have it at least as good as
      amortized _O(n*m*k²)_, where *n* is number of lines, *m* is deepest
      nesting level of blocks, and *k* is length of the longest paragraph (more
      strictly, _[text span
      sequence](http://www.vfmd.org/vfmd-spec/specification/#identifying-span-elements)_).
      But I *absolutely* haven't confirmed yet if the current code has such
      efficiency characteristics.)

## More detailed characteristics

- **Extensible syntax** (thanks to the vfmd spec) ― both for block- and
  span-level markup;
    - As an example, subpackage
      [x/mdgithub](https://godoc.org/gopkg.in/akavel/vfmd.v1/x/mdgithub)
      provides some extensions from [GitHub-flavored
      Markdown](https://help.github.com/articles/github-flavored-markdown/):
      strikethrough with `~~` and fenced code blocks with triple backtick. The
      [cmd/vfmd](https://godoc.org/gopkg.in/akavel/vfmd.v1/cmd/vfmd) sample
      application shows how to enable those (when executed with `--github`
      flag).
    - __TODO:__ add tables support from GH-flavored MD too.
- **Quite well-tested** (thanks to the vfmd testsuite);
- __*Does not* support inline HTML__ (at least currently; this is arguably a
  feature for some use cases, like desktop editors or comment systems);
- __*Does not* support inline HTML entities__ (like `&amp;` etc.) ― Unicode should
  make up for that;
- __TODO:__ the QuickHTML renderer does not currently filter URLs in links to
  protect against e.g. JavaScript "bookmarklet" attacks;
- __FIXME:__ detect md.HardBreak tag for lines ending with `"  \n"`;
- __FIXME:__ godoc
- __FIXME:__ example in README
- __FIXME:__ add tests for GitHub-flavored Markdown extensions;
- __FIXME:__ true Region information in spans (vfmd.v2?)
- __TODO:__ make DefaultDetectors comparable?
- __TODO:__ add SmartyPants extensions (also, `<a name="..." />` anchors if not there);
- __TODO:__ add [tests from Blackfriday](https://github.com/russross/blackfriday/tree/master/testdata) too;



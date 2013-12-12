Morpheme Extraction System
==========================

This software allows for the programmatic extraction of morphemes
from a corpus into a defined morpheme-list location.

Command-line Arguments:

Verbose Mode:      --verbose
Serial Processing: --serial or --sequential --process-sequentially
Full Processing:   --process
Output File:       --output-file REL-FILE-PATH
Corpus Dir:        --corpus-dir  REL-CORPUS-PATH

where REL-FILE-PATH and REL-CORPUS-PATH are relative paths to a
desired filename and/or corpus directory.

Verbose Mode gives more visual output, however it impacts speed.

Serial Processing yields data results for each file process as
    opposed to a conglomorate data processing experience :)

Full Processing yields serial and sequential results as if you
    were to have run the program with --serial the first time
    and then a second time without that flag.

Output File is the place in which data results are appended
    (it won't overwrite existing data).

Corpus Dir is the place where all the files that need to be
    processed reside.

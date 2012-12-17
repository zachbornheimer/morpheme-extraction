#!/usr/loal/bin/perl6

use v6;
:autoflush; # Forces immediate flush.

constant PreviousWords = 4; # How many words behind the currently analyzed word should have their positions recorded
constant SubsequentWords = 4; # How many words in front of the currently analyzed word shave their positions recorded

my $restoreFile = "restore.zy";
my @letters;
my @words = ();
my @dictionary = ();
my $nlpLiteratureDir = "nlp-literature"; # No trailing slash
my %wordsOnEitherSide = ();
my @filesToParse;
my %frequencyHash;
my %procedureHash;
my %wordsInSimilarContexts;

grammar english {
    token punctuation {
        <- [a..zA..Z\d\s]>+($|\s+|<- [a..zA..Z\d]>)
    }
}

grammar Lang {
    rule TOP {
        <prefix>*<stem>+<suffix>*
    }
    proto token prefix {*}
    proto token stem {*}
    proto token suffix {*}
    proto token infix {*}
}

grammar replGrammar {
    rule TOP {
        <command>
    }
    token command {
        <keyword>\s+<command> | <keyword>\s+(<expression> +)+ | <expression> | <keyword>
    }
    token expression {
        <- keyword>*
    }
    proto token keyword { * }
    token keyword:sym<help> { <sym> }
    token keyword:sym<show me> { <sym> }
    token keyword:sym<display> { <sym> }
    token keyword:sym<parse> { <sym> }
    token keyword:sym<learn> { <sym> }
    token keyword:sym<exit> { <sym> }
    token keyword:sym<process> { <sym> }
    token keyword:sym<store> { <sym> }
    token keyword:sym<load> { <sym> }
    token keyword:sym<demo> { <sym> }
}


sub run($val, $failMessage = 'Code failed to evaluate successfully') {
    # The purpose of this is to generate a simplistic error catching interface
    my $evalResult = $val.eval;
    CATCH {
        say $failMessage if $failMessage;
        $evalResult = 0;
    }
    return $evalResult;
}

sub typecheck($object, $type) {
    return ~$object.WHICH ~~ m:i/^$type\|/;
}

sub parse(Str $filename) {
    # Eventually will allow for directories.
    learn ([$filename]);
}

multi sub learn() {
    recursiveDir($nlpLiteratureDir, @filesToParse);
    learn(@filesToParse);
}
multi sub learn(@filesToParse) {
    # Read all text into an array of words.
    my $numFilesProcessed = 0;
    for (@filesToParse) {
        if $_.IO ~~ :e  {
            my @uniqueWords;
            my $acqWordsCounter = 0;
            my $acqWordsIndex = 0;
            say "\nReading $_";
            print "\tParsing $_ ...";
            my Str $text = slurp $_;
            $text ~~ s:g/\n/\ /;
            my Int $numWords = $text.split(' ').elems - 1; # index 0 
            say "done";
            print "\tProcessing the text ...\r";
            my @acqWords = $text.split(' '); # acquired words
            for (@acqWords) {
                print "\tProcessing the text ..." ~ ($acqWordsIndex / $numWords) * 100 ~ "%                 \r";
                $_ ~~ s:g/<english::punctuation>/ /; # Needs to replace the punctuation with a space 
                $_ ~~ s:g/\ //;

                ###################################################################################
                # The ending data structure for the following code will look something like this:
                # This example is done for 4 words.
                # %hash = ( $word => (@wordsBefore, @wordsAfter) );
                # @wordsBefore is the  words before that word. 
                # @wordsBefore = ( %_4wordsBefore, %_3wordsBefore, %_2wordsBefore, %_1wordBefore );
                # %_4wordsBefore = ("another"=>3, "word"=>2, "frequency"=>14, "table"=>9);
                # %_NUMwordsBefore (where NUM is any number) is similar to %_4wordsBefore
                # @wordsAfter looks like @wordsBefore
                # Note: frequency for surrounding words only happens after running the stats.
                ###################################################################################

                if ($acqWordsIndex >= 0) {
                    for 0 .. 1 -> $iterator {
                        %wordsOnEitherSide{$_}[$iterator] = [] if !(typecheck %wordsOnEitherSide{$_}[$iterator], 'List');
                    }

                    my $wordsBefore = item %wordsOnEitherSide{$_}[0];
                    my $wordsAfter = item %wordsOnEitherSide{$_}[1];

                    # Because we have not yet incremented $acqWordsIndex, it will automatically be 1 less than the current word
                    my $startingNum = $acqWordsIndex - PreviousWords;
                    if $startingNum < 0 {
                        $startingNum = 0;
                    }

                    my $endingNum = $acqWordsIndex - 1;
                    my $array = $wordsBefore;
                    for 0 .. 1 -> $iteration {
                        for $startingNum .. $endingNum {
                            my $i = $endingNum - $_;
                            @$array[$i].push(@acqWords[$_]) if defined @acqWords[$_] && @acqWords[$_] ne "" ;
                        }
                        if $iteration == 0 {
                            $startingNum = $acqWordsIndex + 2; # $acqWordsIndex + 1 = current_word
                            $endingNum = $acqWordsIndex + SubsequentWords + 1; # SubsequentWord is base 0
                            $array = $wordsAfter;
                        }
                    }

                    $endingNum = PreviousWords - 1;
                    $array = $wordsBefore;
                    for 0 .. 1 -> $iteration {
                        for 0 .. $endingNum -> $val {
                            my @arr;
                            my $t = @$array[$val][$iteration];
                            push @arr, @$array[$val][$iteration] if (defined $t) && $t;
                            push %wordsOnEitherSide{$_}[$iteration][$val], @arr;
                        }
                        if $iteration == 0 {
                            $endingNum = SubsequentWords - 1;
                            $array = $wordsAfter;
                        }
                    }

                    for ($wordsBefore, $wordsAfter) -> $array {
                        @$array = unique(@$array);
                    }

                    if !defined %wordsOnEitherSide{$_}[0] {
                        for 0 .. PreviousWords - 1 {
                            %wordsOnEitherSide{$_}[0].push('[]');
                        }
                    }
                    if !defined %wordsOnEitherSide{$_}[1] {
                        for 0 .. SubsequentWords - 1 {
                            %wordsOnEitherSide{$_}[1].push('[]');
                        }
                    }

                    $acqWordsIndex++;

                }
            }
            for (@acqWords) {
                if $_ ne ""  {
                    @words.push($_);
                }
            }
            print "\tProcessing the text ...100%                            \r";
            say "\tProcessing the text ...done";
            say "Done Reading $_";
            say "Learned " ~ @acqWords.elems ~ " words\n";
            $numFilesProcessed++;
        }
    }
    if $numFilesProcessed > 0 {
        print "Creating the dictionary database by removing duplicate words in the word database ...";
        for unique(@words) {
            @dictionary.push($_);
        }
        @dictionary = unique @dictionary;
        say "done";
    }

}
sub removeEmpty(@array) {
    my @a;
    for (0 .. @array.elems - 1) {
        push @a, @array[$_] if @array[$_] ne "";
    }
    return @a;
}


sub stats($kind is rw, $num? = 0) {
    if ($kind eq "wordFrequency") {
        return wFrequency(@words, $num);
    } elsif ($kind eq "surroundingWordFrequency") {
        # We do this for each one
        for (%wordsOnEitherSide.keys) {
            # Two Elements, 0 and 1
            # We must get the arrays and replace them with hashes.
            # @alpha is a throwaway reference in both the blocks below.
            for %wordsOnEitherSide{$_}[0] -> @item is rw {
                for 0 .. @item.elems-1 -> $index {
                    for @item[$index] -> @alpha is rw {
                        @alpha = sFrequency(@alpha);
                    }
                }
            }
            for %wordsOnEitherSide{$_}[1] -> @item is rw {
                for 0 .. @item.elems-1 -> $index {
                    for @item[$index] -> @alpha is rw {
                        @alpha = sFrequency(@alpha);
                    }
                }
            }
        }
    } elsif ($kind eq "context") {
        # General Procedure:
        #
        # Analyze the Following:
        #     P(W0|Wb0_in_pos_1); P(W0|Wa0_in_pos_1); P(W1|Wb1_in_pos_2); P(W0|Wa1_in_pos_2); 
        #     P(W0|Wbn_in_pos_n+1); P(W0|Wan_in_pos_n+1);
        #     Where W0 is the Current word and Wb0 is the the first word before and Wa0 is the first word after
        # Then analyze the following:
        #     P(W0|Wbn_in_pos_n+1_AND_Wbn+1+in_pos_n+2) etc for each category and word.
        #     Eventually we need to see both before and after so: P(W0|wb0_in_pos_1_AND_wa0_in_pos_1) etc for
        #         additional words AND additional columns. 
        #
        # We then create a new hash, %context which will contain a lot of data.
        #
        # The data structure looks like this:
        #
        # %context = { <word> => %dataHash };
        # %dataHash = { <surrounding word> => %hash2 }
        # %hash2 = { <surrounding word => @array }
        # @array = @array[2] = ( %hash3, %hash4 );
        # %hash3 = { <before = b, after = a> ~ <column index>) => <P(W0|WB0)> }
        # %hash4 = { @array, <P(W|@array.join(' ')) } Note, if @array = (1,2,3), in this case,
        #     @array as the key would look like @array.join(' ')
        #
        # It is probably necessary to call a recursive function to generate the insane amount of possibilities.
        #     we will call that function procStat.  See procStat for implementation details.
        procStat(%frequencyHash);
    }
}


sub procStat(%frequencyHash) {
    # For now, this will compare words before and after to similar words.  The similar words are based on
    # the percent of words it has in common with the other word (in a particular position).
    # Ex.  If "the" is wb0 and "is" is wa0, then what other words have "the" in the wb0 array and "is" in wa0?
    #      We will store those words in a hash.
    my %similarWords;
    my @similarWords = [];
    print "Processing context...\r";
    my Int $i = 0;
    for %wordsOnEitherSide.keys -> $w {
        print "Processing context ..." ~ ($i / %frequencyHash.keys.elems) * 100 ~ "% complete                        \r";
        my @wordsBefore = %wordsOnEitherSide{$w}[0];
        my @wordsAfter  = %wordsOnEitherSide{$w}[1];
        for %wordsOnEitherSide.keys -> $a {
            if $a ne $w && defined(%wordsOnEitherSide{$a}) && arraysAreSimilar(@wordsBefore, %wordsOnEitherSide{$a}[0], 50) && arraysAreSimilar(@wordsAfter, %wordsOnEitherSide{$a}[1], 50) {
                @similarWords.push($a);
            }
        }
        %similarWords{$w}[0] = @similarWords if @similarWords ne "";
        $i++;
    }

    print "Processing context...finishing                                                 \r ";
    my Int $totalLength = 0;
    for @words {
        $totalLength += $_.chars;
    }
    my Int $avgWordLength = round ($totalLength / @words.elems);
    my $u = 0;
    my %allMorphemes;
    for %similarWords.keys {
        $u++;
        %similarWords{$_}[0] = unique(%similarWords{$_}[0], $_);
        my @wordsToExamine;
        for 0 .. %similarWords{$_}[0].elems - 1 -> $num {
            @wordsToExamine.push(%similarWords{$_}[0][$num]);
        }
        @wordsToExamine.push($_);
        say "Analyzing morphemes..."~$u~"/"~%similarWords.keys.elems~ "                                     ";
        %similarWords{$_}[1] = getMorphemicStructure(@wordsToExamine, $avgWordLength, %allMorphemes);
    }
    say "Processing context...done                                 ";
    my $fh = open 'grammar.zy', :w;
    $fh.say('grammar Morphemes is Lang {');
    for %allMorphemes.keys {
        if %allMorphemes{$_}[0] ne 'NOTaMorpheme' {
            $fh.say("     token " ~ %allMorphemes{$_}[0] ~ ":sym<"~$_~'> { ' ~ %allMorphemes{$_}[2] ~ ' }');
        }
    }
    $fh.say('}');
    $fh.close();

    say "I have stored all the data I have learned."#  If you want me to refresh the data I have now, tell me: load grammar.zy";
    #    my $doc = slurp 'grammar.zy';
    #run (say Morphemes.parse('shadows'));
}

sub round($num) {
    return floor($num + 0.5);
}


sub arraysAreSimilar(@array1 = (), @array2 = (), Int $percentAgreement = 50) {
    my %arrayHash;
    my Int $totalElements = 0; # base 1 because *.elems is base 1
    for @array1 -> $q {
        $totalElements += $q.elems;
        for 0 .. $q.elems - 1 -> $index {
            for (0 .. $q[$index].elems - 1) -> $i {
                $totalElements += $q[$index].elems - 1;
                my $surroundingWord = $q[$index][$i];
                if defined $surroundingWord && $surroundingWord ne "" {
                    %arrayHash{$surroundingWord} = 1;
                }
            }
        }
    }
    for @array2 -> $q {
        $totalElements += $q.elems;
        for 0 .. $q.elems - 1 -> $index {
            for (0 .. $q[$index].elems - 1) -> $i {
                $totalElements += $q[$index].elems - 1;
                my $surroundingWord = $q[$index][$i];
                if defined $surroundingWord && $surroundingWord ne "" {
                    %arrayHash{$surroundingWord} = 1;
                }
            }
        }
    }
    return 0 if $totalElements == 0;
    return (((%arrayHash.keys.elems)/$totalElements) * 100) >= $percentAgreement; 
}

sub getMorphemicStructure(@alpha, $mostCommonWordLength is rw, %allMorphemes is rw) {
    my $i;
    my %morphemeGrammar = %allMorphemes;
    for %morphemeGrammar.keys {
        %morphemeGrammar{$_}[1] = 0;
    }
    my $alphabet = 'abcdefghijklmnopqrstuvwxyz'; # this will be dynamically defined at a later point
    my @regex; # stores the undetermined morphemes
    my $z = 0;
    my $increment = 0;
    for 0 .. @alpha.elems - 1 {
        for ($_ + 1 .. @alpha.elems - 1) {
            $z++;
        }
    }

    for 0 .. @alpha.elems - 1 {
        my $currentElem = @alpha[$_];
        for ($_ + 1 .. @alpha.elems - 1) {
            print 'Processing morphemes for this context...' ~ ($increment / $z) * 100 ~ "%            \r";
            my $smallElem = @alpha[$_];
            getMorphemes($currentElem, $smallElem, %morphemeGrammar, %allMorphemes, @regex);
            $increment++;
        }
    }

    if ($mostCommonWordLength % 2 != 0) {
        $mostCommonWordLength++;
    }
    # Break each regex into half around the morpheme.
    # If infront of the morpheme, take the last ciel($mostCommonWordLength/2) character classes
    # If behind of the morpheme, thake the first ciel($mostCommonWordLength/2) character classes
    # If all characters in those categories DON'T match all letters, it is a valid infix.

    for @regex {
        my @a = $_.split(')');
        my $morpheme = @a[1];
        $morpheme ~~ s/(.*?)\(.*/$0/;
        my $preMorpheme = @a[0].substr(1);
        my $postMorpheme = @a[1].substr($morpheme.chars + 1) ~ @a[2];
        my $rule = $_;
        if (!defined %morphemeGrammar{$morpheme}) {

            my @preMorphemeSplit = $preMorpheme.split('[');
            @preMorphemeSplit.shift();
            for @preMorphemeSplit -> $_ is rw {
                $_ = '[' ~ $_;
            }
            $preMorpheme = '';
            @preMorphemeSplit = reverse @preMorphemeSplit;
            for 0 .. ($mostCommonWordLength / 2) - 1 {
                if defined @preMorphemeSplit[$_] {
                    $preMorpheme ~= @preMorphemeSplit[$_];
                }
            } 
            $preMorpheme = rev $preMorpheme;
            $preMorpheme ~~ s:g/(\[|\])//;

            my @postMorphemeSplit = $postMorpheme.split('[');
            @postMorphemeSplit.shift();

            for @postMorphemeSplit -> $_ is rw {
                $_ = '[' ~ $_;
            }
            $postMorpheme = '';
            for 0 .. ($mostCommonWordLength / 2) - 1 {
                if defined @postMorphemeSplit[$_] {
                    $postMorpheme ~= @postMorphemeSplit[$_];
                }
            }
            $postMorpheme ~~ s:g/(\[|\])//;

            $preMorpheme = (uniq $preMorpheme.split('')).sort( {lc($^a) cmp lc($^b)} ).join('');
            $postMorpheme = (uniq $postMorpheme.split('')).sort( {lc($^a) cmp lc($^b)} ).join('');
            $preMorpheme = lc($preMorpheme);
            $postMorpheme = lc($postMorpheme);
            if (!($preMorpheme eq $alphabet || $preMorpheme eq $alphabet)) {
                if (defined(@a[2]) && @a[2] eq $morpheme.substr(0,1) eq '*') {
                    addMorphemeToGrammar($morpheme.substr(1), 'stem', %morphemeGrammar);
                    addMorphemeToGrammar($morpheme.substr(1), 'stem', %allMorphemes);
                } else {
                    if $morpheme.substr(0,1) eq '*' {
                        $morpheme = $morpheme.substr(1);
                        $rule ~~ s/\*$morpheme/$morpheme/;
                    }
                    addMorphemeToGrammar($morpheme, 'infix', %morphemeGrammar, $rule);
                    addMorphemeToGrammar($morpheme, 'infix', %allMorphemes, $rule);
                }
            } else {
                addMorphemeToGrammar($morpheme, 'NOTaMorpheme', %morphemeGrammar);
                addMorphemeToGrammar($morpheme, 'NOTaMorpheme', %allMorphemes);
            }
        }
    }
    return %morphemeGrammar;
}

sub addMorphemeToGrammar($morpheme, $type, %morphemeGrammar is rw, $regex = '<sym>') {
    if !defined(%morphemeGrammar{$morpheme}[0]) || (defined(%morphemeGrammar{$morpheme}[0]) && %morphemeGrammar{$morpheme}[0] ne 'stem') {
        %morphemeGrammar{$morpheme}[0] = $type;
        %morphemeGrammar{$morpheme}[2] = $regex;
    }
    if defined %morphemeGrammar{$morpheme}[1] {
        %morphemeGrammar{$morpheme}[1]++;
    } else {
        %morphemeGrammar{$morpheme}[1] = 1;
    }
    return 1;
}

sub getMorphemes($word1, $word2, %morphemeGrammar is rw, %allMorphemes is rw, @regex is rw) {
    # Extracts all possible morphemes
    # from two arguments
    # 
    # Morphemes will be represented in regex
    #
    # Example:
    #     thirsty
    #     hungry
    #     ========== 0 Morphemes
    #     thirsty
    #     firstly
    #     ========== 1 Morpheme /.*rst.*/
    #     alphabet
    #     bet
    #     ========== 1 Morpheme /bet$/
    #     eating
    #     starving
    #     ========== 1 Morpheme /ing$/
    #     
    # Although the data may be incorrect, this can lead to better data...
    # especially when combined with dictionary based checknig for the prob.
    # that the word exists in English.

    # Minimum match length is 2 (phase 1: minimum length: 1)
    # The length for phase 1 is shorter because it allows matching morphemes like {a-} and {-s}.  Without phase 1, we can't match those morphemes.

    # Step 1, find the longest match of characters in the same position.

    # %matches structure
    # {'lmFront' => longest matches starting with pos 0 = front, match starting index}
    # {'lmBack' => longest matches starting with pos 0 = back, match starting index}
    # {'secondLV' => [pos in w1, pos in w2, morpheme]}
    my %matches;
    %matches{<lmFront>} = [];
    %matches{<lmBack>} = [];
    %matches{<secondLv>} = [];

    firstLevel($word1, $word2, %matches);

    # Step 2, find all matching strings of characters in any pos (min 2).

    secondLevel($word1, $word2, %matches);

    for <lmFront lmBack secondLv> -> $type {
        if %matches{$type}.elems >= 1 {
            for (0 .. %matches{$type}.elems - 1) {
                regexify(%allMorphemes, %morphemeGrammar, @regex, $word1, $word2, %matches{$type}[$_], $type);
            }
        }
    }
}

sub firstLevel($w1, $w2, %matches is rw) {
    # Execute Step 1

    # lm = LongestMatch 
    # lmFront = Longest Matches starting with pos 0 being at the front.
    # lmBack = Longest Matches starting with pos 0 being at the back.

    my $word1 = $w1;
    my $word2 = $w2;
    my $hashID = <lmFront>;

    if ($word2.chars > $word1.chars) {
        my $tmp = $word1;
        $word1 = $word2; 
        $word2 = $tmp;
    }

    my $maxnum = $word1.chars == $word2.chars?? 0 !! 1;

    my @w1 = $word1.split('');
    my @w2 = $word2.split('');
    my $prevWasMatch = 0;

    for (0 .. $maxnum) {
        for (0 .. @w2.elems - 1) {
            my Int $matchIndex = %matches{$hashID}.elems - 1;
            my Int $matchIndexNew = $matchIndex + 1;
            if $matchIndex == -1 { $matchIndex = 0; }
            if $prevWasMatch || @w1[$_] eq @w2[$_] {
                if (my $match = @w1[$_] eq @w2[$_]) {
                    if $match && (defined %matches{$hashID}[$matchIndex][0]) && (%matches{$hashID}[$matchIndex][1] + %matches{$hashID}[$matchIndex][0] + 1) eq $_ {
                        %matches{$hashID}[$matchIndex][1]++;   
                        %matches{$hashID}[$matchIndex][2] ~= @w1[$_];
                    } else {
                        %matches{$hashID}[$matchIndexNew][0] = $_;
                        %matches{$hashID}[$matchIndexNew][1] = 0;
                        %matches{$hashID}[$matchIndexNew][2] = @w1[$_];
                        # Now that the prev contiguous string has been identified, process regex
                    }
                    $prevWasMatch = 1 if $match;
                }
            } else {
                $prevWasMatch = 0;
            }
        }

        @w1 = reverse @w1;
        @w2 = reverse @w2;
        $prevWasMatch = 0;
        $hashID = <lmBack>;
    }
}

sub secondLevel($w1, $w2, %matches is rw) {

    my $word1 = $w1;
    my $word2 = $w2;
    my $id = "secondLv";
    my $switch = 0;

    #word1 = longest
    if ($word2.chars > $word1.chars) {
        my $tmp = $word1;
        $word1 = $word2; 
        $word2 = $tmp;
        $switch = 1;
    }

    my @w1 = $word1.split('');
    my @w2 = $word2.split('');
    my $skip = 0;
    for (0 .. @w1.elems - 1) -> $w1index {
        my $currChar = @w1[$w1index];
        if ($skip == 0) {
            for (0 .. @w2.elems - 1) -> $w2index {
                if ($currChar eq @w2[$w2index]) {
                    my $morpheme;
                    my $i = 0;
                    while (defined(@w1[$w1index + $i]) && defined(@w2[$w2index + $i]) && @w1[$w1index + $i] eq @w2[$w2index + $i]) {
                        $morpheme ~= @w1[$w1index + $i];
                        $i++;
                    }

                    if $morpheme.chars >= 2  {
                        my $w1I;
                        my $w2I;
                        # starting place in w1, starting place in w2, morpheme
                        if ($switch) {
                            $w1I = $w2index;
                            $w2I = $w1index; 
                        } else {
                            $w1I = $w1index;
                            $w2I = $w2index; 
                        }
                        %matches{$id}[%matches{$id}.elems] = ([$w1I, $w2I, $morpheme]);
                        $skip += $i;
                    }
                }
            }
        } else {
            $skip--;
        }
    }
}

multi sub rev (Str $a) {
    return $a.flip;
}
multi sub rev (Str @a) {
    return reverse @a;
}

sub regexify(%allMorphemes is rw, %morphemeGrammar is rw, @regex is rw,  $w1?, $w2?, @dat?, $type?) {
    # Proc:
    # =====
    # isolate pattern
    # sub all chars for dots (less pattern)
    # summarize pattern
    # get existing pattern in @regex, if exist.
    # if regex_pattern_exists:
    #    make as specific as possible to fit the data AND
    #    as general as reasonably expected.
    #
    #    begrudgingly & bewilderingly => /^be/ && /ingly$/
    #    being & exceedingly => /ing(ly)*$/
    #       For this example, we have to discern the best method
    #       for dealing with multiple suffices.  The most efficeint
    #       method for now is to manually include the additional suffices.
    #       Eventually, we will reclassify the regex.
    #       When all the patterns are translated to regex, we can create
    #       different classes to represent the types of regex.
    #
    #       Seemingly, a prefix will contain /^/ and suffix will contain /$/
    #       Infix will not stand alone in the data set.  If it stands alone
    #       it is a root.  If it does not have a prefix or suffix attached
    #       directly, it is a root.
    #
    #       For now, we must classify everything as literally as possible
    #       and delay creating the abstraction of the regex classes.


    my $proc = "optimize";
    my $regexInProgress;

    my $before1;
    my $morpheme;
    my $end1;
    my $before2;
    my $end2;

    $proc = "gen" if ($w1 && $w2 && @dat && $type)

    if ($proc eq "gen") {
        # Is the pattern at beginning?
        # Is the pattern at the end?
        # If either of those, generalize regex.

        # secondLv will not contain any new morphemes for this category
        if ($type ne "secondLv") {
            if (@dat[0] == 0) {
                $morpheme = @dat[2];
                my $s;
                if ($type eq "lmBack") {
                    $morpheme = rev $morpheme;
                    $s = $morpheme ~ '$';
                } else {
                    $s = '^' ~ $morpheme;
                }
                $regexInProgress = $s;
            }
        }

        if (!$regexInProgress) {

            my $word1 = $w1;
            my $word2 = $w2;
            if ($type eq 'lmBack') {
                $word1 = rev $w1;
                $word2 = rev $w2; 
            }

            my $b1;
            my $m;
            my $e1;
            my $b2;
            my $e2;
            if ($type ne "secondLv") {
                $b1 = $word1.substr(0, @dat[0]);
                $m = $word1.substr(@dat[0], @dat[1] + 1);
                $e1 = $word1.substr(@dat[0] + @dat[1] + 1);
                $b2 = $word2.substr(0, @dat[0]);
                $e2 = $word2.substr(@dat[0] + @dat[1] + 1);
            } else {
                $b1 = $word1.substr(0, @dat[0]);
                $m = @dat[2];
                $e1 = $word1.substr(@dat[0] + @dat[2].chars);
                $b2 = $word2.substr(0, @dat[1]);
                $e2 = $word2.substr(@dat[1] + @dat[2].chars);
            }
            if ($type eq 'lmBack') {
                $before1 = $e1.flip;
                $morpheme = $m.flip;
                $end1 = $b1.flip;
                $before2 = $e2.flip;
                $end2 = $b2.flip;
            } else {
                $before1 = $b1;
                $morpheme = $m;
                $end1 = $e1;
                $before2 = $b2;
                $end2 = $e2;
            }

            my $toAdd;
            my $toAddb;
            my $toAdde;
            $toAdd = '(' ~ $before1 ~ '|' ~ $before2 ~ ')';
            for (0 .. 1) {
                if ($toAdd ~~ /(\(\||\|\))/) {
                    $toAdd ~~ s/\(\|/(/;
                    $toAdd ~~ s/\|\)/)/;
                    $toAdd ~= '*';
                }
                $toAdd ~~ s/\(\)\*//;
                if ($_ ne 1) {
                    $toAddb = $toAdd;
                } else {
                    $toAdde = $toAdd;
                }
                $toAdd = '(' ~ $end1 ~ '|' ~ $end2 ~ ')';

            }
            if $toAddb eq '' {
                $toAddb = '^';
            }
            if $toAdde eq '' {
                $toAdde = '$';
            }
            $regexInProgress = $toAddb ~ $morpheme ~ $toAdde;

        }
    }

    # Now we have extracted the morpheme.
    # Now we need to generalize and extrapolate.

    my $existingRegex = '';
    my $existingBeginning = '';
    my $existingEnd = '';
    my $currentBeginning = '';
    my $currentEnd = '';
    my $mergedBeginning = '';
    my $mergedEnd = '';
    my $morphemeClass = '';
    my $kind = '';

    # If suffix, generalize after suffix
    # If prefix, generalize before prefix
    # If infix, characterize literally.
    #     Generalizations will be made with more data. 
    # Uniquify Regex.

    # How to deal with subset problem ??? :
    #     {-in-} when it is really a subset of {-ing-}
    #
    #     For now, we must ignore this problem because
    #     {-in-} as a subset will appear in the same frequency
    #     as {-ing-}, so we can't know what is truely the morpheme
    #     because {-in-} may be a morpheme in another circumstnace.
    #     This problem will be solved when dealing with spellings to
    #     confirm correct morpheme usage.  If {-in-} never occurs where
    #     {-ing-} does besides in {-ing-}, then we can safely say that it
    #     must be a subset.  This is not yet implemented, thus we must ignore the problem.

    ##########################################
    # Step 1, find regex for the same morpheme
    #
    # Example:
    #     (th|test)in(g)*
    ##########################################

    if $regexInProgress eq '^' ~ $morpheme || $regexInProgress eq $morpheme ~ '$' {
        # If it is standing alone, it must be a stem
        $kind = 'stem';
    }
    if $regexInProgress.substr(0,1) eq '^' && $regexInProgress.substr(*-1) ne '$' {
        if (defined(%morphemeGrammar{$morpheme}) && %morphemeGrammar{$morpheme} eq 'suffix') {
            # If it starts and ends the word, it must be a stem
            $kind = 'stem';
        } else {
            $kind = 'prefix';
        }
    } elsif $regexInProgress.substr(0,1) ne '^' && $regexInProgress.substr(*-1) eq '$' {
        if (defined(%morphemeGrammar{$morpheme}) && %morphemeGrammar{$morpheme} eq 'prefix') {
            # If it starts and ends the word, it must be a stem
            $kind = 'stem';
        } else {
            $kind = 'suffix';
        }
    } else {
        if (!defined %morphemeGrammar{$morpheme}) {
            my $regexIndex = 0;
            repeat {
                my $pushed = 0;
                for (0 .. (@regex.elems - 1)) {
                    my $replaceNum = $_;
                    my $value = @regex[$replaceNum];
                    if ($value ~~ rx/(.*?(\)|\*|\+|\}|\^|\]))$morpheme((\(|\[|\{|\$).*)/) { # is this a matching rule?
                        # Step 2, identify the pre-morpheme rule
                        $existingBeginning = ~$0;

                        # Step 3, identify the post-morpheme rule
                        $value ~~ /$existingBeginning$morpheme(.*)/;
                        $existingEnd = ~$0;
                        $regexIndex++;

                        if ($regexInProgress ~~  rx/(.*?(\)|\*|\+|\}|\^|\]))$morpheme((\(|\[|\{|\$).*)/) {
                            $currentBeginning = ~$0;
                            $regexInProgress ~~ /$currentBeginning$morpheme(.*)/;
                            $currentEnd = ~$0;
                        }

                        # Step 4, merge pre-morpheme rule
                        my Str $content;
                        if ($existingBeginning && $existingBeginning.chars > 1) {
                            my $openingChar = $existingBeginning.substr(0,1);
                            my $closingChar = chr(ord($openingChar) + 1);
                            # If it is a paren, turn it into a class
                            mergeRules($existingBeginning, $currentBeginning, $mergedBeginning, 1);
                        } else {
                            $mergedBeginning = $currentBeginning;
                        }

                        # Step 5, merge post-morpheme rule
                        if ($existingEnd && $existingEnd.chars > 1) {
                            my $openingChar = $existingEnd.substr(0,1);
                            my $closingChar = chr(ord($openingChar) + 1);
                            # If it is a paren, turn it into a class
                            mergeRules($existingEnd, $currentEnd, $mergedEnd);
                        } else {
                            $mergedEnd = $currentEnd;
                        }

                        if $kind eq '' {
                            my $val = $mergedBeginning ~ $morpheme ~ $mergedEnd;
                            @regex[$replaceNum] = $val;
                            $pushed = 1;
                        }
                    }
                }
                if (!$pushed) {
                    @regex.push($regexInProgress);
                    $regexIndex++;
                }
            } while ($regexIndex < @regex.elems );


        }
    }

    if $kind ne '' {
        addMorphemeToGrammar($morpheme, $kind, %morphemeGrammar); 
        addMorphemeToGrammar($morpheme, $kind, %allMorphemes); 
    }
}

sub mergeRules($existing, $current, $merged is rw, $reverse = 0) {
    my $content;
    if $existing.substr(0) eq '(' {
        $content = $existing.substr(1, $existing.chars - 2);
    } else {
        $content = $existing;
    }
    if ($content ne '.*') {
        if ($current.substr(*-1) eq '*') {
            $merged = $current;
            return;
        }
        if ($current.substr(*-1) eq ')') {
            $merged = $current.substr(0, *-1) ~ '|' ~ $content ~ ')';
        }
        my @working = $merged.substr(1,*-1).split('|');
        my @classes = @working.join('').split('[');
        my $classesYN = 0;
        for @classes -> $_ is rw {
            if $_.substr(*-1) eq ']' {
                $_ = '[' ~ $_;
                $classesYN = 1;
            }
        }
        if !$classesYN {
            @classes = '';
        }
        for @working -> $_ is rw {
            $_ ~~ s:g/\[.*\]?//;
        }
        if @classes ne '' {
            for @classes -> $_ is rw {
                if $_.substr(*-1) ne ']' {
                    $_ = '';

                }
            }
        }
        for @working -> $_ is rw {
            if $_.substr(0, 1) eq '[' {
                $_ = '';
            }
            $_ ~~ s/(\(|\))//;
        }
        my @temp = Nil;
        for @classes {
            if $_ ne '' {
                @temp.push($_);
            }
        }
        @classes = @temp;
        @temp = Nil;
        for @working {
            if $_ ne '' {
                @temp.push($_);
            }
        }
        @working = @temp;

        for @classes -> $_ is rw {
            $_ = $_.substr(1, *-1);
        }

        if ($reverse) {
            @classes = reverse @classes;
        }

        for @working -> $elem {
            my $e;
            if ($reverse) {
                $e = rev $elem;
            } else {
                $e = $elem;
            }
            my @splitted = $e.split('');
            my $i = 0;
            for @splitted {
                my @chars;
                if defined @classes[$i] {
                    @chars = @classes[$i].split('');
                }
                @classes[$i] = @chars.push($_).join('');
                $i++;
            }
        }
        if ($reverse) {
            @classes = reverse @classes;
        }
        for @classes -> $_ is rw {
            $_ = (uniq $_.split('')).join('');
        }
        for @classes -> $_ is rw {
            $_ = '[' ~ $_ ~ ']';
        }
        $merged = '(' ~ @classes.join('') ~ ')';

    }

}

sub wFrequency(@array, Int $num? = 0) {
    my %uniqueArrayHash;
    for (@array) {
        if defined %uniqueArrayHash{$_} {
            %uniqueArrayHash{$_} += 1;
        } else {
            %uniqueArrayHash{$_} = 1;
        }
    }

    %uniqueArrayHash = reverse (%uniqueArrayHash.pairs.sort: { $^a.value <=> $^b.value });
    for (0 .. $num) {
        if (defined %uniqueArrayHash.keys[$_]) {
            my $key = %uniqueArrayHash.keys[$_];
            say $key ~ " => " ~ %uniqueArrayHash{$key} ~ " occurances.";
        }
    }
    return %uniqueArrayHash;
}

sub sFrequency(@array is rw, $num?) {
    my %uniqueArrayHash;
    my %freqHash;
    for @array -> @a is rw {
        for 0 .. @a.elems - 1 -> $q {
            my $w = @a[$q];
            if $w ne "" {
                if defined %uniqueArrayHash{$w} {
                    %uniqueArrayHash{$w} += 1;
                } else {
                    %uniqueArrayHash{$w} = 1;
                }
            }
        }
    }

    %uniqueArrayHash = reverse (%uniqueArrayHash.pairs.sort: { $^a.value <=> $^b.value });
    if ($num) {
        for (0 .. $num) {
            %freqHash{(keys %uniqueArrayHash)[$_]} = %uniqueArrayHash{(keys %uniqueArrayHash)[$_]};
        }
        for (0 .. $num) {
            if (defined %uniqueArrayHash.keys[$_]) {
                my $key = %uniqueArrayHash.keys[$_];
                say $key ~ " => " ~ %uniqueArrayHash{$key} ~ " occurances.";
            }
        }
    } else {
        %freqHash = %uniqueArrayHash;
    }
    return %freqHash;
}

sub process($type, $num? = -1) {
    if ($type eq "word_frequency") {
        #if $num == -1 {
            #    my $num = prompt("How many words would you like to see? ");
            #}
            print "Processing Word Frequency ...";
            %frequencyHash = stats("wordFrequency", $num);
            %procedureHash{"processWordFrequency"} = 1;
            say "done.";
        } elsif ($type eq "surrounding_word_frequency") {
            print "Processing Surrounding Word Frequency ...";
            stats("surroundingWordFrequency"); # Modifies the data structure by reference
            %procedureHash{"processSurroundingWordFrequency"} = 1;
            say "done";
        } elsif ($type eq "context") {
            if (!%procedureHash{"processWordFrequency"}) {
                process "word_frequency";
            }
            # stats("context") automatcially sets the surrounding word frequency
            stats("context"); # Modifies the data structure by reference
        }
    }

    sub display_knowledge() {
        say "I have processed " ~ @words.elems ~ " total words.";
        say "I know " ~ @dictionary.elems ~ " unique words.";
    }

    sub display($what) {
        my $s = "display_$what";
        run $s, 'Error, cannot display ' ~ $s;
    }

    sub unique(@array, $ignore = '') {
        my %uniqueArrayHash;
        for (@array) {
            if defined %uniqueArrayHash{$_} {
                %uniqueArrayHash{$_} += 1;
            } else {
                if $ignore ne $_ && $_ ne '' {
                    %uniqueArrayHash{$_} = 1;
                }
            }
        }
        return %uniqueArrayHash.keys;
    }

    sub recursiveDir($path, @array is rw) {
        for (dir $path) {
            if ($_ ~~ :d) {
                recursiveDir($path ~ '/' ~  $_, @array);
            }
            if (!($_ ~~ /\.zip$/)) {
                push @array, $path ~ '/' ~ $_;
            }
        }
        return @array;
    }

    multi sub store(Str $f) {
        store(0, $f);
    }
    multi sub store(Int $defaults? = 0, Str $fileName? = "") {
        my $filename = $fileName;
        if !$defaults {
            if $filename eq "" {
                $filename = prompt("What is the filename where I should save this data? -- Note, this WILL overwrite the data  [$restoreFile]  ");
            }
        }
        if $filename eq "" {
            say $filename; say $fileName;
            $filename = $restoreFile;
        }
        print "Storing data ...";
        my $fh = open $filename, :w;
        for (@letters) {
            if $_ ne "" && $_ ~~ s:g/\"// { 
                $fh.print('@letters.push("' ~ $_ ~ '");' ~ "\n"); 
            }
        }
        for (@words) {
            if $_ ne "" && $_ ~~ s:g/\"// { # Quote to fix vim formatting "
                $fh.print('@words.push("' ~ $_ ~ '");' ~ "\n");    
            }
        }
        for (@dictionary) {
            if $_ ne "" && $_ ~~ s:g/\"// { # Quote to fix vim formatting "
                $fh.print('@dictionary.push("' ~ $_ ~ '");' ~ "\n");    
            }
        }

        for 0 .. 1 -> $i {
            for %wordsOnEitherSide[$i].keys -> $key is rw {
                #$key ~~ s:g/\"//;
                if $key && $key ne "" {
                    for 0 .. 1 -> $j {
                        for %wordsOnEitherSide[$i]{$key}[$j] -> $u {
                            $fh.print( '%wordsOnEitherSide[' ~ $i ~ ']{'~ $key.perl ~'}[' ~ $j ~ '].push(' ~ $u.perl ~ ');' ~ "\n");
                        }
                    }
                }
            }
        }

        for %frequencyHash.keys -> $key {
            $fh.print('%frequencyHash{' ~ $key.perl ~ '}=' ~ %frequencyHash{$key}.perl ~ ';' ~ "\n"); 
        }

        # NEED TO ADD STORAGE FOR CONTEXT

        $fh.close();

        say "done."
    }

    multi sub load(Int $defaults? = 0, Str $fileName? = "") {
        say $fileName;
        my $filename = $fileName;
        if !$defaults {
            if $filename eq "" {
                $filename = prompt("What is the filename which should be loaded?\nNote, this will overwrite any data that has been learned this session.  Make sure this is done first. [$restoreFile]  ");
            }
        }
        if $filename eq "" {
            $filename = $restoreFile;
        }
        print "Restoring data ...";
        my $fh = open $filename;
        for $fh.lines { run $_, 'ERROR: SYNTAX ERROR INSIDE ' ~ $filename; }
        say "done.";
    }

    multi sub load(Str $filename) {
        load(0, $filename);
    }


    sub demo($name) {
        parse $name;
        display "knowledge";
        process "context";

    }


    # Note: repl() MUST be last to enable paren-less sub calls
    sub repl() {
        say "\nWelcome to Z. Bornheimer's Quasi-Artifically Intelligent Computational Linguistics Program.";
        say "Interestingly enough, most rules for acquiring language are not in here, and I will extrapolate them";
        say "If you need any help, ask me.\n";
        my $input;
        while (1) {
            $input = prompt(" >> ");
            my $parserResults = replGrammar.parse($input);
            # Substitutions
            if defined $parserResults<command><command><expression> && $parserResults<command><command><expression> ne "" {
                my $exp = $parserResults<command><command><expression>;
                my $origExp = $exp;
                $exp ~~ s:g/\ /_/;
                $input ~~ s:g/$origExp/$exp/;
                my $code;
                if ($exp ~~ rx/[\W\D]/) {
                    $code = Mu;
                } else {
                    $code = "defined(&$exp)";
                }
                if ($exp ~~ m/\./ || $code && !run $code, Nil) {
                    my $sub = '"' ~ $exp ~ '"';
                    $input ~~ s/$exp/$sub/;
                }
            }
            $input ~~ s/show\sme/say/;
            say "";

            my $keyword;
            if ($parserResults<command><keyword><sym>) {
                if (defined $parserResults<command><keyword><sym>) {
                    $keyword = $parserResults<command><keyword><sym>;
                }
            } else {
                if (defined $parserResults<command><command><sym>) {
                    $keyword = $parserResults<command><command><sym>;
                }
            }
            if defined ($keyword) {
                if $keyword eq "parse" || $keyword eq "demo" {
                    $input ~~ s/($keyword\s+)(.*)/{$0}"{$1}"/;
                    $input ~~ s:g/\"\"/"/;
                }
                if  run('defined(&' ~ $keyword ~ ')',Nil) {
                    run $input, Nil;
                    say "";
                }
            } else {
                if $input ne "" {
                    say "Sorry, but \`$input\` is not recognized.  It was probably a typo that I didn't catch, but if you need help, let me know.\n";
                }
            }
        }
    }
    repl();

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
    for (@filesToParse) {
        my $acqWordsCounter = 0;
        say "\nReading $_";
        print "\tParsing $_ ...";
        my Str $text = slurp $_;
        $text ~~ s:g/\n/\ /;
        my Int $numWords = $text.split(' ').elems - 1; # index 0 
        say "done";
        my @uniqueWords;
        print "\tProcessing the text ...\r";
        my @acqWords = $text.split(' '); # acquired words
        my $acqWordsIndex = 0;
        for (@acqWords) {
            print "\tProcessing the text ..." ~ ($acqWordsIndex / $numWords) * 100 ~ "%                 \r";
            $_ ~~ s:g/<english::punctuation>/ /; # Needs to replace the punctuation with a space 
# because it uses a space as a delimiter
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

            if ($acqWordsIndex > 0) {
                my @wordsBefore;
                my @wordsAfter;

# Because we have not yet incremented $acqWordsIndex, it will automatically be 1 less than the current word
                my $startingNum = $acqWordsIndex - PreviousWords;
                if $startingNum < 0 {
                    $startingNum = 0;
                }
                for $startingNum .. $acqWordsIndex - 1 {
                    my $i = ($acqWordsIndex - 1) - $_;
                    push @wordsBefore[$i], @acqWords[$_] if @acqWords[$_] ne "";
                }

# Because we have not yet incremented $acqWordsIndex, it is 1 lower than the current value, so the current word is at $acqWordsIndex+1
                for $acqWordsIndex + 2 .. $acqWordsIndex + (SubsequentWords + 1) {
                    my $i = ($acqWordsIndex + (SubsequentWords + 1)) - $_;
                    if defined @acqWords[$_] {
                        push @wordsAfter[$i], @acqWords[$_] if @acqWords[$_] ne "";
                    }
                }

                for 0 .. (PreviousWords - 1) -> $val {
                    my $t = @wordsBefore[$val][0];
                    my @arrayBefore;
                    if (defined $t && $t) {
                        push @arrayBefore, @wordsBefore[$val][0];
                    }
                    push %wordsOnEitherSide{$_}[0][$val], @arrayBefore;
                }
                for 0 .. (SubsequentWords - 1) -> $val {
                    my $y = @wordsAfter[$val][0];
                    my @arrayAfter;
                    if (defined $y && $y) {
                        push @arrayAfter, @wordsAfter[$val][0];
                    }
                    push %wordsOnEitherSide{$_}[1][$val], @arrayAfter;
                }
            }
            $acqWordsIndex++;
        }

        #@acqWords = removeEmpty(@acqWords);
        push @words, @acqWords ==> map { if $_ ne "" { $_ } };
        print "\tProcessing the text ...100%                            \r";
        say "\tProcessing the text ...done";
        say "Done Reading $_";
        say "Learned " ~ @acqWords.elems ~ " words\n";
        say "Done reading the text in $nlpLiteratureDir.";
        print "Creating the dictionary database by removing duplicate words in the word database ...";
        @dictionary = unique(@words);
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
    my @similarWords;
    print "Processing context...\r";
    my $i = 1;
    for %wordsOnEitherSide.keys -> $w {
        print "Processing context ..." ~ ($i / %frequencyHash.keys.elems) * 100 ~ "% complete \r";
        my @wordsBefore = %wordsOnEitherSide{$w}[0];
        my @wordsAfter  = %wordsOnEitherSide{$w}[1];
        for %wordsOnEitherSide.keys -> $a {
            if $a ne $w {
                if defined %wordsOnEitherSide{$a} {
                    if arraysAreSimilar(@wordsBefore, %wordsOnEitherSide{$a}[0], 50) && arraysAreSimilar(@wordsAfter, %wordsOnEitherSide{$a}[1], 50) {
                        push @similarWords, $a;
                    }
                }
            }
        }
        if @similarWords ne "" {
             %similarWords{$w} = @similarWords;
        }
        $i++;
    }

    print "Processing context...finishing                                                 \r ";
    for %similarWords.keys {
        %similarWords{$_} = unique(%similarWords{$_}, $_);
    }
    say "Processing context...done          ";
}

sub arraysAreSimilar(@array1 = (), @array2 = (), Int $percentAgreement = 50) {
    my %arrayHash;
    my Int $totalElements = 1; # base 1 due to comarison with *.elems (which is base 1)
    for @array1 -> $q {
        for 0 .. $q.elems - 1 -> $index {
            for (0 .. $q[$index].elems - 1) -> $i {
                my $surroundingWord = $q[$index][$i];
                if defined $surroundingWord && $surroundingWord ne "" {
                    %arrayHash{$surroundingWord} = 1;
                    $totalElements++;
                }
            }
        }
    }
    for @array2 -> $q {
        for 0 .. $q.elems - 1 -> $index {
            for (0 .. $q[$index].elems - 1) -> $i {
                my $surroundingWord = $q[$index][$i];
                if defined $surroundingWord && $surroundingWord ne "" {
                    %arrayHash{$surroundingWord} = 1;
                    $totalElements++;
                }
            }
        }
    }

    my $num = (((%arrayHash.keys.elems)/$totalElements) * 100); 
    return ($num >= $percentAgreement);

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
        if (!%procedureHash{"processSurroundingWordFrequency"}) {
          #  process "surrounding_word_frequency";
        }
        stats("context"); # Modifies the data structure by reference
    }
}

sub display_knowledge() {
    say "I have processed " ~ @words.elems ~ " total words.";
    say "I know " ~ @dictionary.elems ~ " unique words.";
}

sub display($what) {
    my $s = "display_$what";
    eval ($s);
}

sub unique(@array, $ignore?) {
    my %uniqueArrayHash;
    for (@array) {
        chomp($_);
        if defined %uniqueArrayHash{$_} {
            %uniqueArrayHash{$_} += 1;
        } else {
            if $ignore && $ignore ne $_ {
                %uniqueArrayHash{$_} = 1;
            }
        }
    }
    return keys %uniqueArrayHash;
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
    my $filename = $fileName;
    if !$defaults {
        if $filename eq "" {
            $filename = prompt("What is the filename which should be loaded?\nNote, this will overwrite any data that has been learned this session.  Make sure this is done first. [$restoreFile]  ");
        }
    }
    if chomp $filename eq "" {
        $filename = $restoreFile;
    }
    print "Restoring data ...";
    my $fh = open $filename;
    for $fh.lines { eval $_; }
    say "done.";
}

multi sub load(Str $filename) {
    load(0, $filename);
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
            if ($exp ~~ m/\./ || $code && !eval $code) {
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
            if $keyword eq "parse" {
                $input ~~ s/($keyword\s+)(.*)/{$0}"{$1}"/;
                $input ~~ s:g/\"\"/"/;
            }
            if  eval('defined(&' ~ $keyword ~ ')') {
                eval $input;
                say "";
            }
        } else {
            if $input ne "" {
                say "Sorry, but \`$input\` is not recognized.  It was probably a typo that I didn't catch, but if you need help, let me know.\n";
            }
        }
    }
}

#repl();
parse "t";
process "context";

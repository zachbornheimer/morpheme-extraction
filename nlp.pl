#!/usr/local/bin/perl6

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
        my Str @text = lines ($_).IO;
        say "done";
        print "\tProcessing and storing words ...\r";
        for 0 .. (@text.elems - 1) -> $index {
            my $text = @text[$index];
            my @uniqueWords;
            print "\tProcessing the text ..." ~ ($index / @text.elems) * 100 ~ "%                        \r";
            my @acqWords = $text.split(' '); # acquired words
                my $acqWordsIndex = 0;
            for (@acqWords) {
                $_ ~~ s:g/<english::punctuation>/ /; # Needs to replace the punctuation with a space 
# because it uses a space as a delimiter
                $_ ~~ s:g/\ //;

###################################################################################
# The ending data structure for the following code will look something like this:
# This example is done for 4 words.
# %hash = ( @wordsBefore, @wordsAfter);
# @wordsBefore is the  words before that word. 
# @wordsBefore = ( @_4wordsBefore, @_3wordsBefore, @_2wordsBefore, @_1wordBefore );
# @wordsAfter looks like @wordsBefore
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

            @acqWords = removeEmpty(@acqWords);
            @uniqueWords = unique(@acqWords);
            push @words, @uniqueWords;
            $acqWordsCounter += @uniqueWords.elems;
        }
        print "\tProcessing the text ...100%                            \r";
        say "\tProcessing the text ...done";
        say "Done Reading $_";
        say "Learned " ~ $acqWordsCounter ~ " words\n";
    }
    say "Done reading the text in $nlpLiteratureDir.";
    print "Creating the dictionary database by removing duplicate words in the word database ...";
    @dictionary = unique(@words);
    say "done";
}
sub removeEmpty(@array) {
    my @a;
    for (@array) {
        push @a, $_ if $_ ne "";
    }
    return @a;
}

multi sub stats($kind) {
}

multi sub stats($kind, $num) {
    if ($kind eq "wordFrequency") {
        my %uniqueArrayHash;
        my %frequencyHash;
        for (@words) {
            if defined %uniqueArrayHash{$_} {
                %uniqueArrayHash{$_} += 1;
            } else {
                %uniqueArrayHash{$_} = 1;
            }
        }

       %uniqueArrayHash = reverse (%uniqueArrayHash.pairs.sort: { $^a.value <=> $^b.value });
        for (0 .. $num) {
            %frequencyHash{(keys %uniqueArrayHash)[$_]} = %uniqueArrayHash{(keys %uniqueArrayHash)[$_]};
        }
        for (0 .. $num) {
            if (defined %uniqueArrayHash.keys[$_]) {
                my $key = %uniqueArrayHash.keys[$_];
                say $key ~ " => " ~ %uniqueArrayHash{$key} ~ " occurances.";
            }
        }
        return %frequencyHash;
    }
}

sub process($type) {
    if ($type eq "word_frequency") {
        my $num = prompt("How many words would you like to see? ");
        stats("wordFrequency", $num);
    }
}

sub display_knowledge() {
    say "I know " ~ @dictionary.elems ~ " unique words.";
}

sub display($what) {
    my $s = "display_$what";
    eval ($s);
}

sub unique(@array) {
    my %uniqueArrayHash;
    for (@array) {
        chomp($_);
        if defined %uniqueArrayHash{$_} {
            %uniqueArrayHash{$_} += 1;
        } else {
            %uniqueArrayHash{$_} = 1;
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



sub store($defaults? = 1) {
    my $fileName = "";
    if !$defaults {
        $fileName = prompt("What is the filename where I should save this data? -- Note, this WILL overwrite the data  [$restoreFile]  ");
    }
    if chomp $fileName eq "" {
        $fileName = $restoreFile;
    }
    print "Storing data ...";
    my $fh = open $fileName, :w;
    $fh.print('my @letters = ' ~ @letters.perl ~ ';'); 
    $fh.print('my @words = ' ~ @words.perl ~ ';'); 
    $fh.print('my @dictionary = ' ~ @dictionary.perl ~ ';'); 
    $fh.print('my %wordsOnEitherSide = ' ~ %wordsOnEitherSide.perl ~ ';'); 
    $fh.print('my %frequencyHash = ' ~ %frequencyHash.perl ~ ';'); 
    $fh.print('my @letters = ' ~ @letters.perl ~ ';'); 
    $fh.close();

    say "done."
}

sub load($defaults? = 1) {
    my $fileName = "";
    if !$defaults {
        $fileName = prompt("What is the filename which should be loaded?\nNote, this will overwrite any data that has been learned this session.  Make sure this is done first. [$restoreFile]  ");
    }
    if chomp $fileName == "" {
        $fileName = $restoreFile;
    }
    say "Restoring data ...";
    my $fh = open $fileName;
    for $fh.lines { say $_; }
    say "Done restoring data.";
}

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
            if ($code && !eval $code) {
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
        if (defined ($keyword)) {
            if $keyword eq "parse" {
                $input ~~ s/($keyword\s+)(.*)/{$0}"{$1}"/;
            }
            if (eval('defined(&' ~ $keyword ~ ')')) {
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


repl();

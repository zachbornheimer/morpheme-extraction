#!/usr/local/bin/perl6
use v6;
:autoflush; # Forces immediate flush.

constant PreviousWords = 4; # How many words behind the currently analyzed word should have their positions recorded
constant SubsequentWords = 4; # How many words in front of the currently analyzed word shave their positions recorded


my @letters;
my @words = ();
my @dictionary = ();
my $nlpLiteratureDir = "nlp-literature/";
my %wordsOnEitherSide = ();

grammar english {
    token punctuation {
       <- [a..zA..Z\d\s]>+($|\s+|<- [a..zA..Z\d]>)
    }
        
}

# Read all text into an array of words.
for (dir $nlpLiteratureDir) {
    my $acqWordsCounter = 0;
    say "\nReading $_";
    print "\tParsing $_ ...";
    my Str @text = lines ($nlpLiteratureDir ~ $_).IO;
    say "done";
    print "\tProcessing and storing words ...\r";
    for 0 .. (@text.elems - 1) -> $index {
        my $text = @text[$index];
        my @uniqueWords;
        print "\tProcessing the text ..." ~ ($index / @text.elems) * 100 ~ "%\r";
        my @acqWords = $text.split(' '); # acquired words
        my $acqWordsIndex = 0;
        for (@acqWords) {
            $_ ~~ s:g/<english::punctuation>/ /; # Needs to replace the punctuation with a space 
                                                 # because it uses a space as a delimiter
            if $_ ~~ /^\s*$/ {
                $_ = "";
            }

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
                    push @wordsBefore[$i], @acqWords[$_];
                }

                # Because we have not yet incremented $acqWordsIndex, it is 1 lower than the current value, so the current word is at $acqWordsIndex+1
                for $acqWordsIndex + 2 .. $acqWordsIndex + (SubsequentWords + 1) {
                    my $i = ($acqWordsIndex + (SubsequentWords + 1)) - $_;
                    if defined @acqWords[$_] {
                        push @wordsAfter[$i], @acqWords[$_];
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
print "Creating the dictionary database by removing duplicate words my the word database ...";
say "done";

@dictionary = unique(@words);
say "Learned a total of " ~ @dictionary.elems ~ " unique words from those texts.";


sub unique(@array) {
    my %uniqueArrayHash;
    for (@array) {
        chomp($_);
        %uniqueArrayHash{$_} = 1;
    }
    return keys %uniqueArrayHash;
}

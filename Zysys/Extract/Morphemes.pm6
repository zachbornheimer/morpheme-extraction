#!/usr/bin/perl

module Zysys::Extract::Morphemes;

use Zysys::Identification::LexicalCategories;

class NLPClass_M is NLPClass_LC {
    has $.corpusDir;
    has $.zedramOutput is rw;
    has $.humanity;
    has $.explicitDeclarations;
    has $!completedAnalysis;
    has @.acqWords;

    sub recursive_dir($path, @array is rw) {
        for dir $path {
            if $_ ~~ :d {
                recursive_dir($path ~ '/' ~  $_, @array);
            }
            if !($_ ~~ /\.zip$/) {
                @array.push($path ~ '/' ~ $_);
            }
        }
        return @array;
    }

    method assemble_structures() {
        my @filesToParse;
        recursive_dir($nlspLiteratureDir, @filesToParse);

        # Read all text into an array of words.
        my $numFilesProcessed = 0;
        for @filesToParse {
            if $_.IO ~~ :e  {
                my @uniqueWords;
                my $acqWordsCounter = 0;
                my $acqWordsIndex = 0;
                say "\nReading $_";
                print "\tParsing $_ ...";
                my Str $text = slurp $_;
                $text ~~ s:g/\n/\ /;
                my Int $numWords = $text.split(' ').end;
                say "done";
                print "\tProcessing the text ...\r";
                my @acqWords = $text.split(' '); # acquired words
                for @acqWords {
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

                    if $acqWordsIndex >= 0 {
                        for (0 .. 1) -> $iterator {
                            %wordsOnEitherSide{$_}[$iterator] = [] if !typecheck(%wordsOnEitherSide{$_}[$iterator], 'List');
                        }

                        my $wordsBefore = %wordsOnEitherSide{$_}[0].item;
                        my $wordsAfter = %wordsOnEitherSide{$_}[1].item;

                        # Because we have not yet incremented $acqWordsIndex, it will automatically be 1 less than the current word
                        my $startingNum = $acqWordsIndex - PreviousWords;
                        $startingNum = 0 if $startingNum < 0;
                        my $endingNum = $acqWordsIndex - 1;
                        my $array = $wordsBefore;
                        for (0 .. 1) -> $iterator {
                            for $startingNum .. $endingNum {
                                my $i = $endingNum - $_;
                                @$array[$i].push(@acqWords[$_]) if @acqWords[$_].defined && @acqWords[$_] ne "" ;
                            }
                            if $iterator == 0 {
                                $startingNum = $acqWordsIndex + 2; # $acqWordsIndex + 1 = current_word
                                $endingNum = $acqWordsIndex + SubsequentWords + 1; # SubsequentWord is base 0
                                $array = $wordsAfter;
                            }
                        }

                        $endingNum = PreviousWords - 1;
                        $array = $wordsBefore;
                        for (0 .. 1) -> $iterator {
                            for (0 .. $endingNum) -> $val {
                                my $item = @$array[$val][$iterator];
                                %wordsOnEitherSide{$_}[$iterator][$val].push($item) if $item.defined && $item;
                            }
                            if $iterator == 0 {
                                $endingNum = SubsequentWords - 1;
                                $array = $wordsAfter;
                            }
                        }

                        for ($wordsBefore, $wordsAfter) -> $array {
                            @$array = unique(@$array);
                        }

                        my $end = PreviousWords - 1;
                        for (0 .. 1) -> $iterator {
                            if !(%wordsOnEitherSide{$_}[$iterator].defined) {
                                for 0 .. $end {
                                    %wordsOnEitherSide{$_}[$iterator].push('[]');
                                }
                            }
                            $end = SubsequentWords - 1;
                        }

                        $acqWordsIndex++;
                    }
                }
                for @acqWords {
                    if $_ ne "" {
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
            @dictionary = unique(@dictionary);
            say "done";
        }

    }

}

}

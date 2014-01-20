#!/usr/bin/perl
use File::Slurp;
use feature qw/say/;
use List::MoreUtils qw/uniq/;
$| = 1;

my $dir = "test-corpus";
my $output = "languageModel.zdrm";

my @prefixes;
my @stems;
my @suffixes;
my @infixes;

open(MODEL, $output);

while(<MODEL>) {
    push @prefixes, $_ if (/^PREFIX: /);
    push @stems   , $_ if (/^STEM: /);
    push @suffixes, $_ if (/^SUFFIX: /);
    push @infixes , $_ if (/^INFIX: /);
}

close(MODEL);

@stems = map { s/^STEM: // and chomp $_ and $_   } @stems;
@prefixes = map { s/^PREFIX: // and chomp $_ and $_  } @prefixes;
@suffixes = map { s/^SUFFIX: // and chomp $_ and $_ } @suffixes;
@infixes = map { s/^INFIXES: // and chomp $_ and $_ } @infixes;
@suffixes = map { s/\W//g; $_ } @suffixes;

my @corpusWords;

opendir(DIR, $dir);
while(readdir(DIR)) {
    next if /^\./;
    my $string = read_file("$dir/$_");
    push @corpusWords, split(" ", $string);
}
@corpusWords = uniq(@corpusWords);

# Look for Prefixes

PREFIX_SEARCH: foreach my $prefix (@prefixes) {
    my $count = 0;
    foreach my $stem (@stems) {
        my $string = $prefix . $stem;
        $string =~ s/ //g;
        next PREFIX_SEARCH if /^\Q$string\E/i ~~ @stems;
        say "PREFIX: $prefix . $stem" and next PREFIX_SEARCH if (/^\Q$string\E/i ~~ @corpusWords and $count++ == 2);
    }

}

# Look for Prefixes

SUFFIX_SEARCH: foreach my $suffix (@suffixes) {
    my $count = 0;
    foreach my $stem (@stems) {
        my $string = $stem . $suffix;
        $string =~ s/ //g;
        next SUFFIX_SEARCH if /^\Q$string\E$/i ~~ @stems;
        say "SUFFIX: $suffix" and next SUFFIX_SEARCH if (/\Q$string\E$/i ~~ @corpusWords and $count++ == 2);
    }
}

# Look for Supplicated Prefixes

#my @corpusWords_cpy = @corpusWords;
#SUPPLICATE_PREFIX_SEARCH: foreach my $prefix (@prefixes) {
#    foreach my $corpusWord (@corpusWords) {
#        my $removedChars;
#        while($corpusWord ne "") {
#            my $string = $prefix . $corpusWord;
#            $string =~ s/ //g;
#            #     next SUPPLICATE_PREFIX_SEARCH if /^\Q$string\E$/i ~~ @corpusWords;
#            say "SUPPLICATED PREFIX: $prefix . $corpusWord" and next SUPPLICATE_PREFIX_SEARCH if (/^\Q$string\E$/i ~~ @corpusWords);
#            $removedChars .= substr $corpusWord, 0, 1;
#            $corpusWord = substr $corpusWord, 1;
#        }
#    }
#
#}
#@corpusWords = @corpusWords_cpy;

# Look for Supplicated Suffixes

#SUPPLICATE_SUFFIX_SEARCH: foreach my $suffix (@suffixes) {
#    foreach my $stem (@stems) {
#        my $removedChars;
#        while($stem ne "") {
#            my $string = $stem . $suffix;
#            $string =~ s/ //g;
#            #next SUPPLICATE_SUFFIX_SEARCH if /^\Q$string\E$/i ~~ @corpusWords;
#            say "SUPPLICATED SUFFIX: $stem . $suffix . _ $removedChars" and next SUPPLICATE_SUFFIX_SEARCH if (/^\Q$string\E$/i ~~ @corpusWords);
#            $removedChars .= substr $stem, -1;
#            $stem = substr $stem, 0, -1;
#        }
#    }
#
#}

#!/usr/bin/perl6

#my @alpha = <wb3a wb2ac wb1 wb0 w0 wa0 wa1 wa2 wa3>;
my @alpha = <thinking begrudgingly sparingly>;
my $i;
my %similarArray;
for 0 .. @alpha.elems - 1 {
    my $currentElem = @alpha[$_];
    for ($_ + 1 .. @alpha.elems - 1) {
        my $smallElem = @alpha[$_];



        %similarArray{$currentElem}.push(getMorphemes($currentElem, $smallElem));






    }
}
#say %similarArray.perl;


sub getMorphemes($word1, $word2) {
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

# we now have all the possible morphemes.
# now we need to convert the extracted morphemes into valid regex

    my @regex;
    for <lmFront lmBack secondLv> -> $type {
        if %matches{$type}.elems >= 1 {
            for (0 .. %matches{$type}.elems - 1) {
#                say %matches.perl;
                regexify(@regex, $word1, $word2, %matches{$type}[$_], $type);
            }
        }
    }


# Step 3, return a non-unique array of regex found in Steps 1 & 2.





#    exit; # INCOMPLETE!!
}

sub firstLevel($w1, $w2, %matches is rw) {

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

sub rev (Str $a) {
    return (reverse $a.split('')).join('');
}

#regexify(@regex, $word1, $word2, %matches{$type}[$_], $type);
sub regexify(@regex is rw,  $w1?, $w2?, @dat?, $type?) {

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

    if ($w1 && $w2 && @dat && $type) {
        $proc = "gen";
    }

    if ($proc eq "gen") {
# Is the pattern at beginning?
# Is the pattern at the end?
# If either of those, generalize regex.

# secondLv will not contain any new morphemes for this category
        if ($type ne "secondLv") {
            if (@dat[0] == 0) {
                my $s = @dat[2];
                if ($type eq "lmBack") {
                    $s = (reverse $s) ~ '$';
                } else {
                    $s = '^' ~ $s;
                }
                $regexInProgress = '/' ~ $s ~ '/';
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
                $m = $word1.substr(@dat[0], @dat[0] + @dat[1] + 1);
                $e1 = $word1.substr(@dat[0] + @dat[1] + 1);
                $b2 = $word2.substr(0, @dat[0]);
                $e2 = $word2.substr(@dat[0] + @dat[1] + 1);
            } else {
                $b1 = $word1.substr(0, @dat[0]);
                $m = $word1.substr(@dat[0], @dat[0] - 1 + @dat[2].chars - 1);
                $e1 = $word1.substr(@dat[0] + @dat[2].chars);
                $b2 = $word2.substr(0, @dat[1]);
                $e2 = $word2.substr(@dat[1] + @dat[2].chars);
            }
            if ($type eq 'lmBack') {
                $before1 = rev $e1;
                $morpheme = rev $m;
                $end1 = rev $b1;
                $before2 = rev $e2;
                $end2 = rev $b2;
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
            $regexInProgress = $toAddb ~ $morpheme ~ $toAdde;
            say $regexInProgress.perl;

        }

    }

# If suffix, generalize after suffix
# If prefix, generalize before prefix
# If infix and !secondLv, use pos-mapping
# If infix and secondLv, characterize literally.
# Generalizations will be made later.
# Uniquify Regex.

# How to deal with subset problem ??? :
#     {-in-} when it is really a subset of {-ing-}


# Step 1, find regex for the same morpheme
    
# Examples:
#     (th|test)in(g)*
#     (.*)ary
    my $existingRegex;
    my $existingBeginning;
    my $existingEnd;
    my $currentBeginning;
    my $currentEnding;
    for @regex {
        if ($_ ~~ rx/(.*\)|\*|\+|\}|\^|\])$morpheme(\(|\[|\{|\$.*)/) {
# Step 2, identify the pre-morpheme rule
            $existingBeginning = $0;
# Step 3, identify the post-morpheme rule
            $_ ~~ /$existingBeginning$morpheme(.*)/;
            $existingEnd = ~$0;
        }
    }
    if ($regexInProgress ~~ rx/(.*\)|\*|\+|\}|\^|\])$morpheme(\(|\[|\{|\$.*)/) {
        $currentBeginning = ~$0;
        $regexInProgress ~~ /$currentBeginning$morpheme(.*)/;
        $currentEnding = ~$0;
    }
    my Str $content;

# Step 4, merge pre-morpheme rule

    if ($existingBeginning) {

        my $openingChar = $existingBeginning.substr(0,1);
        my $closingChar = chr(ord($openingChar) + 1);
# If it is a paren, turn it into a class
        if ($openingChar ne "[" && $existingBeginning ~~ /^$openingChar.*$closingChar/) {
            $content = $existingBeginning.substr(1, $existingBeginning.chars - 2);
        }
    }

    if ($currentBeginning) {
        if ($currentBeginning.substr(0,1) ne "[" && $currentBeginning ~~ /^{$currentBeginning.substr(0,1)}.*{chr(ord($currentBeginning.substr(0,1)) + 1)}/) {
            $content ~= $currentBeginning.substr(1, $currentBeginning.chars - 2); 
        }
    }
    $content ~~ s/\|// if $content;

# Step 5, merge post-morpheme rule
# Step 6, generalize pre-morpheme rule
# Step 7, generalize post-morpheme rule
# Step 8, store regex



}


#!/usr/bin/perl6

#my @alpha = <wb3a wb2ac wb1 wb0 w0 wa0 wa1 wa2 wa3>;
my @alpha = <thoughts forethought foreshadow shadows begrudgingly sparingly>;
my $i;
my %similarArray;
my %morphemeGrammar;
my $alphabet = 'abcdefghijklmnopqrstuvwxyz';
my @regex; # stores the undetermined morphemes
my $z = 0; # Total Increments
my $y = 0; # Current Increment
# $z and $y are used for determining if this is the last run...dealing with the infix

for 0 .. @alpha.elems - 1 {
    my $currentElem = @alpha[$_];
    for ($_ + 1 .. @alpha.elems - 1) {
        my $smallElem = @alpha[$_];
        %similarArray{$currentElem}.push(getMorphemes($currentElem, $smallElem, %morphemeGrammar, @regex));
    }
}

my $mostCommonWordLength = 6; # This should be set by the data
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
            if ($morpheme.substr(0,1) eq '*') {
                addMorphemeToGrammar($morpheme.substr(1), 'stem', %morphemeGrammar);
            } else {
                addMorphemeToGrammar($morpheme, 'infix', %morphemeGrammar);
            }
        }
    }
}
my $fh = open 'thisIsATest', :w;
$fh.say(%morphemeGrammar.perl);
$fh.close();

sub addMorphemeToGrammar($morpheme, $type, %morphemeGrammar is rw) {
    %morphemeGrammar{$morpheme}[0] = $type;
    if defined %morphemeGrammar{$morpheme}[1] {
        %morphemeGrammar{$morpheme}[1]++;
    } else {
        %morphemeGrammar{$morpheme}[1] = 1;
    }
    return 1;
}

sub getMorphemes($word1, $word2, %morphemeGrammar is rw, @regex is rw) {
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
    for <lmFront lmBack secondLv> -> $type {
        if %matches{$type}.elems >= 1 {
            for (0 .. %matches{$type}.elems - 1) {
                regexify(%morphemeGrammar, @regex, $word1, $word2, %matches{$type}[$_], $type);
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

multi sub rev (Str $a) {
    return (reverse $a.split('')).join('');
}
multi sub rev (Str @a) {
    return reverse @a;
}

sub regexify(%morphemeGrammar is rw, @regex is rw,  $w1?, $w2?, @dat?, $type?) {

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
                $m = $word1.substr(@dat[0], @dat[0] + @dat[2].chars);
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
            if $toAddb eq '' {
                $toAddb = '^';
            }
            if $toAdde eq '' {
                $toAdde = '$';
            }
            $regexInProgress = $toAddb ~ $morpheme ~ $toAdde;

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
    my $existingRegex = '';
    my $existingBeginning = '';
    my $existingEnd = '';
    my $currentBeginning = '';
    my $currentEnd = '';
    my $mergedBeginning = '';
    my $mergedEnd = '';
    my $morphemeClass = '';
    my $kind = '';
    if $regexInProgress.substr(0,1) eq '^' && $regexInProgress.substr(*-1) ne '$' {
        if (defined(%morphemeGrammar{$morpheme}) && %morphemeGrammar{$morpheme} eq 'suffix') {
            $kind = 'stem';
        } else {
            $kind = 'prefix';
        }
    } elsif $regexInProgress.substr(0,1) ne '^' && $regexInProgress.substr(*-1) eq '$' {
        if (defined(%morphemeGrammar{$morpheme}) && %morphemeGrammar{$morpheme} eq 'prefix') {
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


# Step 6, add part of word (prefix, suffix, infix) to grammar

# Step 6a: is prefix?

# Step 6b: is suffix?

# Step 6c: is infix?

# Step 6d: is stem?

# Step 7, store morpheme in @regex so when all morphemes are done, it can be analyzed and stored.

        }
    }

    if $kind ne '' {
        addMorphemeToGrammar($morpheme, $kind, %morphemeGrammar); 
    }
}

sub mergeRules($existing, $current, $merged is rw, $reverse?) {
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

        @classes = reverse @classes;

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
        @classes = reverse @classes;
        for @classes -> $_ is rw {
            $_ = (uniq $_.split('')).join('');
        }
        for @classes -> $_ is rw {
            $_ = '[' ~ $_ ~ ']';
        }
        $merged = '(' ~ @classes.join('') ~ ')';

    }

}

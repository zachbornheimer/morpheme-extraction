#!/usr/bin/perl6

use Zysys::Ai;

my $corpusDir = 'corpus/corpus';
my $zedramOutput = 'languageModel.zdrm';
my $humanInput = humanInput();


my $parser = NLPEngine.new(:corpusDir($corpusDir), :zedramOuput($zedramOutput), :humanity($humanInput), :explictDeclarations(1));

while ($parser.analyze()) {
    $parser.assemble_structures();
    $parser.alphabet_identification();
    $parser.morpheme_extraction();
    $parser.lexical_category_identification();
    $parser.syntax_model_creation();
    $parser.meaning_map_generation();
}

$parser.zedram(:action('optimize-stored-data', 'compress-stored-data'));

sub human_input() {
    # This function should contain methods to access the real world
    # (probably through robotics) or predefined data for the purpose
    # of meaning abstraction.  This is the important part :O)

    constant Humanity = 0;
    return Humanity;
}
